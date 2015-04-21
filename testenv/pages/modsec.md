topic = apache
topic = security
keyword = modsec
keyword = userdata
ModSecurity
===========

Overview
--------

ModSecurity is a Web Application Firewall - like a networking firewall, but instead it operates within the Web Server. There are two different version of ModSecurity that we usually see - `security_module` for Apache version 1.x, and `security2_module` for Apache version 2.x.

To compile ModSecurity on cPanel - either Apache 1.x or 2.x - you just need to run an EasyApache and select `UniqueId` and `mod_security`.

Warning: ModSecurity 1 rules are not compatible with ModSeecurity 2 rules. If you mix them, you will have problems.

Information: If you do about anything on this page, restart Apache.

### Todo list for wiki ###

* figure out how to handle blocks taht don't have a rule id?
* LB rewrites? Explain that?
* Block all traffic not coming from LB/CloudFlare?
* Section on CMC
* Section on cPanel's new ModSecurity plugin

Apache 1.x Specifics
---------------------

If you're running Apache 1.x, the module is loaded with:

```
LoadModule security_module modules/mod_security.so
Include "/usr/local/apache/conf/modsec.conf"
```

On a cPanel server, there's an include to `/usr/local/apache/conf/modsec.conf`, and the following config are relevant:

* `/usr/local/apache/conf/modsec.conf` - The main File, mostly has some cPanel stuff in it and includes `modsec.user.conf`
* `/usr/local/apache/conf/modsec.user.conf` - This file is meant for the end user to put their custom stuff - we include:
 * `/usr/local/apache/conf/modsec/exclude.conf` - This is where we exclude directories across the server
 * `/usr/local/apache/conf/modsec/whitelist.conf` - This is where we whitelist rules across the server
 * `/usr/local/apache/conf/modsec/rootkits.conf` - This is used for rootkit specific rules
 * `/usr/local/apache/conf/modsec/custom.conf` - Any custom stuff for that server goes in here


Then, to install our rules, on CentOS 5 run:

```
lpyum install lp-modsec-rules.noarch
```

On CentOS 6 run:

```
yum install lp-modsec-rules.noarch
```

Apache 2.x Specifics
---------------------

If you're running Apache 2.x, the module is loaded with:

```
LoadModule security2_module modules/mod_security2.so
Include "/usr/local/apache/conf/modsec2.conf"
```

On a cPanel server, there's an include to `/usr/local/apache/conf/modsec2.conf`, and the following config are relevant:

* `/usr/local/apache/conf/modsec2.conf` - The main File, mostly has some cPanel stuff in it and includes `modsec.user.conf`
* `/usr/local/apache/conf/modsec2.user.conf` - This file is meant for the end user to put their custom stuff - we include:
 * `/usr/local/apache/conf/modsec2/exclude.conf` - This is where we exclude directories across the server - updates from repo
 * `/usr/local/apache/conf/modsec2/whitelist.conf` - This is where we whitelist rules across the server
 * `/usr/local/apache/conf/modsec2/rootkits.conf` - This is used for rootkit specific rules
 * `/usr/local/apache/conf/modsec2/custom.conf` - Any custom stuff for that server goes in here

Then, to install our rules, on CentOS 5 run:

```
lpyum install lp-modsec2-rules.noarch
```

On CentOS 6 run:

```
yum install lp-modsec2-rules.noarch
```

Log Files
---------

EasyApache leaves stuff in the following logs:

* `/usr/local/apache/logs/error_log` - the main Apache error log

```
[Wed Dec 26 13:38:42 2012] [error] [client 188.143.232.176] ModSecurity: Access denied with code 406 (phase 2). Pattern match "select.*from.*information_schema\\\\.tables" at REQUEST_URI. [file "/usr/local/apache/conf/modsec2/rootkits.conf"] [line "155"] [id "3000086"] [hostname "fjmc.net"] [uri "/index.php"] [unique_id "UNtEMkPjo@cAAALOTc0AAAAA"]
```

* `/usr/local/apache/logs/audit_log` - the older ModSecurity specific log
* `/usr/local/apache/logs/modsec_audit.log` - the more current ModSeccurity specific log
* `/usr/local/apache/logs/modsec_debug_log` - This is a debugging log for ModSecurity. You don't care about it.

cPanel Whitelisting
------------

Within `/usr/local/apache/conf/userdata/` there is a structure to include stuff into a VirtualHost. When we are whitelisting, we want to disable as little as possible to enable everything to work. The full path/way this is used is:

* `/usr/local/apache/conf/userdata/` - this is the base that is always there to start iwth
* `/usr/local/apache/conf/userdata/SSL or no?/` - this folder choice allows you to work with non-SSL - `std`, or SSL - use `ssl`
* `/usr/local/apache/conf/userdata/std/Apache Version/` - this is the Apache version - 1 or 2
* `/usr/local/apache/conf/userdata/std/2/cPanel Username/` - this is their cPanel username. Drop a file in this folder, and it will apply to all of that cPanel user's VirtualHosts
* `/usr/local/apache/conf/userdata/std/2/username/Domain Name/` - this allows you to put a rule into place for any one domain
* `*.conf` - ALL of your files that you want included have to have a filename that ends in `.conf`. I suggest using `modsec.conf`

Once again, you want to whitelist for as little of an area as possible. Pick your file to whitelist in, and add the following:

```
<IfModule mod_security2.c>
 <LocationMatch "\/regex\/webserver\/path\/to\/script\.php">
  SecRuleRemoveById RULEID
 </LocationMatch>
</IfModule>
```

The path in that include is the URI of the request.

Plesk Whitelisting
------------------

On Plesk, this is even easier.

* `/var/www/vhosts/<domain name>/statistics/logs/error_log` This is the error log that should show you the rule hit
* `/var/www/vhosts/system/<domain name>/conf/siteapp.d/` This is the folder where you put your include
* ` /var/www/vhosts/system/<domain name>/conf/siteapp.d/vhost.conf` This is where you put in your whitelisting
 * **The file has to be called `vhost.conf`**
* Once your changes have been made, run ` /usr/local/psa/admin/sbin/httpdmng --reconfigure-domain <domain_name.tld>`
* IF you would rather hit all includes at once, `/usr/local/psa/admin/sbin/httpdmng --reconfigure-all`
* Restart Apache

Whitelist a Specific IP address
-------------------------------

You can write a custom rule to disable the ModSecurity engine for a given IP address:

```
 SecRule REMOTE_ADDR "^111\.222\.333\.444" "phase:1,nolog,allow,ctl:ruleEngine=off,id:SOMERANDOMNUMBER"
```

Disable ModSecurity
-------------------

If you want to, you can - for a given location - disable ModSecurity altogether. Enclose it like before, and use `SecRuleEngine Off` instead.

PCRE Limits
-----------

> need the actual error here still

PCRE Limits limit the amount of recurison within Apache for Regex Rules. If you're getting these errors, raise these limits:

In `/usr/local/lib/php.ini`:

```
pcre.backtrack_limit = 10000000
pcre.recursion_limit = 10000000
```

In `/usr/local/apache/conf/modsec2/custom.conf`:

```
SecPcreMatchLimit 150000
SecPcreMatchLimitRecursion 150000
```

Request Body
------------

```
[Thu May 09 09:27:23 2013] [error] [client x.x.x.x] ModSecurity: Request body no files data length is larger than the configured limit (1048576).. Deny with code (413) [hostname "www.hostname.com"] [uri "/admin/index.php"] [unique_id "UYtPvEg0uj0AABTCHsoAAAAG"]
```

ModSecurity has a configured maximum Request Body size. You can up it by setting the following in ModSec:

```
SecRequestBodyLimit 2097152
SecRequestBodyNoFilesLimit 2097152
```

Alternatively, you can disable this altogether with:

```
SecRule REQUEST_FILENAME "^/upload/ft2.php$" "phase:1,t:none,allow,nolog,ctl:requestBodyAccess=Off"
```

Response Body
-------------

```
 [Tue Oct 16 09:16:39 2012] [error] [client xxx.xxx.xxx.xxx] ModSecurity: Output filter: Response body too large (over limit of 524288, total length not known). [hostname "www.domain.com"] [uri "/eesys/index.php"] [unique_id "UH2IZ38AAAEAABwlDzsAAAAD"]
```

Similar to the Request Body too large, you can have similar issues with the Response Body. Adjust it with:

```
SecResponseBodyLimit 1572864
```

Upload tmpdir issues
--------------------

If you upload a file on a server with ModSecurity, ModSecurity will actually put that file somewhere, scan it, then give it to the script. Sometimes, as a result, we'll get the error:

```
[Fri Nov 18 14:49:50 2011] [error] [client 72.52.142.215] mod_security: Access denied with code 406. read_post_payload: Failed to create file "/root/tmp/20111118-144950-72.52.142.215-request_body-xGPNPd" because 13("Permission denied") [severity "EMERGENCY"] [hostname "lakedonpedro.org"] [uri "/wp-cron.php?doing_wp_cron"] [unique_id "TsbhJkg0jtcAACYIFDk"]
```

If you see this, this can be fixed by adding to the `php.ini`

```
upload_tmp_dir = /tmp
session.save_path = /tmp
```

And in `modsec2/custom.conf`

```
SecUploadDir /tmp
SecTmpDir /tmp
```