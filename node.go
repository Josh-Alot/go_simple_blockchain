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
	var version Version
	var transaction *Transaction

	if err := gob.NewDecoder(conn).Decode(&message); err != nil {
		return "", err
	}

	switch message.Command {
	case cmdVersion:
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&version)
		fmt.Printf("peer version: %d\n", version.Height)

		node.handleAddr(version.AddrFrom)
	case cmdTransaction:
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
	}
	return message.Command, nil
}

func (node *Node) sendVersion(conn net.Conn) {
	var buffer bytes.Buffer

	version := Version{AddrFrom: node.Address, Height: len(node.Chain.Blocks)}
	if err := gob.NewEncoder(&buffer).Encode(version); err != nil {
		log.Fatal(err)
	}

	payload := buffer.Bytes()
	message := Message{Command: cmdVersion, Payload: payload}
	if err := gob.NewEncoder(conn).Encode(message); err != nil {
		log.Fatal(err)
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

func SendTransaction(address string, transaction *Transaction) error {
	var buffer bytes.Buffer

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := gob.NewEncoder(&buffer).Encode(transaction); err != nil {
		log.Fatal(err)
	}

	payload := buffer.Bytes()
	message := Message{Command: cmdTransaction, Payload: payload}
	if err := gob.NewEncoder(conn).Encode(message); err != nil {
		log.Fatal(err)
	}

	return nil
}
