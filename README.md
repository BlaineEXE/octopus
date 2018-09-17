Octopus
=========
Octopus is a commandline tool for running the same command on multiple remote hosts in parallel.

Theory
--------
Octopus is a simple tool inspired by `pdsh`'s ability to execute commands on multiple hosts in
parallel. In environments where `pdsh` cannot be installed, Go's ability to produce static binaries
is useful; thus, Octopus is written in Go. As long as one has the ability to `scp` or `rsync` files
to a host, a static `octopus` executable can be copied to it. A host may be a cluster admin node,
for example.

Octopus can execute arbitrary commands on multiple hosts in parallel, and hosts are grouped together
into "host groups" in a file which inspired by `pdsh`'s "genders" file. The host groups file for
Octopus is actually a Bash file with groups defined by variable definitions. A file which defines
host groups as Bash variables was chosen so that so that the same file may be used easily by both
Octopus and by user-made scripts.

Under the hood, Octopus uses `ssh` connections, and some `ssh` arguments are reflected in Octopus's
arguments.

**warning:** Octopus does not do verification of the remote hosts (`StrictHostKeyChecking=no`), and
it does not add entries to the user's known hosts file.

Usage
-------
See `octopus -help` for command usage.

An example host groups file can be found in the [config](config) directory.
