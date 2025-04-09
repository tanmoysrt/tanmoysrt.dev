---
title: "Physical Backup & Restoration"
description: "Physical Backup & Restoration"
date: "2025-04-09"
is_redirect: false
redirect_url:
---


---
In Frappe Cloud, every site update triggers an automatic backup - ensuring your data stays safe in case of failures and we can rollback your site to older version as soon as possible.

As of now, we use `mysqldump` to take logical backup your site's database.

```
mysqldump \
  --single-transaction --quick \
  --lock-tables=false \
  -u root -psecret \
  _c678923hj223 \
  > backup.sql
```

This works fine, but as your database grows, there are some challenges that come with using this method.

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


	> BTW, Frappe Cloud gives you an option to skip backup as well. 
	> 
	> But that's like playing with 🔥 in a production site. If something goes wrong, you need to fix it on production yourself. Resulted in more downtime.

    <p align="center">
        <img style="max-width: 300px" src="https://frappe.io/files/meme435384.png" />
    </p>


### Simple Solution - Disk Snapshot

Taking a disk snapshot of the database server, such as an AWS EBS or ZFS Snapshot, is a quick and efficient process. It typically takes just a few seconds and is considered a physical backup of the database. On the surface, this seems like a straightforward and effective solution for creating backups.

However, there’s an issue with this approach. A running database server always has data in memory. If you take a snapshot without stopping the database, it can result in corrupted files. To avoid this, the database needs to be stopped before taking a snapshot.

This is especially challenging on Frappe Cloud, where each database server hosts hundreds of databases. Shutting down the entire server just to back up one database is not feasible.

![](https://frappe.io/files/fc-db-infra.png)



### Solution To Physical Backup

To create a safe physical backup, we need to ensure two things :

1. There should be no activity in that specific database while taking snapshot.
2. The database should commit all data from memory to disk.

#### Solution to Problem 1

To prevent activity, take a **Read-Write lock** on the database. This ensures that no changes occur to the database tables while the snapshot is being taken.

#### Solution to Problem 2

This solution depends on the storage engine being used. MariaDB has multiple storage engines, including InnoDB and MyISAM, which handle storage in different ways.

- **InnoDB:** Each table is a separate tablespace, and every table has an associated \*.ibd file that holds the actual data. InnoDB also supports Transportable Tablespace, allowing you to export and import tablespaces even in a live database.
- **MyISAM:** Each table consists of two files: \*.MYI (index) and \*.MYD (data).

MariaDB provides a single SQL query that can solve both problems:
```
FLUSH TABLES `tabUser`, `__global_search`, `tabVersion` FOR EXPORT
```

> Make sure to keep the terminal open until the backup is complete.

This command locks the database and flushes the data to disk, making it safe to copy. Once done, you can backup the required files from the database folder. For example:
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

Note that the \*.cfg* file is generated for InnoDB tables and must be backed up. It’s required for importing the tablespace into another database server.

However, you’re still missing one crucial thing. Inside the `/var/lib/mysql/` folder, there are \.frm* files that hold the structure and column definitions of the tables, but they are not importable in a live database. To backup the table structure, use `mysqldump`:

```
mysqldump _cdsd32dsvn92 --no-data > schema.sql
```

Once you've backed up the data and structure, you can either take a disk snapshot or copy the files to another directory, disk, or remote server.

**Note:** In Frappe Cloud, we currently use AWS EBS Snapshots for physical backups because they are time-consistent and take only 3-5 seconds to create.

🎉 With that, we now have a complete physical backup, including both data and table definitions.

### Time For Physical Restoration

Let’s start with a simplest approach to restoration, and later, we’ll address each problem and its solution in more detail.

#### Steps

1. **Download Backup :** Ensure you have the physical backup ready before starting the restoration.
	- For **EBS/Disk Snapshots**, create a volume from the snapshot and mount it on your server.
	- For other backup types, download the backup files to your server and mount them as needed.

2. **Drop Tables:** Drop the tables you want to import by running the following command:

	```
	DROP TABLE <table-name>
	``` 
	⚠️ Caution: This permanently deletes the table and its data.

3. **Recreate Tables from schema:** If you followed the backup steps, you would have dumped the database schema using `mysqldump`. To restore the table structure, we need to find the `CREATE TABLE` queries for the required tables in the `schema.sql` file and run them.

	If you need help extracting the `CREATE TABLE` query using regex, refer to [code](https://github.com/frappe/agent/blob/69084c59fda37e5ee08854e0a893d01116c2b303/agent/database_physical_restore.py#L387-L403).

4. **Restore Data :** The restoration process depends on the storage engine. Let’s look at how we can restore data for each engine:
	1. **InnoDB Restoration (Using Transportable Tablespaces)** 
		- Discard the tablespace :
			```
			ALTER TABLE tabUser DISCARD TABLESPACE;
			```
			After running this, you’ll see that the `tabUser.ibd` file is removed from `/var/lib/mysql/<database-name>`.
		- Copy the backup files :
		  Paste `tabUser.ibd` (data) + `tabUser.cfg` (metadata) from backup to the same folder.
	    - **Import tablespace :**
			```sql
			ALTER TABLE tabUser IMPORT TABLESPACE;	
			```
		 If this step is successful, the data is restored to the live database.

	2. **MyISAM Restoration :**
		- Acquire a write lock on the table.
		- Copy the backup files (`*.MYI` for index and `*.MYD` for data) to the `/var/lib/mysql/<database-name>` folder.
5. Physical  Restoration Completed 🤞

---

### Problems and Solutions

While physical backup and restoration offer performance benefits and seem straightforward in theory, we faced several real-world challenges when implementing it on Frappe Cloud for medium-sized sites. During our initial tests, unexpected failures began to occur.

To thoroughly test the method, we developed a Bulk Backup and Restoration Tool in Frappe Cloud. This allowed us to identify and resolve potential bugs in the process.

Every day, we took 40-50 backups and restore the data multiple times on a dummy site to ensure everything works smoothly.

![](https://frappe.io/files/bulk-restore-tool.png)



#### Slow File Copy Due to Lazy Loading

When restoring physical backups, copying files from the backup disk to the main disk is a critical process. During testing, we noticed two things that were holding us back:

- 10GB of files took around 16 minutes.
- The max transfer speed was capped at about 15MBPS.

**Root Cause**

The issue stemmed from AWS’s provisioning of volumes from snapshots. While the volume is created instantly, it doesn’t provide the expected performance due to lazy loading. Initially, the EBS volume is in an uninitialized state, so it loads disk blocks only when requested.

![](https://frappe.io/files/ebs-lazy-loading.png)

This meant we couldn't cross 15MBPS read speed, which was a big bottleneck.

**Failed Experiments**

We tried a few things to fix the issue but didn’t get the results we wanted:

1. **AWS’s Recommended Pre-Warming Process**: AWS suggests using `dd` or `fio` to pre-warm your volume. However, we couldn’t push the speed beyond 15-20 MBPS due to the sequential read nature of `dd`.
2. **Parallel `dd` Processes**: Next, we thought about running 5 `dd` processes in parallel to get to around 80MBPS. But this approach led to:
        - Severe CPU overhead
        - High iowait congestion

**The Solution: io_uring**

After some digging, we came across a solution - io_uring, a Linux feature that supports asynchronous I/O operations. To leverage this, we built a custom tool in golang with a python wrapper for easy integration.

With io_uring, we managed to achieve nearly 300MBPS disk throughput at pre-warm stage (almost 90% of what the EBS gp3 disk was capable of) without the heavy CPU load.

**Results**

The warm-up time for 10GB of files dropped dramatically from 16 minutes to just 1 minute.

If you want, you can install the library from PyPI and tryout: [filewarmer](https://pypi.org/project/filewarmer/). Also, if you want to dive deeper into io_uring, check out this  [blog post by Mattermost](https://mattermost.com/blog/iouring-and-go/).
#### The Snapshot Availability Bottleneck

While AWS EBS snapshots are created instantly, they’re unusable until the status changes from **`Pending`** → **`Available`**.

**What’s Happening Behind the Scenes?**

- AWS takes a snapshot of your disk and stores the data in Cold Storage/S3.
- Duration depends on amount of changes in files on the disk since the last snapshot

![](https://frappe.io/files/snapshot.png)

At Frappe Cloud, we already keep 24-hour and 48-hour disk snapshots, which speeds up the process a bit - but not enough.

To further reduce the time, we implemented a **Rolling Snapshot** strategy.

Every 2 hours, Frappe Cloud takes a snapshot of all database server disks. Once the new snapshot becomes available, the older one is deleted. This approach has helped us reduce snapshot availability time by about **50%**.

#### Missing \*.cfg files in Physical Backup

We started encountering issues where \*.cfg* files were missing, preventing InnoDB table imports. This was puzzling, especially since AWS EBS Snapshots are supposed to be time-consistent.

After some debugging, we identified the problem :

- MariaDB uses `fdatasync` to flush data to disk, but it doesn’t flush the metadata.
- While the \*.cfg files were written to disk, their metadata remained in buffers.
- As a result, snapshots captured 0-byte \*.cfg files.

To resolve this, we added an extra step: performing an `fsync` on the \*.cfg* files before taking the snapshot.


```python
# Force metadata flush using fsync
# Ref: https://docs.python.org/3/library/os.html#os.fsync
with open(file_path, "rb", buffering=0) as f:
	os.fsync(f.fileno())
```

dditionally, during backups, store exact file sizes and checksums for all \*.cfg files so that during restoration it can verify file sizes and checksums. If the process finds any mismatch, it will abort the restoration.
#### Handle Broken MyISAM Tables

MyISAM tables are prone to corruption. They can get corrupted easily, but luckily, they are just as easy to fix! 😅

**Common reasons of corruption -**

1. Improperly closed database connections
2. Unexpected database server crashes
3. Hardware failure

In most cases, the index file (\.MYI*) becomes corrupted, but the data file (\*.MYD) stays safe.

MariaDB provides a utility called [`myisamchk`](https://mariadb.com/kb/en/myisamchk/) to check and fix corruption, and you don’t even need the database running to use it. Simply run:

```
myisamchk -r /var/lib/mysql/_c8383gdu339jd8/__global_search
```

This tool can solve 99% of the issues. So, before attempting restoration, we always check for MyISAM table corruption and fix it. If the issue persists, we won’t attempt the restoration.

After restoration, we sometimes encounter instances where MyISAM tables are still marked as corrupted.

![](https://frappe.io/files/myisam-corruption.png)

Since the MyISAM restoration procedure only involves copying files (without updating index positions or metadata), we also check for corruption afterward using `CHECK TABLE`. If we find any issue in that use `REPAIR TABLE <table-name> USE_FRM` to fix the issues,

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
