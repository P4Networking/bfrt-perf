# BFRuntime Go Client

Client library and flow write performance tester

## Setup (one time)
*Install Go* (>= 1.13.3)
https://golang.org/doc/install

## Building binaries

```
make
```

## Test with Tofino device

Start BFRuntime on your Tofino switch

Then, you can run the test:
```
./bfrt_test_tofino \
 -iterations 1000 \
 -batchSize 1000  \
 -numThreads 1 \
 -p4Name tna_simple_router
```

<img src="https://github.com/P4Networking/bfrt-perf/raw/master/test_tofino.gif" width="688px" height="342px" />

Notes:
- Remember to update the target string to match the IP of your switch
- Remember to change the table and action which you want to test
- Update GOOS to match the operating system of where you will run the test binary
