`loadwatch`
===========

Loadwatch. Oh loadwatch. How we love thee.

Loadwatch is nothing more than a list of commands that automatically run on a server when the server is experiencing high load. The output from these commands are then stored in a file for analysis later - if needed.

Loadwatch has two important requirements:

* It cannot cause load
* It has to run quickly

Installation
------------

```bash
mkdir -p /root/loadwatch /root/bin
cat <<EOM >/root/bin/loadwatch.sh
#!/bin/bash
FILE=loadwatch.`date +%F.%H.%M`.txt
DIR=/root/loadwatch
#Load Threshold for doing a dump.
THRESH=12
SQLTHRESH=25
APACHETHRESH=120

LOAD=`cat /proc/loadavg | awk '{print $1}' | awk -F '.' '{print $1}'`

CURSQL=`/usr/bin/mysqladmin stat|awk '{print $4}'`

HISTSQL=`/usr/bin/mysql -Bse 'show global status LIKE "Max_used_connections";'|awk '{print $2}'`

APACHECONN=$(/usr/sbin/httpd status|awk '/requests\ currently\ being\ processed,/ {print $1}')


echo `date +%F.%X` - Load: $LOAD MySQL conn: $CURSQL Highest MySQL conn: $HISTSQL Current httpd conn: $APACHECONN >> $DIR/checklog

if [ $LOAD -gt $THRESH ] || [ $CURSQL -gt $SQLTHRESH ] || [ $APACHECONN -gt $APACHETHRESH ]
then
	echo Loadwatch tripped, dumping info to $DIR/$FILE >> $DIR/checklog
	echo `date +%F.%H.%M` > $DIR/$FILE
	free -m >> $DIR/$FILE
	uptime >> $DIR/$FILE
	mysqladmin processlist stat >> $DIR/$FILE
	/bin/netstat -nut|awk '$4 ~ /:(80|443)/ {gsub(/:[0-9]*$/, "", $5); print $5, $6}'|sort|uniq -c|sort -n|tail -n50 >> $DIR/$FILE
	top -bcn1 >> $DIR/$FILE
	ps auxf >> $DIR/$FILE
	/sbin/service httpd fullstatus >> $DIR/$FILE
fi
EOM
chmod 700 /root/bin/loadwatch.sh
echo "*/3 * * * * /root/bin/loadwatch.sh > /dev/null 2>&1" >> /var/spool/cron/root
echo "0 1 * * * /usr/bin/find /root/loadwatch -type f -mtime +30 -delete" >> /var/spool/cron/root
echo "0 2 * * 0 /usr/bin/tail -n100000 /root/loadwatch/checklog > /root/loadwatch/checklog.new; mv -f /root/loadwatch/checklog.new /root/loadwatch/checklog" >> /var/spool/cron/root
```

Once you've run that block, you will likely want to go back into `/root/bin/loadwatch.sh` and adjust the threshold settings. You will also want to check over `/var/spool/cron/root` and remove any duplicate cron tasks.

Reading Output
--------------

Loadwatch spits out command output. They're the output from the commands. Go read the files and learn what's going on.

If you really cannot log into the server, you can mount `/root/loadwatch/` locally on your computer via sshfs:

```bash
sshfs root@127.0.0.1:/root/loadwatch/ ~/local_mountpoint/
```

Emailing Output
---------------

Some customers want to have the output from `loadwatch` emailed to them. This is a bad idea. This creates an email and attempts to deliver it to the customer during a period of high load - a bad idea.

Still, you said they wanted it. W/e. It's a bad idea. I said it's a bad idea. Let's move on. Add the following line at the bottom of the file `/root/bin/loadwatch.sh`:

```bash
cat $DIR/$FILE | mail -s "Loadwatch Report" "the_customers@email.address" >> $DIR/$FILE 
```