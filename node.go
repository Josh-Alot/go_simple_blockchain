package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"slices"
	"sync"
)

const cmdVersion = "version"
const cmdTransaction = "transaction"
const cmdInv = "inv"
const cmdGetData = "getdata"
const cmdBlock = "block"
const cmdGetBlock = "getblock"

type Node struct {
	Address string
	mu      sync.Mutex
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

type GetDataRequest struct {
	Hash     []byte
	AddrFrom string
}

type GetBlockRequest struct {
	Height int
}

func StartNode(port, connectTo string, chain *Blockchain) {
	node := Node{Address: "localhost:" + port, Chain: chain}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	if connectTo != "" {
		go node.connectToPeer(connectTo)
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

	for {
		_, err = node.handleMessage(conn)
		if err != nil {
			log.Printf("%v", err)
			break
		}
	}
}

func (node *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		command, err := node.handleMessage(conn)
		if err != nil {
			log.Printf("%v", err)
			break
		}

		if command == cmdVersion {
			node.sendVersion(conn)
		}
	}

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

		if version.Height > node.getChainLength() {
			if err := node.syncChain(conn, version.Height); err != nil {
				log.Printf("%v", err)
			}
		}

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
			_, err := node.getBlock(hash)
			if err != nil {
				log.Println("block not found, fetching block data on chain")

				getData := GetDataRequest{Hash: hash, AddrFrom: node.Address}
				if err := sendTo(inv.AddrFrom, cmdGetData, getData); err != nil {
					log.Printf("%v\n", err)
				}
			}
		}

	case cmdGetData:
		var getData *GetDataRequest
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&getData)

		block, err := node.getBlock(getData.Hash)
		if err != nil {
			log.Printf("%v\n", err)
			break
		}

		if err := sendTo(getData.AddrFrom, cmdBlock, block); err != nil {
			log.Printf("%v\n", err)
		}

	case cmdGetBlock:
		var getBlock *GetBlockRequest
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&getBlock)

		block, err := node.getBlockAt(getBlock.Height)
		if err != nil {
			log.Printf("%v\n", err)
			break
		}

		if err := sendMessage(cmdBlock, conn, block); err != nil {
			log.Printf("%v\n", err)
		}

	case cmdBlock:
		var block *Block
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&block)

		err := node.appendAndSaveBlock(block)
		if err != nil {
			log.Printf("%v\n", err)
			break
		}

		node.broadcastInv(block)
	}

	return message.Command, nil
}

func (node *Node) appendAndSaveBlock(block *Block) error {
	node.mu.Lock()
	defer node.mu.Unlock()

	err := node.Chain.AddMinedBlock(block)
	if err != nil {
		return err
	}

	err = node.Chain.SaveToFile()
	if err != nil {
		return err
	}

	return nil
}

func (node *Node) getChainLength() int {
	node.mu.Lock()
	defer node.mu.Unlock()

	return len(node.Chain.Blocks)
}

func (node *Node) getBlock(hash []byte) (*Block, error) {
	node.mu.Lock()
	defer node.mu.Unlock()

	return node.Chain.GetBlock(hash)
}

func (node *Node) getBlockAt(height int) (*Block, error) {
	node.mu.Lock()
	defer node.mu.Unlock()

	return node.Chain.GetBlockAt(height)
}

func (node *Node) syncChain(conn net.Conn, peerHeight int) error {
	for {
		height := node.getChainLength()
		if height >= peerHeight {
			break
		}

		getBlockRequest := GetBlockRequest{Height: height}
		if err := sendMessage(cmdGetBlock, conn, getBlockRequest); err != nil {
			return err
		}

		var message Message
		err := gob.NewDecoder(conn).Decode(&message)
		if err != nil {
			return err
		}

		var block Block
		err = gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&block)
		if err != nil {
			return err
		}

		err = node.appendAndSaveBlock(&block)
		if err != nil {
			return err
		}
	}

	return nil
}

func (node *Node) sendVersion(conn net.Conn) {
	version := Version{AddrFrom: node.Address, Height: node.getChainLength()}
	err := sendMessage(cmdVersion, conn, version)
	if err != nil {
		log.Printf("%v\n", err)
	}
}

func (node *Node) handleAddr(address string) {
	node.mu.Lock()
	defer node.mu.Unlock()

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

	// the unlock won't be deferred because it cannot lock all the goroutines on the network on slow I/) syncs
	node.mu.Lock()
	peers := slices.Clone(node.Peers)
	node.mu.Unlock()

	for _, peer := range peers {
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
