package data

import (
	"crypto/rsa"
	"encoding/json"
)

type Identity struct {
	Name    string
	Age     int32
	Address string
	Email   string
	Phone   string
}

type Merits struct {
	Skills     []string
	Education  []string
	Experience []string
}

type Submission struct {
	Nonce  int32
	Id     Identity
	Merit  Merits
	PubKey rsa.PublicKey
}

type Registration struct {
	CompanyName string
	PubKey      rsa.PublicKey
}

func NewIdentity(name string, age int32, address string, email string, phone string) *Identity {
	return &Identity{Name: name, Age: age, Address: address, Email: email, Phone: phone}
}

func NewMerits(skills []string, education []string, experience []string) *Merits {
	return &Merits{Skills: skills, Education: education, Experience: experience}
}

func DecodeSubmissionJson(jsonString []byte) (Submission, error) {
	var sub Submission
	err := json.Unmarshal(jsonString, &sub)
	if err != nil {
		return Submission{}, err
	}
	return sub, nil
}
