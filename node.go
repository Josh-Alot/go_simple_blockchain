package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

type Message struct {
	Command string
	Payload []byte
}

type Version struct {
	Height int
}

func StartNode(port, connectTo string, chain *Blockchain) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	if connectTo != "" {
		connectToPeer(connectTo, chain)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go handleConnection(conn, chain)
	}
}

func connectToPeer(address string, chain *Blockchain) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	sendVersion(conn, chain)
	err = handleMessage(conn)
	if err != nil {
		log.Fatal(err)
	}
}

func handleConnection(conn net.Conn, chain *Blockchain) {
	err := handleMessage(conn)
	if err != nil {
		log.Fatal(err)
	}

	sendVersion(conn, chain)
	conn.Close()
}

func handleMessage(conn net.Conn) error {
	var message Message
	var version Version

	if err := gob.NewDecoder(conn).Decode(&message); err != nil {
		return err
	}

	switch message.Command {
	case "version":
		gob.NewDecoder(bytes.NewReader(message.Payload)).Decode(&version)
		fmt.Printf("peer version: %d\n", version.Height)
	}
	return nil
}

func sendVersion(conn net.Conn, chain *Blockchain) {
	var buffer bytes.Buffer

	version := Version{Height: len(chain.Blocks)}
	if err := gob.NewEncoder(&buffer).Encode(version); err != nil {
		log.Fatal(err)
	}

	payload := buffer.Bytes()
	message := Message{Command: "version", Payload: payload}
	if err := gob.NewEncoder(conn).Encode(message); err != nil {
		log.Fatal(err)
	}
}
