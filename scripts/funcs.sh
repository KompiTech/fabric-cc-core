#!/bin/bash

function init_fabric_repo {
  if [[ -z "${FABRIC_SAMPLES}" ]]; then
    echo "FABRIC_SAMPLES env variable is not defined. Do \"export FABRIC_SAMPLES=<path-to-repo>\" and try again."
    exit 1
  fi

  # Define env necessary for peer binary to work, using crypto material from fabric's test-network
  local CC_NAME=local-cc
  local CHANNEL_NAME=mychannel

  export COMPOSE_PROJECT_NAME=${CC_NAME}
  export IMAGE_TAG=latest
  export SYS_CHANNEL=system-channel
  export FABRIC_CFG_PATH=${FABRIC_SAMPLES}/config/
  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID=Org1MSP
  export CORE_PEER_TLS_ROOTCERT_FILE=${FABRIC_SAMPLES}/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${FABRIC_SAMPLES}/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
  export CORE_PEER_ADDRESS=localhost:7051
  export LOCAL_CAFILE=${FABRIC_SAMPLES}/test-network/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  export LOCAL_TLSCERT_ORG1=${FABRIC_SAMPLES}/test-network/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
  export LOCAL_TLSCERT_ORG2=${FABRIC_SAMPLES}/test-network/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
  export LOCAL_CC_NAME=${CC_NAME}
  export LOCAL_CHANNEL_NAME=${CHANNEL_NAME}

  export GONOSUMDB="github.com/KompiTech/*,github.com/cucumber/godog" # this package has issues with go.sum checksum, TODO remove KompiTech once repos are public
  export GOPRIVATE="github.com/KompiTech/*" # TODO remove once public
}

function gen_collection_config {
  go run ../../cmd/collgen/*.go -registryDir ${PWD}/asset -templateFile ${PWD}/../../templates/collection_config_template.yaml > ${PWD}/collection_config.json
}

function gen_cc_init {
  # Generates chaincode init data to a file cc_init.json. It contains all currently available registry items and singletons except blacklisted ones.
  go run ../../cmd/initgen/*.go -registryDir ${PWD}/asset -singletonDir ${PWD}/singleton -singletonBlacklist org_roles,identity_public_key > ${PWD}/cc_init.json
}

function gen_cc_metainf {
  # Generates couchDB index metadata in src/META-INF subdir. They are necessary for both chaincode operation and tests also use them.
  rm -Rf ${PWD}/src/META-INF
  go run ../../cmd/metainfgen/*.go -registryDir ${PWD}/asset -outputDir ${PWD}/src/META-INF
}

function cc_init {
  local INITARG='{"Args":["Init", "'$(cat ${PWD}/cc_init.json | jq -c '.init_manager = "Admin@org1.example.com"' | sed 's/"/\\"/g')'"]}'
  peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${LOCAL_CAFILE} -C ${LOCAL_CHANNEL_NAME} -n ${LOCAL_CC_NAME} --peerAddresses localhost:7051 --tlsRootCertFiles ${LOCAL_TLSCERT_ORG1} --peerAddresses localhost:9051 --tlsRootCertFiles ${LOCAL_TLSCERT_ORG2} -c "${INITARG}"
}

function start_local {
  init_fabric_repo
  gen_collection_config
  gen_cc_metainf
  local THISPWD=${PWD}
	cd ${FABRIC_SAMPLES}/test-network
	./network.sh up -s couchdb
	./network.sh createChannel -s couchdb
	./network.sh deployCC -ccn local-cc -ccp ${THISPWD}/src/ -ccl go -cccg ${THISPWD}/collection_config.json
	cd -
	gen_cc_init
	cc_init
}

function stop_local {
  init_fabric_repo
  rm -rf ${PWD}/itsm_init.json ${PWD}/collection_config.json ${PWD}/vendor/ ${PWD}/src/META-INF/
  cd ${FABRIC_SAMPLES}/test-network
  ./network.sh down
  cd -
}

function start_rest {
  init_fabric_repo
  go run ${PWD}/../../cmd/micro-rest/*.go 10000 &
}
