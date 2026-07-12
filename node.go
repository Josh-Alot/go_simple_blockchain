package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

const cmdVersion = "version"
const cmdTransaction = "transaction"
const cmdInv = "inv"
const cmdGetData = "getdata"
const cmdBlock = "block"

type Node struct {
	Address string
	Chain   *Blockchain
	Peers   []string
}

type Message struct {
	Command string
	Payload []byte
}

type Version struct {
	AddrFrom string
	Height   int
}

type Inv struct {
	Hashes   [][]byte
	AddrFrom string
}

type GetData struct {
	Hash     []byte
	AddrFrom string
}

func StartNode(port, connectTo string, chain *Blockchain) {
	node := Node{Address: "localhost:" + port, Chain: chain}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	if connectTo != "" {
		node.connectToPeer(connectTo)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go node.handleConnection(conn)
	}
}

func (node *Node) connectToPeer(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("%v", err)
		return
	}
	defer conn.Close()

	node.sendVersion(conn)
	_, err = node.handleMessage(conn)
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

func (node *Node) handleConnection(conn net.Conn) {
	command, err := node.handleMessage(conn)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	if command == cmdVersion {
		node.sendVersion(conn)
	}

	conn.Close()
}

func (node *Node) handleMessage(conn net.Conn) (string, error) {
	var message Message

	if err := gob.NewDecoder(conn).Decode(&message); err != nil {
		return "", err
	}

	switch message.Command {
	case cmdVersion:
		var version Version
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&version)

		fmt.Printf("peer version: %d\n", version.Height)
		node.handleAddr(version.AddrFrom)

	case cmdTransaction:
		var transaction *Transaction
		gob.NewDecoder((bytes.NewReader(message.Payload))).Decode(&transaction)

		transactions := []*Transaction{transaction}
		err := node.Chain.AddBlock(transactions)
		if err != nil {
			return "", err
		}

		err = node.Chain.SaveToFile()
		if err != nil {
			return "", err
		}

		newBlock := node.Chain.Blocks[len(node.Chain.Blocks)-1]
		node.broadcastInv(newBlock)

	case cmdInv:
		var inv *Inv
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&inv)

		for _, hash := range inv.Hashes {
			_, err := node.Chain.GetBlock(hash)
			if err != nil {
				log.Println("block not found, fetching block data on chain")

				getData := GetData{Hash: hash, AddrFrom: node.Address}
				if err := sendTo(inv.AddrFrom, cmdGetData, getData); err != nil {
					log.Printf("%v\n", err)
				}
			}
		}

	case cmdGetData:
		var getData *GetData
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&getData)

		block, err := node.Chain.GetBlock(getData.Hash)
		if err != nil {
			log.Printf("%v\n", err)
			break
		}

		if err := sendTo(getData.AddrFrom, cmdBlock, block); err != nil {
			log.Printf("%v\n", err)
		}

	case cmdBlock:
		var block *Block
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&block)

		err := node.Chain.AddMinedBlock(block)
		if err != nil {
			log.Printf("%v\n", err)
			break
		}

		err = node.Chain.SaveToFile()
		if err != nil {
			log.Printf("%v\n", err)
			break
		}

		node.broadcastInv(block)
	}

	return message.Command, nil
}

func (node *Node) sendVersion(conn net.Conn) {
	version := Version{AddrFrom: node.Address, Height: len(node.Chain.Blocks)}
	err := sendMessage(cmdVersion, conn, version)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func (node *Node) handleAddr(address string) {
	knownAddr := false

	for _, peer := range node.Peers {
		if peer == address {
			knownAddr = true
			break
		}
	}

	if !knownAddr {
		node.Peers = append(node.Peers, address)
	}

	fmt.Printf("known peers: %v\n", node.Peers)
}

func sendMessage(command string, conn net.Conn, payload any) error {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(payload); err != nil {
		return err
	}

	message := Message{Command: command, Payload: buffer.Bytes()}
	if err := gob.NewEncoder(conn).Encode(message); err != nil {
		return err
	}
	return nil
}

func SendTransaction(address string, transaction *Transaction) error {
	return sendTo(address, cmdTransaction, transaction)
}

func (node *Node) broadcastInv(newBlock *Block) {
	inv := Inv{Hashes: [][]byte{newBlock.Hash}, AddrFrom: node.Address}
	for _, peer := range node.Peers {
		if err := sendTo(peer, cmdInv, inv); err != nil {
			log.Printf("%v\n", err)
		}
	}
}

func sendTo(address, command string, payload any) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	return sendMessage(command, conn, payload)
}
