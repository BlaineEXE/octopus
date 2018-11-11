export NUM_HOSTS ?= 3
export HOST_BASENAME ?= octopus-test-host
export RUNNER_NAME ?= octopus-test-runner

export IMAGE_TAG ?= octopus-test

export GROUPFILE_DIR ?= groupfile
export GROUPFILE ?= $(GROUPFILE_DIR)/_node-list
$(shell mkdir -p groupfile)

# teardown before and after to prevent any unexpected errors
all: teardown build setup run
	@ $(MAKE) teardown

build:
	@ docker build --tag $(IMAGE_TAG) .

setup:
	@ echo "Setting up test hosts ..."
	@ ./setup-test-hosts.sh

run:
	@ ./run-tests.sh

teardown:
	@ echo "Tearing down test hosts..."
	@ for i in $$(seq 1 $(NUM_HOSTS)); do \
		host=$(HOST_BASENAME)-$$i ; \
		docker stop $$host ; \
	done || true
	@ rm $(GROUPFILE) || true