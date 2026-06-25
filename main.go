package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

type Block struct {
	Data         []byte
	Hash         []byte
	PreviousHash []byte
	Nonce        int
}

type Blockchain struct {
	Blocks []*Block
}

const difficulty = 3

func (block *Block) CalculateHash(nonce int) []byte {
	blockData := bytes.Join([][]byte{block.PreviousHash, block.Data, []byte(strconv.Itoa(nonce))}, []byte{})
	hash := sha256.Sum256(blockData)

	return hash[:]
}

func (block *Block) Mine() {
	nonce := 0
	goal := strings.Repeat("0", difficulty)

	for {
		hash := block.CalculateHash(nonce)
		hashHex := fmt.Sprintf("%x", hash)

		if strings.HasPrefix(hashHex, goal) {
			block.Hash = hash
			block.Nonce = nonce
			break
		} else {
			nonce++
		}

	}
}

func (chain *Blockchain) AddBlock(data string) {
	previousBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := Block{Data: []byte(data), PreviousHash: previousBlock.Hash}
	newBlock.Mine()

	chain.Blocks = append(chain.Blocks, &newBlock)
}

func main() {
	genesisBlock := Block{Data: []byte("Genesis block")}
	genesisBlock.Mine()

	chain := Blockchain{Blocks: []*Block{&genesisBlock}}
	chain.AddBlock("João pagou 10 para Maria")
	chain.AddBlock("Pedro pagou 15 para João")

	fmt.Println("**** Blockchain ****")
	for i, block := range chain.Blocks {
		fmt.Printf("Block id: %d\n", i)
		fmt.Printf("Block Hash: %x\n", block.Hash)
		fmt.Printf("Block data: %s\n", block.Data)
		fmt.Printf("Previous Block: %x\n\n", block.PreviousHash)
	}
}
