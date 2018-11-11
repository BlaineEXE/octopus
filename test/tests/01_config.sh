#!/usr/bin/env bash
source tests/shared.sh

echo "Running CLI tests ..."

# move ssh keys away from default location for testing identity-file arg
mv "$HOME/.ssh" "$HOME/ssh-keys"
assert_success "with --host-groups and --identity-file args" \
  octopus --host-groups all --identity-file "$HOME/ssh-keys"/id_rsa run hostname
assert_success "with -g and -i args" octopus -g all -i "$HOME/ssh-keys"/id_rsa run hostname
mv "$HOME/ssh-keys" "$HOME/.ssh"

# move group file away from default location for testing group-file arg
mkdir -p "$HOME/work"
mv "$GROUPFILE" "$HOME/work/groups-file"
assert_success "with --groups-file arg" \
  octopus -g all --groups-file "$HOME/work/groups-file" run hostname
assert_success "with -f arg" octopus -g all -f "$HOME/work/groups-file" run hostname
mv "$HOME/work/groups-file" "$GROUPFILE"


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
