#!/bin/bash

ROOTDIR=$(cd "$(dirname "$0")" && pwd)
export PATH=${ROOTDIR}/../bin:${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}/../config/
export ORDERER_CA=${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer.eurisko/msp/tlscacerts/tlsca.eurisko-cert.pem

CHANNEL_NAME=mychannel
. ./scripts/utils.sh

mkdir -p orderer/orderer3

function orderer3-ca(){
    echo "turning orderer3-ca container on"
    docker-compose -f ./compose/compose-orderer3-ca.yaml up -d
}

function hosp1Init(){
    export CORE_PEER_TLS_ENABLED=true
    export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/hosp1.arez/users/Admin@hosp1.arez/msp
    export CORE_PEER_ADDRESS=localhost:7051
    export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/hosp1.arez/peers/peer0.hosp1.arez/tls/ca.crt
    export CORE_PEER_LOCALMSPID=Hosp1MSP
}

function tlsInit(){
    export ORDERER_ADMIN_TLS_SIGN_CERT=${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer3.eurisko/tls/server.crt
    export ORDERER_ADMIN_TLS_PRIVATE_KEY=${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer3.eurisko/tls/server.key
}

function ordererInit(){
    export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/ordererOrganizations/eurisko/users/Admin@eurisko/msp
    export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer.eurisko/msp/tlscacerts/tlsca.eurisko-cert.pem
    export CORE_PEER_LOCALMSPID=OrdererMSP    
}

function createOrderer3(){
    echo "running register enroll script"
    set -x
    . organizations/fabric-ca/orderer.sh 3
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi
}

function runOrderer3Container(){
    hosp1Init
    echo "turning orderer3 container on"
    set -x
    docker-compose -f ./compose/compose-orderer3.yaml up -d
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi
}


function invokeChannel(){
    hosp1Init
    echo "invoking channel"
    set -x
    peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.eurisko --tls --cafile "${ORDERER_CA}" -C mychannel -n basic --peerAddresses localhost:7051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/hosp1.arez/peers/peer0.hosp1.arez/tls/ca.crt" --peerAddresses localhost:9051 --tlsRootCertFiles "${PWD}/organizations/peerOrganizations/hosp2.rizk/peers/peer0.hosp2.rizk/tls/ca.crt" -c '{"function":"InitLedger","Args":[]}'
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi
}

function joinChannel(){
    tlsInit  
	# Poll in case the raft leader is not set yet
	local rc=1
	local COUNTER=1
	while [ $rc -ne 0 -a $COUNTER -lt 5 ] ; do
		set -x
		osnadmin channel join --channelID $CHANNEL_NAME --config-block ./channel-artifacts/${CHANNEL_NAME}.block -o localhost:9053 --ca-file "$ORDERER_CA" --client-cert "$ORDERER_ADMIN_TLS_SIGN_CERT" --client-key "$ORDERER_ADMIN_TLS_PRIVATE_KEY" >&log.txt
		res=$?
		{ set +x; } 2>/dev/null
		let rc=$res
		COUNTER=$(expr $COUNTER + 1)
	done
	cat log.txt
    rm log.txt
}

function fetchChannelConfig(){
    set -x
  peer channel fetch config orderer/orderer3/config_block.pb -o localhost:7050 -c $CHANNEL_NAME --tls --cafile "$ORDERER_CA"
  res=$?
  { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi

  echo "Decoding config block to JSON and isolating config to orderer/orderer3/config.json"
  set -x
  configtxlator proto_decode --input orderer/orderer3/config_block.pb --type common.Block | jq '.data.data[0].payload.data.config' > orderer/orderer3/config.json
  res=$?
  { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi
export TLS_FILE=${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer3.eurisko/tls/server.crt

echo "{\"client_tls_cert\": \"$(cat $TLS_FILE | base64 | awk 1 ORS='')\",
                    \"host\": \"orderer3.eurisko\",
                    \"port\": 7050,
                    \"server_tls_cert\": \"$(cat $TLS_FILE | base64 | awk 1 ORS='')\"}" > ./orderer/orderer3/orderer3.json

jq ".channel_group.groups.Orderer.values.ConsensusType.value.metadata.consenters += [$(cat orderer/orderer3/orderer3.json)]" orderer/orderer3/config.json > orderer/orderer3/modified_config.json
}

function encode(){
    set -x
    configtxlator proto_encode --input "orderer/orderer3/config.json" --type common.Config --output orderer/orderer3/config.pb
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi

    configtxlator proto_encode --input "orderer/orderer3/modified_config.json" --type common.Config --output orderer/orderer3/modified_config.pb
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi
}

function updateAndDecode(){
    set -x
    configtxlator compute_update --channel_id mychannel --original "orderer/orderer3/config.pb" --updated "orderer/orderer3/modified_config.pb" --output orderer/orderer3/config_update.pb
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi

    configtxlator proto_decode --input orderer/orderer3/config_update.pb --type common.ConfigUpdate --output orderer/orderer3/config_update.json
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi

    echo "{\"payload\":{\"header\":{\"channel_header\":{\"channel_id\":\"mychannel\",\"type\":2}},\"data\":{\"config_update\":"$(cat orderer/orderer3/config_update.json)"}}}" | jq . > orderer/orderer3/config_update_in_envelope.json

    configtxlator proto_encode --input "orderer/orderer3/config_update_in_envelope.json" --type common.Envelope --output orderer/orderer3/config_update_in_envelope.pb
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi
}

function updateChannel(){
    ordererInit
    set -x
    peer channel update -f orderer/orderer3/config_update_in_envelope.pb -c mychannel -o localhost:7050 --tls true --cafile $ORDERER_CA
    res=$?
    { set +x; } 2>/dev/null
    if [ "$res" != "0" ]; then
        fatalln "Failed to run script..."
        exit 1
    fi
}

function ordererDown(){
    ./network.sh down
    docker-compose -f ./compose/compose-orderer3.yaml down --volumes --remove-orphans    
}

# first run networkup and deploycc
if [ "$1" == "up" ]; then
    createOrderer3
    runOrderer3Container
    joinChannel
    fetchChannelConfig
    encode
    updateAndDecode
    updateChannel
elif [ "$1" == "down" ]; then
    ordererDown
else
    echo "choose man! choose"
    exit 1
fi

