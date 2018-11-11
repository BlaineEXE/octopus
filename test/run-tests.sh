#!/usr/bin/env bash
set -u

# Env vars should be exported by makefile
echo ""
echo "Running tests with config:"
echo "  NUM_HOSTS: $NUM_HOSTS"
echo "  HOST_BASENAME: $HOST_BASENAME"
echo "  GROUPFILE_DIR: $GROUPFILE_DIR"
echo ""

num_suites=0
num_failures=0
function run_test () { # 1) test_name
  docker run --rm --name $RUNNER_NAME \
        --user tester \
        --volume $PWD/$GROUPFILE_DIR:/home/tester/$GROUPFILE_DIR:ro \
        --env NUM_HOSTS=$NUM_HOSTS \
        --env HOST_BASENAME=$HOST_BASENAME \
        --env GROUPFILE_DIR=$GROUPFILE_DIR \
      $IMAGE_TAG bash "tests/$1"
  num_failures=$((num_failures + $?)) # retval is number of failed tests
  num_suites=$((num_suites + 1))
  echo ""
}

run_test 01_config.sh
run_test 02_run.sh
run_test 03_copy.sh

# stop the first test host to simulate a node being unreachable
# host can't be restarted except by creating a new one, so make sure to do this very last
host=$HOST_BASENAME-1
docker stop "$host" 1> /dev/null
run_test 04_node-offline.sh

echo "Test suites run: $num_suites"
echo "Test failures: $num_failures"
echo ""
exit $num_failures
