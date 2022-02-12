# How To Migrate To A New Server

### Step 1
We need to ensure that we have a lot of storage, mounted on `/mnt/sec` because we will have all the DB and Files in that directory.

The `/mnt/sec/db` will have the database, and as of now, it is 79 GB.

The `/mnt/sec/cache` will have the Index (RSS), ZIP, and Data Files, and it is 

The `/mnt/sec/unzipped_cache` will have the uncompressed content from the ZIP files inside `/mnt/sec/cache`.

So firstly we need to ensure that we have enough space to have all this data.

### Step 2
We need to load the DB. We have the DB backed up in the `secprod` compressed using `pg_basebackup` so need to load the Database from there

### Step 3
Add the IP Address in the Ansible Playbook and run it on the new server.

### Step 4
Configure the DNS to point to the IP Address of the new server.
