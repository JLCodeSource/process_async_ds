#!/bin/bash
# Description: This script processes a list of files in the async_processed dataset generated from FileGet.jar
# It cleanses the output by dropping any non-backup files, any extracted files or any files in gbfs and
# creates lists of processed files within the staging area on  each of the nodes 
# Input: Pipe separated file with file name|create time|fan ip|fan uri|file size|backup file|file id|file hash|backupkv status
# Usage: 
## Set parameters including source files, output files, and user id & file to process
## Run process_node_async_process_list.sh script on nodes locally with the output file


# Parameters
## Set to work on testing
test=
## Set to input file user id
user=
dir="/home/JLCodeSource/shell/process_async_ds"
outdir=$dir/out
testfile=test_processed_files.out
file=async_processed_files.out


# Constants
ftp='ftp://'
## Regex for ip address & port 2121
ftp_ip_and_port='@([0-9]{1,3}\.){3}[0-9]{1,3}:2121'
## Regex for backup name guid
backup_guid='^[a-fA-F0-9]{8}(-[a-fA-F0-9]{8}){5}\|'
## fanstaging_type ordered by last to first
fanstaging_type=("fan_c2:" "fan_c1:" "fan_c0:" "fan:")
## staging_path ordered by last to first 
staging_path=("/data3/staging" "/data2/staging" "/data1/staging" "/mb/FAN")
# Sites/ip segments for node array
sites=("site1" "site2")
segment=("41" "49")
declare -A nodes
ipstart=101



# --- Environment Prep
## Check for testing & create shorter file to work with if set
if [ ! -z $test ]; then
heap -n10000 $file > $testfile 
file="$testfile" 
fi

## Check for outdir & create if not
if [ ! -d $outdir ]; then
mkdir -p $outdir
fi

## --- Creating nodes = ip to node hashmap
### 2 sites mapping to 2 segments
for i in {0..1};
do 
       for nodenum in {11..26};
       do
                node="${sites[i]}-n${nodenum}";
                lastoctet=$(( $ipstart + $nodenum ))
		        ipaddr="10.${segment[i]}.28.${lastoctet}"
		        nodes[${ipaddr}]="${node}"
        done;
done;

## --- Verifying nodes hashmap
echo "---Hashmap nodes"
echo "---Hashmap size ${#nodes[@]}" 
for ip in "${!nodes[@]}"
do 
    echo "$ip - ${nodes[$ip]}"
done


# --- Data cleansing
## Function to cleanse data file
## Input: Arg1 - source file
## Output: Cleansed file
function data_cleanse_file () {

    # Set Arg1 to local Working File
    local file=$1

    echo Total Lines: $(wc -l $file)
    
    # -- Drop first line
    tail -n+2 $file > tmp_file
    mv tmp_file $file

    # -- Drop Unnecessary Files
    echo Total Files: $(wc -l $file)

    ## Drop files that aren't backup guids
    ### Dump non backup guid files to outdir
    cat $file | grep -v -E "$backup_guid" >> $outdir/dumped_non_backup_guids.out
    ### Filter out non backup guids
    cat $file | grep -E "$backup_guid" > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files without Backup GUID name: $(wc -l $file)

    ## Drop files with a hash (and therefore in gbfs)
    ### Column 8 is hash
    ### Dump files with hash to outdir
    awk -F'|' '$8!="" {print}' $file >> $outdir/dumped_files_with_hash.out
    ### Filter out files with hash
    awk -F'|' '$8=="" {print}' $file > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files with a Hash: $(wc -l $file)

    ## Drop files with no fanip
    ### Column 3 is fan ip
    ### Dump files with null in fan ip to outdir
    awk -F'|' '$3=="null" { print }' $file >> $outdir/dumped_files_with_no_fanip.out 
    ### Filter out files with null in fan ip
    awk -F'|' '$3!="null" { print }' $file > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files with No FanIP: $(wc -l $file)

    ## Drop files with no fanuri
    ### Column 4 is fan uri
    ### Dump files with null in fan uri to outdir
    awk -F'|' '$4=="null" { print }' $file >> $outdir/dumped_files_with_no_fanuri.out
    ### Filter out files with null in fan ip
    awk -F'|' '$4!="null" { print }' $file > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files with No FanURI: $(wc -l $file)

    ## Drop files with Extracted in backupkv (set based on the existence or not of backupkv data object)
    ### Dump files with extracted set to outdir
    cat $file | grep "backupkv Extracted" >> $outdir/dumped_files_with_extracted.out
    ### Filter out files with backupkv Extracted
    cat $file | grep -v "backupkv Extracted" > tmp_file
    mv tmp_file $file
    echo Total Files After Removing Files with Backup Info: $(wc -l $file)


    # -- Tidy FanURI

    ## Strip ftp://
    sed -i "s;$ftp;;g" $file

    ## Strip user id
    sed -i "s;$user;;g" $file

    ## Strip ip and port
    sed -i -E "s;$ftp_ip_and_port;;g" $file

    # Switch fan/fan_cold[1,2] to staging path
    for i in {0..3}; do
        echo swap: "${fanstaging_type[$i]}" fan_to_staging: "${staging_path[$i]}"
        sed -i "s;${fanstaging_type[$i]};${staging_path[$i]};g" $file
    done
    echo Swapped fanpath to stagingpath

    # Reorder & drop unnecessary fields
    ## Column 4 is fan uri
    ## Column 2 is date
    ## Column 5 is file size
    ## Column 7 is file id
    ## Column 3 is fan ip
    awk -F'|' '{ print $4 "|" $2 "|" $5 "|" $7 "|" $3 }' $file > tmp_file
    mv tmp_file $file
    echo Reordered and removed fields

    # Sort by file size (new "key" 3) largest first
    sort -t'|' -rnk3 $file > tmp_file
    mv tmp_file $file
    echo Ordered list by size - largest first


    echo Cleaned File
}

# --- Splitting by node
## Function to split the cleansed file into per node files
## Input: Arg1 - source file
## Output: Processed file list per node in outdir
function node_split () {

# Set Arg1 to local Working File
local file=$1

# Step through nodes array and write to file
for ip in "${!nodes[@]}"; 
do 
    echo "--- $ip"
    cat $file | grep $ip > tmp_file
    ## Where the ip column is the node ip output (new) columns 1-4 to outdir/ip.out
    awk -F'|' '$5="$ip" { print $1 "|" $2 "|" $3 "|" $4 }' tmp_file > "$outdir/$ip".out
    echo $(wc -l "$outdir/$ip.out")
done

}

# --- Clean Up
## Function to clean up after run
function clean_up () {

    rm tmp_file

}


# --- Main
## Function to init the script 
function main () {

# Make copy of file
cp $file $outdir/$file
file=$outdir/$file

data_cleanse_file $file

node_split $file

clean_up

}


# Begin
main
