# How To Migrate To A New Server

# Introduction

The purpose of the migration is to get more disk space, and to prove we are
doing the backups correctly.
The idea is that there should be nothing copied from machine to machine.
The purpose here is to only take stuff from the backup storage, and if there's
anything missing, we need to start storing it on the backup storage.
For example: if it proves that the recovery of the DB isn't possible, because
we are missing some configuration file of the DB, we should ensure we're backing
this file up.

## Machines

Old machine `secprod` (let's call it `OLD`)

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


### Step 1
Ensure that we have all the necessary data, such as the backups. 

Each backup is in a compressed format. For the Database, we need to extract it and run certain commands to ensure it works fine and is ready to be used. For the Files, we need to extract them and then run a process that compares the files listed as existing in the database and check if they are actually there.

We also need to ensure that the configuration files we have are working properly and are complete for all services we use in the system. (We can test it by running it on a Vagrant machine and ensuring that it works fine.)

### Step 2
We need to ensure that the new server is up and running with the same OS as the previous one to make sure that we do not face any bugs.

If needed, we run the commands in the `soyoustart.sh` file to have more storage on `/mnt/sec` where we will be having all the files and also the database.

We need to ensure that we have a lot of storage, mounted on `/mnt/sec` because we will have all the DB and Files in that directory.

The `/mnt/sec/db` will have the database, and as of now, it is 79 GB.

The `/mnt/sec/cache` will have the Index (RSS), ZIP, and Data Files, and it is 

The `/mnt/sec/unzipped_cache` will have the uncompressed content from the ZIP files inside `/mnt/sec/cache`.

So firstly we need to ensure that we have enough space to have all this data.

### Step 3
After ensuring the backups and configurations work fine, we can copy the backups to the new server in order to later run the restore process.

### Step 4
Add the IP Address in the Ansible Playbook and run it on the new server.

### Step 5
Since we have Postgres installed after running the Playbook, we can start the process to restore the database from the backup file we retrieved earlier.

### Step 6
Copy over the SEC Binary and start running tests to ensure everything works fine, these tests include:
- Commands work fine
- Server runs properly
- Database is up and running
- Files can be accessed from the CLI and the website

### Step 7
Configure the DNS to point to the IP Address of the new server.
