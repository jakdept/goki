setting up hooks
================

* Add the two required users:

```
useradd -s /sbin/nologin -M www-data 
useradd git
```

* Make sure you have `git` installed.

* Install Gitolite.

```
su - git
mkdir -p ~/bin
git clone git://github.com:sitaramac/gitolite
gitolite/install -ln ~/bin
gitolite setup -pk yourkey.pub
```

* Add `~/bin` to `git`'s path in `/etc/bashrc or similar

* Edit `~/.gitolite.rc`

 * Under the section `%RC = ` uncomment the line:

```
 LOCAL_CODE                =>  "$rc{GL_ADMIN_BASE}/local",
```

 * Or add a similar line. Further down in the same file, uncomment/add the line:

```
'repo-specific-hooks',
```

* Clone the gitolite-admin repository, and within it, first add the folders:

```
mkdir -p local/commmands local/triggers local/VREF local/hooks/repo-specific
```

* In your clone, drop your hooks into:

```
local/hooks/repo-specific/
```

 * The name should be named for that reppository, or what it does, not post-recieve. An example deploy script for the content directory is below:

```
#!/bin/sh
git --work-tree=/var/wiki --git-dir=/home/git/repositories/wiki-content.git checkout -f
find /var/wiki -mindepth 1 ! -group www-data -exec chgrp www-data {} \;
find /var/wiki -mindepth 1 ! -perm 640 -exec chmod 640 {} \;
````

 * Note, the following version does not currently work, but will be supported in the future

```
#!/bin/sh
git --work-tree=/var/backend --git-dir=/home/git/repositories/wiki-backend.git checkout -f
find /var/wiki -mindepth 1 ! -group www-data -exec chgrp www-data {} \;
systemctl reload gnosis.service
```

* Make sure your destination directories are created:

```
/var/wiki
/var/backend
```

* Push the repositories up to the server so you have stuff to associate.

* Enable your systemctl script service

```
systemctl enable /var/wiki-backend/gnosis.service
```

 * The current example script for this is:

 ```
 [Unit]
Description=gnosis wiki server
After=firewalld.service
Requires=firewalld.service
After=rsyslog.service
Requires=rsyslog.service
After=network.target
Requires=network.target

[Service]
User=www-data
Group=www-data
WorkingDirectory=/var/backend
ExecStart=/var/wiki-backend/webserver --config=/var/wiki-backend/config.json
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Reference for the `systemd` `gnosis.service`: http://www.freedesktop.org/software/systemd/man/systemd.unit.html
