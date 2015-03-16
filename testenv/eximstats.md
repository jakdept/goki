Eximstats
=========

Overview
--------

Eximstats parses exim_mainlog and maillog to insert their content into tables in the `eximstats` databases. This table is then queried to provide the Mail Usage Statistics in WHM.

Since we don't use the Mail Usage Statistics, it doesn't much help us.

Disable Eximstats
-----------------

We don't use Eximstats. It's a waste of time. So why not disable it?

```bash
/usr/local/cpanel/bin/tailwatchd --disable=Cpanel::TailWatch::Eximstats
mysql -Bse 'truncate table defers; truncate table failures; truncate table sends; truncate table smtp;' eximstats
rm /var/cpanel/sql/eximstats.sql
```

exim_tidydb
-------------

Normally, on a cPanel server, the script `exim_tidydb` runs nightly. Sometimes, it fails to run. Sometimes it gets behind. If you want to, you can clean this up by running `/scripts/exim_tidydb`. If the command fails, it's probably because files are missing. If they are, you can create them with:

```bash
for file in $(/scripts/exim_tidydb |grep "No such file or directory"|awk '{print $8}'|tr -d ':'); do touch $file && chown mailnull.mail $file && chmod 640 $file; done
```

Reenable Eximstats
------------------

If you disabled Eximstats before, and you now need it reenabled, it can be reenabled with:

```
/usr/local/cpanel/bin/tailwatchd --enable=Cpanel::TailWatch::Eximstats
```

The password for the MySQL user for Eximstats is contained in the file `/var/cpanel/eximstatspass` - in case you need to recreate the user or update the password.

Recreate Eximstats DB
---------------------

Sometimes databases get lost. It happens. If you ever need to recreate it, just create an empty database and then import the file `/usr/local/cpanel/etc/eximstats_db.sql`.