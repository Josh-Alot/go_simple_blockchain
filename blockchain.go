package main

import (
	"bytes"
	"errors"
)

const difficulty = 3
const reward = 50

type Blockchain struct {
	Blocks []*Block
}

func InitBlockchain(addr []byte) *Blockchain {
	genesisTransaction := CoinbaseTx(addr)
	genesisBlock := Block{Transactions: []*Transaction{genesisTransaction}}
	genesisBlock.Mine()

	chain := Blockchain{Blocks: []*Block{&genesisBlock}}
	return &chain
}

func (chain *Blockchain) AddBlock(transactions []*Transaction) error {
	for _, transaction := range transactions {
		if !transaction.ValidateTransaction(chain) {
			return errors.New("failed to validate transactions")
		}
	}

	previousBlock := chain.Blocks[len(chain.Blocks)-1]
	newBlock := Block{Transactions: transactions, PreviousHash: previousBlock.Hash}
	newBlock.Mine()

	chain.Blocks = append(chain.Blocks, &newBlock)

	return nil
}

func (chain *Blockchain) AddMinedBlock(incomingBlock *Block) error {
	incomingHash := incomingBlock.CalculateBlockHash(incomingBlock.Nonce)
	previousBlock := chain.Blocks[len(chain.Blocks)-1]

	for _, transaction := range incomingBlock.Transactions {
		if !transaction.ValidateTransaction(chain) {
			return errors.New("failed to validate incoming transactions")
		}
	}

	if !bytes.Equal(incomingBlock.PreviousHash, previousBlock.Hash) {
		return errors.New("failed to verify incoming block previous hash")
	}

	if !bytes.Equal(incomingHash, incomingBlock.Hash) {
		return errors.New("failed to verify incoming block hash")
	}

	if verified := VerifyBlockHashDifficulty(incomingHash); !verified {
		return errors.New("failed to verify incoming block PoW")
	}

	chain.Blocks = append(chain.Blocks, incomingBlock)
	return nil
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

func (chain *Blockchain) GetBlockAt(height int) (*Block, error) {
	if height >= len(chain.Blocks) || height < 0 {
		return nil, errors.New("block doesn't exist on chain")
	}

	return chain.Blocks[height], nil
}

func (chain *Blockchain) GetBlock(hash []byte) (*Block, error) {
	for _, block := range chain.Blocks {
		if bytes.Equal(hash, block.Hash) {
			return block, nil
		}
	}

	return nil, errors.New("block hash not found!")
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

func (chain *Blockchain) findTransaction(txId []byte) (*Transaction, error) {
	for _, block := range chain.Blocks {
		for _, transaction := range block.Transactions {
			if bytes.Equal(txId, transaction.ID) {
				return transaction, nil
			}
		}
	}

	return nil, errors.New("transaction not found!")
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
