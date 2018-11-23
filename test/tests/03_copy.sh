#!/usr/bin/env bash
source tests/shared.sh

echo "Running 'octopus copy' tests ..."

# Set up a test file tree for copying
mkdir work/
mkdir work/dirA
mkdir work/dirB
mkdir work/dirWO && chmod 333 work/dirWO # write-only dir
head --bytes=1K /dev/urandom > work/fileA
head --bytes=2K /dev/urandom > work/fileB
head --bytes=1M /dev/urandom > work/dirA/fileAA && chmod 600 work/dirA/fileAA # perms 0600
head --bytes=1K /dev/urandom > work/dirA/fileAB
head --bytes=3K /dev/urandom > work/dirB/fileBA
head --bytes=2K /dev/urandom > work/dirB/fileBWO && chmod -r work/dirB/fileBWO # write-only file
head --bytes=1K /dev/urandom > work/dirWO/fileWOA # readable file in write-only dir

function get_md5sum () {
  md5sum "$1" | awk '{print $1}'
}

assert_success 'with successful file copy to all nodes' \
  octopus -g all copy work/fileA work/fileB /tmp/1/
assert_success '  and file md5sums match' octopus -g all run 'md5sum /tmp/1/*'
assert_output_count "$(get_md5sum work/fileA)" $NUM_HOSTS
assert_output_count "$(get_md5sum work/fileB)" $NUM_HOSTS
for host in $HOSTNAMES; do # hostnames should be reported
  assert_output_count "$host" 1
done

assert_retcode 'with failure to copy a dir without setting --recursive' $NUM_HOSTS \
  octopus -g all copy work/fileA work/dirA/ work/fileB /tmp/2/
assert_success '  and file md5sums match for files not in dir' octopus -g all run 'md5sum /tmp/2/*'
assert_output_count "$(get_md5sum work/fileA)" $NUM_HOSTS
assert_output_count "$(get_md5sum work/fileB)" $NUM_HOSTS

assert_success 'with successful recursive dir copy to all nodes' \
  octopus -g all copy --recursive work/dirA /tmp/3/
assert_success '  and file md5sums match' octopus -g all run 'md5sum /tmp/3/dirA/*'
assert_output_count "$(get_md5sum work/dirA/fileAA)" $NUM_HOSTS
assert_output_count "$(get_md5sum work/dirA/fileAB)" $NUM_HOSTS
assert_success '  and fileAA has perms 0600' octopus -g all run 'ls -l /tmp/3/dirA/fileAA'
assert_output_count '-rw-------' $NUM_HOSTS

assert_retcode 'with failure to copy unreadable file in readable dir' $NUM_HOSTS \
  octopus -g all copy -r work/dirB /tmp/4
assert_success '  and file md5sums match for readable files' octopus -g all run 'md5sum /tmp/4/dirB/*'
assert_output_count "$(get_md5sum work/dirB/fileBA)" $NUM_HOSTS
assert_output_count "/tmp/4/dirB/fileBWO" 0

assert_retcode 'with failure to copy unreadable dir' $NUM_HOSTS \
  octopus -g all copy -r work/dirWO /tmp/5
assert_retcode '  and dir does not exist on host' $NUM_HOSTS octopus -g all run 'ls /tmp/5/dirWO'

assert_retcode 'with failure to copy to file on host' $NUM_HOSTS \
  octopus -g all copy -r work/dirA /dev/null

assert_success 'with successful copy to one node' \
  octopus -g one copy -r work/dirA /tmp/6
assert_success '  and file md5sums match' octopus -g one run 'md5sum /tmp/6/dirA/*'
assert_output_count "$(get_md5sum work/dirA/fileAA)" 1
assert_output_count "$(get_md5sum work/dirA/fileAB)" 1
assert_retcode '  and no files on rest of nodes' 2 octopus -g all run 'ls /tmp/6'

#
# SFTP copy throughput settings
# It is possible that this test could report a false failure due to fluctuations in network speeds,
# but in testing it has never happened.
#
paths="work/file* work/dirA"

# default buffer = 32kib, requests per file = 64
start_time=$(date +%s%N) # nanoseconds
assert_success "baseline copy speed" octopus -g all copy -r $paths /tmp/7
end_time=$(date +%s%N)
baseline_time=$(( end_time - start_time ))

start_time=$(date +%s%N)
assert_success "slow copy" octopus -g all copy -r $paths /tmp/8 --buffer-size 1 --requests-per-file 1
end_time=$(date +%s%N)
slow_time=$(( end_time - start_time ))

# slow copy should be ~2048 times slower, but there is also the overhead of Octopus's processing,
# so the best we can say is that this should be slower
if [[ ! $slow_time -gt $baseline_time ]]; then
  fail "slow time $slow_time should be greater than medium time $baseline_time"
else
  pass "slow time $slow_time is greater than medium time $baseline_time"
fi

# buffer size greater than 256 kib should cause a failure with OpenSSH
assert_failure "ludicrous settings" octopus -g all copy -B 1024 -R 1024

# Remove all the files we wrote from the test hosts
octopus -g all run 'rm -rf /tmp/*' 1> /dev/null
