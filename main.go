package main

import (
	"encoding/json"
	"fmt"

	"github.com/satori/go.uuid"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// ObjectType defines single type od data stored in the ledger
const ObjectType = "DATA"

// ExampleCC implements a simple chaincode to share channel config updates
type ExampleCC struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *ExampleCC) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// one of the: 'put', 'update', 'list', 'read' methods
// defined by the function name passed to Invoke
func (t *ExampleCC) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	var result []byte
	var err error

	switch fn {
	case "put":
		result, err = put(stub, args)
	case "update":
		result, err = update(stub, args)
	case "list":
		result, err = list(stub, args)
	default:
		result, err = read(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(result)
}

// put creates entry in the ledger
func put(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	var key, id string
	logger := shim.NewLogger("putlogger")

	if len(args) <= 1 && len(args) >= 2 {
		logger.Errorf("FNPUT: Incorrect number of arguments %v", len(args))
		return nil, fmt.Errorf("Incorrect number of arguments %v", len(args))
	}

	if len(args) == 2 {
		id = args[1]
	} else {
		id = uuid.NewV4().String()
	}

	key, err = stub.CreateCompositeKey(ObjectType, []string{id})
	if err != nil {
		logger.Errorf("FNPUT: %v", err.Error())
		return nil, fmt.Errorf("Failed to create composite key, err: %v", err.Error())
	}

	if err = stub.PutState(key, []byte(args[0])); err != nil {
		logger.Errorf("FNPUT: %v", err.Error())
		return nil, fmt.Errorf("Failed to put data: %v, err: %v", args[0], err.Error())
	}

	return []byte(key), nil
}

// update updates entry in the ledger
func update(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error
	logger := shim.NewLogger("updatelogger")

	if len(args) != 2 {
		logger.Errorf("FNUPDATE: %v", "Incorrect arguments. Expecting key and data")
		return nil, fmt.Errorf("Incorrect arguments. Expecting key and data")
	}

	key, err := stub.CreateCompositeKey(ObjectType, []string{args[0]})
	if err != nil {
		logger.Errorf("FNUPDATE: %v", err.Error())
		return nil, fmt.Errorf("Failed to create composite key, err: %v", err.Error())
	}

	err = stub.PutState(key, []byte(args[1]))
	if err != nil {
		logger.Errorf("FNUPDATE: %v", err.Error())
		return nil, fmt.Errorf("Failed to update data: %v, err: %v", args[1], err.Error())
	}

	return []byte(key), nil
}

// read reads entry from the ledger
func read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger := shim.NewLogger("readlogger")
	if len(args) != 1 {
		logger.Errorf("FNREAD: %v", "Incorrect arguments. Expecting a key")
		return nil, fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	key, err := stub.CreateCompositeKey(ObjectType, []string{args[0]})
	if err != nil {
		logger.Errorf("FNREAD: %v", err.Error())
		return nil, fmt.Errorf("Failed to create composite key, err: %v", err.Error())
	}

	data, err := stub.GetState(key)
	if err != nil {
		logger.Errorf("FNREAD: %v", err.Error())
		return nil, fmt.Errorf("Failed to get data: %v, err: %v", args[0], err.Error())
	}

	if data == nil {
		logger.Errorf("FNREAD: %v", err.Error())
		return nil, fmt.Errorf("Data not found: %v, err: %v", args[0], err.Error())
	}

	return data, nil
}

// list returns collection of entries from the ledger
func list(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger := shim.NewLogger("listlogger")
	ret := [][]byte{}

	value, err := stub.GetStateByPartialCompositeKey(ObjectType, []string{})
	if err != nil {
		logger.Errorf("FNLIST: %v", err.Error())
		return nil, fmt.Errorf("Failed to list data, err: %v", err)
	}

	if value == nil {
		logger.Errorf("FNLIST: %v", err.Error())
		return nil, fmt.Errorf("No data found")
	}

	for value.HasNext() {
		kv, err := value.Next()
		if err != nil {
			logger.Errorf("FNLIST: %v", err.Error())
			return nil, fmt.Errorf("Failed to parse data, err: %v", err)
		}
		ret = append(ret, kv.GetValue())
	}

	result, err := json.Marshal(ret)
	if err != nil {
		logger.Errorf("FNLIST: %v", err.Error())
		return nil, fmt.Errorf("Failed marshal JSON: %v, err: %v", ret, err.Error())
	}

	return result, nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(ExampleCC)); err != nil {
		fmt.Printf("Error starting ExampleCC chaincode: %v", err)
	}
}
