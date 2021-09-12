

Lease info

	Expiry date
	12 Mar 2022 Renew
	Datacentre
	BHS2 - Rack: T02C41 - Server ID: 240157
	OS
	Ubuntu Server 20.04 "Focal Fossa" (64bits)
	Boot
	Boot from hard drive (no netboot)
	Main IP
	192.99.161.20
	Commercial name
	E3-SAT-1-16 Server - Xeon E3-1225v2 (4c/4th) - 16GB DDR3 1333 MHz - SoftRaid 3x2To SATA
	Reverse
	ns505854.ip-192-99-161.net
	Backup Storage
	0 / 100 GB


We're not running in a soft-RAID. We're running the system from 1 2TB disk.
We merge 2 2TB disks into 1 4TB filesystem.

To do this, right after 1st login I did:

	pvcreate /dev/sdb
	pvcreate /dev/sdc
	vgcreate my_volume_group  /dev/sdb /dev/sdc
	vgdisplay
	lvcreate -n my_logical_volume -l 100%FREE my_volume_group
	mke2fs /dev/my_volume_group/my_logical_volume
	mkdir /mnt/sec
	mount /dev/my_volume_group/my_logical_volume /mnt/sec
