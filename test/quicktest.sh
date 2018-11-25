#!/usr/bin/env bash
set -Eeuo pipefail

# quicktest checks that the current system can call basic Octopus functionality against test hosts.

# Env vars should be exported by makefile
echo ""
echo "Running quicktest with config:"
echo "  NUM_HOSTS: $NUM_HOSTS"
echo "  HOST_BASENAME: $HOST_BASENAME"
echo "  GROUPFILE: $GROUPFILE"
echo ""

set -x

# Most basic test is to copy a dir to all hosts and then run 'ls' on the dir which should now have
# been created on the hosts.
_output/octopus -i "$PWD"/.ssh/id_rsa -f "$GROUPFILE" -g all copy -r "$PWD"/.ssh/ /root/quicktest/
_output/octopus -i "$PWD"/.ssh/id_rsa -f "$GROUPFILE" -g all run 'ls -alFh /root/quicktest/.ssh/'
