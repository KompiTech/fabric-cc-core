# Prerequisites

1. Follow steps on https://hyperledger-fabric.readthedocs.io/en/release-2.3/prereqs.html
2. Follow steps on https://hyperledger-fabric.readthedocs.io/en/release-2.3/install.html
3. Test that you have ```peer``` binary accessible in your PATH
4. Clone the fabric-samples repo to some location on your disk (```git clone https://github.com/hyperledger/fabric-samples.git```), ```cd``` to cloned repo and ```git checkout 51f76977b0ee102ea7efc17875f2694c42823777```   
5. Export ```FABRIC_SAMPLES``` env var and set it to ```fabric-samples``` repo location that you previously cloned. This is necessary, as we are re-using test-network there.
