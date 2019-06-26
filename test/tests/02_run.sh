#!/usr/bin/env bash
source tests/shared.sh

echo "Running 'octopus run' tests ..."

# All containers should have the same content in /bin
ls='ls --width=0 /bin | md5sum'
ls_expected="$(bash -c "$ls")"
assert_success 'with success on all nodes' octopus -g all run "$ls"
assert_output_count "$ls_expected" $NUM_HOSTS
for host in $HOSTNAMES; do
  assert_output_count "$host" 1
done

assert_retcode 'with failure on all nodes' $NUM_HOSTS octopus -g all run 'thisisnotacommand'
# all nodes should report a command failure which includes the name of the failed command
assert_output_count 'thisisnotacommand' $NUM_HOSTS
for host in $HOSTNAMES; do
  assert_output_count "$host" 1
done

assert_success 'with success on one node' octopus -g one run 'hostname'
assert_output_count "$HOST_BASENAME" 2  # once for hostname header, once for command result

assert_success 'with success on rest of nodes' octopus -g rest run 'hostname'
assert_output_count "$HOST_BASENAME" $(( (NUM_HOSTS - 1) * 2 ))


# Move the 'hostname' binary to simulate failed hostname commands
hostname_location="/usr/bin/hostname"  # opensuse
# hostname_location="/bin/hostname"  # ubuntu
octopus -g all run "mv ${hostname_location} ${hostname_location}.bkp" 1> /dev/null

assert_retcode 'with hostname failure on all nodes' 0 octopus -g all run 'ls'
assert_output_count "$HOST_BASENAME" 0

octopus -g all run "mv ${hostname_location}.bkp ${hostname_location}" 1> /dev/null


# move the 'ls' binary for 'one' node to simulate failed command on subset of nodes
ls_location="/usr/bin/ls"  # opensuse
# ls_location="/bin/ls"  # ubuntu
octopus -g one run "mv ${ls_location} ${ls_location}.bkp" 1> /dev/null
assert_retcode 'with command failure on subset of nodes' 1 octopus -g all run 'ls'
for host in $HOSTNAMES; do # command failure should still report hostnames
  assert_output_count "$host" 1
done
octopus -g one run "mv ${ls_location}.bkp ${ls_location}" 1> /dev/null
