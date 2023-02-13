#!/bin/bash

# Purpose: This script processes a list of files in the async_processed dataset

# Vars
test=
user=nmr
dir="/home/JLCodeSource/shell/process_async_ds"
outdir=$dir/out
testfile=test_processed_files.out
file=async_processed_files.out


# Consts
ftp='ftp://'
ftp_ip_and_port=([0-9]{1,3}\.){3}[0-9]{1,3}:2121
backup_guid='[a-fA-F0-9]{8}(-[a-fA-F0-9]{8}){5}'
fanstaging_type=("fan_c2:" "fan_c1:" "fan_c0:" "fan:")
staging_path=("/data3/staging" "/data2/staging" "/data1/staging" "/mb/FAN")
sites=("site1" "site2")
segment=("41" "49")
declare -A nodes
ipstart=101



# --- Env Prep

## Testing
if [ ! -z $test ]; then
heap -n10000 $file > $testfile 
file="$testfile" 
fi

if [ ! -d $outdir ]; then
mkdir -p $outdir
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
		nodes[${ipaddr}]="${node}"
        done;
done;

## --- Verifying nodes hashmap
echo "---Hashmap nodes"
echo "---Hashmap size ${#nodes[@]}" 
for ip in "${!nodes[@]}"; do echo "$ip - ${nodes[$ip]}"; done


# --- Data cleansing
function data_cleanse_file () {

    # Working File
    local file=$1

    # -- Drop first line
    echo Total Lines: $(wc -l $file)
    tail -n+2 $file > tmp_file
    mv tmp_file $file

    # -- Drop Unnecessary Files
    # N.B. Preprocess to remove top line
    echo Total Files: $(wc -l $file)

    ## Drop files that aren't backup guids
    cat $file | grep -v -E "$backup_guid" > $outdir/dumped_non_backup_guids.out
    cat $file | grep -E "$backup_guid" > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files without Backup GUID name: $(wc -l $file)

    ## Drop files with a hash
    awk -F'|' '$8!="" {print}' $file > $outdir/dumped_files_with_hash.out
    awk -F'|' '$8=="" {print}' $file > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files with a Hash: $(wc -l $file)

    ## Drop files with no fanip
    awk -F'|' '$3=="null" { print }' $file > $outdir/dumped_files_with_no_fanip.out 
    awk -F'|' '$3!="null" { print }' $file > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files with No FanIP: $(wc -l $file)

    ## Drop files with no fanuri
    awk -F'|' '$4=="null" { print }' $file > $outdir/dumped_files_with_no_fanuri.out
    awk -F'|' '$4!="null" { print }' $file > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files with No FanURI: $(wc -l $file)

    ## Drop files with Extracted
    cat $file | grep "backupkv Extracted" > $outdir/dumped_files_with_extracted.out
    cat $file | grep -v "backupkv Extracted" > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files with Backup Info: $(wc -l $file)


    # -- Tidy FanURI

    ## Strip ftp
    sed -i "s;$ftp;;g" $file

    ## Strip user
    sed -i "s;$user;;g" $file

    ## Strip ip
    sed -i -E "s;$ftp_ip_and_port;;g" $file

    # Switch fan/fan_cold[1,2] to staging path
    # Order
    for i in {0..3}; do
        echo swap: "${fanstaging_type[$i]}" fan_to_staging: "${staging_path[$i]}"
        sed -i "s;${fanstaging_type[$i]};${staging_path[$i]};g" $file
    done
    echo Swapped fanpath to stagingpath

    # Reorder & drop unnecessary fields
    awk -F'|' '{ print $4 "|" $2 "|" $5 "|" $7 "|" $3 }' $file > tmp_file
    mv tmp_file $file
    echo Reordered and removed fields

    # Sort by file size
    sort -t'|' -rnk3 $file > tmp_file
    mv tmp_file $file
    echo Ordered list by size - largest first


    echo Cleaned File
}

# --- Splitting by node
function node_split () {

local file=$1

for ip in "${!nodes[@]}"; 
do 
    echo "--- $ip"
    cat $file | grep $ip > tmp_file
    awk -F'|' '$5="$ip" { print $1 "|" $2 "|" $3 "|" $4 }' tmp_file > "$outdir/$ip".out
    echo $(wc -l "$outdir/$ip.out")
done

}

# Clean Up
function clean_up () {

    rm tmp_file

}


function main () {

# Make copy of file
cp $file $outdir/$file
file=$outdir/$file

data_cleanse_file $file

node_split $file

clean_up

}

main
