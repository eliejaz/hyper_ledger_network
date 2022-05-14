 export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/ordererOrganizations/eurisko

  ORDERER=$1
  mkdir -p "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/msp/"
  mkdir -p "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls/"

  echo "Registering orderer"
  set -x
  fabric-ca-client register --caname ca-orderer --id.name "orderer${ORDERER}" --id.secret "orderer${ORDERER}pw" --id.type orderer --tls.certfiles "${PWD}/organizations/fabric-ca/ordererOrg/tls-cert.pem"
  { set +x; } 2>/dev/null

  echo "Registering the orderer admin"
  set -x
  fabric-ca-client register --caname ca-orderer --id.name "orderer${ORDERER}Admin" --id.secret "orderer${ORDERER}Adminpw" --id.type admin --tls.certfiles "${PWD}/organizations/fabric-ca/ordererOrg/tls-cert.pem"
  { set +x; } 2>/dev/null

  echo "Generating the orderer msp"
  set -x
  fabric-ca-client enroll -u https://"orderer${ORDERER}":"orderer${ORDERER}pw"@localhost:9054 --caname ca-orderer -M "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/msp" --csr.hosts "orderer${ORDERER}.eurisko" --csr.hosts localhost --tls.certfiles "${PWD}/organizations/fabric-ca/ordererOrg/tls-cert.pem"
  { set +x; } 2>/dev/null

  cp "${PWD}/organizations/ordererOrganizations/eurisko/msp/config.yaml" "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/msp/config.yaml"

  echo "Generating the orderer-tls certificates"
  set -x
  fabric-ca-client enroll -u https://"orderer${ORDERER}":"orderer${ORDERER}pw"@localhost:9054 --caname ca-orderer -M "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls" --enrollment.profile tls --csr.hosts "orderer${ORDERER}.eurisko" --csr.hosts localhost --tls.certfiles "${PWD}/organizations/fabric-ca/ordererOrg/tls-cert.pem"
  { set +x; } 2>/dev/null

  cp "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls/tlscacerts/"* "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls/ca.crt"
  cp "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls/signcerts/"* "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls/server.crt"
  cp "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls/keystore/"* "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls/server.key"

  mkdir -p "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/msp/tlscacerts"
  cp "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/tls/tlscacerts/"* "${PWD}/organizations/ordererOrganizations/eurisko/orderers/orderer${ORDERER}.eurisko/msp/tlscacerts/tlsca.eurisko-cert.pem"