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
	"encoding/json"
	"errors"
	"fmt"
	"strings"


	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type LoginInfo struct {
	Username	string	`json:"userName"`
	Password	string	`json:"password"`
}

type User struct {
	Username 		string `json:"userName"`
	Role		 		string `json:"role"`
	Password		string `json:"password"`
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
	fmt.Println("Init firing. Function will be ignored: " + function)

	// Initialize the collection of commercial paper keys
	fmt.Println("Initializing user accounts")
	t.createUser(stub, []string{"importerBank", "importerBank", "Importer Bank"})

	fmt.Println("Initialization complete")
	return nil, nil
}

func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

    // Handle different functions
    if function == "read" {                            //read a variable
        return t.read(stub, args)
    } else if function == "login" {
			return t.Login(stub, args)
		}
    fmt.Println("query did not find func: " + function)

    return nil, errors.New("Received unknown function query: " + function)
}

func (t *SimpleChaincode) Login(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Inside Login method")
	var username, password, jsonResp string
	var err error

    if len(args) < 2 {
        return nil, errors.New("Incorrect number of arguments. Expecting username and password")
    }
    username = args[0]
		password = args[1]

    valAsbytes, err := stub.GetState(username)
		if err != nil {
        jsonResp = "{\"Error\":\"Username not found " + username + "\"}"
        return nil, errors.New(jsonResp)
    }
		var existingUser User
		json.Unmarshal(valAsbytes, &existingUser)

		if existingUser.Password != password {
			jsonResp = "{\"Error\":\"Password does not match for " + username + "\"}"
			return nil, errors.New(jsonResp)
		}
		fmt.Println("Login complete")
    return valAsbytes, nil
}

func (t *SimpleChaincode) createUser(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	fmt.Println("Creating user")

	// Obtain the username to associate with the account
	if len(args) != 3 {
		fmt.Println("Error obtaining username/password/role")
		return nil, errors.New("createUser accepts 3 arguments (username, password, role)")
	}
	username := args[0]
	password := args[1]
	role := args[2]

	// Build an user object for the user
	var user = User{Username: username, Password: password, Role: role}
	userBytes, err := json.Marshal(&user)
	if err != nil {
		fmt.Println("error creating user" + user.Username)
		return nil, errors.New("Error creating user " + user.Username)
	}

	fmt.Println("Attempting to get state of any existing account for " + user.Username)
	existingBytes, err := stub.GetState(user.Username)
	if err == nil {

		var existingUser User
		err = json.Unmarshal(existingBytes, &existingUser)
		if err != nil {
			fmt.Println("Error unmarshalling user " + user.Username + "\n--->: " + err.Error())

			if strings.Contains(err.Error(), "unexpected end") {
				fmt.Println("No data means existing user found for " + user.Username + ", initializing user.")
				err = stub.PutState(user.Username, userBytes)

				if err == nil {
					fmt.Println("created user" + user.Username)
					return nil, nil
				} else {
					fmt.Println("failed to create initialize user for " + user.Username)
					return nil, errors.New("failed to initialize an account for " + user.Username + " => " + err.Error())
				}
			} else {
				return nil, errors.New("Error unmarshalling existing account " + user.Username)
			}
		} else {
			fmt.Println("Account already exists for " + user.Username + " " + existingUser.Username)
			return nil, errors.New("Can't reinitialize existing user " + user.Username)
		}
	} else {

		fmt.Println("No existing user found for " + user.Username + ", initializing user.")
		err = stub.PutState(user.Username, userBytes)

		if err == nil {
			fmt.Println("created user" + user.Username)
			return nil, nil
		} else {
			fmt.Println("failed to create initialize user for " + user.Username)
			return nil, errors.New("failed to initialize an user for " + user.Username + " => " + err.Error())
		}

	}
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "init" {//initialize the chaincode state, used as reset
		return t.Init(stub, "init", args)
	} else if function == "write" {
    return t.write(stub, args)
  } else if function == "createUser" {
		return t.createUser(stub, args)
	} else if function == "read" {
		return t.read(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation: " + function)
}

func (t *SimpleChaincode) write(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

	key = args[0]                            //rename for fun
	value = args[1]
	err = stub.PutState(key, []byte(value))  //write the variable into the chaincode state
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

    return valAsbytes, nil
}
