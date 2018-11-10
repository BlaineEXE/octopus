#!/usr/bin/env bash
source tests/shared.sh

echo "Running config file tests ..."

# Make our working dir something different
mkdir /root/work
pushd /root/work &> /dev/null || exit 1

configfile_name="config.yaml"
configfile_placement_options="/etc/octopus /root/.octopus ./.octopus"

# Test successful connection to all hosts with config file in all config file placements
for placement in $configfile_placement_options; do
  mkdir -p "$placement"
  cat > "$placement/$configfile_name" << EOF
groups-file: $GROUPFILE
EOF
  assert_success "with config file in $placement" octopus -g all run hostname
  rm -rf "$placement"
done
