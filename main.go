package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var chain *Blockchain
	var wallets *Wallets
	var err error

	// cli commands
	startnode := flag.NewFlagSet("startnode", flag.ExitOnError)
	createwallet := flag.NewFlagSet("createwallet", flag.ExitOnError)
	createblockchain := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printchain := flag.NewFlagSet("printchain", flag.ExitOnError)
	getbalance := flag.NewFlagSet("getbalance", flag.ExitOnError)
	send := flag.NewFlagSet("send", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("usage: [not defined yet]") // name if after
		os.Exit(1)
	}

	switch os.Args[1] {
	case "startnode":
		port := startnode.String("port", "", "the IP port number")
		connect := startnode.String("connect", "", "the IP to connect")
		startnode.Parse(os.Args[2:])

		if *port == "" {
			fmt.Println("provide a valid port (example: 8080)")
			os.Exit(1)
		}

		chain, err := LoadChainFromFile()
		if err != nil {
			log.Fatal(err)
		}

		StartNode(*port, *connect, chain)
	case "createwallet":
		createwallet.Parse(os.Args[2:])

		wallets, err = LoadOrCreateWallets()
		if err != nil {
			log.Fatal(err)
		}

		address := wallets.AddWallet()
		err = wallets.SaveToFile()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("New wallet address: %s\n", address)

		os.Exit(0)
	case "createblockchain":
		address := createblockchain.String("address", "", "a wallet address")
		createblockchain.Parse(os.Args[2:])

		if *address == "" {
			fmt.Printf("give a wallet address to create a blockchain")
			os.Exit(1)
		}

		decodedAddr, err := hex.DecodeString(*address)
		if err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stat(chainFile); os.IsNotExist(err) {
			chain = InitBlockchain(decodedAddr)
		} else {
			fmt.Println("blockchain already exists")
			os.Exit(1)
		}

		fmt.Printf("Blockchain created with the address %s", *address)
		err = chain.SaveToFile()
		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	case "printchain":
		printchain.Parse(os.Args[2:])

		if _, err = os.Stat(chainFile); os.IsNotExist(err) {
			fmt.Printf("blockchain not found, use \"createblockchain\" to create a blockchain\n")
			os.Exit(1)
		} else {
			chain, err = LoadChainFromFile()
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

		os.Exit(0)
	case "getbalance":
		getbalance.Parse(os.Args[2:])
		address := getbalance.Arg(0)

		if getbalance.NArg() == 0 {
			fmt.Println("a wallet address is required")
			os.Exit(1)
		}

		if _, err = os.Stat(chainFile); os.IsNotExist(err) {
			fmt.Printf("blockchain not found, use \"createblockchain\" to create a blockchain\n")

			os.Exit(1)
		} else {
			chain, err = LoadChainFromFile()
			if err != nil {
				log.Fatal(err)
			}

			isFileValid := chain.IsValid()
			if !isFileValid {
				log.Fatal("Invalid blockchain file, file is corrupted")
			}
		}

		decodedAddr, err := hex.DecodeString(address)
		if err != nil {
			log.Fatal(err)
		}

		balance := chain.Balance(decodedAddr)
		fmt.Printf("wallet balance: %d", balance)

		os.Exit(0)
	case "send":
		// args validation
		node := send.String("node", "", "the node IP address running")
		from := send.String("from", "", "the origin address")
		to := send.String("to", "", "the destiny address")
		amount := send.Int("amount", 0, "the amount the origin send to destiny")

		send.Parse(os.Args[2:])

		if *node == "" {
			fmt.Printf("provide the node IP address (example: localhost:3000)\n")
			os.Exit(1)
		}

		if *from == "" {
			fmt.Printf("give the wallet sender address\n")
			os.Exit(1)
		}

		if *to == "" {
			fmt.Printf("give the wallet destiny address\n")
			os.Exit(1)
		}

		if *amount <= 0 {
			fmt.Printf("give a positive value amount you want to send\n")
			os.Exit(1)
		}

		// blockchain validation
		if _, err = os.Stat(chainFile); os.IsNotExist(err) {
			fmt.Printf("blockchain not found, use \"createblockchain\" to create a blockchain\n")
			os.Exit(1)
		} else {
			chain, err = LoadChainFromFile()
			if err != nil {
				log.Fatal(err)
			}

			isFileValid := chain.IsValid()
			if !isFileValid {
				log.Fatal("Invalid blockchain file, file is corrupted")
			}
		}

		// create the transaction
		wallets, err = LoadOrCreateWallets()
		if err != nil {
			log.Fatal(err)
		}

		sender, err := hex.DecodeString(*from)
		if err != nil {
			log.Fatal(err)
		}

		destiny, err := hex.DecodeString(*to)
		if err != nil {
			log.Fatal(err)
		}

		wallet, err := wallets.GetWallet(*from)
		if err != nil {
			log.Fatal(err)
		}

		transaction, err := NewTransaction(chain, sender, destiny, *amount)
		if err != nil {
			fmt.Printf("%v", err)
			os.Exit(1)
		}

		// signs and submits the block to the node
		transaction.Sign(wallet.PrivateKey)
		SendTransaction(*node, transaction)

		os.Exit(0)
	default:
		fmt.Printf("command not found\n") // name if after
		os.Exit(1)
	}
}
