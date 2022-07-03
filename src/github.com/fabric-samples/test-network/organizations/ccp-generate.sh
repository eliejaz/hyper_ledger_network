#!/bin/bash

function one_line_pem {
    echo "`awk 'NF {sub(/\\n/, ""); printf "%s\\\\\\\n",$0;}' $1`"
}

function json_ccp {
    local PP=$(one_line_pem $6)
    local CP=$(one_line_pem $7)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${P0PORT}/$2/" \
        -e "s/\${CAPORT}/$3/" \
        -e "s/\${HOSP}/$4/" \
        -e "s/\${P1PORT}/$5/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        organizations/ccp-template.json
}

function yaml_ccp {
    local PP=$(one_line_pem $6)
    local CP=$(one_line_pem $7)
    sed -e "s/\${ORG}/$1/" \
        -e "s/\${P0PORT}/$2/" \
        -e "s/\${CAPORT}/$3/" \
        -e "s/\${HOSP}/$4/" \
        -e "s/\${P1PORT}/$5/" \
        -e "s#\${PEERPEM}#$PP#" \
        -e "s#\${CAPEM}#$CP#" \
        organizations/ccp-template.yaml | sed -e $'s/\\\\n/\\\n          /g'
}

ORG=1
HOSP=arez
P0PORT=7051
P1PORT=10051
CAPORT=7054
PEERPEM=organizations/peerOrganizations/hosp1.arez/tlsca/tlsca.hosp1.arez-cert.pem
CAPEM=organizations/peerOrganizations/hosp1.arez/ca/ca.hosp1.arez-cert.pem

echo "$(json_ccp $ORG $P0PORT $CAPORT $HOSP $P1PORT $PEERPEM $CAPEM)" > organizations/peerOrganizations/hosp1.arez/connection-hosp1.json
echo "$(yaml_ccp $ORG $P0PORT $CAPORT $HOSP $P1PORT $PEERPEM $CAPEM)" > organizations/peerOrganizations/hosp1.arez/connection-hosp1.yaml

ORG=2
HOSP=rizk
P0PORT=9051
P1PORT=12051
CAPORT=8054
PEERPEM=organizations/peerOrganizations/hosp2.rizk/tlsca/tlsca.hosp2.rizk-cert.pem
CAPEM=organizations/peerOrganizations/hosp2.rizk/ca/ca.hosp2.rizk-cert.pem

echo "$(json_ccp $ORG $P0PORT $CAPORT $HOSP $P1PORT $PEERPEM $CAPEM)" > organizations/peerOrganizations/hosp2.rizk/connection-hosp2.json
echo "$(yaml_ccp $ORG $P0PORT $CAPORT $HOSP $P1PORT $PEERPEM $CAPEM)" > organizations/peerOrganizations/hosp2.rizk/connection-hosp2.yaml
