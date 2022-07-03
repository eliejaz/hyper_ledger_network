package chaincode

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type Prescription struct {
	ID             string `json:"id"`
	MedicationName string `json:"medicationName"`
	Quantity       string `json:"quantity"`
}

type Illness struct {
	ID string `json:"id"`
}

type PatientPrivateDetails struct {
	Bill         float32 `json:"bill"`
	PatientID    string  `json:"patientID"`
	ContractType string  `json:"contractType"`
}

// PaginatedQueryResult structure used for returning paginated query results and metadata
type PaginatedQueryResult struct {
	Records             []*Patient `json:"records,omitempty" metadata:"records,optional" `
	FetchedRecordsCount int32      `json:"fetchedRecordsCount"`
	Bookmark            string     `json:"bookmark"`
}

type Patient struct {
	ID               string      `json:"id"`
	FirstName        string      `json:"firstName"`
	LastName         string      `json:"last_name"`
	Email            string      `json:"email"`
	Description      string      `json:"description"`
	GroupType        string      `json:"groupType"`
	Allergies        []string    `json:"allergies,omitempty" metadata:"allergies,optional"`
	EmergencyContact string      `json:"emergencyContact"`
	Diagnosis        []Diagnosis `json:"diagnosis,omitempty" metadata:"diagnosis,optional" `
	DoctorsID        []string    `json:"doctorsId,omitempty" metadata:"doctorsId,optional"`
}

type Diagnosis struct {
	ID           string         `json:"id"`
	Description  string         `json:"description"`
	Illness      string         `json:"illness"`
	Prescription []Prescription `json:"prescriptions,omitempty" metadata:"prescriptions,optional"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Patient{
		{
			ID:               "1",
			FirstName:        "John",
			LastName:         "Doe",
			Email:            "j@gmail.com,",
			Description:      "",
			GroupType:        "",
			Allergies:        []string{},
			EmergencyContact: "",
			Diagnosis:        []Diagnosis{},
			DoctorsID:        []string{},
		},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

//TODO to be deprecated by next version
func (s *SmartContract) GetAllPatients(ctx contractapi.TransactionContextInterface) ([]*Patient, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Patient
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Patient
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// //----------------------------------------------ADMIN FUNCTIONS ----------------------------------------------//
func (s *SmartContract) CreatePatient(ctx contractapi.TransactionContextInterface, patient Patient) error {

	// transientMap, err := ctx.GetStub().GetTransient()
	// if err != nil {
	// 	return fmt.Errorf("error getting transient: %v", err)
	// }

	// type PatientAllDetails struct {
	// 	Bill         float32 `json:"bill"`
	// 	ContractType string  `json:"contractType"`
	// }

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	// transientAssetJSON, ok := transientMap["asset_properties"]
	// if !ok {
	// 	//log error to stdout
	// 	return fmt.Errorf("asset not found in the transient map input")
	// }

	// var patientBill PatientAllDetails
	// err = json.Unmarshal(transientAssetJSON, &patientBill)
	// if err != nil {
	// 	return fmt.Errorf("failed to unmarshal JSON: %v", err)
	// }

	if patient.ID == "" {
		return fmt.Errorf("the patient ID cannot be empty")
	}
	exists, err := s.PatientExists(ctx, patient.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the patient %s already exists", patient.ID)
	}
	if patient.FirstName == "" {
		return fmt.Errorf("the patient FirstName cannot be empty")
	}
	if patient.LastName == "" {
		return fmt.Errorf("the patient LastName cannot be empty")
	}
	if patient.Email == "" {
		return fmt.Errorf("the patient Email cannot be empty")
	}
	if patient.Description == "" {
		return fmt.Errorf("the patient Description cannot be empty")
	}
	if patient.GroupType == "" {
		return fmt.Errorf("the patient GroupType cannot be empty")
	}
	if patient.EmergencyContact == "" {
		return fmt.Errorf("the patient EmergencyContact cannot be empty")
	}

	// Marshal patient public info to JSON
	patientJSON, err := json.Marshal(patient)
	if err != nil {
		return err
	}

	// Persist patient to world state.
	err = ctx.GetStub().PutState(patient.ID, patientJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	// patientPrivateDetails := PatientPrivateDetails{
	// 	PatientID:    patient.ID,
	// 	Bill:         patientBill.Bill,
	// 	ContractType: patientBill.ContractType,
	// }

	// Marshal patient private info to JSON
	// patientJSON, err = json.Marshal(patientPrivateDetails)
	// if err != nil {
	// 	return err
	// }

	// orgCollection, err := s.getCollectionName(ctx)
	// if err != nil {
	// 	return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	// }

	// // Persist patient private info to world state.
	// err = ctx.GetStub().PutPrivateData(orgCollection, patient.ID, patientJSON)
	// return err
	return nil
}

func (s *SmartContract) DeletePatient(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.PatientExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the patient %s does not exist", id)
	}

	patientJson, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to get patient %s from world state. %v", id, err)
	}

	var patient Patient
	err = json.Unmarshal(patientJson, &patient)

	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	error := s.ValidateDoctorsID(ctx, patient.DoctorsID)

	orgCollection, err := s.getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	patientBill, err := ctx.GetStub().GetPrivateData(orgCollection, id)
	if error != nil {
		return fmt.Errorf("%v", error)
	}

	if patientBill == nil {
		return fmt.Errorf("the patient %s does not exist", id)
	}

	err = ctx.GetStub().DelState(id)
	if err != nil {
		return fmt.Errorf("failed to delete from world state. %v", err)
	}

	// Delete patient from world state.
	err = ctx.GetStub().DelPrivateData(orgCollection, id)
	if err != nil {
		return fmt.Errorf("failed to delete from private state. %v", err)
	}
	return nil
}

// GetAssetsByRangeWithPagination performs a range query based on the start and end key,
// page size and a bookmark.
// The number of fetched records will be equal to or lesser than the page size.
// Paginated range queries are only valid for read only transactions.
// Example: Pagination with Range Query
func (*SmartContract) GetPatientsByRangeWithPagination(ctx contractapi.TransactionContextInterface, pageSize int, bookmark string) (*PaginatedQueryResult, error) {

	resultsIterator, responseMetadata, err := ctx.GetStub().GetStateByRangeWithPagination("", " ", int32(pageSize), bookmark)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	assets, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	return &PaginatedQueryResult{
		Records:             assets,
		FetchedRecordsCount: responseMetadata.FetchedRecordsCount,
		Bookmark:            responseMetadata.Bookmark,
	}, nil
}

// //----------------------------------------------Doctor FUNCTIONS ----------------------------------------------//
func (s *SmartContract) ReadPatientById(ctx contractapi.TransactionContextInterface, id string) (*Patient, error) {
	patientJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if patientJSON == nil {
		return nil, fmt.Errorf("the patient %s does not exist", id)
	}

	var patient Patient
	err = json.Unmarshal(patientJSON, &patient)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize patient" + id)
	}

	error := s.ValidateDoctorsID(ctx, patient.DoctorsID)

	if error != nil {
		return nil, fmt.Errorf("%v", error)
	}

	return &patient, nil
}

func (s *SmartContract) CreateDiagnosis(ctx contractapi.TransactionContextInterface,
	diagnosisJson string, patientId string) (*Patient, error) {

	var diagnosis Diagnosis
	err := json.Unmarshal([]byte(diagnosisJson), &diagnosis)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal diagnosis JSON: %v", err)
	}

	patient, err := s.ReadPatientById(ctx, patientId)
	if err != nil {
		return nil, err
	}

	patient.Diagnosis = append(patient.Diagnosis, diagnosis)

	patientJSON, err := json.Marshal(patient)

	if err != nil {
		return nil, fmt.Errorf("Failed to serialize patient")
	}

	ctx.GetStub().PutState(patientId, patientJSON)

	return patient, nil
}

func (s *SmartContract) UpdateDiagnosis(ctx contractapi.TransactionContextInterface,
	patientId string,
	diagnosisJson string) error {

	var diagnosis Diagnosis
	err := json.Unmarshal([]byte(diagnosisJson), &diagnosis)
	if err != nil {
		return fmt.Errorf("failed to unmarshal diagnosis JSON: %v", err)
	}

	patient, err := s.ReadPatientById(ctx, patientId)
	if err != nil {
		return fmt.Errorf("failed to get patient: %v", err)
	}

	for i, d := range patient.Diagnosis {
		if d.ID == diagnosis.ID {
			patient.Diagnosis[i].Description = diagnosis.Description
			patient.Diagnosis[i].Illness = diagnosis.Illness
			patient.Diagnosis[i].Prescription = diagnosis.Prescription
			break
		}
	}

	patientJSON, err := json.Marshal(patient)
	if err != nil {
		return fmt.Errorf("failed to serialize patient")
	}

	ctx.GetStub().PutState(patientId, patientJSON)

	return nil
}

// //----------------------------------------------COMMON DOCTORS&ADMINS FUNCTIONs ----------------------------------------------//
func (s *SmartContract) UpdatePatient(ctx contractapi.TransactionContextInterface,
	patientJson string) error {

	var patient Patient
	err := json.Unmarshal([]byte(patientJson), &patient)
	if err != nil {
		return fmt.Errorf("failed to unmarshal patient JSON: %v", err)
	}

	exists, err := s.PatientExists(ctx, patient.ID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the patient %s does not exist", patient.ID)
	}

	patientJSON, err := json.Marshal(patient)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(patient.ID, patientJSON)
}

func (s *SmartContract) ReadPatientsByDoctorsID(ctx contractapi.TransactionContextInterface, doctorId string) ([]*Patient, error) {
	queryString := fmt.Sprintf(`{"selector": {"doctorsID": {"$elemMatch": {"$eq": "%s"}}}}`, doctorId)
	return getQueryResultForQueryString(ctx, queryString)
}

// //----------------------------------------------NURSE&&SECRETARY FUNCTIONS ----------------------------------------------//
func (s *SmartContract) ReadPatientPerscriptions(ctx contractapi.TransactionContextInterface, patientId string) ([]Prescription, error) {
	patient, err := s.ReadPatientById(ctx, patientId)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get patient: %v", err)
	}

	prescription := make([]Prescription, 0)

	for _, d := range patient.Diagnosis {
		prescription = append(prescription, d.Prescription...)
	}
	return prescription, nil
}

func (s *SmartContract) ReadPatientBill(ctx contractapi.TransactionContextInterface, patientId string) (PatientPrivateDetails, error) {
	collectionName, err := s.getCollectionName(ctx)
	if err != nil {
		return PatientPrivateDetails{}, err
	}

	patient, err := s.ReadPatientById(ctx, patientId)

	if err != nil {
		return PatientPrivateDetails{}, err
	}

	error := s.ValidateDoctorsID(ctx, patient.DoctorsID)

	if error != nil {
		return PatientPrivateDetails{}, fmt.Errorf("%v", error)
	}

	billAsBytes, error := ctx.GetStub().GetPrivateData(collectionName, patientId)
	if error != nil {
		return PatientPrivateDetails{}, fmt.Errorf("failed to get patient: %v", error)
	}
	if billAsBytes == nil {
		return PatientPrivateDetails{}, fmt.Errorf("the patient %s does not exist", patientId)
	}

	// unmarshal the bill
	var bill PatientPrivateDetails
	err = json.Unmarshal(billAsBytes, &bill)
	if err != nil {
		return PatientPrivateDetails{}, fmt.Errorf("failed to deserialize patient: %v", err)
	}

	return bill, nil
}

func (s *SmartContract) UpdatePatientBill(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		//log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	var bill PatientPrivateDetails
	err = json.Unmarshal(transientAssetJSON, &bill)
	if err != nil {
		return fmt.Errorf("failed to unmarshal asset: %v", err)
	}

	patient, err := s.ReadPatientById(ctx, bill.PatientID)
	if err != nil {
		return fmt.Errorf("failed to get patient: %v", err)
	}

	error := s.ValidateDoctorsID(ctx, patient.DoctorsID)

	if error != nil {
		return fmt.Errorf("%v", error)
	}

	collectionName, err := s.getCollectionName(ctx)
	if err != nil {
		return err
	}

	billAsBytes, err := json.Marshal(bill)
	if err != nil {
		return fmt.Errorf("failed to serialize bill: %v", err)
	}

	return ctx.GetStub().PutPrivateData(collectionName, bill.PatientID, billAsBytes)
}

// //----------------------------------------------(*************___*************)----------------------------------------------\\
// func (s *SmartContract) TransferPatient(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
// 	patient, err := s.ReadPatientById(ctx, id)
// 	if err != nil {
// 		return err
// 	}

// 	patient.Owner = newOwner
// 	patientJSON, err := json.Marshal(patient)
// 	if err != nil {
// 		return err
// 	}

// 	return ctx.GetStub().PutState(id, patientJSON)
// }

// //----------------------------------------------(*************_u_t_i_l_*************)----------------------------------------------\\
func (s *SmartContract) GetOrganization(ctx contractapi.TransactionContextInterface) (string, error) {
	return ctx.GetClientIdentity().GetMSPID()
}

func (s *SmartContract) getDoctorID(ctx contractapi.TransactionContextInterface) (value string, err error) {
	value, found, err := ctx.GetClientIdentity().GetAttributeValue("doctorId")

	if err != nil {
		return "", fmt.Errorf("failed to get doctorId: %v", err)
	}

	if !found || value == "" {
		return "", fmt.Errorf("doctorId not found")
	}

	return value, nil
}

func (s *SmartContract) ValidateDoctorsID(ctx contractapi.TransactionContextInterface, doctorsID []string) (err error) {
	// value, found, err := ctx.GetClientIdentity().GetAttributeValue("doctorId")

	// if err != nil {
	// 	return fmt.Errorf("failed to get doctorId: %v", err)
	// }

	// if !found || value == "" {
	// 	return fmt.Errorf("doctorId not found")
	// }

	// if !s.contains(doctorsID, value) {
	// 	return fmt.Errorf("the doctor %s does not belong to the group", value)
	// }

	return nil
}

func (s *SmartContract) PatientExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	PatientJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return PatientJSON != nil, nil
}

func (s *SmartContract) getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	// Create the collection name
	orgCollection := clientMSPID + "PrivateCollection"

	return orgCollection, nil
}

func (s *SmartContract) contains(list []string, elem string) bool {
	for _, e := range list {
		if e == elem {
			return true
		}
	}
	return false
}

// getQueryResultForQueryString executes the passed in query string.
// The result set is built and returned as a byte array containing the JSON results.
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Patient, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Patient, error) {
	var patients []*Patient
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var patient Patient
		err = json.Unmarshal(queryResult.Value, &patient)
		if err != nil {
			return nil, err
		}
		patients = append(patients, &patient)
	}

	return patients, nil
}

// //----------------------------------------------(*************_QUERY_HISTORY_*************)----------------------------------------------\\
// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *Patient  `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

func (s *SmartContract) GetPatientHistory(ctx contractapi.TransactionContextInterface, patientID string) ([]HistoryQueryResult, error) {
	s.ReadPatientById(ctx, patientID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(patientID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var patient Patient
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &patient)
			if err != nil {
				return nil, err
			}
		} else {
			patient = Patient{
				ID: patientID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &patient,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}
