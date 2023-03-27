#/bin/bash
# Description: Script to process the output from the process_async_processed.sh to be run per node
# It verifies files in the list for various validations and if they pass all checks, if on dry run
# it reports the command it will execute; if on Execute_Move, it moves files from /mount/current_dir
# to /mount/current_dir.processed
# Input: Pipe separated file with local staging path|human readable date|file size|file id
# Usage:
## Default - dry run on all lines
## Argument 1 - input file
## Argument 2 - number of days ago time limit
## Argument 3 - set to Execute_Move to move


# Parameters
in_file=$1

# Constants
cqlsh=/usr/lib64/GB/DCF/JServices/MbService/bin/cqlsh
select="select * from storage.datasets_by_name"
async_processed_files_dataset=$($cqlsh $(hostname -I) 21205 -e "$select" | grep "ASYNCH PROCESSED FILES FOR" | cut -d"|" -f2 | xargs)

# Variables
now=$(date +"%s")
ip=$(hostname -I | xargs)

# timestamp
function timestamp () {
   echo $(date -u '+%Y-%m-%dT%H:%M:%S UTC')
}


# Arguments
## Setting the input file
if [ ! -z "$1" ]
then
   in_file=$1
   if [[ ! -f "$in_file" ]]
   then
      # Incorrect usage
      echo "Usage: Arg1 - input file; Arg2 - number of days ago to process (integer); Arg3 - use 'Execute_Move' to cancel dry run default"
      exit
   fi
else
      # Incorrect usage
      echo "Usage: Arg1 - input file; Arg2 - number of days ago to process (integer); Arg3 - use 'Execute_Move' to cancel dry run default"
      exit
fi

## Setting the time limit
if [ ! -z "$2" ]
then
   days="$2"
   re='^[0-9]+$'
   if ! [[ "$days" =~ $re ]]
   then
      # Incorrect usage
      echo "Usage: Arg1 - input file; Arg2 - number of days ago to process (integer); Arg3 - use 'Execute_Move' to cancel dry run default"
      exit
   else
      time_limit=$(("$now"-("$days"*86400)))
   fi
   echo "INFO [$(timestamp)]: Time limit set to $days days ago which is $time_limit in epoch time" 
else
   # Default argument
   echo "WARN [$(timestamp)]: No time limit set; processing all processed files"
fi

## Switching off dry run
if [ "$3" = "Execute_Move" ]
then
   echo "WARN [$(timestamp)]: Argument Execute_Move; setting dryrun to false" 
   dryrun=
else
   echo "INFO [$(timestamp)]: No Execute_Move argument; setting dryrun to true"
   dryrun=true
fi

# Function to process each line in the list
# Input: file of list of processed files (ordered by size first) 
# Output: Either report (if dry run), or files moved into xxx.processed (if set to Execute_Move)
function get_lines () {

# Set input argument to processed_file_list
processed_file_list=$1

# While block processes processed_file_list line by line
while read line;
do
   # Output current line from processed_file_list
   echo INFO [$(timestamp)]: Processing $line
   
   # Set working variables

   ## target_file is staging_path
   target_file=$(echo $line | cut -d"|" -f2)
   
   ### basename is the full basename from staging_path including the {gbtmp-xxx} value
   basename=$(basename $target_file)

   ### orig_basename is the original basename in SMB, excluding the {gbtmp-xxx} value
   orig_basename=$(echo $line | cut -d"|" -f1)
   ### dirname is the path from the staging_path
   dirname=$(dirname $target_file)

   ## create_time is set based on line input
   create_time=$(echo $line | cut -d"|" -f3)
   ### create_time_epoch converts create_time to epoch time
   ### create_time_epoch=$(date --date "$create_time" +"%s")

   ## file_size is set based on line input
   file_size=$(echo $line | cut -d"|" -f4)
   ## file_id is set based on line input
   file_id=$(echo $line | cut -d"|" -f5)

   ## fan_ip is set based on the line input
   fan_ip=$(echo $line | cut -d"|" -f6)

   # Process lines

   ## Filter out files on other servers
   if [[ "$fan_ip" == "$ip" ]]
   then
      echo "INFO [$(timestamp)]: Fan IP $fan_ip matches local IP $ip"
   else
      echo "WARN [$(timestamp)]: Fan IP $fan_ip does not match local IP $ip; skipping file $target_file"
      continue
   fi


   ## Filter out files from before time_limit
   if [[ $create_time -lt $time_limit ]]
   then
      echo "WARN [$(timestamp)]: Create time (epoch) $create_time is before time limit (epoch) $time_limit; skipping file $target_file"
      continue
   else
      echo "INFO [$(timestamp)]: Create time (epoch) $create_time is after time limit (epoch) $time_limit for $target_file"
   fi

   ## Verify file is in async processed dataset
   ### Use gbr to list file by file id with details, grep for parent id and strip preceding
   dataset=$(gbr file ls -i $file_id -d | grep "parent id" | cut -c25-)
   ### Filter out any files not in async_processed_files_dataset
   if [ "$dataset" = "$async_processed_files_dataset" ]
   then
      echo "INFO [$(timestamp)]: file id $file_id is in async processed files dataset $async_processed_files_dataset for $target_file"
   else
      echo "WARN [$(timestamp)]: file id $file_id is in dataset $dataset, rather than the async processed files dataset $async_processed_files_dataset; skipping file $target_file"
      continue
   fi


   ## Verify file exists
   if [ -f "$target_file" ]
   then
      echo "INFO [$(timestamp)]: file $target_file exists"
   else
      echo "WARN [$(timestamp)]: file does not exist; skipping file $target_file"
      continue
   fi

   ## Verify file size matches
   ### Get file size from stat
   staging_file_size=$(stat --format='%s' $target_file)
   ### Filter out any files where size does not match
   if [ "$file_size" -eq "$staging_file_size" ]
   then
      echo "INFO [$(timestamp)]: file size $file_size matches staging file size $staging_file_size for $orig_basename"
   else
      echo "WARN [$(timestamp)]: file size $file_size does not match staging file size $staging_file_size; skipping file $orig_basename"
      continue
   fi

   ## Verify create time matches
   ### Get file create time from stat
   staging_create_time=$(stat --format='%Y' $target_file)
   ### Filter out any files where create time does not match
   if [ "$create_time" = "$staging_create_time" ]
   then
      echo "INFO [$(timestamp)]: create time $create_time matches staging create time $staging_create_time for $orig_basename"
   else
      echo "WARN [$(timestamp)]: create time $create_time does not match staging create time $staging_create_time; skipping file $orig_basename"
      continue
   fi

   ## Verify file_id matches with file_name
   ### Use gbr to check filename
   file_id_checked_name=$(gbr file ls -i $file_id | cut -d" " -f3)
   ### Filter out any file whose original SMB name does not match the name by file_id
   if [ "$file_id_checked_name" = "$orig_basename" ]
   then
      echo "INFO [$(timestamp)]: basename $orig_basename matches file id checked name $file_id_checked_name"
   else
      echo "WARN [$(timestamp)]: basename $orig_basename does not match file id checked name $file_id_checked_name; skipping file"
      continue
   fi

   # File has been verified and is ready for migration
   echo "INFO [$(timestamp)]: $target_file verified as ready to be migrated in preparation for removal"


   # Process files for migration
   
   ## Hash file
   echo "INFO [$(timestamp)]: starting initial sha256 for $target_file"
   staging_file_hash=$(sha256sum $target_file | cut -d" " -f1)
   echo "INFO [$(timestamp)]: sha256 $staging_file_hash for $target_file completed"

   ## Move file
   ### Check if mb or datavX staging folder
   if $(echo $dirname | grep -q mb)
   then
      old_path=FAN
      new_path=FAN.processed
   else
      old_path=staging
      new_path=staging.processed
   fi

   ### sed pathname to swap old_path for new_path
   new_path_name=$(echo $dirname | sed "s/$old_path/$new_path/g")
   echo "INFO [$(timestamp)]: Staging path $dirname; new path $new_path_name"

   ### Create target folder
   if [ ! -d $new_path_name ]
   then
       mkdir -p $new_path_name
       echo "INFO [$(timestamp)]: Created $new_path_name"
   fi

   ### Create move command for target_file
   cmd="mv $dirname/$basename $new_path_name/$basename"
   echo "INFO [$(timestamp)]: Command $cmd"

   ### Check for dry run
   if [ ! -z $dryrun ]
   then
      ### If dry run skip
      echo "INFO [$(timestamp)]: Dryrun skipping execute"
      continue
   else
      ### If not dry run execute move
      echo "WARN [$(timestamp)]: Executing command $cmd"
      eval $cmd
   fi

   ### Recheck hash
   echo "INFO [$(timestamp)]: starting check sha256 for $new_path_name/$basename"
   moved_file_hash=$(sha256sum $new_path_name/$basename | cut -d" " -f1)
   echo "INFO [$(timestamp)]: sha256 $moved_file_hash for $new_path_name/$basename completed"
   

   ### Compare moved file hash and original hash
   if [ "$staging_file_hash" = "$moved_file_hash" ]
   then
      ### If hashes match
      echo "INFO [$(timestamp)]: $orig_basename hash after move $moved_file_hash matches original hash $staging_file_hash"
      ### Confirm ready for process
      echo "INFO [$(timestamp)]: $new_path_name/$basename is ready for processing"
   else
      ### If no match, error out and exit, as this shouldn't be possible
      echo "ERROR [$(timestamp)]: $orig_basename hash after move $moved_file_hash does not match original hash $staging_file_hash"
      exit
   fi

done < $processed_file_list

}

# Execute get_lines function to begin processing
get_lines $in_file

