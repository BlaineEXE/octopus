#!/usr/bin/env bash
source tests/shared.sh

echo "Running tests where node 'one' is unreachable ..."

NUM_REMAINING=$(( NUM_HOSTS - 1 ))

assert_retcode "with run command successes on remaining nodes" 1 octopus run -g all 'hostname'
assert_output_count "$HOST_BASENAME" $(( NUM_REMAINING * 2 ))

assert_success "with no errors running commands on the 'rest' group" octopus run -g rest 'hostname'
assert_output_count "$HOST_BASENAME" $(( NUM_REMAINING * 2 ))


assert_retcode "with copy successes on remaining nodes" 1 \
  octopus copy -g all -r '$HOME/tests' /tmp/1
assert_output_count "$HOST_BASENAME" $NUM_REMAINING
# get md5sums of only files, then get the md5sum of the list of md5sums to keep test output short
expected_md5sums="$(md5sum test/* 2>/dev/null | awk '{print $1}' | md5sum)"
assert_success "  and with file md5sums matching" \
  octopus run -g rest 'md5sum /tmp/1/test/* 2>/dev/null | awk "{print \$1}" | md5sum'
assert_output_count "$expected_md5sums" $NUM_REMAINING

assert_success "with no errors copying files to the 'rest' group" \
  octopus copy -g rest -r '$HOME/tests' /tmp/2
assert_success "  and with file md5sums matching" \
  octopus run -g rest 'md5sum /tmp/2/test/* 2>/dev/null | awk "{print \$1}" | md5sum'
assert_output_count "$expected_md5sums" $NUM_REMAINING
