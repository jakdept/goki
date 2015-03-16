Jack's Dumping Ground
=====================

Storm Create Pause
------------------

Used when you have a Storm Create that keeps failing on "failed to SSH after 120 attempts"

```bash
pid=''; while [ -z $pid ]; do echo 'not yet boss, looking again'; pid=$(ps faux|grep XenWaitForBoot|grep -v grep|awk '{print $2}'); sleep 5; done; kill -19 $pid; echo 'I CAUGHT THAT CREATE 4 U'
```

Then once ur done, grab the pid for `XenWaitForBoot` and send it `kill -18`.

Nuke a directory and contents
-----------------------------

Warning: This is very destructive.

Go into the directory and run this:

```bash
find . -type d |sort -r|xargs -L1 -I {} echo perl -e "unlink(glob({}));"
```

Delete old Exim emails (1 day)
------------------------------

```bash
exiqgrep -o 86400|xargs -L1 exim -Mrm
```

Alternate stat alias
--------------------

```bash
alias stat='stat -c"a:%x m:%y c:%z %n"'
```

CWD lines from `exim_mainlog`
-----------------------------

```bash
awk '$4 ~ /^cwd=/ {gsub('/cwd=/', "", $4); print $4}' /var/log/exim_mainlog|sort|uniq -c|sort
```

Find Symlink Hax
----------------

```bash
awk '/DocumentRoot/ {print $2}' /usr/local/apache/conf/httpd.conf|xargs -I{} find {} -type l|awk -F'/' 'BEGIN {OFS="/"} ; {$NF=""; print $0}'|sort|uniq|xargs -L1 stat -c"a:%x m:%y c:%z %n"|sort -k3
```

Wordpress Shite
---------------

```bash
UPDATE wp_options SET option_value = replace(option_value, 'http://cookieandkate.com', 'http://lwdev.cookieandkate.com') WHERE option_name IN ('siteurl', 'home');
UPDATE wp_posts SET guid = replace(guid, 'http://cookieandkate.com','http://lwdev.cookieandkate.com') WHERE guid LIKE "%cookieandkate.com%";
UPDATE wp_posts SET post_content = replace(post_content, 'http://cookieandkate.com','http://lwdev.cookieandkate.com') WHERE guid LIKE "%cookieandkate.com%";
 
SELECT option_value from wp_options WHERE option_name = 'active_plugins';
UPDATE wp_options SET option_value = 'a:0:{}' WHERE option_name = 'active_plugins';
 
SELECT * from wp_options WHERE option_name IN ('stylesheet','template');
 
 
UPDATE wp_users as u INNER JOIN `wp_usermeta` um ON (u.`ID` = um.`user_id`) SET u.`user_pass`=MD5('pa55w0rd') where um.`meta_value` like "%administrator%";
```

Incoming Apache Connection States
---------------------------------

```bash
netstat -nut|awk '$4 ~ /:(80|443)/ {gsub(/:[0-9]*$/, "", $5); print $5, $6}'|sort|uniq -c|sort -n
```

Repeated Files
--------------

```bash
find . -name "*.php"|xargs md5sum|sort|awk '{print $2,$1}'|uniq -f1 -c|awk '{print $1, $2}'|sort -n
```

Multiple opening PHP Tags in first 10 lines
-------------------------------------------

```bash
find . -type f|xargs awk '/<\?php/ {print FILENAME; if(NR > 10) {nextfile}}'|sort|uniq -c|sort -n
```

Maldet reports
--------------

md5sum everything

```bash
cat /usr/local/maldetect/sess/session.$(cat /usr/local/maldetect/sess/session.last)|tail -n +11|head -n -2|awk '{print $3}'|xargs md5sum
````

stat everything

```bash
cat /usr/local/maldetect/sess/session.$(cat /usr/local/maldetect/sess/session.last)|tail -n +11|head -n -2|awk '{print $3}'|xargs stat -c"p:%a a:%x m:%y c:%z %n"|sort -k3
```

cPanel specific
---------------

cpipv6 (cPanel's IPv6 implementation) stores it's IPs in `/etc/userdatadomains`

Random Server Shite
-------------------

History tweak:

```bash
HISTCONTROL=ignoreboth
```

servers with bad line width:

```bash
shopt -s checkwinsize
```

kill console spam

```bash
echo '3 4 1 7' > /proc/sys/kernel/printk
```

