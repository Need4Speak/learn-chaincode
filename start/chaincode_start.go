/*
Copyright IBM Corp 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"time"
	"encoding/json"
	"strings"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

// 病人信息结构体
type Patient struct {
   Name string
   Sex string
   Age int
   Illness string
   Records string
}

// 病人就诊记录结构体
type Record struct {
   Illness string
   recordTime string
}

// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	err := stub.PutState("hello_world", []byte(args[0]))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {													//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "addPatient" {
		return t.addPatient(stub, args)
	} else if function == "addRecord" {
		return t.addRecord(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {											//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleChaincode) addPatient(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
    var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0]
	value = args[1]
	err = stub.PutState(key, []byte(value))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// 根据病人id插入其某次看病的记录
func (t *SimpleChaincode) addRecord(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value, curTime string
    var err error
	fmt.Println("running addRecord()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0]
	curTime = time.Now().Format("200601020304")
	// newKey: 就诊记录的 key, 为string类型
	newKey := key + curTime

	// 根据key获取就诊病人的信息，在其Records中加入新的就诊记录后，再重新存储
	valAsbytes, err := stub.GetState(key)
    if err != nil {
        jsonResp := "{\"Error\":\"Failed to get state for " + key + "\"}"
        return nil, errors.New(jsonResp)
    }

	var patient Patient
	json.Unmarshal(valAsbytes, &patient)
	patient.Records = patient.Records + ", " + newKey
	b, _ := json.Marshal(&patient)
	err = stub.PutState(key, b)
	if err != nil {
		return nil, err
	}

	value = args[1]
	err = stub.PutState(newKey, []byte(value))
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
    var key, jsonResp string
    var err error

    if len(args) != 1 {
        return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
    }

    key = args[0]
    valAsbytes, err := stub.GetState(key)
    if err != nil {
        jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
        return nil, errors.New(jsonResp)
    }

	var patient Patient
	json.Unmarshal(valAsbytes, &patient)
	records := patient.Records
	recordArr := strings.Split(records, ", ")
	// records 字符串以" ,"开头，故需要去掉第一个元素
	recordArr = recordArr[1:len(recordArr)]

	var recordsContent string
	for _, ele := range recordArr {
		 valAsbytes, err := stub.GetState(ele)
		if err != nil {
			jsonResp = "{\"Error\":\"Failed to get state for " + ele + "\"}"
			return nil, errors.New(jsonResp)
		}
		recordsContent = recordsContent + "; \n 就诊时间：" + ele[len(ele)-12:len(ele)] + ", 病情：" + string(valAsbytes)
	}

    // return valAsbytes, nil
	return []byte(recordsContent+string(valAsbytes)), nil
}
