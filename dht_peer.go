package dht

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	// "os"
	"strconv"
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
	Src     Contact
	Dest    Contact
}

var predecessor *Contact
var contact *Contact
var finger []*DHTFinger
var networkSize int
var answerChannel chan Contact

func MakeDHTNode(nodeId *string, ip string, port string, nwSize int) Contact {
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
	answerChannel = make(chan Contact)
	go listen(contact)

	return *contact
}

func addToRing(ring *Contact) {
	fmt.Println("Player 2 join the game")
	initFingerTable(ring)
	fmt.Println("finger table init")
	updateOthers()
	fmt.Println("Ok c'est fait!")
	printFingerTable()
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
func initFingerTable(ring *Contact) {
	finger[1].responsibleNode = findSuccessor(finger[1].start)

	predecessor = requestFinger(finger[1].responsibleNode, -1)
	setRemoteFinger(finger[1].responsibleNode, -1, contact)
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
		message.Type = "updateFinger"
		message.Src = *contact
		message.Dest = *nodeToUpdate
		message.Payload = strconv.Itoa(i)
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
		message.Type = "updateFinger"
		message.Src = *newNode
		message.Dest = *predecessor
		message.Payload = strconv.Itoa(index)
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

	return requestFinger(idPredecessor, 1)
}

func findPredecessor(id string) *Contact {
	nodeIterator := contact
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

func listen(contact *Contact) {
	udpAddress, err := net.ResolveUDPAddr("udp4", contact.String())
	if err != nil {
		fmt.Println("Error while resolving", udpAddress)
	}

	connection, err := net.ListenUDP("udp4", udpAddress)
	if err != nil {
		fmt.Println("Error while listening to", udpAddress)
		fmt.Println(err)
	}
	defer connection.Close()

	decoder := json.NewDecoder(connection)

	fmt.Println("allo oui j'écoute?")
	for {
		message := new(Message)
		err = decoder.Decode(message)
		if err != nil {
			fmt.Println("Unvalid message format")
		}

		switch message.Type {
		case "requestFinger":
			fmt.Println("asking for a finger")
			index, _ := strconv.Atoi(message.Payload)
			answer := new(Message)
			answer.Type = "answerFinger"
			answer.Src = *(finger[index].responsibleNode)
			answer.Dest = message.Src
			answer.Payload = ""

		case "updateFinger":
			fmt.Println("requesting to update a finger")
			index, _ := strconv.Atoi(message.Payload)
			updateFingerTable(&(message.Src), index)

		case "setFinger":
			fmt.Println("requesting to set a finger")
			index, _ := strconv.Atoi(message.Payload)
			finger[index].responsibleNode = &(message.Src) //TODO Pb here maybe

		case "joinRing":
			fmt.Println("I will join the ring")
			addToRing(&(message.Src))

		case "answerFinger":
			answerChannel <- message.Src
		}
	}
}

func send(message *Message) {
	udpAddress, err := net.ResolveUDPAddr("udp4", message.Dest.String())
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
	message := new(Message)
	message.Type = "requestFinger"
	message.Src = *contact
	message.Dest = *peer
	message.Payload = strconv.Itoa(fingerIndex)
	send(message)

	//TODO timeout that shit
	answer := <-answerChannel
	return &answer
}

func setRemoteFinger(peer *Contact, fingerIndex int, newContact *Contact) {
	message := new(Message)
	message.Type = "setFinger"
	message.Src = *newContact
	message.Dest = *peer
	message.Payload = strconv.Itoa(fingerIndex)
	send(message)
}
