# pam-usher

A configurable Linux PAM designed to setup user sessions upon login

## Motivation

We needed a way to create directories automatically upon a new user session.
Normally `pam_mkhomedir` would be sufficient, but these directories were in
places outside of the home directory. For instance, if you have a network share
at `/shares/users` and user directories inside there, we needed
`/shares/users/$USER`. 

While `bash` scripting could have solved this, we wanted a way that was agnostic
to the terminal used and had the potential to be added onto later. 

## Design

We decided to use `go` for the potential expansion into authentication modules.
All calls are error checked and it will return proper error states. 

Consideration for performance was also taken into account. For the average user,
the least amount of syscalls will be executed (only one to check for the exist
of a directory). This reduces the latency for logging into multiple sessions.

## Setup

### Building

To build, have golang installed on your machine and use `make`.

### Installing

For ubuntu, add the resulting `.so` to `/usr/lib/x86_64-linux-gnu/security`.
Then update `/etc/pam.d/common-session` to include the following:

```
session optional pam_usher.so
```


### Configuring

Now add a configuration file at `/etc/nibious/config.yaml` that contains the following:

```yaml
UserDirectories:
  - /tmp/users
```

The user directory upon session creation will be created like
`/tmp/users/$USER`. You can have multiple directories created.

