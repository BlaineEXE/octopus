#!/usr/bin/env bash
source tests/shared.sh

echo "Running CLI tests ..."

# move ssh keys away from default location for testing identity-file arg
mv "$HOME/.ssh" "$HOME/ssh-keys"
assert_failure "with identity file not found" octopus -g all run hostname
assert_success "with --host-groups and --identity-file args" \
  octopus --host-groups all --identity-file "$HOME/ssh-keys"/id_rsa run hostname
assert_success "with -g and -i args" octopus -g all -i "$HOME/ssh-keys"/id_rsa run hostname
mv "$HOME/ssh-keys" "$HOME/.ssh"

# move group file away from default location for testing group-file arg
mkdir -p "$HOME/work"
mv "$GROUPFILE" "$HOME/work/groups-file"
assert_failure "with groups file not found" octopus -g all run hostname
assert_success "with --groups-file arg" \
  octopus -g all --groups-file "$HOME/work/groups-file" run hostname
assert_success "with -f arg" octopus -g all -f "$HOME/work/groups-file" run hostname
mv "$HOME/work/groups-file" "$GROUPFILE"

# test unreadable config files
mkdir -p "$HOME/work"
touch "$HOME/work/unreadable-file"
chmod 333 "$HOME/work/unreadable-file" # write-only
assert_failure "with unreadable identity file" \
  octopus -g all -i "$HOME/work/unreadable-file" run hostname
assert_failure "with unreadable groups file" \
  octopus -g all -f "$HOME/work/unreadable-file" run hostname

# Start a second ssh daemon on hosts listening on port 3022 to test connecting to different port
assert_success "with ssh daemon running on port 3022" \
  octopus -g all run '/usr/sbin/sshd -p 3022'
assert_success "  and daemon is verified running on port 3022" \
  octopus -g all run 'ps aux | grep "/usr/sbin/sshd -p 3022"'
assert_failure "  ... with invalid port" octopus -g all -p 5555 all run hostname
assert_success "  ... with --port arg" octopus -g all --port 3022 run hostname
assert_success "  ... with -p arg" octopus -g all -p 3022 run hostname
assert_success "  and finally killing ssh daemon on port 3022" \
  octopus -g all run 'pkill --full "/usr/sbin/sshd -p 3022"'

# alternate user
assert_success "with default user (root)" octopus -g all run 'id --user --name'
assert_output_count "root" $NUM_HOSTS
assert_failure "with invalid user" octopus -g all -u 'invalid' run 'ls'
assert_success "with --user arg" octopus -g all --user 'tester' run 'id --user --name'
assert_output_count "tester" $NUM_HOSTS
assert_success "with -u arg" octopus -g all -u 'tester' run 'id --user --name'
assert_output_count "tester" $NUM_HOSTS

echo ""
echo "Running config file tests ..."

# Make our working dir something different
mkdir -p "$HOME/work"
pushd "$HOME/work" &> /dev/null || exit 1

configfile_name="config.yaml"
configfile_placement_options="/etc/octopus $HOME/.octopus ./.octopus"

# Test successful connection to all hosts with config file in all config file placements
for placement in $configfile_placement_options; do
  mkdir -p "$placement"
  cat > "$placement/$configfile_name" << EOF
groups-file: $GROUPFILE
EOF
  assert_success "with config file in $placement" octopus -g all run hostname
  rm -rf "$placement/$configfile_name"
done

popd 1> /dev/null

echo ""
echo "Running 'octopus host-groups' tests ..."

assert_success "with default groups" octopus host-groups
assert_output_count "one" 1
assert_output_count "rest" 1

# create custom, more complex group file for this test
mkdir -p "$HOME/work"
mv "$GROUPFILE" "$HOME/work/groups-file"
assert_failure "with groups file not found" octopus host-groups
cat > "$GROUPFILE" << EOF
one="1.1.1.1"
two="2.2.2.2"
three="3.3.3.3"

first="$one"
rest="$two"
rest="$rest $three"
EOF
assert_success "with custom groups file" octopus host-groups
assert_output_count "one" 1
assert_output_count "two" 1
assert_output_count "three" 1
assert_output_count "first" 1
assert_output_count "rest" 1
assert_num_output_lines_with_text 5
mv "$HOME/work/groups-file" "$GROUPFILE"
