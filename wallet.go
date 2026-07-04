package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"log"
	"os"
)

const walletsFile = "wallets.dat"

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
}

type Wallets struct {
	Hashes map[string]*Wallet
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

func NewWallets() *Wallets {
	return &Wallets{Hashes: make(map[string]*Wallet)}
}

func LoadWalletsFromFile() (*Wallets, error) {
	wallets := Wallets{}

	file, err := os.Open(walletsFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := gob.NewDecoder(file).Decode(&wallets); err != nil {
		return nil, err
	}

	return &wallets, nil
}

func LoadOrCreateWallets() (*Wallets, error) {
	wallets, err := LoadWalletsFromFile()
	if os.IsNotExist(err) {
		return NewWallets(), nil
	}

	if err != nil {
		return nil, err
	}

	return wallets, nil
}

func (wallets *Wallets) AddWallet() string {
	wallet := CreateWallet()
	address := wallet.GetAddress()
	wallets.Hashes[hex.EncodeToString(address)] = &wallet

	return hex.EncodeToString(address)
}

func (wallets *Wallets) GetWallet(address string) (Wallet, error) {
	wallet, found := wallets.Hashes[address]
	if !found {
		return Wallet{}, errors.New("wallet not found")
	}

	return *wallet, nil
}

func (wallets *Wallets) SaveToFile() error {
	file, err := os.Create(walletsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(wallets)
}

func (wallet *Wallet) GobEncode() ([]byte, error) {
	return wallet.PrivateKey.Bytes()
}

func (wallet *Wallet) GobDecode(data []byte) error {
	privKey, err := ecdsa.ParseRawPrivateKey(elliptic.P256(), data)
	if err != nil {
		return err
	}

	wallet.PrivateKey = privKey
	return nil
}
