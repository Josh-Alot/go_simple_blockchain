package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const difficulty = 3
const reward = 50
const chainFile = "blockchain.dat"

type Block struct {
	Transactions []*Transaction
	Hash         []byte
	PreviousHash []byte
	Nonce        int
}

type Blockchain struct {
	Blocks []*Block
}

type TxInput struct {
	TransactionHash []byte
	TxOutputIndex   int
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
			fmt.Printf("Hash: %s, block not found\n", hashHex)
			nonce++
		}

	}
}

func InitBlockchain() *Blockchain {
	owner := sha256.Sum256([]byte("John Doe"))
	genesisTransaction := CoinbaseTx(owner[:])
	genesisBlock := Block{Transactions: []*Transaction{genesisTransaction}}
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
	// temporary until transactions are implemented
	owner := sha256.Sum256([]byte("Jane Doe"))
	genesisTransaction := CoinbaseTx(owner[:])

	previousBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := Block{Transactions: []*Transaction{genesisTransaction}, PreviousHash: previousBlock.Hash}
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

		if hash := currentBlock.CalculateBlockHash(currentBlock.Nonce); !bytes.Equal(hash, currentBlock.Hash) {
			return false
		}
	}

	return true
}

func CoinbaseTx(owner []byte) *Transaction {
	tx := Transaction{Outputs: []TxOutput{{Amount: reward, Owner: owner}}}
	tx.ID = tx.CalculateTxID()

	return &tx
}

func NewTransaction(from []byte, to []byte, amount int) {
	chain, err := LoadFromFile()
	if err != nil {
		log.Fatal(err)
	}

	spentOutputs := make(map[string][]int)

	for _, block := range chain.Blocks {
		for _, tx := range block.Transactions {
			for _, input := range tx.Inputs {
				hash := string(input.TransactionHash)
				spentOutputs[hash] = append(spentOutputs[hash], input.TxOutputIndex)
			}
		}
	}

	accumulatedAmount := 0
	var inputs []TxInput

	for _, block := range chain.Blocks {
	Collect:
		for _, tx := range block.Transactions {
			for i, output := range tx.Outputs {
				if !bytes.Equal(from, output.Owner) {
					continue
				}

				isSpent := false
				for _, spent := range spentOutputs[string(tx.ID)] {
					if spent == i {
						isSpent = true
						break
					}
				}

				if isSpent {
					continue
				}

				newInput := TxInput{TransactionHash: tx.ID, TxOutputIndex: i}
				accumulatedAmount += output.Amount
				inputs = append(inputs, newInput)

				if accumulatedAmount >= amount {
					break Collect
				}
			}
		}
	}

	if accumulatedAmount < amount {
		fmt.Errorf("Not enough cash to spend on this transaction: %d, requested: %d", accumulatedAmount, amount)
	}
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

func main() {
	var chain *Blockchain

	if _, err := os.Stat(chainFile); os.IsNotExist(err) {
		chain = InitBlockchain()
	} else {
		chain, err = LoadFromFile()
		if err != nil {
			log.Fatal(err)
		}

		isFileValid := chain.IsValid()
		if !isFileValid {
			log.Fatal("Invalid blockchain file, file is corrupted")
		}
	}

	fmt.Println("**** Blockchain ****")
	for i, block := range chain.Blocks {
		fmt.Printf("Block id: %d\n", i)
		fmt.Printf("Block Hash: %x\n", block.Hash)
		fmt.Printf("Previous Block: %x\n\n", block.PreviousHash)

		fmt.Printf("\nBlock transactions\n")
		for _, tx := range block.Transactions {
			fmt.Printf("Transaction ID: %x\n", tx.ID)
			fmt.Printf("Transaction Inputs: %x\n", tx.Inputs)
			fmt.Printf("Transaction Outputs: %x\n", tx.Outputs)
		}
	}

	err := chain.SaveToFile()
	if err != nil {
		log.Fatal(err)
	}
}
