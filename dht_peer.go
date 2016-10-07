package dht

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	// "os"
	// "strconv"
)

type Contact struct {
	nodeId string
	ip     string
	port   string
}

type DHTFinger struct {
	start           string
	end             string
	responsibleNode *Contact
}

type Message struct {
	Type    string
	Payload string
	Src     string
	Dest    string
}

var predecessor *Contact
var contact *Contact
var finger []*DHTFinger
var networkSize int

func MakeDHTNode(nodeId *string, ip string, port string, nwSize int) {
	contact := new(Contact)
	contact.ip = ip
	contact.port = port
	networkSize = nwSize

	if nodeId == nil {
		genNodeId := generateNodeId()
		contact.nodeId = genNodeId
	} else {
		contact.nodeId = *nodeId
	}

	predecessor = contact

	nodeIdBytes, _ := hex.DecodeString(contact.nodeId)

	finger = make([]*DHTFinger, networkSize+1)

	finger0 := new(DHTFinger)
	finger0.responsibleNode = contact
	finger0.start = contact.nodeId
	finger0.end = contact.nodeId
	finger[0] = finger0

	for i := 1; i < networkSize+1; i++ {
		newFinger := new(DHTFinger)
		newFinger.responsibleNode = contact
		newFinger.start, _ = calcFinger(nodeIdBytes, i, networkSize)
		newFinger.end, _ = calcFinger(nodeIdBytes, i+1, networkSize)
		finger[i] = newFinger
	}
}

func addToRing(peer *Contact) {
	initFingerTable(peer)
	updateOthers()
}

func isResponsible(key string) bool {
	if between(predecessor.nodeId, contact.nodeId, key, false, true) {
		return true
	}

	return false
}

func isAlone() bool {
	return finger[1].responsibleNode.nodeId == contact.nodeId
}

// -------------------------------------------------------------------------------Finger table util
func initFingerTable(ring *Contact) { // TODO: replace findSuccesor call by sending a message
	finger[1].responsibleNode = findSuccessor(finger[1].start)

	predecessor = requestFinger(finger[1].responsibleNode, -1)
	setFinger(finger[1].responsibleNode, -1, contact)
	for i := 1; i < networkSize; i++ {
		if between(contact.nodeId, finger[i].responsibleNode.nodeId, finger[i+1].start, false, true) {
			finger[i+1].responsibleNode = finger[i].responsibleNode
		} else {
			finger[i+1].responsibleNode = findSuccessor(finger[i+1].start)
		}
	}
	printFingerTable()
}

func updateOthers() {
	var nodeToUpdate *Contact
	var newNodeIdInt, power, toSubstract, result, modulo, addressSpace big.Int
	(&newNodeIdInt).SetString(contact.nodeId, 16)
	for i := 1; i < networkSize+1; i++ {
		power = *big.NewInt(int64(i - 1))
		(&toSubstract).Exp(big.NewInt(2), &power, nil)
		(&result).Sub(&newNodeIdInt, &toSubstract)
		(&addressSpace).Exp(big.NewInt(2), big.NewInt(int64(networkSize)), nil)
		(&modulo).Mod(&result, &addressSpace)
		targetNode := hex.EncodeToString(modulo.Bytes())
		if targetNode == "" {
			targetNode = "00"
		}
		//targetNode = (newNode - 2 ^ (i-1)) mod 2 ^ m
		nodeToUpdate = findPredecessor(targetNode)

		message := new(Message)
		message.Dest = nodeToUpdate.String()
		send(message)
		// nodeToUpdate.updateFingerTable(newNode, i) // TODO: send a message to nodeToUpdate
	}
}

func updateFingerTable(newNode *Contact, index int) {
	if newNode.nodeId == contact.nodeId {
		return
	}

	if between(contact.nodeId, finger[index].responsibleNode.nodeId, newNode.nodeId, false, true) {
		finger[index].responsibleNode = newNode
		message := new(Message)
		send(message)
		// predecessor.updateFingerTable(newNode, index) //TODO: send a message
	}
}

// ------------------------------------------------------------------------------------------Lookup
func findSuccessor(id string) *Contact {
	idPredecessor := findPredecessor(id)

	// TODO QuickFix:
	if idPredecessor.nodeId == id {
		return idPredecessor
	}

	return requestFinger(idPredecessor, 1) // TODO: send message
}

func findPredecessor(id string) *Contact {
	nodeIterator := contact // TA MERE
	nodeIteratorSuccessor := finger[1].responsibleNode

	for !between(nodeIterator.nodeId, nodeIteratorSuccessor.nodeId, id, true, false) {
		nodeIterator = closestPrecedingFinger(nodeIterator, id)
		nodeIteratorSuccessor = requestFinger(nodeIterator, 1)
	}

	// TODO: Quick fix
	if id == "" {
		id = "00"
	}
	if nodeIteratorSuccessor.nodeId == id {
		nodeIterator = nodeIteratorSuccessor
	}

	return nodeIterator
}

func closestPrecedingFinger(dhtNode *Contact, id string) *Contact {
	for i := networkSize; i > 0; i-- {
		ithFinger := requestFinger(dhtNode, i)
		if between(dhtNode.nodeId, id, ithFinger.nodeId, false, true) {
			return ithFinger
		}
	}
	return dhtNode // Should never happen
}

// ----------------------------------------------------------------------------------------Printing
func printFingerTable() {
	fmt.Println("Node", contact.nodeId, "finger table is:")
	for i := 1; i < networkSize+1; i++ {
		fmt.Println("finger", i, "start:", finger[i].start, "end: ", finger[i].end, "responsible node:", finger[i].responsibleNode.nodeId)
	}
}

// func (dhtNode *DHTNode) printRing() {
// 	firstNode := dhtNode.findSuccessor("00") // Find first node
// 	currentNode := firstNode

// 	for currentNode.finger[1].responsibleNode != firstNode {
// 		fmt.Println(currentNode.nodeId)
// 		currentNode = currentNode.finger[1].responsibleNode
// 	}
// 	fmt.Println(currentNode.nodeId)
// }

// -------------------------------------------------------------------------------Net communication

func (contact *Contact) String() string {
	return contact.ip + ":" + contact.port
}

func listen() {
	udpAddress, err := net.ResolveUDPAddr("udp4", contact.String())
	if err != nil {
		fmt.Println("Error while resolving", udpAddress)
	}
	connection, err := net.ListenUDP("udp4", udpAddress)
	defer connection.Close()
	if err != nil {
		fmt.Println("Error while listening to", udpAddress)
	}

	decoder := json.NewDecoder(connection)

	for {
		message := new(Message)
		err = decoder.Decode(message)
		fmt.Println("Unvalid message format")

		switch message.Type {
		case "Join":
			fmt.Println("Someone wants to join!")
		case "Lookup":
			fmt.Println("Someone is looking for a value!")
		}
	}
}

func send(message *Message) {
	udpAddress, err := net.ResolveUDPAddr("udp4", message.Dest)
	if err != nil {
		fmt.Println("Error while resolving", udpAddress)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		fmt.Println("Error while connecting to", udpAddress)
		return
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)

	err = encoder.Encode(message)
	if err != nil {
		fmt.Println("Error while sending a message", udpAddress)
		return
	}
}

func requestFinger(peer *Contact, fingerIndex int) *Contact {
	// send message
	// parse answer
	// Create Contact
	return peer
}

func setFinger(peer *Contact, fingerIndex int, newContact *Contact) {
	// send message
}
