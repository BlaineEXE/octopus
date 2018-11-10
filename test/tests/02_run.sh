#!/usr/bin/env bash
source tests/shared.sh

echo "Running basic run tests ..."

# All containers should have the same content in /bin
ls='ls --width=0 /bin | md5sum'
ls_expected="$(bash -c "$ls")"
assert_success 'with success on all nodes' octopus -g all run "$ls"
assert_output_count "$ls_expected" 3
for host in $HOSTNAMES; do
  assert_output_count "$host" 1
done

assert_retcode 'with failure on all nodes' 3 octopus -g all run 'thisisnotacommand'
# all nodes should report a command failure which includes the name of the failed command
assert_output_count 'thisisnotacommand' 3
for host in $HOSTNAMES; do
  assert_output_count "$host" 1
done

assert_success 'with success on one node' octopus -g one run 'hostname'
assert_output_count "$HOST_BASENAME" 2  # once for hostname header, once for command result

assert_success 'with success on rest of nodes' octopus -g rest run 'hostname'
assert_output_count "$HOST_BASENAME" $(( (NUM_HOSTS - 1) * 2 ))


# Move the 'hostname' binary to simulate failed hostname commands
octopus -g all run 'mv /usr/bin/hostname /usr/bin/hostname.bkp' &> /dev/null

assert_retcode 'with hostname failure on all nodes' 0 octopus -g all run 'ls'
assert_output_count "$HOST_BASENAME" 0

octopus -g all run 'mv /usr/bin/hostname.bkp /usr/bin/hostname' &> /dev/null


# move the 'ls' binary for 'one' node to simulate failed command on subset of nodes
octopus -g one run 'mv /usr/bin/ls /usr/bin/ls.bkp' &> /dev/null
assert_retcode 'with command failure on subset of nodes' 1 octopus -g all run 'ls'
for host in $HOSTNAMES; do # command failure should still report hostnames
  assert_output_count "$host" 1
done
octopus -g one run 'mv /usr/bin/ls.bkp /usr/bin/ls' &> /dev/null
