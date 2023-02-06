#/bin/bash

# Vars
in_file=/home/admin/10.49.28.170.out
cqlsh=/usr/lib64/GB/DCF/JServices/MbService/bin/cqlsh
select="select * from storage.datasets_by_name"
async_processed_files_dataset=$($cqlsh $(hostname -I) 21205 -e "$select" | grep "ASYNCH PROCESSED FILES FOR" | cut -d"|" -f2 | xargs)

function get_lines () {

while read line;
do
   echo INFO: Processing $line
   target_file=$(echo $line | cut -d"|" -f1)
   basename=$(basename $target_file | cut -d"{" -f1)
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
      echo "INFO: $basename file size $file_size matches staging file size $staging_file_size"
   else
      echo "WARN: $basename file size $file_size does not match staging file size $staging_file_size; skipping file"
      continue
   fi

   # Check file time matches
   staging_create_time=$(ls -l --time-style="+%Y-%m-%dT%H:%M:%S%z" $target_file | cut -d" " -f6)
   # echo Staging Create Time: $staging_create_time 

   if [ "$create_time_formatted" = "$staging_create_time" ]
   then
      echo "INFO: $basename create time $create_time_formatted matches staging create time $staging_create_time"
   else
      echo "WARN: $basename create time $create_time_formatted does not match staging create time $staging_create_time; skipping file"
      continue
   fi

   # Check file id matches with file name
   file_id_checked_name=$(gbr file ls -i $file_id | cut -d" " -f3)
   
   if [ "$file_id_checked_name" = "$basename" ]
   then
      echo "INFO: $basename basename $basename matches file id checked name $file_id_checked_name"
   else
      echo "WARN: basename $basename does not match file id checked name $file_id_checked_name; skipping file"
      continue
   fi

   echo "INFO: $target_file verified as ready for removal"


   # Hash file

   # Move file   

   # Recheck hash

   # Confirm ready for process


done < $1

}


get_lines $in_file

