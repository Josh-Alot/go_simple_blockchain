package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type TxInput struct {
	TransactionHash []byte
	TxOutputIndex   int
	Signature       []byte
	PublicKey       []byte
}

type TxOutput struct {
	Amount int
	Owner  []byte
}

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func CoinbaseTx(owner []byte) *Transaction {
	tx := Transaction{Outputs: []TxOutput{{Amount: reward, Owner: owner}}}
	tx.ID = tx.CalculateTxID()

	return &tx
}

func NewTransaction(chain *Blockchain, from []byte, to []byte, amount int) (*Transaction, error) {
	spentOutputs := chain.findSpentOutputs()
	inputs, accumulatedAmount := chain.findSpendableInputs(from, amount, spentOutputs)

	if accumulatedAmount < amount {
		return nil, errors.New("not enough cash to spend on this transaction!")
	}

	var outputs []TxOutput
	payment := TxOutput{Amount: amount, Owner: to}
	outputs = append(outputs, payment)

	if accumulatedAmount > amount {
		change := TxOutput{Amount: accumulatedAmount - amount, Owner: from}
		outputs = append(outputs, change)
	}

	transaction := Transaction{Inputs: inputs, Outputs: outputs}
	transaction.ID = transaction.CalculateTxID()
	return &transaction, nil
}

func (transaction *Transaction) CalculateTxID() []byte {
	tx := struct {
		Inputs  []TxInput
		Outputs []TxOutput
	}{transaction.Inputs, transaction.Outputs}

	serial, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}

	id := sha256.Sum256(serial)
	return id[:]
}
func (transaction *Transaction) Sign(privKey *ecdsa.PrivateKey) {
	if len(transaction.Inputs) == 0 {
		return
	}

	// the SignASN1 does not use the reader, however, the function requires it
	// this is the only reason that i'm sending the reader
	// in future Go versions, i'll remove it if becomes deprecated
	signature, err := ecdsa.SignASN1(rand.Reader, privKey, transaction.ID)
	if err != nil {
		log.Fatal(err)
	}

	pubKeyBytes, err := privKey.PublicKey.Bytes()
	if err != nil {
		log.Fatal(err)
	}

	for i := range transaction.Inputs {
		transaction.Inputs[i].Signature = signature
		transaction.Inputs[i].PublicKey = pubKeyBytes
	}
}

func (transaction *Transaction) VerifySignature() bool {
	if len(transaction.Inputs) == 0 {
		return true
	}

	for _, input := range transaction.Inputs {
		pubKey, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), input.PublicKey)
		if err != nil {
			fmt.Printf("Failed to verify signature: %s\n", err)
			return false
		}

		validSignature := ecdsa.VerifyASN1(pubKey, transaction.ID, input.Signature)
		if !validSignature {
			return false
		}
	}
	return true
}

func (transaction *Transaction) VerifyTxOwnership(chain *Blockchain) bool {
	if len(transaction.Inputs) == 0 {
		return true
	}

	for _, input := range transaction.Inputs {
		tx, err := chain.findTransaction(input.TransactionHash)
		if err != nil {
			fmt.Printf("%v\n", err)
			return false
		}

		output := tx.Outputs[input.TxOutputIndex]
		hash := sha256.Sum256(input.PublicKey)

		if !bytes.Equal(hash[:], output.Owner) {
			return false
		}

	}

	return true
}

func (transaction *Transaction) ValidateTransaction(chain *Blockchain) bool {
	return transaction.VerifySignature() && transaction.VerifyTxOwnership(chain)
}
