package dht

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"strings"
	// "os"
	"strconv"
)

type Contact struct {
	NodeId string
	Ip     string
	Port   string
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
var answerChannel chan Contact

func MakeDHTNode(NodeId *string, Ip string, Port string, nwSize int) Contact {
	contact := new(Contact)
	contact.Ip = Ip
	contact.Port = Port
	networkSize = nwSize
	if NodeId == nil {
		genNodeId := generateNodeId()
		contact.NodeId = genNodeId
	} else {
		contact.NodeId = *NodeId
	}

	predecessor = contact

	NodeIdBytes, _ := hex.DecodeString(contact.NodeId)

	finger = make([]*DHTFinger, networkSize+1)

	finger0 := new(DHTFinger)
	finger0.responsibleNode = contact
	finger0.start = contact.NodeId
	finger0.end = contact.NodeId
	finger[0] = finger0

	for i := 1; i < networkSize+1; i++ {
		newFinger := new(DHTFinger)
		newFinger.responsibleNode = contact
		newFinger.start, _ = calcFinger(NodeIdBytes, i, networkSize)
		newFinger.end, _ = calcFinger(NodeIdBytes, i+1, networkSize)
		finger[i] = newFinger
	}
	answerChannel = make(chan Contact)
	go listen(contact)

	return *contact
}

func addToRing(ring *Contact) {
	initFingerTable(ring)
	updateOthers()
}

func isResponsible(key string) bool {
	if between(predecessor.NodeId, contact.NodeId, key, false, true) {
		return true
	}

	return false
}

func isAlone() bool {
	return finger[1].responsibleNode.NodeId == contact.NodeId
}

// -------------------------------------------------------------------------------Finger table util
func initFingerTable(ring *Contact) {
	finger[1].responsibleNode = findSuccessor(finger[1].start, ring)
	predecessor = requestFinger(finger[1].responsibleNode, -1)
	setRemoteFinger(finger[1].responsibleNode, -1, contact)
	setRemoteFinger(predecessor, 1, contact)
	for i := 1; i < networkSize; i++ {
		if between(contact.NodeId, finger[i].responsibleNode.NodeId, finger[i+1].start, false, true) {
			finger[i+1].responsibleNode = finger[i].responsibleNode
		} else {
			finger[i+1].responsibleNode = findSuccessor(finger[i+1].start, ring)
		}
	}
	printFingerTable()
}

func updateOthers() {
	var nodeToUpdate Contact
	var newNodeIdInt, power, toSubstract, result, modulo, addressSpace big.Int
	(&newNodeIdInt).SetString(contact.NodeId, 16)
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
		findPredecessor(contact.ContactToString(), targetNode)
		nodeToUpdate = <-answerChannel

		message := new(Message)
		message.Type = "updateFinger"
		message.Src = (contact.ContactToString())
		message.Dest = ((&nodeToUpdate).ContactToString())
		message.Payload = strconv.Itoa(i)
		send(message)
		// nodeToUpdate.updateFingerTable(newNode, i) // TODO: send a message to nodeToUpdate
	}
}

func updateFingerTable(newNode *Contact, index int) {
	if newNode.NodeId == contact.NodeId {
		return
	}

	if between(contact.NodeId, finger[index].responsibleNode.NodeId, newNode.NodeId, false, true) {
		finger[index].responsibleNode = newNode
		message := new(Message)
		message.Type = "updateFinger"
		message.Src = (newNode.ContactToString())
		message.Dest = (predecessor.ContactToString())
		message.Payload = strconv.Itoa(index)
		send(message)
		// predecessor.updateFingerTable(newNode, index) //TODO: send a message
	}
}

// ------------------------------------------------------------------------------------------Lookup
func findSuccessor(id string, dest *Contact) *Contact {
	message := new(Message)
	message.Type = "findPredecessor"
	message.Src = contact.ContactToString()
	message.Dest = dest.ContactToString()
	message.Payload = id
	send(message)

	idPredecessor := <-answerChannel

	// TODO QuickFix:
	if idPredecessor.NodeId == id {
		return &idPredecessor
	}
	return requestFinger(&idPredecessor, 1)
}

func findPredecessor(source, id string) {
	if !between(contact.NodeId, finger[1].responsibleNode.NodeId, id, true, false) {
		message := new(Message)
		message.Type = "findPredecessor"
		message.Src = source
		message.Dest = closestPrecedingFinger(id).ContactToString()
		message.Payload = id
		send(message)
		return
	}

	// TODO: Quick fix
	if id == "" {
		id = "00"
	}

	var responsibleNode string
	if finger[1].responsibleNode.NodeId == id {
		responsibleNode = finger[1].responsibleNode.ContactToString()
	} else {
		responsibleNode = contact.ContactToString()
	}

	message := new(Message)
	message.Type = "answerPredecessor"
	message.Src = responsibleNode
	message.Dest = source
	message.Payload = ""
	send(message)
}

func closestPrecedingFinger(id string) *Contact {
	for i := networkSize; i > 0; i-- {
		if between(contact.NodeId, id, finger[i].responsibleNode.NodeId, false, true) {
			return finger[i].responsibleNode
		}
	}
	return contact // Should never happen
}

// ----------------------------------------------------------------------------------------Printing
func printFingerTable() {
	fmt.Println("Node", contact.NodeId, "finger table is:")
	for i := 1; i < networkSize+1; i++ {
		fmt.Println("finger", i, "start:", finger[i].start, "end: ", finger[i].end, "responsible node:", finger[i].responsibleNode.NodeId)
	}
}

// func (dhtNode *DHTNode) printRing() {
// 	firstNode := dhtNode.findSuccessor("00") // Find first node
// 	currentNode := firstNode

// 	for currentNode.finger[1].responsibleNode != firstNode {
// 		fmt.Println(currentNode.NodeId)
// 		currentNode = currentNode.finger[1].responsibleNode
// 	}
// 	fmt.Println(currentNode.NodeId)
// }

// -------------------------------------------------------------------------------Net communication

func (contact *Contact) String() string {
	return contact.Ip + ":" + contact.Port
}

func (message *Message) String() string {
	return message.Type + " - " + message.Src + " - " + message.Dest + " - " + message.Payload
}

func (contact *Contact) ContactToString() string {
	return contact.Ip + "-" + contact.Port + "-" + contact.NodeId
}

func StringToContact(stringContact string) Contact {
	res := strings.Split(stringContact, "-")
	contact := new(Contact)
	contact.Ip = res[0]
	contact.Port = res[1]
	contact.NodeId = res[2]
	return *contact
}

func listen(self *Contact) {
	contact = self
	udpAddress, err := net.ResolveUDPAddr("udp4", self.String())
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

	for {
		message := new(Message)
		err = decoder.Decode(message)
		if err != nil {
			fmt.Println("Unvalid message format", self.Port)
		}
		switch message.Type {
		case "requestFinger":
			fmt.Println("asking for a finger")
			index, _ := strconv.Atoi(message.Payload)
			answer := new(Message)
			answer.Type = "answerFinger"
			if index >= 0 {
				answer.Src = finger[index].responsibleNode.ContactToString()
			} else {
				answer.Src = predecessor.ContactToString()
			}
			answer.Dest = message.Src
			answer.Payload = ""
			send(answer)

		case "updateFinger":
			fmt.Println("requesting to update a finger")
			index, _ := strconv.Atoi(message.Payload)
			source := StringToContact(message.Src)
			go updateFingerTable(&source, index)

		case "setFinger":
			fmt.Println("requesting to set a finger")
			index, _ := strconv.Atoi(message.Payload)
			source := StringToContact(message.Src)
			if index >= 0 {
				finger[index].responsibleNode = &source //TODO Pb here maybe
			} else {
				predecessor = &source
			}

		case "joinRing":
			fmt.Println(StringToContact(message.Dest), "will join the ring")
			source := StringToContact(message.Src)
			go addToRing(&source)

		case "answerFinger":
			answerChannel <- StringToContact(message.Src)

		case "findPredecessor":
			go findPredecessor(message.Src, message.Payload)

		case "answerPredecessor":
			answerChannel <- StringToContact(message.Src)
		}
	}
}

func send(message *Message) {
	dest := StringToContact(message.Dest)
	udpAddress, err := net.ResolveUDPAddr("udp4", (&dest).String())
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
	message.Src = contact.ContactToString()
	message.Dest = peer.ContactToString()
	message.Payload = strconv.Itoa(fingerIndex)
	send(message)

	//TODO timeout that shit
	answer := <-answerChannel
	return &answer
}

func setRemoteFinger(peer *Contact, fingerIndex int, newContact *Contact) {
	message := new(Message)
	message.Type = "setFinger"
	message.Src = newContact.ContactToString()
	message.Dest = peer.ContactToString()
	message.Payload = strconv.Itoa(fingerIndex)
	send(message)
}
