package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	var chain *Blockchain
	john := CreateWallet()
	jane := CreateWallet()

	if _, err := os.Stat(chainFile); os.IsNotExist(err) {
		chain = InitBlockchain(john.GetAddress())
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

	fmt.Printf("Address 1 balance before transaction: %d\n", chain.Balance(john.GetAddress()))
	fmt.Printf("Address 2 balance before transaction: %d\n", chain.Balance(jane.GetAddress()))

	newTransaction, err := NewTransaction(chain, john.GetAddress(), jane.GetAddress(), 30)
	if err != nil {
		log.Fatal(err)
	}

	newTransaction.Sign(jane.PrivateKey)

	transactions := []*Transaction{newTransaction}
	err = chain.AddBlock(transactions)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	fmt.Printf("Address 1 balance after transaction: %d\n", chain.Balance(john.GetAddress()))
	fmt.Printf("Address 2 balance after transaction: %d\n", chain.Balance(jane.GetAddress()))

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

	err = chain.SaveToFile()
	if err != nil {
		log.Fatal(err)
	}
}
