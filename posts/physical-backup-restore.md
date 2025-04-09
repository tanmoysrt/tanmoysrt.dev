---
title: "Physical Backup & Restoration"
description: "Physical Backup & Restoration"
date: "2025-04-09"
is_redirect: false
redirect_url:
---


In Frappe Cloud, every **Site Update** triggers an automatic backup - ensuring your data stays safe in case of failures and we can rollback your site to older version as soon as possible.

As of now, we use `mysqldump` to take logical backup your site's database.

```
mysqldump \
  --single-transaction --quick \
  --lock-tables=false \
  -u root -psecret \
  _c678923hj223 \
  > backup.sql
```

This works fine !

### Issues

1. **High Disk I/O :** When backing up your large database, MariaDB needs to read all data from disk and convert each table/row into SQL queries. As your database grows:
	- The database need to perform more work
	- Required more disk i/o
	- Impact overall server performance

2. **Backup Time :** In `mysqldump` based backup process, the duration depends completely on database size.

	Each operation requires :
	- Reading data from disk
	- Converting each table definition and rows to SQL queries
	- Write those data to \*.sql file.

	So, the backup duration grows proportionally with database size.

	Take a glimpse of that -
	![](https://frappe.io/files/logical-backup.png)

	If you have some site with more than 10GB of database size, you need to wait ~10 minutes before even starting with actual update.


	> BTW, Frappe Cloud gives you an option to **Skip Backup** as well. 
	> 
	> But that's like playing with 🔥 in a production site. If something goes wrong, you need to fix it on production yourself. Resulted in more downtime.

    <p align="center">
        <img style="max-width: 300px" src="/files/meme435384.png" />
    </p>


### Simple Solution - Take Disk Snapshot

Just take disk snapshot of database server because AWS EBS Snapshot, ZFS Snapshot is just a few seconds operation.

Technically, we have just did **Physical Backup** of database.

But there is an **issue** with this approach -

Running Database Server always has some data in memory. So, if we just take snapshot without stopping the database server, we will end-up with some corrupted files.

So, we need to stop database for this.

Meanwhile, on Frappe Cloud on each database server, we have hundreds of database running.

![](https://frappe.io/files/fc-db-infra.png)

So, It's impossible to shut down database server to backup one database.

### Solution To Physical Backup

To create a safe physical backup, we need to ensure two things -

1. There should be no activity in that specific database while taking snapshot.
2. Database should commit all data from memory to disk.

**Solution to Problem 1** -

Take Read Write lock on the database to ensure no activity on database tables.

**Solution to Problem 2** - 

This depends on storage engine. MariaDB has different storage engines like InnoDB, MyISAM, Area, NBD which handle the storage mechanisms.

Out of these engines, most popular ones for MariaDB are - 

- InnoDB
- MyISAM

In InnoDB, every table is considered as tablespace. Each InnoDB tables has one file in disk
- **\*.ibd** - This file hold actual data of your tables. This is the file we need to backup mainly.

InnoDB has support for **Transportable Tablespace**. That means we can export and import tablespace in live database.

For MyISAM, there is no concept of tablespace. Each MyISAM table contains two files
- **\*.MYI** - Index for the stored data
- **\*.MYD** - Actual data

MariaDB provides a single SQL query, which can solve problem 1 & 2 together.

Open a terminal and run,
```
FLUSH TABLES `tabUser`, `__global_search`, `tabVersion` FOR EXPORT
```

> Don't close the terminal or connection until unless you have backed up the files.

This will take lock on database and flush on-the-fly data to disk. So that we can copy the data.

At this point, if you take a look at the database folder you will find different kind of files.
```list
/var/lib/mysql/
	|- _cdsd32dsvn92
		|- __global_search.frm  <- No need to backup
 		|- __global_search.MYI  <- Backup
		|- __global_search.MYD  <- Backup
		|- tabUser.cfg          <- Backup
		|- tabUser.ibd          <- Backup
		|- tabUser.frm          <- No need to backup
		...
```

Out of this, **\*.cfg** file is the new one for InnoDB. It's important to import the tablespace back in another database server. That file holds information about your **\*ibd** files.

We have now all the data with us. But, we are missing one thing still. If you just look inside your database servers `/var/lib/mysql/` folder, you will find many **\*.frm** file in sub-directories.

This **\*.frm** files hold the structure and column definitions of a table but these are not importable in live database 😢.

 So, we need to take backup of table structure to recreate them back later. We can use mysqldump for that purpose.

```
mysqldump _cdsd32dsvn92 --no-data > schema.sql
```

You can either now take disk snapshot or copy those files to another directory/disk/remote server whatever you prefer .

_FYI, In Frappe Cloud, currently we are using AWS EBS Snapshot for physical backup purpose. Because EBS Snapshot is time consistent and take 3~5 seconds to create._

🎉 Finally, we have our physical backup ready which consists the data + table definitions.

### Time For Physical Restoration

We will first take the simple approach to do the restoration and later will discuss about each problem and their solution.

**Steps -**

1. **Download Backup :** Before restoration, ensure you have the physical backup ready.
	- If using an **EBS / Disk Snapshot**, create a volume from it and mount it on your server.
	- For other backup types, download the files to your sever and mount it somewhere.

2. **Drop Table :** Whichever table you want to import, you need to drop those tables first one by one.

	```
	DROP TABLE <table-name>
	``` 
	⚠️ Caution: This permanently deletes the table and its data.

3. **Recreate Tables from schema:** If you remember, we have dump the database schema to a file using `mysqldump` utility during physical backup. 

	You need to find the schema for required tables from that `schema.sql` file and run that SQL query.
	
	 If you need help for extracting `CREATE TABLE` query for a table with regex, take a look [here](https://github.com/frappe/agent/blob/69084c59fda37e5ee08854e0a893d01116c2b303/agent/database_physical_restore.py#L387-L403). 

4. **Restore Data :** Data restoration varies by storage engine. Let's see one by one.
	1. **InnoDB Restoration (Using Transportable Tablespaces)** 
		- **Discard the tablespace :** 
			```
			ALTER TABLE tabUser DISCARD TABLESPACE;
			```
			If you check inside the `/var/lib/mysql/<database-name>` folder, you will notice `tabUser.ibd` has been disappeared.
		- **Copy backup files :**
		  Paste `tabUser.ibd` (data) + `tabUser.cfg` (metadata) from backup to the same folder.
	    - **Import tablespace :**
			```sql
			ALTER TABLE tabUser IMPORT TABLESPACE;	
			```
		 If the above step gets completed, that means you have successfully imported the data to that live database.

	2. **MyISAM Restoration :**
		- Acquire a write lock on the table.
		- Copy these files from backup to `/var/lib/mysql/<database-name>`:
			- `*.MYI` (index)
			- `*.MYD` (data)

6. Physical  Restoration Completed 🤞

---

### Problems and Solutions

While physical backup/restore solves performance issues and seems straightforward in theory, we quickly discovered real-world challenges when implementing it on Frappe Cloud for medium-sized sites. Unexpected failures began occurring during our initial tests.

To properly test the method, We ended up writing a Bulk Backup and Restoration Tool in FC for finding out maximum possible bugs in this process. 

Everyday we take 40~50 backups and restore the data multiple times in a dummy site.

![](https://frappe.io/files/bulk-restore-tool.png)


#### Missing \*.cfg files in Physical Backup

We started seeing some failures where \*.cfg files are missing - preventing InnoDB table imports. This was confusing since AWS EBS Snapshots are supposed to be time consistent.

After some debugging, we found the root cause :

- MariaDB calls `fdatasync` after disk writes to flush data but it doesn't flush the metadata.
- While the database wrote the \*.cfg files, their metadata remained in memory buffers.
- Result: Snapshots captured 0-byte \*.cfg files

**Solution -** To mitigate this issue, we did `fsync` for the required `cfg` files before taking the snapshot.

```python
# Force metadata flush using fsync
# Ref: https://docs.python.org/3/library/os.html#os.fsync
with open(file_path, "rb", buffering=0) as f:
	os.fsync(f.fileno())
```

**Added Safeguards -** 

1. During Backups:
	- Store exact file sizes for all files
	- Calculate checksums for all \*.cfg files
2. During Restoration:
	- Verify file sizes first
	- Verify checksum of \*.cfg files
	- Abort Restoration if any inconsistencies are detected

#### Handle Broken MyISAM Tables

MyISAM tables are prone to corruption. They corrupt easily but fortunately are just as easy to fix! 😅 

**Common reasons of corruption -**

1. Improperly closed database connections
2. Unexpected database server crashes
3. Hardware failure

In most of the cases,

- The index file (**\*MYI**) got corrupted.
- The data file (**\*MYD**) stays safe.


MyISAM comes with a utility called [`myisamchk`](https://mariadb.com/kb/en/myisamchk/) to check and fix corruption and you don't need to run database to do the fixtures.

Just run,
```
myisamchk -r /var/lib/mysql/_c8383gdu339jd8/__global_search
```

It can solve 99% of problem. So, before attempting restoration we check for corruption in MyISAM table files and fix corruption.

If that fails, we will not attempt restoration.

After restoration also, we found some instances where MyISAM Table was marked as corrupted.

![](https://frappe.io/files/myisam-corruption.png)

In this MyISAM restoration procedure, we just copy files. That doesn't update the index position and some metadatas.

So after restoration we check for corruption using

```
CHECK TABLE <table-name>
```

If we find any error, we will repair that using
```
REPAIR TABLE <table-name>
```

> We are using `CHECK TABLE` instead of `myisamchk` in post-restoration phase because it's unsafe to use `myismchk` tool while database is running.

#### The FULLTEXT Index Restoration Challenge

Unlike Logical Restoration, In Physical Restoration we have the ability to restore table indexes as well. We don't need to recreate those.

But, FULLTEXT indexes are not like other B-Tree indexes -

- It has its own data structure.
- Store index data in `FTS_xxxxxxxx.ibd` files.
- Doesn't support `FLUSH TABLES ... FOR EXPORT`, so it can't be restored back ([docs](https://dev.mysql.com/doc/refman/8.4/en/innodb-table-import.html)).

So the initial solution was -

1. Drop the FULLTEXT Index
2. Recreate it

**Result:** Instant database crash 💀 at the time of fulltext index recreation (even with 4GB `innodb_buffer_pool_size` for a 12MB table!).

![](https://frappe.io/files/innodb-fts-index-issue.png)

**The Root Cause**

- `DROP INDEX` only updates metadata ([docs](https://dev.mysql.com/doc/refman/8.4/en/innodb-table-import.html))
- Orphaned FULLTEXT index metadata remained in the tablespace
- Recreating the index collided with some metadata in tablespace and end-up crashing the database server.

**The Fix**
```sql
-- 1. Remove the corrupted index
ALTER TABLE your_table DROP INDEX fulltext_index_name;

-- 2. Force to fix corruption in innodb table or indexes
OPTIMIZE TABLE your_table;

-- 3. Add FULLTEXT index back
ALTER TABLE your_table ADD FULLTEXT(fulltext_index_name) (columns);
```

**Why it works ?**

- `OPTIMIZE TABLE` repairs/rebuild innodb table and fix corrupted indexes
- Then it acts as a "clean slate" for FULLTEXT index recreation

This solution worked flawlessly 😀.

#### The Snapshot Availability Bottleneck

While AWS EBS snapshots are created instantly, they’re unusable until the status changes from **`Pending`** → **`Available`**.

**What’s Happening Behind the Scenes?**

- AWS snapshot your disk and moves snapshot data to Cold Storage / S3
- Duration depends on **delta** from the last snapshot

![](https://frappe.io/files/snapshot.png)

Since Frappe Cloud already keeps **24hr** & **48hr** disk snapshots that make this process bit faster, but not much.

To make it more faster, we have implemented **Rolling Snapshot**. 

Every 2hr, FC will take a snapshot of all the database servers disk and once the new snapshot become available, we will delete the older one.

This process helps to reduce the Snapshot availability time by ~50%

#### Slow File Copy Due to Lazy Loading

Copying file from backup disk to main disk is one of the main process of physical backup restoration. We start noticing two things :

- 10GB of files took ~16 minutes
- Max speed capped at ~5MBPS

**Root Cause**

AWS provision volume from snapshot instantly but it can't provide expected performance due to lazy loading.

Initially, EBS volume is on uninitialized state. It loads the disk blocks as on demand.

![](https://frappe.io/files/ebs-lazy-loading.png)

For this, we are not able to cross more than ~5MBPS read speed. 

**Failed Experiments**

1. **AWS-recommended process for pre-warming -** AWS suggests to use `dd` or `fio` to pre-warm your volume ([ref](https://docs.aws.amazon.com/ebs/latest/userguide/ebs-initialize.html)). But, we are not able to cross more than 15~20 MBPS read speed due to sequential read of `dd`.
2. **Parallel `dd` processes -**  We thought let's run 5 dd process parallaly to get ~80MBPS speed. But, that causes -
        - Severe CPU overhead
        - High iowait congestion

**The Solution: io_uring**

We built a custom library leveraging Linux’s [`io_uring`](https://en.wikipedia.org/wiki/Io_uring) which provides support for asynchronous IO.

The tool is written in **Golang** and we wrote a python wrapper to make it easy to integrate.

Thus, we are able to cross almost ~300MBPS (90% of EBS gp3 disk) speed at pre-warm stage without much cpu overhead.

**Results**

Warmup time for 10GB files reduced from **16min** → **1min**

We have published the library on pypi, you can take a look here - [filewarmer](https://pypi.org/project/filewarmer/)

If you want to know more about io_uring, check this [blog by mattermost](https://mattermost.com/blog/iouring-and-go/).



### Benefits of Physical Backup

1. **Lightning-Fast Backups (15-50 Seconds)** : For larger sites (DB size > 1GB), physical backup eliminate size-dependent backup delays.

	Logical Backup - 
	![](https://frappe.io/files/logical-backup.png)
	The duration scales with database size

	Physical Backup -
	![](https://frappe.io/files/physical-backup.png)

	Duration almost stays same (15~50 secs) for any database size.


2. **Low CPU and Disk I/O Overhead :** 
  Logical backup force databases to :
   - Scan the files from disk
   - Convert the binary data to SQL queries
   - That can take spike CPU or I/O usage.

  For example, 
  **Backup of 32GB Database on 8vCPU/64GB Server** 

  ***Logical Backup:***
  ![](https://frappe.io/files/logical-backup-cpu.png)
  ![](https://frappe.io/files/logical-backup-disk-io.png)
  ![](https://frappe.io/files/logical-backup-io-usage.png)

  In case of ***Physical Backup*** of the same site, the stats looks like this -	
  ![](https://frappe.io/files/physical-backup-cpu.png)
  ![](https://frappe.io/files/physical-backup-disk-io.png)
  ![](https://frappe.io/files/physical-backup-io-usage.png)

  Very low overhead on CPU or Disk IO usage while using Physical Backup.

### Benefits Of Physical Restoration - 

**1. Minimal System Impact**

- Low CPU / Disk IO overhead during restoration
- Faster recovery for large databases

**2. Comparable Time to Logical Restoration**

Physical Restoration is going to take almost same time as logical restoration because of **Snapshot availability delays**

Snapshot availability time depends on multiple factors -

- AWS backend processes.
- Delta from previous snapshot.
- Network condition of server as well.

**It's a win-win situation**, because -

- **92% of large sites (>1GB DB)**: Migrate successfully with *no restoration* required, so 92% sites going to see huge benefit.
- **8% needing restoration**: Takes the same time as logical backups (not worse than before)
- **Critical improvement**: Backups now take ~30 sec on avg (vs. couple of mins to hours for large sites)
- **No risky compromises**: Users no longer need to opt for `Skipped Backups` during `Site Update` to minimize downtime.

**We’re Still Optimizing!**

We'll soon add this solution as the standard backup/restore method for large sites (opt-in for users).

---

**Learning Resources -**

- [MariaDB Documentation on Importable Tablespace](https://mariadb.com/kb/en/innodb-file-per-table-tablespaces/)
- [AWS EBS Initialization](https://docs.aws.amazon.com/ebs/latest/userguide/ebs-initialize.html)
- [MySQL Documentation on FULLTEXT Index](https://dev.mysql.com/doc/refman/8.4/en/innodb-fulltext-index.html)
- [myisamchk Documentation](https://mariadb.com/kb/en/myisamchk/)
- [Mattermost's blog on io_uring](https://mattermost.com/blog/iouring-and-go/)
- [Source Code of Implementation of Physical Backup in FC](https://github.com/frappe/agent/blob/master/agent/database_physical_backup.py)
- [Source Code of Implementation of Physical Restore in FC](https://github.com/frappe/agent/blob/master/agent/database_physical_restore.py)
