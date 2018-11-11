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
	@ # Run some test hosts, and generate a node list with group "one" the first ip, "rest" the
	@ # rest of the ips and "all" set to all of the ips
	@ echo '#!/usr/bin/env bash' > $(GROUPFILE)
	@ echo '' >> $(GROUPFILE)
	@ for i in $$(seq 1 $(NUM_HOSTS)); do \
		host=$(HOST_BASENAME)-$$i ; \
		echo "running test host $$host" ; \
		docker run --rm --name $$host --hostname $$host --detach $(IMAGE_TAG); \
		ip=$$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $$host) ; \
		echo "ip: $$ip" ; \
		if [ $$i = 1 ] ; then \
			echo "one='$$ip'" >> $(GROUPFILE) ; \
			echo "rest='" >> $(GROUPFILE) ; \
		else \
			echo $$ip >> $(GROUPFILE) ; \
		fi ; \
	done
	@ echo "'" >> $(GROUPFILE)
	@ echo 'all="$$one $$rest"' >> $(GROUPFILE)
	@ echo '' >> $(GROUPFILE)
	@ echo "  $(GROUPFILE) file:"
	@ cat $(GROUPFILE)

run:
	@ ./run-tests.sh

teardown:
	@ echo "Tearing down test hosts..."
	@ for i in $$(seq 1 $(NUM_HOSTS)); do \
		host=$(HOST_BASENAME)-$$i ; \
		docker stop $$host ; \
	done || true
	@ rm $(GROUPFILE) || true
