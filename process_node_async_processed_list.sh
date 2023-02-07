#/bin/bash

# Vars
in_file=/home/admin/10.49.28.170.out
cqlsh=/usr/lib64/GB/DCF/JServices/MbService/bin/cqlsh
select="select * from storage.datasets_by_name"
async_processed_files_dataset=$($cqlsh $(hostname -I) 21205 -e "$select" | grep "ASYNCH PROCESSED FILES FOR" | cut -d"|" -f2 | xargs)


# Args
if [ "$1" = "Execute_Move" ]
then
   echo "WARN: Argument Execute_Move; setting dryrun to false" 
   dryrun=
else
   echo "INFO: No Execute_Move argument; setting dryrun to true"
   dryrun=true
fi


function get_lines () {

while read line;
do
   echo INFO: Processing $line
   target_file=$(echo $line | cut -d"|" -f1)
   basename=$(basename $target_file)
   orig_basename=$(basename $target_file | cut -d"{" -f1)
   dirname=$(dirname $target_file)

   # echo Target File: $target_file
   create_time=$(echo $line | cut -d"|" -f2)
   create_time_formatted=$(date -d "$create_time" +"%Y-%m-%dT%H:%M:%S%z")
   # echo Create Time: $create_time_formatted
   file_size=$(echo $line | cut -d"|" -f3)
   # echo File Size: $file_size
   file_id=$(echo $line | cut -d"|" -f4)
   # echo File Id: $file_id


   # Verify file is in async processed dataset
   dataset=$(gbr file ls -i $file_id -d | grep "parent id" | cut -c25-)
   if [ "$dataset" = "$async_processed_files_dataset" ]
   then
      echo "INFO: file id $file_id for $target_file is in async processed files dataset $async_processed_files_dataset"
   else
      echo "WARN: file id $file_id for $target_file is in dataset $dataset, rather than the async processed files dataset $async_processed_files_dataset; skipping file"
      continue
   fi


   # Check file exists

   if [ -f "$target_file" ]
   then
      echo "INFO: $target_file exists"
   else
      echo "WARN: $target_file does not exist; skipping file"
      continue
   fi

   # Check file size matches
   staging_file_size=$(ls -l $target_file | cut -d" " -f5)
   # echo Staging File Size: $staging_file_size

   if [ "$file_size" -eq "$staging_file_size" ]
   then
      echo "INFO: $orig_basename file size $file_size matches staging file size $staging_file_size"
   else
      echo "WARN: $orig_basename file size $file_size does not match staging file size $staging_file_size; skipping file"
      continue
   fi

   # Check file time matches
   staging_create_time=$(ls -l --time-style="+%Y-%m-%dT%H:%M:%S%z" $target_file | cut -d" " -f6)
   # echo Staging Create Time: $staging_create_time

   if [ "$create_time_formatted" = "$staging_create_time" ]
   then
      echo "INFO: $orig_basename create time $create_time_formatted matches staging create time $staging_create_time"
   else
      echo "WARN: $orig_basename create time $create_time_formatted does not match staging create time $staging_create_time; skipping file"
      continue
   fi

   # Check file id matches with file name
   file_id_checked_name=$(gbr file ls -i $file_id | cut -d" " -f3)

   if [ "$file_id_checked_name" = "$orig_basename" ]
   then
      echo "INFO: basename $orig_basename matches file id checked name $file_id_checked_name"
   else
      echo "WARN: basename $orig_basename does not match file id checked name $file_id_checked_name; skipping file"
      continue
   fi

   echo "INFO: $target_file verified as ready for removal"


   # Hash file
   staging_file_hash=$(sha256sum $target_file | cut -d" " -f1)

   # Move file
   ## Check if mb or datavX staging folder
   if $(echo $dirname | grep -q mb)
   then
      old_path=FAN
      new_path=FAN.processed
   else
      old_path=staging
      new_path=staging.processed
   fi

   ## sed pathname
   new_path_name=$(echo $dirname | sed "s/$old_path/$new_path/g")
   echo "INFO: Staging path $dirname; new path $new_path_name"

   ## Create target folder
   if [ ! -d $new_path_name ]
   then
       mkdir -p $new_path_name
       echo "INFO: Created $new_path_name"
   fi
   ## mv target_file
   cmd="mv $dirname/$basename $new_path_name/$basename"
   echo "INFO: Command $cmd"

   if [ ! -z $dryrun ]
   then
       echo "INFO: Dryrun skipping execute"
       continue
   else
      echo "WARN: Executing command $cmd"
      eval $cmd
   fi

   # Recheck hash

   moved_file_hash=$(sha256sum $new_path_name/$basename | cut -d" " -f1)
   if [ "$staging_file_hash" = "$moved_file_hash" ]
   then
       echo "INFO: $orig_basename hash after move $moved_file_hash matches original hash $staging_file_hash"
       # Confirm ready for process
       echo "INFO: $target_file is ready for processing"
   else
       echo "ERROR: $orig_basename hash after move $moved_file_hash does not match original hash $staging_file_hash"
       exit
   fi


done < $1

}


get_lines $in_file

