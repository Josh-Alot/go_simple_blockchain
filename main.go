package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"os"
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
const chainFile = "blockchain.dat"

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

			fmt.Printf("Hash: %s, block found\n", hashHex)
			break
		} else {
			fmt.Printf("Hash: %s, block not found\n", hashHex)
			nonce++
		}

	}
}

func InitBlockchain() *Blockchain {
	genesisBlock := Block{Data: []byte("Genesis block")}
	genesisBlock.Mine()

	chain := Blockchain{Blocks: []*Block{&genesisBlock}}
	return &chain
}

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

func (chain *Blockchain) AddBlock(data string) {
	previousBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := Block{Data: []byte(data), PreviousHash: previousBlock.Hash}
	newBlock.Mine()

	chain.Blocks = append(chain.Blocks, &newBlock)
}

func (chain *Blockchain) IsValid() bool {
	for i := len(chain.Blocks) - 1; i > 0; i-- {
		currentBlock := chain.Blocks[i]
		previousBlock := chain.Blocks[i-1]

		if !bytes.Equal(currentBlock.PreviousHash, previousBlock.Hash) {
			return false
		}

		if hash := currentBlock.CalculateHash(currentBlock.Nonce); !bytes.Equal(hash, currentBlock.Hash) {
			return false
		}
	}

	return true
}

func main() {
	var chain *Blockchain

	if _, err := os.Stat(chainFile); os.IsNotExist(err) {
		chain = InitBlockchain()
	} else {
		chain, err = LoadFromFile()
		if err != nil {
			log.Fatal(err)
		}

		// chain.Blocks[1].Data = []byte("Sindios, grupo rakiado")

		isFileValid := chain.IsValid()
		if !isFileValid {
			log.Fatal("Invalid blockchain file, file is corrupted")
		}
	}

	// chain.AddBlock("João pagou 10 para Maria")
	// chain.AddBlock("Pedro pagou 15 para João")

	fmt.Println("**** Blockchain ****")
	for i, block := range chain.Blocks {
		fmt.Printf("Block id: %d\n", i)
		fmt.Printf("Block Hash: %x\n", block.Hash)
		fmt.Printf("Block data: %s\n", block.Data)
		fmt.Printf("Previous Block: %x\n\n", block.PreviousHash)
	}

	err := chain.SaveToFile()
	if err != nil {
		log.Fatal(err)
	}
}
