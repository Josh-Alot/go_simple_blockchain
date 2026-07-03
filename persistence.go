package main

import (
	"encoding/gob"
	"os"
)

const chainFile = "blockchain.dat"

func LoadFromFile() (*Blockchain, error) {
	chain := Blockchain{}

	file, err := os.Open(chainFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := gob.NewDecoder(file).Decode(&chain); err != nil {
		return nil, err
	}

	return &chain, nil
}

func (chain *Blockchain) SaveToFile() error {
	file, err := os.Create(chainFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	return encoder.Encode(chain)
}
