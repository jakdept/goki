Additional OpenSSH
==================

Overview
--------

There is a bunch of information with OpenSSH that we don't typically use on servers. Most of it isn't really necessary. So it hides here.

SSH/SSHD Config
----------

On servers, your configuration file for OpenSSH server is typically located at `/etc/ssh/sshd_config`. If you need to adjust something for the server, you can likely adjust it there.

On computers that you're going to SSH from (you're going to use the SSH client) the main SSH configuration is usually at `/etc/ssh/ssh_config`. However, anything you can do there, you can do in `~/.ssh/config` - so it's probably best to just make your changes there.

Host Sections
-------------

Within your SSH config file, you can set options outside of Host sections. But you probably don't want to. So, don't. Split up your SSH connections by Host sections, and then you can configure different things for different things.

Write-Protected/Broken Pipe
---------------------------

Sometimes, on an idle SSH connection you get disconnected you say? Use `ServerAliveInterval` to send a packet every so often so you don't get automatically disconnected. You can do this with:

```
#Value below in seconds.
ServerAliveInterval 30
#Value below is the number of times ServerAliveInterval is done per session.
ServerAliveCountmax 1200
```

Shared SSH Sockets
------------------

Normally, when you connect to a server via SSH, it connects via a network socket just for that SSH connection. Instead of doing that, you can configure SSH to create the network socket when SSH'ing to a server, then share that between different SSH connections to the same server. Then, if you connect to the server a second time with SSH, it will use the already open connection - think `rsync`, `scp`, `vncviewer` in addition to SSH.

To set this up, add the following lines to your SSH config in a Host section:

```
ControlPath ~/.ssh/.master-%r@%h:%p
ControlMaster auto
ControlPersist 1
```

Hashed Public Keys
------------------

If your `~/.ssh/known_hosts` looks like it's hashed like so:

```
|1|rNMjbSOiccDtv2c8qfbDENzLMKh=|nU-aqGDAXASMnKreS4gBPC8mBiY= ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAvcXoPQ3SSxG9XjWdcQsRNSRBu9R0fWaQapcesMrja1t0Z9mXyU6igqZHxu3adi3OQSxDLb/UsIYiSZScoQnE0SyGy1XpzNuCMvQ9WUVv+6BkVx1VYTbhWDU+t/pT9D3tOuORoqo5HdcKub2ppmjWNqiifI3UVUFMWAfAVbqFG+PYphBcxHA75ZYu+Fwhn5Gsoz+grCpmnxf0cI3O1MDGEY7TeK/pertBPlKirX4lXxUdQ8fp/TTXl+GyHk21BQaRCm1SHE0MqyyJieQpcPRTliIFg5PLFcZtS9I7g7RWxv5l5tuhna+c9jzoZZkYmoxrlX63sa2XDwUHGWxP10yQa4w=
```

This is because you have `HashKnownHosts` set to `yes`.

Strict Host Key Checking
------------------------

When you SSH to a location you haven't been to before, you might see a message like:

```bash
jack@jack:~$ ssh root@deering
The authenticity of host 'deering (67.225.255.50)' can't be established.
RSA key fingerprint is cb:45:20:ac:e8:17:0b:7a:4f:63:d8:6d:78:91:6a:94.
Are you sure you want to continue connecting (yes/no)?
```

This is letting you know that you haven't been there, and prompting you before connecting and remebering that location. You can disable this notification by adding `StrictHostKeyChecking no` to your SSH configuration.

### Disabling Host Key Checking ###

If you want to go the next step, at home, you can disable host key checking all together by running:

```bash
rm -f ~/.ssh/known_hosts
ln -s /dev/null ~/.ssh/known_hosts
```

Warning: You shouldn't do this at Liquid Web - we want to know when Host Keys change. This circumvents security we have in place here. If you do this here at LW, you should probably just quit.

Shared Server Host Sections
---------------------------
To make it easeir to get into our shared servers, it's likely wise to add the following to your ssh configuration:

```
#Shared servers
Host adama.liquidweb.com aerelon.liquidweb.com alpha.liquidweb.com arachnid.liquidweb.com archbishop.liquidweb.com ardala.liquidweb.com ari.liquidweb.com attar.liquidweb.com aura.liquidweb.com bandicoot.liquidweb.com barfolomew.liquidweb.com barin.liquidweb.com beast.liquidweb.com bender.liquidweb.com bester.liquidweb.com book.liquidweb.com calculon.liquidweb.com caprica.liquidweb.com carbon.liquidweb.com centauri.liquidweb.com cerberus.liquidweb.com chaos.liquidweb.com cobb.liquidweb.com colossus.liquidweb.com cornelius.liquidweb.com corona.liquidweb.com cortana.liquidweb.com crais.liquidweb.com crichton.liquidweb.com dallas.liquidweb.com darko.liquidweb.com deering.liquidweb.com delenn.liquidweb.com dizzy.liquidweb.com dodge.liquidweb.com dot.liquidweb.com dozer.liquidweb.com efram.liquidweb.com electron.liquidweb.com elzar.liquidweb.com falcon.liquidweb.com filter.liquidweb.com flexo.liquidweb.com formosa.liquidweb.com franklin.liquidweb.com fry.liquidweb.com galen.liquidweb.com
user=root

Host gambit.liquidweb.com gamera.liquidweb.com garibaldi.liquidweb.com ghidora.liquidweb.com ghost.liquidweb.com gibson.liquidweb.com g-kar.liquidweb.com goodfellow.liquidweb.com grig.liquidweb.com hawk.liquidweb.com helo.liquidweb.com hesh.liquidweb.com honorious.liquidweb.com huer.liquidweb.com hydra.liquidweb.com hydrogen.liquidweb.com hydro.sourcedns.com hypnotoad.liquidweb.com ibanez.liquidweb.com impulse.liquidweb.com inara.liquidweb.com indigo.liquidweb.com infinity.liquidweb.com insomnia.liquidweb.com jade.liquidweb.com jenkins.liquidweb.com julius.liquidweb.com kala.liquidweb.com kane.liquidweb.com karubi.liquidweb.com kaylee.liquidweb.com klytus.liquidweb.com koala.liquidweb.com kobol.liquidweb.com kodan.liquidweb.com lambert.liquidweb.com landon.liquidweb.com lava.liquidweb.com lennier.liquidweb.com lepton.liquidweb.com lightning.liquidweb.com lonestarr.liquidweb.com lucius.liquidweb.com luro.liquidweb.com magnum.liquidweb.com mal.liquidweb.com matrix.liquidweb.com
user=root

Host maximus.liquidweb.com mcp.liquidweb.com ming.liquidweb.com minister.liquidweb.com mollari.liquidweb.com morbo.liquidweb.com mordor.liquidweb.com mothra.liquidweb.com moya.liquidweb.com munson.liquidweb.com nitro.liquidweb.com nomad.liquidweb.com nova.liquidweb.com nucleon.liquidweb.com oak.liquidweb.com orga.liquidweb.com phoenix.liquidweb.com picon.liquidweb.com pilot.liquidweb.com platinum.liquidweb.com pod6.liquidweb.com poseidon.sourcedns.com quantum.liquidweb.com quark.liquidweb.com rails.liquidweb.com rasczek.liquidweb.com rico.liquidweb.com ripley.liquidweb.com rogan.liquidweb.com rogers.liquidweb.com rogue.liquidweb.com rygel.liquidweb.com rylos.liquidweb.com sabre.sourcedns.com sandar.liquidweb.com shepard.liquidweb.com sheridan.liquidweb.com solar.liquidweb.com sonic.liquidweb.com speed.liquidweb.com sputnik.liquidweb.com storm.liquidweb.com synapse.liquidweb.com tachyon.liquidweb.com talia.liquidweb.com tam.liquidweb.com thade.liquidweb.com theopolis.liquidweb.com
user=root

Host thun.liquidweb.com titanium.liquidweb.com tomcat.liquidweb.com tungsten.liquidweb.com underworld.liquidweb.com vapor.liquidweb.com varek.liquidweb.com venom.liquidweb.com vespa.liquidweb.com vir.liquidweb.com vision.liquidweb.com vultan.liquidweb.com warlock.liquidweb.com wash.liquidweb.com wintermute.liquidweb.com wolf.liquidweb.com xenon.liquidweb.com xur.liquidweb.com zack.liquidweb.com zaius.liquidweb.com zarkov.liquidweb.com zeus.liquidweb.com zim.liquidweb.com zira.liquidweb.com zoe.liquidweb.com ztestupdate.liquidweb.com bartertown.liquidweb.com bassey.liquidweb.com clunk.liquidweb.com dealgood.liquidweb.com entity.liquidweb.com masterblaster.liquidweb.com maxx.liquidweb.com nickcagefs.liquidweb.com nix.liquidweb.com sarse.liquidweb.com slake.liquidweb.com thunderdome.liquidweb.com adama aerelon alpha arachnid archbishop ardala ari attar aura bandicoot barfolomew barin beast bender bester book calculon caprica carbon centauri cerberus chaos cobb colossus cornelius corona cortana
user=root

Host crais crichton dallas darko deering delenn dizzy dodge dot dozer efram electron elzar falcon filter flexo formosa franklin fry galen gambit gamera garibaldi ghidora ghost gibson g-kar goodfellow grig hawk helo hesh honorious huer hydra hydrogen hydro.sourcedns.com hypnotoad ibanez impulse inara indigo infinity insomnia jade jenkins julius kala kane karubi kaylee klytus koala kobol kodan lambert landon lava lennier lepton lightning lonestarr lucius luro magnum mal matrix maximus mcp ming minister mollari morbo mordor mothra moya munson nitro nomad nova nucleon oak orga phoenix picon pilot platinum pod6 poseidon.sourcedns.com quantum quark rails rasczek rico ripley rogan rogers rogue rygel rylos sabre.sourcedns.com sandar shepard sheridan solar sonic speed sputnik storm synapse tachyon talia tam thade theopolis thun titanium tomcat tungsten underworld vapor varek venom vespa vir vision vultan warlock wash wintermute wolf xenon xur zack zaius zarkov zeus zim zira zoe ztestupdate bartertown
user=root

Host bassey clunk dealgood entity masterblaster maxx nickcagefs nix sarse slake thunderdome  
user=root

Host srs001.sourcedns.com srs002.sourcedns.com srs001 srs002
user=root
```

This will set your user to root for these connections. This is broken into multiple `Host` sections to account for the maximum line width.

Other Internal Host Section
---------------------------

The following usually makes it easier to get into Storm - just make sure that you change your username below.

```
#storm
Host *.xvps  *.xvps.liquidweb.com
	user=jhayhurst
	 
#misc other internal
Host dc1* dc2* dc3* vz* *.liquidweb.com 
	user=jhayhurst
```

Customer server's Host Section
------------------------------

I usually find it's easiest to have a section in your SSH config just for customer boxes - and putting my ControlPath stuff there. You can negate a Host match with a `!` in front of the host name, like so:

```
#default
Host * !*.perboner.com !perboner.com !jack* !vps !colo
	Port 22
	User root
	ControlPath ~/.ssh/.master-%r@%h:%p
	ControlMaster auto
	ControlPersist 1
```

OSX Changing the SSH port
-------------------------

On OSX, `LaunchDaemon` lauches `OpenSSH` and overrides any port number you put in there. As such, you have to set this somewhere else - in the file `/System/Library/LaunchDaemons/ssh.plist`. In this file, I'm configured to use port `3006`.

```
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Disabled</key>
	<true/>
	<key>Label</key>
	<string>com.openssh.sshd</string>
	<key>Program</key>
	<string>/usr/libexec/sshd-keygen-wrapper</string>
	<key>ProgramArguments</key>
	<array>
		<string>/usr/sbin/sshd</string>
		<string>-i</string>
	</array>
	<key>Sockets</key>
	<dict>
		<key>Listeners</key>
		<dict>
			<key>SockServiceName</key>
			<string>3006</string>
			<key>Bonjour</key>
			<array>
				<string>ssh</string>
				<string>sftp-ssh</string>
			</array>
		</dict>
	</dict>
	<key>inetdCompatibility</key>
	<dict>
		<key>Wait</key>
		<false/>
	</dict>
	<key>StandardErrorPath</key>
	<string>/dev/null</string>
	<key>SHAuthorizationRight</key>
	<string>system.preferences</string>
	<key>POSIXSpawnType</key>
	<string>Interactive</string>
</dict>
</plist>
```