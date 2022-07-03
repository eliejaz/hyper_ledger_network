#!/bin/bash
#
# SPDX-License-Identifier: Apache-2.0




# default to using Hosp1
ORG=${1:-Hosp1}

# Exit on first error, print all commands.
set -e
set -o pipefail

# Where am I?
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

ORDERER_CA=${DIR}/test-network/organizations/ordererOrganizations/eurisko/tlsca/tlsca.eurisko-cert.pem
PEER0_ORG1_CA=${DIR}/test-network/organizations/peerOrganizations/hosp1.arez/tlsca/tlsca.hosp1.arez-cert.pem
PEER0_ORG2_CA=${DIR}/test-network/organizations/peerOrganizations/hosp2.rizk/tlsca/tlsca.hosp2.rizk-cert.pem
PEER0_ORG3_CA=${DIR}/test-network/organizations/peerOrganizations/org3.example.com/tlsca/tlsca.org3.example.com-cert.pem


if [[ ${ORG,,} == "hosp1" || ${ORG,,} == "digibank" ]]; then

   CORE_PEER_LOCALMSPID=Hosp1MSP
   CORE_PEER_MSPCONFIGPATH=${DIR}/test-network/organizations/peerOrganizations/hosp1.arez/users/Admin@hosp1.arez/msp
   CORE_PEER_ADDRESS=localhost:7051
   CORE_PEER_TLS_ROOTCERT_FILE=${DIR}/test-network/organizations/peerOrganizations/hosp1.arez/tlsca/tlsca.hosp1.arez-cert.pem

elif [[ ${ORG,,} == "hosp2" || ${ORG,,} == "magnetocorp" ]]; then

   CORE_PEER_LOCALMSPID=Hosp2MSP
   CORE_PEER_MSPCONFIGPATH=${DIR}/test-network/organizations/peerOrganizations/hosp2.rizk/users/Admin@hosp2.rizk/msp
   CORE_PEER_ADDRESS=localhost:9051
   CORE_PEER_TLS_ROOTCERT_FILE=${DIR}/test-network/organizations/peerOrganizations/hosp2.rizk/tlsca/tlsca.hosp1.arez-cert.pem

else
   echo "Unknown \"$ORG\", please choose Hosp1/Digibank or Hosp2/Magnetocorp"
   echo "For example to get the environment variables to set upa Hosp2 shell environment run:  ./setOrgEnv.sh Hosp2"
   echo
   echo "This can be automated to set them as well with:"
   echo
   echo 'export $(./setOrgEnv.sh Hosp2 | xargs)'
   exit 1
fi

# output the variables that need to be set
echo "CORE_PEER_TLS_ENABLED=true"
echo "ORDERER_CA=${ORDERER_CA}"
echo "PEER0_ORG1_CA=${PEER0_ORG1_CA}"
echo "PEER0_ORG2_CA=${PEER0_ORG2_CA}"
echo "PEER0_ORG3_CA=${PEER0_ORG3_CA}"

echo "CORE_PEER_MSPCONFIGPATH=${CORE_PEER_MSPCONFIGPATH}"
echo "CORE_PEER_ADDRESS=${CORE_PEER_ADDRESS}"
echo "CORE_PEER_TLS_ROOTCERT_FILE=${CORE_PEER_TLS_ROOTCERT_FILE}"

echo "CORE_PEER_LOCALMSPID=${CORE_PEER_LOCALMSPID}"
