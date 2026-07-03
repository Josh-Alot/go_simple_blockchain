package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

type Block struct {
	Transactions []*Transaction
	Hash         []byte
	PreviousHash []byte
	Nonce        int
}

func (block *Block) CalculateBlockHash(nonce int) []byte {
	transactions := block.HashTransactions()

	blockData := bytes.Join([][]byte{block.PreviousHash, transactions, []byte(strconv.Itoa(nonce))}, []byte{})
	hash := sha256.Sum256(blockData)

	return hash[:]
}

func (block *Block) HashTransactions() []byte {
	var txIDs [][]byte

	for _, tx := range block.Transactions {
		txIDs = append(txIDs, tx.ID)
	}

	return bytes.Join(txIDs, []byte{})
}

func (block *Block) Mine() {
	nonce := 0
	goal := strings.Repeat("0", difficulty)

	for {
		hash := block.CalculateBlockHash(nonce)
		hashHex := fmt.Sprintf("%x", hash)

		if strings.HasPrefix(hashHex, goal) {
			block.Hash = hash
			block.Nonce = nonce

			fmt.Printf("Hash: %s, block found\n", hashHex)
			break
		} else {
			// fmt.Printf("Hash: %s, block not found\n", hashHex)
			nonce++
		}

	}
}
