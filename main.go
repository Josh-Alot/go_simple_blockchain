package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"errors"
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

func (chain *Blockchain) Balance(owner []byte) int {
	spentOutputs := chain.findSpentOutputs()
	amount := 0

	for _, block := range chain.Blocks {
		for _, tx := range block.Transactions {
			for i, output := range tx.Outputs {
				isSpent := false
				for _, spent := range spentOutputs[string(tx.ID)] {
					if spent == i {
						isSpent = true
						break
					}
				}

				if bytes.Equal(output.Owner, owner) && !isSpent {
					amount += output.Amount
				}
			}
		}
	}

	return amount
}

func CoinbaseTx(owner []byte) *Transaction {
	tx := Transaction{Outputs: []TxOutput{{Amount: reward, Owner: owner}}}
	tx.ID = tx.CalculateTxID()

	return &tx
}

func (chain *Blockchain) findSpentOutputs() map[string][]int {
	spentOutputs := make(map[string][]int)

	for _, block := range chain.Blocks {
		for _, tx := range block.Transactions {
			for _, input := range tx.Inputs {
				hash := string(input.TransactionHash)
				spentOutputs[hash] = append(spentOutputs[hash], input.TxOutputIndex)
			}
		}
	}

	return spentOutputs
}

func (chain *Blockchain) findSpendableInputs(from []byte, amount int, spentOutputs map[string][]int) ([]TxInput, int) {
	accumulatedAmount := 0
	var inputs []TxInput

Collect:
	for _, block := range chain.Blocks {
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

	return inputs, accumulatedAmount
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
