#/bin/bash
# Description: Script to process the output from the process_async_processed.sh to be run per node
# It verifies files in the list for various validations and if they pass all checks, if on dry run
# it reports the command it will execute; if on Execute_Move, it moves files from /mount/current_dir
# to /mount/current_dir.processed
# Input: Pipe separated file with local staging path|human readable date|file size|file id
# Usage:
## Set in_file parameter
## Default - dry run on all lines
## Argument 1 - number of days ago time limit
## Argument 2 - set to Execute_Move to move


# Parameters
in_file=/home/admin/10.41.28.170.out

# Constants
cqlsh=/usr/lib64/GB/DCF/JServices/MbService/bin/cqlsh
select="select * from storage.datasets_by_name"
async_processed_files_dataset=$($cqlsh $(hostname -I) 21205 -e "$select" | grep "ASYNCH PROCESSED FILES FOR" | cut -d"|" -f2 | xargs)

# Variables
now=$(date +"%s")

# Arguments
## Setting the time limit
if [ ! -z "$1" ]
then
   days="$1"
   re='^[0-9]+$'
   if ! [[ "$days" =~ $re ]]
   then
      # Incorrect usage
      echo "Usage: Arg1 - number of days ago to process (integer); Arg2 - use 'Execute_Move' to cancel dry run default"
      exit
   else
      time_limit=$(("$now"-("$days"*86400)))
   fi
   echo "INFO: Time limit set to $days days ago which is $time_limit in epoch time" 
else
   # Default argument
   echo "WARN: No time limit set; processing all processed files"
fi

## Switching off dry run
if [ "$2" = "Execute_Move" ]
then
   echo "WARN: Argument Execute_Move; setting dryrun to false" 
   dryrun=
else
   echo "INFO: No Execute_Move argument; setting dryrun to true"
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
   echo INFO: Processing $line
   
   # Set working variables

   ## target_file is staging_path
   target_file=$(echo $line | cut -d"|" -f1)
   ### basename is the full basename from staging_path including the {gbtmp-xxx} value
   basename=$(basename $target_file)
   ### orig_basename is the original basename in SMB, excluding the {gbtmp-xxx} value
   orig_basename=$(basename $target_file | cut -d"{" -f1)
   ### dirname is the path from the staging_path
   dirname=$(dirname $target_file)

   ## create_time is set based on line input
   create_time=$(echo $line | cut -d"|" -f2)
   ### create_time_epoch converts create_time to epoch time
   create_time_epoch=$(date --date "$create_time" +"%s")

   ## file_size is set based on line input
   file_size=$(echo $line | cut -d"|" -f3)
   ## file_id is set based on line input
   file_id=$(echo $line | cut -d"|" -f4)

   # Process lines

   ## Filter out files from before time_limit
   if [[ $create_time_epoch -lt $time_limit ]]
   then
      echo "WARN: Create time (epoch) $create_time_epoch is before time limit (epoch) $time_limit; skipping file $target_file"
      continue
   else
      echo "INFO: Create time (epoch) $create_time_epoch is after time limit (epoch) $time_limit for $target_file"
   fi

   ## Verify file is in async processed dataset
   ### Use gbr to list file by file id with details, grep for parent id and strip preceeding
   dataset=$(gbr file ls -i $file_id -d | grep "parent id" | cut -c25-)
   ### Filter out any files not in async_processed_files_dataset
   if [ "$dataset" = "$async_processed_files_dataset" ]
   then
      echo "INFO: file id $file_id is in async processed files dataset $async_processed_files_dataset for $target_file"
   else
      echo "WARN: file id $file_id is in dataset $dataset, rather than the async processed files dataset $async_processed_files_dataset; skipping file $target_file"
      continue
   fi


   ## Verify file exists
   if [ -f "$target_file" ]
   then
      echo "INFO: file $target_file exists"
   else
      echo "WARN: file does not exist; skipping file $target_file"
      continue
   fi

   ## Verify file size matches
   ### Get file size from stat
   staging_file_size=$(stat --format='%s' $target_file)
   ### Filter out any files where size does not match
   if [ "$file_size" -eq "$staging_file_size" ]
   then
      echo "INFO: file size $file_size matches staging file size $staging_file_size for $orig_basename"
   else
      echo "WARN: file size $file_size does not match staging file size $staging_file_size; skipping file $orig_basename"
      continue
   fi

   ## Verify create time matches
   ### Get file create time from stat
   staging_create_time=$(stat --format='%Y' $target_file)
   ### Filter out any files where create time does not match
   if [ "$create_time_epoch" = "$staging_create_time" ]
   then
      echo "INFO: create time $create_time_epoch matches staging create time $staging_create_time for $orig_basename"
   else
      echo "WARN: create time $create_time_epoch does not match staging create time $staging_create_time; skipping file $orig_basename"
      continue
   fi

   ## Verify file_id matches with file_name
   ### Use gbr to check filename
   file_id_checked_name=$(gbr file ls -i $file_id | cut -d" " -f3)
   ### Filter out any file whose original SMB name does not match the name by file_id
   if [ "$file_id_checked_name" = "$orig_basename" ]
   then
      echo "INFO: basename $orig_basename matches file id checked name $file_id_checked_name"
   else
      echo "WARN: basename $orig_basename does not match file id checked name $file_id_checked_name; skipping file"
      continue
   fi

   # File has been verified and is ready for migration
   echo "INFO: $target_file verified as ready to be migrated in preparation for removal"


   # Process files for migration
   
   ## Hash file
   staging_file_hash=$(sha256sum $target_file | cut -d" " -f1)

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
   echo "INFO: Staging path $dirname; new path $new_path_name"

   ### Create target folder
   if [ ! -d $new_path_name ]
   then
       mkdir -p $new_path_name
       echo "INFO: Created $new_path_name"
   fi

   ### Create move command for target_file
   cmd="mv $dirname/$basename $new_path_name/$basename"
   echo "INFO: Command $cmd"

   ### Check for dry run
   if [ ! -z $dryrun ]
   then
      ### If dry run skip
      echo "INFO: Dryrun skipping execute"
      continue
   else
      ### If not dry run execute move
      echo "WARN: Executing command $cmd"
      eval $cmd
   fi

   ### Recheck hash
   moved_file_hash=$(sha256sum $new_path_name/$basename | cut -d" " -f1)
   
   ### Compare moved file hash and original hash
   if [ "$staging_file_hash" = "$moved_file_hash" ]
   then
      ### If hashes match
      echo "INFO: $orig_basename hash after move $moved_file_hash matches original hash $staging_file_hash"
      ### Confirm ready for process
      echo "INFO: $new_path_name/$basename is ready for processing"
   else
      ### If no match, error out and exit, as this shouldn't be possible
      echo "ERROR: $orig_basename hash after move $moved_file_hash does not match original hash $staging_file_hash"
      exit
   fi

done < $processed_file_list

}

# Execute get_lines function to begin processing
get_lines $in_file

