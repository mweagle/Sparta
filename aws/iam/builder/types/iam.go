package types

import "encoding/json"

// IAMPolicyDocument represents an IAM policy document
type IAMPolicyDocument struct {
	Version   string `json:",omitempty"`
	Statement []IAMPolicyStatement
}

// Avoid infinite loops when we just want to unmarshal the struct normally.
type iamPolicyDocumentCopy IAMPolicyDocument

// iamPolicyDocumentSingleStatement is used for parsing policy documents with a
// single statement.
type iamPolicyDocumentSingleStatement struct {
	Version   string `json:",omitempty"`
	Statement IAMPolicyStatement
}

// ToJSON returns the JSON representation of the policy document or
// panics if the object cannot be marshaled.
func (i IAMPolicyDocument) ToJSON() string {
	buf, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

// UnmarshalJSON sets the object from the provided JSON representation. This has
// been added to handle the special case of a single statement versus an array.
func (i *IAMPolicyDocument) UnmarshalJSON(data []byte) error {
	// Handle single statement policy documents
	var v iamPolicyDocumentSingleStatement
	err := json.Unmarshal(data, &v)
	if err == nil {
		i.Version = v.Version
		i.Statement = []IAMPolicyStatement{v.Statement}
		return nil
	}

	// Handle multiple statements
	var v2 iamPolicyDocumentCopy
	err = json.Unmarshal(data, &v2)
	if err != nil {
		return err
	}

	i.Version = v2.Version
	i.Statement = v2.Statement

	return nil
}

// IAMPrincipal represents a principal in an IAM policy
type IAMPrincipal struct {
	AWS           []string `json:",omitempty"`
	CanonicalUser []string `json:",omitempty"`
	Federated     []string `json:",omitempty"`
	Service       []string `json:",omitempty"`
}

// IAMPolicyStatement represents an IAM policy statement
type IAMPolicyStatement struct {
	Sid          string        `json:",omitempty"`
	Effect       string        `json:",omitempty"`
	Principal    *IAMPrincipal `json:",omitempty"`
	NotPrincipal *IAMPrincipal `json:",omitempty"`
	Action       []string      `json:",omitempty"`
	NotAction    []string      `json:",omitempty"`
	Resource     []string      `json:",omitempty"`
	Condition    interface{}   `json:",omitempty"`
}

// Avoid infinite loops when we just want to marshal the struct normally.
type iamPrincipalCopy IAMPrincipal

// MarshalJSON returns a JSON representation of the object. This has been added
// to handle the special case of "*" as the Principal value.
func (i IAMPrincipal) MarshalJSON() ([]byte, error) {
	// Special case for "*"
	if i.AWS != nil && len(i.AWS) == 1 && i.AWS[0] == "*" {
		return json.Marshal(i.AWS[0])
	}

	c := iamPrincipalCopy(i)

	return json.Marshal(c)
}

// UnmarshalJSON sets the object from the provided JSON representation. This has
// been added to handle the special case of "*" as the Principal value.
func (i *IAMPrincipal) UnmarshalJSON(data []byte) error {
	// Handle single string values like "*"
	var v string
	err := json.Unmarshal(data, &v)
	if err == nil {
		i.AWS = []string{v}
		i.CanonicalUser = nil
		i.Federated = nil
		i.Service = nil
		return nil
	}

	// Handle all other values
	var v2 iamPrincipalCopy
	err = json.Unmarshal(data, &v2)
	if err != nil {
		return err
	}

	i.AWS = v2.AWS
	i.CanonicalUser = v2.CanonicalUser
	i.Federated = v2.Federated
	i.Service = v2.Service

	return nil
}
