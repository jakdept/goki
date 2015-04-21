OpenSSH
=======

OpenSSH is the webserver that we use on most all of our servers. It's pretty secure, and for the most part, is trouble free.

This wiki's going to approach this large subject in three parts:

* SSH Keys
* The OpenSSH Server
* The OpenSSH Client

Warning: You should never update OpenSSH. It doesn't need to be updated.

There is a [second page](openssh_additional.md) for more information that we do not usually see on customer servers.

SSH Keys Overview
-----------------

The basic idea with SSH keys is one private key can generate multiple public keys. However, getting the private key from the public key is a difficult problem - and can take weeks or years of work (depending on the size of your server cluster).

OpenSSH allows you to log in with these keys. Put the public key in `~/.ssh/authorized_keys` for the user that you're going to log in as. Thus, for root, this would be the file `/root/.ssh/authorized_keys` in most cases. Then, you provide the private key when logging in.

Generating an SSH key
---------------------

notice: CentOS 5 does not support ECDSA keys. Until it goes EOL, we are going to be using RSA keys by default.

To generate a key, run the following command wherever the private key will be installed:

```bash
ssh-keygen -t rsa -b 4096 -C "$(whoami)@$(hostname)-$(date -u +%Y-%m-%d-%H:%M:%S%z)"
```

info: You will need to pick where to save the key. Name the file the same as the customer's hostname.

Warning: There is no benefit to entering a passphrase here - the passphrase will go right into the sticky note right next to the key, negating any benefit of having the key password protected.

This should create two files - an example of each is below:

* The public key - usually ends in .pub

```
ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAIEApO5wdFfTNpqWMz/xBzYq1FBLy398v9BcyjoePSNPXITSRnSF55CdmOG7s19auld/pe1a/fXmQIsNj7qwyN2Plw5hQS2WRiQ7hwJq4ERaneheYEp7vkIUMFjjSl+DvOJqQVdsvSMgDfpTkZFod99g1XwsfMgPEuddQfwM++D50Bk= jack@jack.wks.liquidweb.com-2015-01-20-03:25:23+0000
```

* The private key - usually ends in .key

```
-----BEGIN RSA PRIVATE KEY-----
MIICWgIBAAKBgQCk7nB0V9M2mpYzP/EHNirUUEvLf3y/0FzKOh49I09chNJGdIXn
kJ2Y4buzX1q6V3+l7Vr99eZAiw2PurDI3Y+XDmFBLZZGJDuHAmrgRFqd6F5gSnu+
QhQwWONKX4O84mpBV2y9IyAN+lORkWh332DVfCx8yA8S511B/Az74PnQGQIBIwKB
gD1CnswgnuhlTbtDoqroO8szxGGHH7T1ns7FISymtxO8ThorR63IAAWVrB4NeXhp
pHDUgOH8P5RQ58ervgF23Y9IDoOoUeRBG+63aX3j7MeiK//SQW9WPu97imiaibXg
T02atOENDoCqJ0MaVJ0i+GpRH842QDbHWhKqc2OG2eCLAkEA08wBTsy35sQlwmPJ
35+Bl9esjvwXi/6QUx/ek5o/9d9waqnQG2vv7PYQAPVQGZ6H5g6bP4xPcZtuhlXC
CBpU+QJBAMdaevI72PCiVLSv/oju23SGZXG7yboxG0MAA2bs4gbI//15e25lr/R9
pYSjCo7+G6gDU+LCL/qUPr6AkFoevCECQBg0km9n2oDF9a/Q498K6j09OEra+2B0
3UtUGW/0XxTJFClyfi8FBXoqwAAcCSd/1QRZcNQQCRRMRyVLoSV/WisCQQCDAPG1
IAOW0RMXptp3PeCrqMZSDbBzClO+UHdDovruhBXvtjsrSiMrowZeecUcI1QAsbrI
NndM5RNJ/LahnysrAkAIlNST2840KUI5HLiA7grHNzUAbacVVvwtOh3gcz/l3UqG
70+/x1cE6GYNv8Re/K8PVfsR2HEjsbRAcprQ8Pb6
-----END RSA PRIVATE KEY-----
```

#### Installing a Public Key on a Server ####

* Open the file `~/.ssh/authorized_keys` for the user that you're going to SSH in as.
* Put the key on one line.
* Make sure the file is owned by the user that is going to be used with.
* Make sure permissions are at least read + write for the user in question.

#### Creating the Private Key note ####

Add some stuff to the top and bottom of the key:

```
Copy the following commands on your workstation terminal
 
cat <<EOM >~/.ssh/host.server.com.key
```

The key goes here

```
EOM
chmod 600 ~/.ssh/host.server.com.key
ssh root@host.server.com -i ~/.ssh/host.server.key
```

The end result should be something that looks like:

```
Copy the following commands on your workstation terminal
 
cat <<EOM >~/.ssh/host.server.com.key
-----BEGIN RSA PRIVATE KEY-----
MIICWgIBAAKBgQCk7nB0V9M2mpYzP/EHNirUUEvLf3y/0FzKOh49I09chNJGdIXn
kJ2Y4buzX1q6V3+l7Vr99eZAiw2PurDI3Y+XDmFBLZZGJDuHAmrgRFqd6F5gSnu+
QhQwWONKX4O84mpBV2y9IyAN+lORkWh332DVfCx8yA8S511B/Az74PnQGQIBIwKB
gD1CnswgnuhlTbtDoqroO8szxGGHH7T1ns7FISymtxO8ThorR63IAAWVrB4NeXhp
pHDUgOH8P5RQ58ervgF23Y9IDoOoUeRBG+63aX3j7MeiK//SQW9WPu97imiaibXg
T02atOENDoCqJ0MaVJ0i+GpRH842QDbHWhKqc2OG2eCLAkEA08wBTsy35sQlwmPJ
35+Bl9esjvwXi/6QUx/ek5o/9d9waqnQG2vv7PYQAPVQGZ6H5g6bP4xPcZtuhlXC
CBpU+QJBAMdaevI72PCiVLSv/oju23SGZXG7yboxG0MAA2bs4gbI//15e25lr/R9
pYSjCo7+G6gDU+LCL/qUPr6AkFoevCECQBg0km9n2oDF9a/Q498K6j09OEra+2B0
3UtUGW/0XxTJFClyfi8FBXoqwAAcCSd/1QRZcNQQCRRMRyVLoSV/WisCQQCDAPG1
IAOW0RMXptp3PeCrqMZSDbBzClO+UHdDovruhBXvtjsrSiMrowZeecUcI1QAsbrI
NndM5RNJ/LahnysrAkAIlNST2840KUI5HLiA7grHNzUAbacVVvwtOh3gcz/l3UqG
70+/x1cE6GYNv8Re/K8PVfsR2HEjsbRAcprQ8Pb6
-----END RSA PRIVATE KEY-----
EOM
chmod 600 ~/.ssh/host.server.com.key
ssh root@host.server.com -i ~/.ssh/host.server.key
```

This should be in a billing sticky note. Next, in their auth settings for that subaccount, under special instructions, add the SSH command in - so it reads something like:

```
ssh -i ~/.ssh/host.server.key root@host.server.com
```

Changing Passphrases on Private Keys
------------------------------------

Run the following command and you should be prompted for:

* The private key who's passphrase you want to change
* The old passphrase
* The new passphrase
* The new passphrase again (to confirm)

```
ssh-keygen -p
```

OpenSSH Server
--------------

The OpenSSH server configuration file is typically at `/etc/ssh/sshd_config`. Some often used options:

* `Port` - the port that OpenSSH listens on
 * If you change this option, you should update the firewall, and disable SSH monnitoring for the server
* `Protocol` - which SSH protocols SSH should allow - this should be set to 2
* `PasswordAuthentication` - allows you to use a Password to log into the server
 * If you disable this option, you have to already have an SSH key set up
* `PermitRootLogin` - allow root to log in via SSH or no
 * If you disable this option, you have to already have an alternate SSH user set up.
* `PubkeyAuthentication` - allows SSH keys to login
* `AuthorizedKeysFile` - allows you to specifiy an alternate location
 * Default is `%h/.ssh/authorized_keys` - `%h` is replaced with your user's homedir.

There's not a lot of options there - that's because for the most part OpenSSH just works.

