.PHONY: check_env_ build

all: bmv2 stratum_bf stratum_bfrt

check_env_:
ifndef TARGET
	$(error TARGET undefined)
endif

build: check_env_
	go build -tags "$(TARGET)" -o p4rt-test-$(TARGET) ./bin/main.go

bmv2:
	TARGET="bmv2" $(MAKE) build

stratum_bf:
	TARGET="stratum-bf" $(MAKE) build

stratum_bfrt:
	TARGET="stratum-bfrt" $(MAKE) build
