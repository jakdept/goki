category = internal
keyword = whitelist
Internal IP Ranges
==================

CSF/APF
-------

Please add the following to `/etc/csf/csf.allow` or `/etc/apf/allow_hosts.rules`

```
## BEGIN LIQUIDWEB ##
# DC1 office range:
## IPv4
10.10.4.0/23
#DC2 office range
## IPv4
10.20.4.0/22
#Dc3 office ranges:
## IPv4
10.30.4.0/22
10.30.2.0/24
10.30.104.0/24
# Liquidweb VPN
10.20.7.0/24
# Workbench
69.167.129.192/28
## IPv6:
##2607:fad0:32:a01::/64
##2607:fad0:32:a03::/64
##2607:fad0:32:a02::/64
#Backup Server Range
10.2.0.1/24
#DNS resolvers
209.59.157.254
69.167.128.254
10.10.10.10
# syspackages.sourcedns.com
10.254.254.254
# BELOW SYNTAX WORKS FOR CSF AND NEWER VERSIONS OF APF 
#NTP servers (NTP outgoing port only)
udp:out:d=123:d=209.59.139.28
udp:out:d=123:d=64.91.251.155
udp:out:d=123:d=209.59.139.7
udp:out:d=123:d=209.59.139.82
#Monitoring servers (MySQL ports only)
tcp:in:d=3306:s=10.20.9.0/24
tcp:in:d=3306:s=10.30.9.0/24
tcp:in:d=3306:s=10.40.11.0/28
#Monitoring ICMP - to avoid rate limit.
icmp:in:d=ping:s=10.20.9.0/24
icmp:in:d=ping:s=10.30.9.0/24
icmp:in:d=ping:s=10.40.11.0/28
##END LIQUIDWEB##
```

Please also add the following to either `/etc/csf/csf.ignore` or `/usr/local/bfd/ignore.hosts`

```
#LiquidWeb Monitoring IP range
10.20.9.0/24
10.30.9.0/24
10.40.11.0/28
#End LiquidWeb Monitoring
```

Hosts Access
------------

```
# Offices
ALL : 10.10.4.0/255.255.254.0 : allow
ALL : 10.20.4.0/255.255.252.0 : allow
ALL : 10.30.2.0/255.255.255.0 : allow
ALL : 10.30.104.0/255.255.255.0 : allow
ALL : 10.30.4.0/255.255.252.0 : allow
ALL : 10.20.7.0/255.255.255.0 : allow
# Workbench
ALL : 69.167.129.192/255.255.255.240 : allow
# Infrastructure
ALL : 209.59.157.254 : allow
ALL : 69.167.128.254 : allow
ALL : 10.10.10.10 : allow
ALL : 10.254.254.254 : allow
# Monitoring
ALL : 10.20.9.0/255.255.255.0 : allow
ALL : 10.30.9.0/255.255.255.0 : allow
ALL : 10.40.11.0/255.255.255.240 : allow
```

iptables
--------

In the file, make sure to add the `BEGIN LiquidWeb` to `End LiquidWeb` sections into the file `/etc/sysconfig/iptables`

```
# Firewall configuration written by system-config-securitylevel
# Manual customization of this file is not recommended.
*filter
:INPUT ACCEPT [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:RH-Firewall-1-INPUT - [0:0]
-A INPUT -j RH-Firewall-1-INPUT
-A FORWARD -j RH-Firewall-1-INPUT
-A RH-Firewall-1-INPUT -i lo -j ACCEPT
-A RH-Firewall-1-INPUT -p icmp --icmp-type any -j ACCEPT
-A RH-Firewall-1-INPUT -p 50 -j ACCEPT
-A RH-Firewall-1-INPUT -p 51 -j ACCEPT
-A RH-Firewall-1-INPUT -p udp --dport 5353 -d 224.0.0.251 -j ACCEPT
-A RH-Firewall-1-INPUT -p udp -m udp --dport 631 -j ACCEPT
-A RH-Firewall-1-INPUT -p tcp -m tcp --dport 631 -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -m tcp -p tcp --dport 22 -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -m tcp -p tcp --dport 80 -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -m tcp -p tcp --dport 443 -j ACCEPT
#Begin LiquidWeb
#LiquidWeb Access
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.10.4.0/23 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.20.4.0/22 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.20.4.0/22 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.30.4.0/22 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.30.2.0/24 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.30.104.0/24 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 69.167.129.192/28 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.2.0.1/24 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 209.59.157.254 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 69.167.128.254 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.10.10.10 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.254.254.254 -m tcp -p tcp -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.20.7.0/24 -m tcp -p tcp -j ACCEPT
#LiquidWeb Monitoring
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.20.9.0/24 -m tcp -p tcp --dport 3306 -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.30.9.0/24 -m tcp -p tcp --dport 3306 -j ACCEPT
-A RH-Firewall-1-INPUT -m state --state NEW -s 10.40.11.0/28 -m tcp -p tcp --dport 3306 -j ACCEPT
#End LiquidWeb
-A RH-Firewall-1-INPUT -j REJECT --reject-with icmp-host-prohibited
COMMIT
```

firewalld
---------

Run the following commands:

```bash
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.10.4.0/23" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.20.4.0/22" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.30.4.0/22" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.30.2.0/22" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.30.104.0/24" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="69.167.129.192/28" protocol value="tcp" accept"

firewall-cmd --permanent --add-rich-rule="rule family="ipv6" source address="2607:fad0:32:a01::/64" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv6" source address="2607:fad0:32:a02::/64" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv6" source address="2607:fad0:32:a03::/64" protocol value="tcp" accept"

firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.20.1.0/24" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="209.59.157.254" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="69.167.128.254" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.10.10.10" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.254.254.254" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.20.7.0/23" protocol value="tcp" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="209.59.139.28" port protocol="udp" port="123" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="209.59.139.7" port protocol="udp" port="123" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="209.59.139.82" port protocol="udp" port="123" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.20.9.0/24" port protocol="tcp" port="3306" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.30.9.0/24" port protocol="tcp" port="3306" accept"
firewall-cmd --permanent --add-rich-rule="rule family="ipv4" source address="10.40.11.0/28" port protocol="tcp" port="3306" accept"

firewall-cmd --reload
```

Hardware Firewall
-----------------

```
object-group network LWOffice
 network-object 10.10.4.0 255.255.254.0
 network-object 10.20.4.0 255.255.252.0
 network-object 10.30.4.0 255.255.252.0
 network-object 10.20.9.0 255.255.255.0
 network-object 10.30.9.0 255.255.255.240
 network-object 10.40.11.0 255.255.255.240
 network-object 10.20.7.0 255.255.255.0
```

Allow ping:

```
access-list IN extended permit icmp object-group full-access object-group internal
```

External Office IPs
-------------------

Our offices are behind NAT for external traffic. The IPs for this are:

* 209.59.139.141
* 67.225.149.136
* 69.167.130.5
* 69.167.130.9
* 69.167.130.11
* 69.167.130.12
* 69.167.130.13