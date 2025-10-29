package abi

import (
	"encoding/json"

	"github.com/rxtech-lab/smart-contract-cli/internal/errors"
)

type AbiArray []ABIElement

type AbiObject struct {
	Abi      AbiArray
	Bytecode string
	Metadata map[string]any
}

// ParseAbi parse an abi string which can be in array or object format or
// Abi object format and returns an AbiArray.
func ParseAbi(abi string) (AbiArray, error) {
	var abiArray AbiArray
	var abiObject AbiObject

	err := json.Unmarshal([]byte(abi), &abiArray)
	// check if error is not nil, try to unmarshal as an object
	if err != nil {
		err = json.Unmarshal([]byte(abi), &abiObject)
		if err != nil {
			return nil, errors.WrapABIError(err, errors.ErrCodeInvalidABIFormat, "failed to parse ABI: invalid JSON format")
		}
		return abiObject.Abi, nil
	}

	return abiArray, nil
}
