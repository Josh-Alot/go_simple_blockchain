package main

import (
	"crypto/sha256"
	"fmt"
)

type Block struct {
	Data         []byte
	Hash         []byte
	PreviousHash []byte
}

type Blockchain struct {
	Blocks []*Block
}

func (block *Block) CalculateHash() {
	hash := sha256.Sum256(append(block.PreviousHash, block.Data...))
	block.Hash = hash[:]
}

func (chain *Blockchain) AddBlock(data string) {
	previousBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := Block{Data: []byte(data), PreviousHash: previousBlock.Hash}
	newBlock.CalculateHash()

	chain.Blocks = append(chain.Blocks, &newBlock)
}

func main() {
	generisBlock := Block{Data: []byte("Genesis block")}
	generisBlock.CalculateHash()

	chain := Blockchain{Blocks: []*Block{&generisBlock}}
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
