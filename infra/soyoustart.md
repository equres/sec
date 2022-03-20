

## Lease info

	Expiry date
	20 Mar 2022 Renew
	Datacentre
	BHS4 - Rack: T04B26 - Server ID: 252472
	OS
	Ubuntu Server 20.04 "Focal Fossa" (64bits)
	Boot
	Boot from hard drive (no netboot)
	Main IP
	158.69.54.118
	Commercial name
	SYS-1-SAT-32 Server - Intel Xeon D1520 - 32GB DDR3 ECC 2133MHz - 4x 2To HDD Soft RAID
	Reverse
	ns519217.ip-158-69-54.net
	Backup Storage
	-

We're not running in a soft-RAID.
We're running the system from 1 2TB disk.
We merge all other disks into one large filesystem without any redundancy (striping).

## Root/swap Disk setup

If the system is newly setup, you must remember that the labels on the root/swap filesystems
must be set properly.
Don't depend on UUIDs because they change if you change disk/server.
We should only depend on labels, because this ensures that provisioning the system from the
Ansible playbook will result in this system being operational.
Otherwise booting the system won't be possible.

Let's list available labels:

	root@ca1:/home/ubuntu# ls -la /dev/disk/by-label/
	total 0
	drwxr-xr-x 2 root root 120 Mar 19 21:10 .
	drwxr-xr-x 8 root root 160 Mar 19 21:10 ..
	lrwxrwxrwx 1 root root  10 Mar 19 21:11 EFI_SYSPART -> ../../sda1
	lrwxrwxrwx 1 root root  10 Mar 19 21:11 config-2 -> ../../sda4
	lrwxrwxrwx 1 root root  10 Mar 19 21:18 root -> ../../sda2
	lrwxrwxrwx 1 root root  10 Mar 19 21:11 swap-sda3 -> ../../sda3

Root is there and looks OK. The swap is weird. Let's change it:

	root@ca1:/home/ubuntu# swaplabel /dev/sda3
	LABEL: swap-sda3
	UUID:  fa823922-00df-4e9b-b99e-138d3922ef56
	root@ca1:/home/ubuntu# swaplabel -L swap /dev/sda3
	root@ca1:/home/ubuntu# swaplabel /dev/sda3
	LABEL: swap
	UUID:  fa823922-00df-4e9b-b99e-138d3922ef56

## Data volume setup

To do this, right after 1st login I did:

	pvcreate /dev/sdb
	pvcreate /dev/sdc
	vgcreate my_volume_group  /dev/sdb /dev/sdc
	vgdisplay
	lvcreate -n my_logical_volume -l 100%FREE my_volume_group
	mke2fs /dev/my_volume_group/my_logical_volume
	mkdir /mnt/sec
	mount /dev/my_volume_group/my_logical_volume /mnt/sec


	root@ca1:/home/ubuntu# fdisk -l

	....

	Disk /dev/sda: 1.84 TiB, 2000398934016 bytes, 3907029168 sectors
	Disk model: HGST HUS726020AL
	Units: sectors of 1 * 512 = 512 bytes
	Sector size (logical/physical): 512 bytes / 512 bytes
	I/O size (minimum/optimal): 512 bytes / 512 bytes
	Disklabel type: gpt
	Disk identifier: 3E78A84C-13C5-41D4-90F6-6E691E8ED545

	Device          Start        End    Sectors  Size Type
	/dev/sda1        2048    1048575    1046528  511M EFI System
	/dev/sda2     1048576 3905972223 3904923648  1.8T Linux filesystem
	/dev/sda3  3905972224 3907020799    1048576  512M Linux filesystem
	/dev/sda4  3907025072 3907029134       4063    2M Linux filesystem


	Disk /dev/sdb: 1.84 TiB, 2000398934016 bytes, 3907029168 sectors
	Disk model: HGST HUS726020AL
	Units: sectors of 1 * 512 = 512 bytes
	Sector size (logical/physical): 512 bytes / 512 bytes
	I/O size (minimum/optimal): 512 bytes / 512 bytes


	Disk /dev/sdc: 1.84 TiB, 2000398934016 bytes, 3907029168 sectors
	Disk model: HGST HUS726020AL
	Units: sectors of 1 * 512 = 512 bytes
	Sector size (logical/physical): 512 bytes / 512 bytes
	I/O size (minimum/optimal): 512 bytes / 512 bytes


	Disk /dev/sdd: 1.84 TiB, 2000398934016 bytes, 3907029168 sectors
	Disk model: HGST HUS726020AL
	Units: sectors of 1 * 512 = 512 bytes
	Sector size (logical/physical): 512 bytes / 512 bytes
	I/O size (minimum/optimal): 512 bytes / 512 bytes


	Disk /dev/mapper/sec_data-sec_lv: 5.47 TiB, 6001193385984 bytes, 11721080832 sectors
	Units: sectors of 1 * 512 = 512 bytes
	Sector size (logical/physical): 512 bytes / 512 bytes
	I/O size (minimum/optimal): 512 bytes / 512 bytes

Looking at the output, we see that the correctly merged volume is ~5.5TB.
Let's see what is its label:

	root@ca1:/home/ubuntu# e2label /dev/mapper/sec_data-sec_lv

Looks like the label is empty. Let's make it something more reasonable.

	root@ca1:/home/ubuntu# e2label /dev/mapper/sec_data-sec_lv sec_data
	root@ca1:/home/ubuntu# e2label /dev/mapper/sec_data-sec_lv
	sec_data
	root@ca1:/home/ubuntu# ls /dev/disk/by-label/
	EFI_SYSPART  config-2  root  sec_data  swap

Now we have all the labels.
