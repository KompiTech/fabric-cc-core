# Prerequisites

1. Follow steps on https://hyperledger-fabric.readthedocs.io/en/release-2.3/prereqs.html
2. Follow steps on https://hyperledger-fabric.readthedocs.io/en/release-2.3/install.html
3. Test that you have ```peer``` binary accessible in your PATH
4. Clone the fabric-samples repo to some location on your disk (```git clone https://github.com/hyperledger/fabric-samples.git```), ```cd``` to cloned repo and ```git checkout 51f76977b0ee102ea7efc17875f2694c42823777```   
5. Export ```FABRIC_SAMPLES``` env var and set it to ```fabric-samples``` repo location that you previously cloned. This is necessary, as we are re-using test-network there.

# Common issues

First step when encountering issues is to _always_ use ```bash stop.sh``` to wipe previous network

## Issue 1

```
Error: failed to endorse chaincode install: rpc error: code = ResourceExhausted desc = trying to send message larger than max (116083851 vs. 104857600)
Chaincode installation on peer0.org1 has failed
Deploying chaincode failed
```
Make sure that fabric-samples repo is cloned somewhere outside from this repo on filesystem

## Issue 2

```
2021-04-12 12:08:21.407 CEST [main] InitCmd -> ERRO 001 Cannot run peer because error when setting up MSP of type bccsp from directory /home/user/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp: Setup error: nil conf reference
```
Make sure you have correct versions of Fabric Docker images (steps explained here: https://hyperledger-fabric.readthedocs.io/en/release-2.3/install.html)

## Issue 3

```
Error: Post "https://localhost:7053/participation/v1/channels": x509: certificate signed by unknown authority (possibly because of "x509: ECDSA verification failure" while trying to verify candidate authority ce
rtificate "tlsca.example.com")
Channel creation failed 
```
Make sure you have correct versions of Fabric binaries, and they are available in your PATH. You can check for correct version by running:

```
peer version
```

You should get this result:

```
peer:
 Version: 2.3.1
 Commit SHA: 2f69b4222
 Go version: go1.14.12
 OS/Arch: linux/amd64
 Chaincode:
  Base Docker Label: org.hyperledger.fabric
  Docker Namespace: hyperledger
```