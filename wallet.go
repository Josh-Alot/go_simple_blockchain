package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
}

func CreateWallet() Wallet {
	var err error
	wallet := Wallet{}

	wallet.PrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	return wallet
}

func (wallet *Wallet) GetAddress() []byte {
	publicKey := wallet.PrivateKey.PublicKey

	pubKeyBytes, err := publicKey.Bytes()
	if err != nil {
		log.Fatal(err)
	}

	hash := sha256.Sum256(pubKeyBytes)

	return hash[:]
}
