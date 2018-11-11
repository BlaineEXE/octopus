#!/usr/bin/env bash

#
# FUNCTIONS
#

# Automatically report the number of failures any time a test script exits
trap 'exit $num_failures' EXIT

printarray () {
  for arg in "$@"; do
      # testing if the argument contains space(s)
      if [[ $arg =~ \  ]]; then
        # enclose in double quotes if it does
        arg=\"$arg\"
      fi
      echo -n "$arg "
  done
}

num_passes=0
function pass () {
  num_passes=$((num_passes + 1))
  printf "      PASS: "
  printarray "$@"
  echo ''
}

num_failures=0
function fail () {
  num_failures=$((num_failures + 1))
  printf "      FAIL: "
  printarray "$@"
  echo ''
}

function assert_retcode () { # 1) test name 2) retcode  ...) command
  echo "  $1"
  "${@:3}" &> /tmp/output
  if [ "$?" = "$2" ]; then
    pass "${@:3}"
  else
    cat /tmp/output
    fail "${@:3}"
  fi
}

function assert_success () { # 1) test name  ...) command
  assert_retcode "$1" 0 "${@:2}"
}

function assert_output_count () { # 1) output desired 2) desired count
  if [ "$(grep --count "$1" /tmp/output)" == "$2" ]; then
    pass "'$1' is in output exactly $2 times"
  else
    cat /tmp/output
    fail "'$1' is in output $(grep --count "$1" /tmp/output) times; expected $2"
  fi
}

#
# SETUP
#

# Double check that our home dir is /root
if [ "$HOME" != "/home/tester" ]; then
  echo "Home $HOME is not /home/tester"
  exit 1
fi

GROUPFILE="$HOME/_node-list"
cp "$HOME/$GROUPFILE_DIR/_node-list" "$GROUPFILE"
chown "$(id --user --name)" "$GROUPFILE"

# Generate the HOSTNAMES variable which lists all our hostnames
HOSTNAMES=""
for i in $(seq 1 $NUM_HOSTS); do
  HOSTNAMES="$HOSTNAMES $HOST_BASENAME-$i"
done
