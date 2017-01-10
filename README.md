# xen-cli

A thing for managing multiple XenServer hosts through XAPI on unix-like
desktops.

__THIS IS A WORK IN PROGRESS__

## Install

Install the daemon:

```
go build github.com/johnprather/xen-cli/xen-daemon
```

Install the CLI:

```
go build github.com/johnprather/xen-cli/xen
```

### Usage

The daemon runs in the background, and the CLI is used to exchange information
with it.

#### Daemon

Run the daemon:

```
xen-daemon
```

Trigger daemon to re-open logfile (should log rotation be desired):

```
killall -HUP xen-daemon
```

Launch daemon in debug mode (process does not fork to background, and logging
  is sent to stdout):

```
xen-daemon -debug
```

Specify a custom base dir, where the daemon creates its files (default
is $HOME/.xen-cli/):

```
xen-daemon -base.dir=/path/to/basedir
```

Specify a custom socket path:

```
xen-daemon -socket.path=/path/to/socket
```

Specify a custom logfile:

```
xen-daemon -log.file=/path/to/logfile
```

Specify a custom save file for the xapi servers list:

```
xen-daemon -servers.file=/path/to/servers.json
```

#### CLI

Set a default XAPI password for use when adding new servers:

```
xen -password
```

Use the CLI as an interactive shell:

```
xen
```

Use the CLI for single or scripted/batched commands:

```
xen help
xen help server
```

Specify a custom base directory (default is $HOME/.xen-cli/):

```
xen -base.dir=/path/to/basedir
```

Specify a custom socket path:

```
xen -socket.path=/path/to/socket
```

### XAPI Servers

Add and remove XAPI servers as necessary.

Server list is saved in JSON format when modified so that it can be loaded
the next time the process launches.

#### List XAPI Servers

List currently configured XAPI servers:

```
server list
```

#### Add An XAPI Server

Add an XAPI server to the list of servers.

```
server add <hostname>
```

The default xapi password will be attempted at the time the server is added.
If successful, the password is saved (associated with the server's IP address)
to your system's keychain/keyring/wallet system, and will not be affected by
later changes to the default xapi password.

#### Remove An XAPI Server

Remove an XAPI server from the list of servers:

```
server remove <hostname>
```

Polling of data for this server ceases, existing data for this server is
deleted, and the server is removed from the list of known xapi servers.
