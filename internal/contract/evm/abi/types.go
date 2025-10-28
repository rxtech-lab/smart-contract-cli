package abi

import "encoding/json"

// StateMutability represents the state mutability of a function
type StateMutability string

const (
	StateMutabilityPure       StateMutability = "pure"
	StateMutabilityView       StateMutability = "view"
	StateMutabilityNonPayable StateMutability = "nonpayable"
	StateMutabilityPayable    StateMutability = "payable"
)

// ABIElement represents a single element in an ABI array
type ABIElement struct {
	Type            string     `json:"type"`
	Name            string     `json:"name,omitempty"`
	Inputs          []ABIParam `json:"inputs,omitempty"`
	Outputs         []ABIParam `json:"outputs,omitempty"`
	StateMutability string     `json:"stateMutability,omitempty"`
	Constant        bool       `json:"constant,omitempty"`
	Payable         bool       `json:"payable,omitempty"`
	Anonymous       bool       `json:"anonymous,omitempty"`
	Components      []ABIParam `json:"components,omitempty"`
	InternalType    string     `json:"internalType,omitempty"`
}

// ABIParam represents a parameter in a function or event
type ABIParam struct {
	Name         string     `json:"name,omitempty"`
	Type         string     `json:"type"`
	Indexed      bool       `json:"indexed,omitempty"`
	Components   []ABIParam `json:"components,omitempty"`
	InternalType string     `json:"internalType,omitempty"`
}

// ABIArray represents a standard ABI as an array of elements
type ABIArray []ABIElement

// ABIObject represents an object that contains an ABI field
type ABIObject struct {
	ABI      ABIArray       `json:"abi"`
	Bytecode string         `json:"bytecode,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ABI is a wrapper that can handle both ABI array and object formats
type ABI struct {
	elements ABIArray
}

// UnmarshalJSON implements custom unmarshaling to handle both formats
func (a *ABI) UnmarshalJSON(data []byte) error {
	// First, try to unmarshal as an array
	var arr ABIArray
	if err := json.Unmarshal(data, &arr); err == nil {
		a.elements = arr
		return nil
	}

	// If that fails, try to unmarshal as an object
	var obj ABIObject
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	a.elements = obj.ABI
	return nil
}

// MarshalJSON implements custom marshaling
func (a *ABI) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.elements)
}

// Elements returns the ABI elements
func (a *ABI) Elements() ABIArray {
	return a.elements
}

// SetElements sets the ABI elements
func (a *ABI) SetElements(elements ABIArray) {
	a.elements = elements
}

// GetStateMutability returns the state mutability as an enum
func (a *ABIElement) GetStateMutability() StateMutability {
	return StateMutability(a.StateMutability)
}

// IsReadOnly returns true if the function is view or pure
func (a *ABIElement) IsReadOnly() bool {
	sm := a.GetStateMutability()
	return sm == StateMutabilityPure || sm == StateMutabilityView
}

// IsWriteOperation returns true if the function modifies state
func (a *ABIElement) IsWriteOperation() bool {
	return !a.IsReadOnly()
}

func (a *ABIElement) IsWritable() bool {
	sm := a.GetStateMutability()
	return sm == StateMutabilityNonPayable || sm == StateMutabilityPayable
}

func (a *ABIElement) IsReadable() bool {
	sm := a.GetStateMutability()
	return sm == StateMutabilityView || sm == StateMutabilityPure
}

func (a *ABIElement) IsPayable() bool {
	return a.GetStateMutability() == StateMutabilityPayable
}
