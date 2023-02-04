#!/bin/bash

# Vars
test=1
user=nmr
dir="/home/JLCodeSource/shell/process_async_ds"
outdir=$dir/out
testfile=test_processed_files.out
file=async_processed_files.out


# Consts
ftp='ftp://'
ftp_ip_and_port=([0-9]{1,3}\.){3}[0-9]{1,3}:2121
backup_guid='[a-fA-F0-9]{8}(-[a-fA-F0-9]{8}){5}'
fanstagingtype=("fan_c2:" "fan_c1:" "fan_cold" "fan")
stagingpath=("/data3/staging" "/data2/staging" "/data1/staging" "/mb/FAN")
sites=("site1" "site2")
segment=("41" "49")
declare -A nodes
declare -A fan_to_staging
ipstart=101



# --- Env Prep

## Testing
if [ ! -z $test ]; then
heap -n10000 $file > $testfile 
file="$testfile" 
fi

if [ ! -d $outdir ]; then
mkdir $outdir
fi

## --- Creating nodes = ip to node hashmap

## 2 sites mapping to 2 segments
for i in {0..1};
do 
       for nodenum in {11..26};
       do
                node="${sites[i]}-n${nodenum}";
                # echo $node;
		lastoctet=$(( $ipstart + $nodenum ))
		# echo $lastoctet
		ipaddr="10.${segment[i]}.28.${lastoctet}"
		# echo $ipaddr
		nodes["${ipaddr}"]="${node}"
        done;
done;

## --- Verifying nodes hashmap
echo "---Hashmap nodes"
echo "---Hashmap size ${#nodes[@]}" 
for ip in "${!nodes[@]}"; do echo "$ip - ${nodes[$node]}"; done

## --- Creating fan_to_staging hashmap

for i in {0..3}; 
do
    fan_to_staging["${fanstagingtype[i]}"]=${stagingpath[i]}
done
echo "---Hashmap fan_to_staging"
for staging in "${!fan_to_staging[@]}"; do echo $staging - ${fan_to_staging[$staging]}; done


# --- Data cleansing
function data_cleanse_file () {

    # -- Drop Unnecessary Files

    echo Total Files: $(wc -l $1)

    ## Drop files that aren't backup guids
    cat $1 | grep -v -E "$backup_guid" > $outdir/dumped_non_backup_guids.out
    cat $1 | grep -E "$backup_guid" > tmp_file
    mv tmp_file $1
    echo Total Files After Removing Files without Backup GUID name: $(wc -l $1)

    ## Drop files with a hash
    awk -F'|' '$8!="" {print}' $1 > $outdir/dumped_files_with_hash.out
    awk -F'|' '$8=="" {print}' $1 > tmp_file
    mv tmp_file $1
    echo Total Files After Removing Files with a Hash: $(wc -l $1)

    ## Drop files with no fanip
    awk -F'|' '$3=="null" { print }' $1 > $outdir/dumped_files_with_no_fanip.out 
    awk -F'|' '$3!="null" { print }' $1 > tmp_file
    mv tmp_file $1
    echo Total Files After Removing Files with No FanIP: $(wc -l $1)

    ## Drop files with no fanuri
    awk -F'|' '$4=="null" { print }' $1 > $outdir/dumped_files_with_no_fanuri.out
    awk -F'|' '$4!="null" { print }' $1 > tmp_file
    mv tmp_file $1
    echo Total Files After Removing Files with No FanURI: $(wc -l $1)

    ## Drop files with Extracted
    cat $1 | grep "backupkv Extracted" > $outdir/dumped_files_with_extracted.out
    cat $1 | grep -v "backupkv Extracted" > tmp_file
    mv tmp_file $1
    echo Total Files After Removing Files with Backup Info: $(wc -l $1)


    # -- Tidy FanURI

    ## Strip ftp
    sed -i "s;$ftp;;g" $1

    ## Strip user
    sed -i "s;:$user;;g" $1

    ## Strip ip
    sed -i -E "s;$ftp_ip_and_port;;g" $1

    # Switch fan/fan_cold[1,2] to staging path
    for swap in "${!fan_to_staging[@]}"; do
        echo swap: $swap fan_to_staging: ${fan_to_staging[$swap]}
        sed -i "s;$swap;${fan_to_staging[$swap]};g" $1
    done

    # Reorder & drop unnecessary fields
    awk -F'|' '{ print $4 "|" $2 "|" $5 "|" $7 "|" $3 }' $1 > tmp_file
    mv tmp_file $1

    echo Cleaned File
}

# --- Splitting by node
function node_split () {

for ip in "${!nodes[@]}"; 
do 
    echo "--- $ip"
    cat $1 | grep $ip > tmp_file
    awk -F'|' '$5="$ip" { print $1 "|" $2 "|" $3 "|" $4 }' tmp_file > "$outdir/$ip".out
    echo $(wc -l "$outdir/$ip.out")
done

}

# Clean Up
function clean_up () {

    rm tmp_file

}


function main () {

data_cleanse_file $file

node_split $file

clean_up

}

main


# function get_variables () {

# local IFS="|"

# count=0

# while read lines; do 
#     echo line $count
#     #get_variables $lines
#     line=($lines)
#     filename=${line[0]}
#     echo filename = $filename
#     created_date=${line[1]}
#     echo created_date = $created_date
#     fan_ip=${line[2]}
#     echo fan_ip = $fan_ip
#     fan_uri=$(echo ${line[3]} | sed "s;${ftp}${fanstagingtype[i]}${NMR};${stagingpath[i]}/;g")
#     echo fan_uri = $fan_uri
#     file_size=${line[4]}
#     echo file_size = $file_size
#     is_backup_file=${line[5]}
#     echo is_backup_file = $is_backup_file
#     file_id=${line[6]}
#     echo file_id = $file_id
#     file_hash=${line[7]}
#     echo file_hash = $file_hash
#     has_backup_info=${line[8]}
#     echo has_backup_info = $has_backup_info

#     let count=(count+1)

# done < $1

# }

# get_variables $file