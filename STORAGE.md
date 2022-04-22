# File System Options

We have a very large number of files that we need to keep in order to properly serve the necessary data for our website. 

The main files we have in our system are the following:
- **Index Files**: These files include the main data, such as the filings, the companies, and the file links.
- **Filing/Data Files**: These files include the actual data, such as the company details.
- **ZIP Files**: These files are compressed ZIP files that include the files for that specific filing.

In our case, we need to mainly have the Index files as well as the Filing/Data Files. So need to figure our how to shape your file system.

_Note: We will assume that Index files will be downloaded and will exist regardless of the options we have because we need those files in  order to retrieve any data._

## Option 1
The first option, which is what we have right now, is to focus on downloading ZIP files and extracting them, then whatever file we are missing from there, we download them individually, such as in `sec dow data`.

Advantages:
- We can reduce the amount of time we need to wait in order to download the data, that is because we download a compressed version of, let us say, 6 files, instead of downloading them individually.
- If we use `rsync` or mirroring in order to backup our data to a different server, then it would be faster if we move the ZIP files, and then extracting them again on the new server.

Disadvantages:
- We cannot properly compress the files because there are already ZIP files in the directory.


## Option 2
The second option is to download the files individually without downloading any ZIP files. This can be done by running `sec dow data`.

Advantages:
- The application would be simpler, since we do not have to worry about downloading the ZIP files and making sure they have the correct files.
- It would be easier to backup the data, since we can simply compress the directory and move it to a different server.

Disadvantages:
- The download process would take much longer since we would have to download the files individually.

## Option 3
This option can be used along with either of these two options mentioned above. It involves using ZFS, which is a more sophisticated file system that uses things such a snapshots and data integrity checks.

We would divide our data into three parts:
- SEC Files
- Database
- Backups

And inside the backups, we would have snapshots of the ZFS file system that we can use to restore the data. We can also download and set up a package called `zfs-auto-snapshot` to automate that for us.

## Implementation

The below has been implemented on `ca2` server:

There are 4 disks, each being 2 TB of storage.

1. First one (`/dev/sda`) was kept as it is for the system data (mounted at `/`)
2. Second two were added together and mounted on `/mnt/sec` for the SEC files and Database
3. Third one was mounted on `/mnt/backups` so that it keeps the files as backup

Before we can setup the file system, we need to make sure that the disks do not have partitions, we can run the below command for each disk we need to setup:

```
dd if=/dev/zero bs=1m count=1 of=/dev/sdb
```

To make sure it worked fine, run the command `fdisk -l` and make sure that the disk you ran the `dd` command on does not have any partitions. Once all that is done, we can start setting up the file system.


To setup the `/mnt/sec` directory:
```
    pvcreate /dev/sdb
	pvcreate /dev/sdc
	vgcreate sec  /dev/sdb /dev/sdc
	vgdisplay
	lvcreate -n data -l 100%FREE sec
	mke2fs /dev/sec/data
	mkdir /mnt/sec
	mount /dev/sec/data /mnt/sec
```

To mount the `/mnt/backups` directory:
```
    mkdir /mnt/backups
    mount /dev/sdd /mnt/backups
```

This should be the end result:
```
ubuntu@ca2:~$ sudo lsblk -l
NAME     MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
loop0      7:0    0 61.9M  1 loop /snap/core20/1376
loop1      7:1    0 61.9M  1 loop /snap/core20/1405
loop2      7:2    0 67.9M  1 loop /snap/lxd/22526
loop3      7:3    0 67.8M  1 loop /snap/lxd/22753
loop4      7:4    0 43.6M  1 loop /snap/snapd/15177
loop5      7:5    0 44.7M  1 loop /snap/snapd/15534
sda        8:0    0  1.8T  0 disk 
sda1       8:1    0  511M  0 part /boot/efi
sda2       8:2    0  1.8T  0 part /
sda3       8:3    0  512M  0 part [SWAP]
sda4       8:4    0    2M  0 part 
sdb        8:16   0  1.8T  0 disk 
sdc        8:32   0  1.8T  0 disk 
sdd        8:48   0  1.8T  0 disk /mnt/backups
sec-data 253:0    0  3.7T  0 lvm  /mnt/sec
```


The idea of the last disk is that we will first manually run the backing up, in this case archiving, (after all initial downloading is done) in order to get the full files for 2017-2021. Then, we can start running the archiving every month after the new data has been downloaded. 

Along with that, we will also be using `rsync` to sync the files daily to another server as another source of backup. Therefore we have both ways in order to guarantee having the latest data and in two different locations.