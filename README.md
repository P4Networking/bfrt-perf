# P4Runtime Go Client

Client library and flow write performance tester

## Setup (one time)
*Install Go* (>= 1.13.3)
https://golang.org/doc/install

## Building binaries

```
make
```

## Test with bmv2(stratum_bmv2)

Run Stratum BMv2:
```
docker run --privileged --rm -it -p 50001:50001 opennetworking/mn-stratum
```

Then, you can run the test:
```
./p4rt_test_bmv2 \
 -target localhost:50001 \
 -p4info test/bmv2/p4info.txt \
 -deviceConfig test/bmv2/bmv2.json \
 -iterations 1000 \
 -batchSize 100 \
 -numThreads 1 \
 -verbose
```

<img src="https://github.com/Yi-Tseng/p4r-perf/raw/master/test_bmv2.gif" width="688px" height="342px" />

## Test with Tofino device

Start Stratum on your Tofino switch

Then, you can run the test if you are using stratum_bf:
```
./p4rt_test_stratum_bf \
 -target localhost:28000 \
 -p4info p4info.txt \
 -deviceConfig tofino.bin,context.json \
 -iterations 1000 \
 -batchSize 100 \
 -numThreads 1 \
 -verbose
```

Or, using the following command if you are using the stratum_bfrt:
```
./p4rt_test_stratum_bfrt \
 -target localhost:28000 \
 -p4info test/montara/p4info.txt \
 -deviceConfig path/to/device/config \
 -iterations 1000 \
 -batchSize 100 \
 -numThreads 1 \
 -verbose
```

Notes:
- If you use the test files, they were compiled against SDE 9.2.0
- Remember to update the target string to match the IP of your switch (or run the test on the box)
- Update GOOS to match the operating system of where you will run the test binary
- You can use any P4 program/compiler version that you want, just be sure to update the paths
