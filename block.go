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

	for {
		hash := block.CalculateBlockHash(nonce)
		hashVerified := VerifyBlockHashDifficulty(hash)

		if hashVerified {
			block.Hash = hash
			block.Nonce = nonce

			fmt.Printf("Hash: %x, block found\n", hash)
			break
		} else {
			nonce++
		}

	}
}

func VerifyBlockHashDifficulty(hash []byte) bool {
	hashHex := fmt.Sprintf("%x", hash)
	return strings.HasPrefix(hashHex, strings.Repeat("0", difficulty))
}
