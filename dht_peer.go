package dht

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"
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
var secondPredecessor *Contact
var contact *Contact
var finger []*DHTFinger
var networkSize int
var answerChannel chan Contact
var pongChannel chan bool
var healedCircleChannel chan bool
var fileChannel chan []byte

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

	fmt.Println("Hey, i'm", contact.NodeId, "and run on port", Port)

	predecessor = contact
	secondPredecessor = contact

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
	pongChannel = make(chan bool)
	healedCircleChannel = make(chan bool)
	fileChannel = make(chan []byte)
	go listen(contact)

	return *contact
}

func addToRing(ring *Contact) {
	initFingerTable(ring)
	updateOthers()
	balanceStorage()
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
	secondPredecessor = requestFinger(predecessor, -1)
	secondSuccessor := requestFinger(finger[1].responsibleNode, 1)
	setRemoteFinger(secondSuccessor, -2, contact)
	for i := 1; i < networkSize; i++ {
		if between(contact.NodeId, finger[i].responsibleNode.NodeId, finger[i+1].start, false, true) {
			finger[i+1].responsibleNode = finger[i].responsibleNode
		} else {
			finger[i+1].responsibleNode = findSuccessor(finger[i+1].start, ring)
		}
	}
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

		message := createMessage("updateFinger", contact.ContactToString(), (&nodeToUpdate).ContactToString(), strconv.Itoa(i))
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
		message := createMessage("updateFinger", newNode.ContactToString(), predecessor.ContactToString(), strconv.Itoa(index))

		send(message)
		// predecessor.updateFingerTable(newNode, index) //TODO: send a message
	}
}

// ------------------------------------------------------------------------------------------Lookup
func findSuccessor(id string, dest *Contact) *Contact {
	message := createMessage("findPredecessor", contact.ContactToString(), dest.ContactToString(), id)
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
		message := createMessage("findPredecessor", source, closestPrecedingFinger(id).ContactToString(), id)
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

	message := createMessage("answerPredecessor", responsibleNode, source, "")
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

//pseduo csv
func (message *Message) String() string {
	sep := ","
	return message.Type + sep + message.Src + sep + message.Dest + sep + message.Payload
}

func (contact *Contact) ContactToString() string {
	return contact.Ip + "-" + contact.Port + "-" + contact.NodeId
}

func StringToMessage(s string) *Message {
	data := strings.SplitN(s, ",", 4)

	message := new(Message)
	message.Type = data[0]
	message.Src = data[1]
	message.Dest = data[2]
	message.Payload = data[3]
	return message
}

func StringToContact(stringContact string) Contact {
	res := strings.SplitN(stringContact, "-", 3)
	contact := new(Contact)
	contact.Ip = res[0]
	contact.Port = res[1]
	contact.NodeId = res[2]
	return *contact
}

func createMessage(msgType, source, dest, payload string) *Message {
	message := new(Message)
	message.Type = msgType
	message.Src = source
	message.Dest = dest
	message.Payload = payload

	return message
}

func listen(self *Contact) {
	contact = self
	tcpAddress, err := net.ResolveTCPAddr("tcp4", self.String())
	if err != nil {
		fmt.Println("Error while resolving", tcpAddress)
	}

	tcpListener, err := net.ListenTCP("tcp4", tcpAddress)
	if err != nil {
		fmt.Println("Error while listening to", tcpAddress)
		fmt.Println(err)
	}
	defer tcpListener.Close()

	for {
		connection, err := tcpListener.Accept()
		if err != nil {
			fmt.Println("Error while accepting connection")
		}

		var buf bytes.Buffer
		io.Copy(&buf, connection)
		connection.Close()
		if err != nil {
			fmt.Println("Error while reading incoming message")
		}
		message := StringToMessage(string(buf.Bytes()))
		switch message.Type {
		case "requestFinger":
			index, _ := strconv.Atoi(message.Payload)
			source := StringToContact(message.Src)
			go requestFingerHandler(&source, index)

		case "answerFinger":
			answerChannel <- StringToContact(message.Src)

		case "updateFinger":
			index, _ := strconv.Atoi(message.Payload)
			source := StringToContact(message.Src)
			go updateFingerTable(&source, index)

		case "setFinger":
			index, _ := strconv.Atoi(message.Payload)
			source := StringToContact(message.Src)
			if index >= 0 {
				finger[index].responsibleNode = &source //TODO Pb here maybe
			} else if index == -1 {
				predecessor = &source
			} else if index == -2 {
				secondPredecessor = &source
			}

		case "findPredecessor":
			go findPredecessor(message.Src, message.Payload)

		case "answerPredecessor":
			answerChannel <- StringToContact(message.Src)

		case "requestId":
			source := StringToContact(message.Src)
			go requestIdHandler(&source)

		case "answerId":
			answerChannel <- StringToContact(message.Src)

		case "storeData":
			data := strings.SplitN(message.Payload, "/", 2)
			go storeDataHandler(data[0], data[1])

		case "storeKeyValue":
			keyValue := strings.SplitN(message.Payload, "-", 2)
			source := StringToContact(message.Src)
			go storeKeyValue(keyValue[0], keyValue[1], &source)

		case "ping":
			source := StringToContact(message.Src)
			go pingHandler(&source)

		case "pong":
			pongChannel <- true

		case "printRing":
			go printRingHandler(message.Payload)

		case "changeResponsibility":
			go changeResponsibilityHandler()

		case "requestFile":
			source := StringToContact(message.Src)
			go requestFileHandler(message.Payload, &source)

		case "answerFile":
			fileChannel <- []byte(message.Payload)

		case "deleteFile":
			go deleteFileHandler(message.Payload)

		case "checkPredecessor":
			source := StringToContact(message.Src)
			go checkPredecessor(&source)

		case "nodeLeave":
			source := StringToContact(message.Src)
			data := strings.SplitN(message.Payload, "/", 2)
			index, _ := strconv.Atoi(data[0])
			deadNode := StringToContact(data[1])
			go nodeLeaveHandler(&source, &deadNode, index)

		case "circleHealed":
			healedCircleChannel <- true
		}
	}
}

func send(message *Message) {
	if message.Type != "ping" && message.Type != "pong" && message.Type != "requestId" {
		ping(message.Src, message.Dest)
	}

	dest := StringToContact(message.Dest)
	tcpAddress, err := net.ResolveTCPAddr("tcp4", (&dest).String())
	if err != nil {
		fmt.Println("Error while resolving", tcpAddress)
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		fmt.Println("Error while connecting to", tcpAddress)
		return
	}
	defer conn.Close()

	buf := bytes.NewBuffer([]byte(message.String()))
	io.Copy(conn, buf)
	// _, err = conn.Write([]byte(message.String()))
	if err != nil {
		fmt.Println("Error while sending a message", tcpAddress)
		fmt.Println("Error is", err)
		return
	}
}

func ping(src, dest string) {
	timeoutChannel := make(chan bool)
	pingMessage := createMessage("ping", contact.ContactToString(), dest, "")
	send(pingMessage)
	go timeout(timeoutChannel, 3)
	select {
	case <-pongChannel:
		break
	case <-timeoutChannel:
		fmt.Println(dest, ": host is dead")
		if dest == predecessor.ContactToString() {
			healCircle(src)
		} else {
			checkPredecessor(contact)
			<-healedCircleChannel
			ping(src, dest)
		}
	}
}

func healCircle(src string) {
	deadNode := predecessor
	predecessor = secondPredecessor
	secondPredecessor = requestFinger(predecessor, -1)

	// Take responisbility for previous interval
	copyDir := "storage/" + contact.NodeId + "/Copy/"
	mainDir := "storage/" + contact.NodeId + "/MainStorage/"
	if !exists(mainDir) {
		os.Mkdir(mainDir, 744)
	}
	files, _ := ioutil.ReadDir(copyDir)
	for _, file := range files {
		key := file.Name()
		os.Rename(copyDir+key, mainDir+key)
	}

	var nodeToUpdate Contact
	var deadNodeIdInt, power, toSubstract, result, modulo, addressSpace big.Int
	(&deadNodeIdInt).SetString(contact.NodeId, 16)
	for i := 1; i < networkSize+1; i++ {
		power = *big.NewInt(int64(i - 1))
		(&toSubstract).Exp(big.NewInt(2), &power, nil)
		(&result).Sub(&deadNodeIdInt, &toSubstract)
		(&addressSpace).Exp(big.NewInt(2), big.NewInt(int64(networkSize)), nil)
		(&modulo).Mod(&result, &addressSpace)
		targetNode := hex.EncodeToString(modulo.Bytes())
		if targetNode == "" {
			targetNode = "00"
		}
		//targetNode = (deadNode - 2 ^ (i-1)) mod 2 ^ m
		findPredecessor(contact.ContactToString(), targetNode)
		nodeToUpdate = <-answerChannel

		message := createMessage("nodeLeave", contact.ContactToString(), (&nodeToUpdate).ContactToString(), strconv.Itoa(i)+"/"+deadNode.ContactToString())
		send(message)
	}

	message := createMessage("circleHealed", contact.ContactToString(), src, "")
	send(message)
}

func nodeLeaveHandler(source, deadNode *Contact, index int) {
	if finger[index].responsibleNode.ContactToString() == deadNode.ContactToString() {
		finger[index].responsibleNode = source
		message := createMessage("nodeLeave", source.ContactToString(), predecessor.ContactToString(), strconv.Itoa(index)+"/"+deadNode.ContactToString())
		send(message)

		if index == 1 { //That means you were the predecessor of the dead node
			dirname := "storage/" + contact.NodeId + "/MainStorage/"
			files, _ := ioutil.ReadDir(dirname)
			for _, file := range files {
				key := file.Name()
				valueBytes, _ := ioutil.ReadFile(dirname + key)
				value := string(valueBytes)
				message := createMessage("storeKeyValue", contact.ContactToString(), finger[1].responsibleNode.ContactToString(), key+"/"+value)
				send(message)
			}
		}
	}
}

func checkPredecessor(source *Contact) {
	message := createMessage("checkPredecessor", source.ContactToString(), predecessor.ContactToString(), "")
	send(message)
}

func requestFinger(peer *Contact, fingerIndex int) *Contact {
	message := createMessage("requestFinger", contact.ContactToString(), peer.ContactToString(), strconv.Itoa(fingerIndex))
	send(message)

	answer := <-answerChannel
	return &answer
}

func requestFingerHandler(source *Contact, index int) {
	answer := createMessage("answerFinger", predecessor.ContactToString(), source.ContactToString(), "")
	if index >= 0 {
		answer.Src = finger[index].responsibleNode.ContactToString()
	}
	send(answer)
}

func requestIdHandler(source *Contact) {
	answer := createMessage("answerId", contact.ContactToString(), source.ContactToString(), "")
	send(answer)
}

func setRemoteFinger(peer *Contact, fingerIndex int, newContact *Contact) {
	message := createMessage("setFinger", newContact.ContactToString(), peer.ContactToString(), strconv.Itoa(fingerIndex))
	send(message)
}

func storeDataHandler(filename, value string) {
	key := generateHashCode(filename)
	responsibleNode := findSuccessor(key, contact)
	message := createMessage("storeKeyValue", contact.ContactToString(), responsibleNode.ContactToString(), (key + "-" + value))
	send(message)

	// Store the data as you are the backup responsible node
	if *predecessor == *responsibleNode {
		go storeKeyValue(key, value, contact)
	}
}

func pingHandler(source *Contact) {
	pongMessage := createMessage("pong", contact.ContactToString(), source.ContactToString(), "")
	send(pongMessage)
}

func printRingHandler(list string) {
	addresses := strings.Split(list, "\n")
	if addresses[0] == contact.NodeId {
		fmt.Println(list)
	} else {
		message := createMessage("printRing", contact.ContactToString(), finger[1].responsibleNode.ContactToString(), list+contact.NodeId+"\n")
		send(message)
	}
}

func changeResponsibilityHandler() {
	// Hand over the files the new node is responsible for
	dirname := "storage/" + contact.NodeId + "/MainStorage/"
	files, _ := ioutil.ReadDir(dirname)
	predecessorId := predecessor.NodeId
	for _, file := range files {
		key := file.Name()
		if !between(predecessorId, contact.NodeId, key, false, true) {
			valueByte, _ := ioutil.ReadFile(dirname + key)
			value := string(valueByte)
			go storeKeyValue(key, value, contact)
			message := createMessage("storeKeyValue", contact.ContactToString(), predecessor.ContactToString(), key+"-"+value)
			send(message)
			os.Remove(dirname + key)
		}
	}

	// Hand over the replication of the previous predecessor's file
	dirname = "storage/" + contact.NodeId + "/Copy/"
	files, _ = ioutil.ReadDir(dirname)
	for _, file := range files {
		key := file.Name()
		valueByte, _ := ioutil.ReadFile(dirname + key)
		value := string(valueByte)
		message := createMessage("storeKeyValue", contact.ContactToString(), predecessor.ContactToString(), key+"-"+value)
		send(message)
		os.Remove(dirname + key)
	}
}

func requestFileHandler(key string, source *Contact) {
	if isResponsible(key) {
		dirname := "storage/" + contact.NodeId + "/MainStorage/"
		var fileAsString string
		if exists(dirname + key) {
			file, _ := ioutil.ReadFile(dirname + key)
			fileAsString = string(file)
		} else {
			fileAsString = ""
		}
		answer := createMessage("answerFile", contact.ContactToString(), source.ContactToString(), fileAsString)
		send(answer)
	} else {
		responsibleNode := findSuccessor(key, contact)
		message := createMessage("requestFile", source.ContactToString(), responsibleNode.ContactToString(), key)
		send(message)
	}
}

func deleteFileHandler(key string) {
	if isResponsible(key) {
		dirname := "storage/" + contact.NodeId + "/MainStorage/"
		if exists(dirname + key) {
			os.Remove(dirname + key)
		}
	} else {
		responsibleNode := findSuccessor(key, contact)
		message := createMessage("deleteFile", contact.ContactToString(), responsibleNode.ContactToString(), key)
		send(message)
	}
}

// ------------------------------------------------------------------------------------Data Storage
func storeKeyValue(key, value string, source *Contact) {
	byteValue := []byte(value)
	folder, _ := os.Getwd()
	responsible := isResponsible(key)
	if responsible {
		folder += "/storage/" + contact.NodeId + "/MainStorage/"
	} else {
		folder += "/storage/" + contact.NodeId + "/Copy/"
	}
	if !exists(folder) {
		fmt.Println("That does not exist yet")
		os.MkdirAll(folder, 0744)
	}

	filename := folder + key
	err := ioutil.WriteFile(filename, byteValue, 0644)
	if err != nil {
		fmt.Println("Error while storing file")
	}

	// If you get the message from your successor, assume he took care of
	// backuping (to prevent overload during balancing)
	if responsible && *source != *finger[1].responsibleNode {
		message := createMessage("storeKeyValue", contact.ContactToString(), finger[1].responsibleNode.ContactToString(), (key + "-" + value))
		send(message)
	}
}

func balanceStorage() {
	message := createMessage("changeResponsibility", contact.ContactToString(), finger[1].responsibleNode.ContactToString(), "")
	send(message)
}

// ------------------------------------------------------------------------------------Exposed services
func getFile(filename string) (file []byte, err bool) {
	message := createMessage("requestFile", contact.ContactToString(), contact.ContactToString(), generateHashCode(filename))
	send(message)

	file = <-fileChannel
	fileAsString := string(file)

	if fileAsString == "" {
		err = true
	} else {
		err = false
	}
	return file, err
}

func deleteFile(filename string) {
	message := createMessage("deleteFile", contact.ContactToString(), contact.ContactToString(), generateHashCode(filename))
	send(message)
}

func storeFile(filename string, file []byte) {
	fileAsString := string(file)
	message := createMessage("storeData", contact.ContactToString(), contact.ContactToString(), filename+"/"+fileAsString)
	send(message)
}

func joinRing(dest *Contact) bool {
	pingMessage := createMessage("ping", contact.ContactToString(), dest.ContactToString(), "")
	send(pingMessage)
	timeoutChannel := make(chan bool)
	go timeout(timeoutChannel, 3)
	select {
	case <-pongChannel:
		break
	case <-timeoutChannel:
		return false
	}

	idRequest := createMessage("requestId", contact.ContactToString(), dest.ContactToString(), "")
	send(idRequest)

	destWithId := <-answerChannel
	dest = &destWithId
	addToRing(dest)
	return true
}
