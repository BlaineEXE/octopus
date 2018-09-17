SRC_ROOT ?= github.com/BlaineEXE/octopus
OUTPUT_DIR ?= _output

BUILD_OPTS ?= -buildmode pie

dep:
	dep ensure

build: dep
	go build \
			$(BUILD_OPTS) \
			-o $(OUTPUT_DIR)/octopus \
		$(SRC_ROOT)/...

install: dep
	go install $(BUILD_OPTS) $(SRC_ROOT)/...

clean:
	rm -rf $(OUTPUT_DIR)
	rm -rf vendor/
