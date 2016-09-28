package dht

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
)

type Contact struct {
	ip   string
	port string
}

type DHTNode struct {
	nodeId      string
	predecessor *DHTNode
	contact     Contact
	finger      []*DHTFinger
	networkSize int
}

type DHTFinger struct {
	start           string
	end             string
	responsibleNode *DHTNode
}

func makeDHTNode(nodeId *string, ip string, port string, networkSize int) *DHTNode {
	dhtNode := new(DHTNode)
	dhtNode.contact.ip = ip
	dhtNode.contact.port = port
	dhtNode.networkSize = networkSize

	if nodeId == nil {
		genNodeId := generateNodeId()
		dhtNode.nodeId = genNodeId
	} else {
		dhtNode.nodeId = *nodeId
	}

	dhtNode.predecessor = dhtNode

	nodeIdBytes, _ := hex.DecodeString(dhtNode.nodeId)

	dhtNode.finger = make([]*DHTFinger, networkSize+1)

	finger0 := new(DHTFinger)
	finger0.responsibleNode = dhtNode
	finger0.start = dhtNode.nodeId
	finger0.end = dhtNode.nodeId
	dhtNode.finger[0] = finger0

	for i := 1; i < networkSize+1; i++ {
		finger := new(DHTFinger)
		finger.responsibleNode = dhtNode
		finger.start, _ = calcFinger(nodeIdBytes, i, dhtNode.networkSize)
		finger.end, _ = calcFinger(nodeIdBytes, i+1, dhtNode.networkSize)
		dhtNode.finger[i] = finger
	}

	return dhtNode
}

func (dhtNode *DHTNode) addToRing(newDHTNode *DHTNode) {
	newDHTNode.initFingerTable(dhtNode)
	newDHTNode.updateOthers()
}

func (dhtNode *DHTNode) lookup(key string) (responsibleNode *DHTNode) {
	fmt.Println("lookup call, biatch", dhtNode.nodeId) // TODO remove
	if dhtNode.isResponsible(key) {
		responsibleNode = dhtNode
	} else {
		nodeIdBytes, _ := hex.DecodeString(dhtNode.nodeId)
		keyBytes, error := hex.DecodeString(key)
		if error != nil {
			fmt.Println("The key you are looking for is not readable!", key)
		}

		if bytes.Compare(keyBytes, nodeIdBytes) == 1 {
			responsibleNode = dhtNode.finger[1].responsibleNode.lookup(key)
		} else {
			responsibleNode = dhtNode.predecessor.lookup(key)
		}
	}
	return responsibleNode
}

func (dhtNode *DHTNode) acceleratedLookupUsingFingers(key string) (responsibleNode *DHTNode) {
	fmt.Println("acceleratedLookup call, biatch", dhtNode.nodeId) // TODO remove
	if dhtNode.isResponsible(key) {
		fmt.Println("acceleratedLookup: Hey, mais c'est moi!!")
		responsibleNode = dhtNode
	} else {
		fmt.Println("Wait a minute, i'm asking", dhtNode.finger[1].responsibleNode.nodeId)
		responsibleNode = dhtNode.finger[1].responsibleNode.acceleratedLookupUsingFingers(key)
	}
	return responsibleNode
}

func (dhtNode *DHTNode) isResponsible(key string) bool {
	if between(dhtNode.predecessor.nodeId, dhtNode.nodeId, key, false, true) {
		return true
	}

	return false
}

func (dhtNode *DHTNode) testCalcFingers(m int, bits int) {
	idBytes, _ := hex.DecodeString(dhtNode.nodeId)
	fingerHex, _ := calcFinger(idBytes, m, bits)
	fingerSuccessor := dhtNode.lookup(fingerHex)
	fingerSuccessorBytes, _ := hex.DecodeString(fingerSuccessor.nodeId)
	fmt.Println("successor    " + fingerSuccessor.nodeId)

	dist := distance(idBytes, fingerSuccessorBytes, bits)
	fmt.Println("distance     " + dist.String())
}

func (dhtNode *DHTNode) isAlone() bool {
	return dhtNode.finger[1].responsibleNode.nodeId == dhtNode.nodeId
}

func (joiningDhtNode *DHTNode) initFingerTable(ring *DHTNode) {
	// fmt.Println("initFingerTable of", joiningDhtNode.nodeId)

	joiningDhtNode.finger[1].responsibleNode = ring.acceleratedLookupUsingFingers(joiningDhtNode.finger[1].start)

	joiningDhtNode.predecessor = joiningDhtNode.finger[1].responsibleNode.predecessor
	// fmt.Println("initFingerTable: joiningDhtNode.predecessor =", joiningDhtNode.predecessor.nodeId)
	joiningDhtNode.finger[1].responsibleNode.predecessor = joiningDhtNode
	// fmt.Println("initFingerTable: joiningDhtNode.successor.predecessor =", joiningDhtNode.finger[1].responsibleNode.predecessor.nodeId)

	for i := 1; i < joiningDhtNode.networkSize; i++ {
		// fmt.Println("initFingerTable: finger", i)
		// fmt.Println("initFingerTable: finger[i+1].start is", joiningDhtNode.finger[i+1].start, "finger[i] is", joiningDhtNode.finger[i].responsibleNode.nodeId)
		if between(joiningDhtNode.nodeId, joiningDhtNode.finger[i].responsibleNode.nodeId, joiningDhtNode.finger[i+1].start, true, false) {
			joiningDhtNode.finger[i+1].responsibleNode = joiningDhtNode.finger[i].responsibleNode
		} else {
			joiningDhtNode.finger[i+1].responsibleNode = ring.findSuccessor(joiningDhtNode.finger[i+1].start)
		}
	}

	joiningDhtNode.printFingerTable()
}

func (newNode *DHTNode) updateOthers() {
	var nodeToUpdate *DHTNode
	var newNodeIdInt, power, toSubstract, result, modulo, addressSpace big.Int
	(&newNodeIdInt).SetString(newNode.nodeId, 16)
	for i := 1; i < newNode.networkSize+1; i++ {
		power = *big.NewInt(int64(i - 1))
		(&toSubstract).Exp(big.NewInt(2), &power, nil)
		(&result).Sub(&newNodeIdInt, &toSubstract)
		(&addressSpace).Exp(big.NewInt(2), big.NewInt(int64(newNode.networkSize)), nil)
		(&modulo).Mod(&result, &addressSpace)
		targetNode := hex.EncodeToString(modulo.Bytes())
		if targetNode == "" {
			targetNode = "00"
		}
		//targetNode = (newNode - 2 ^ (i-1)) mod 2 ^ m
		nodeToUpdate = newNode.findPredecessor(targetNode)
		fmt.Println("node to update is", targetNode)
		fmt.Println("His predecessor is", nodeToUpdate.nodeId)
		nodeToUpdate.updateFingerTable(newNode, i)
	}
}

func (nodeToUpdate *DHTNode) updateFingerTable(newNode *DHTNode, index int) {
	fmt.Println("updating", nodeToUpdate.nodeId, "concerning finger", index, "resp is actually", nodeToUpdate.finger[index].responsibleNode.nodeId)
	if between(nodeToUpdate.nodeId, nodeToUpdate.finger[index].responsibleNode.nodeId, newNode.nodeId, true, false) {
		fmt.Println("changed to", newNode.nodeId)
		nodeToUpdate.finger[index].responsibleNode = newNode
		predecessor := nodeToUpdate.predecessor
		nodeToUpdate.printFingerTable()
		predecessor.updateFingerTable(newNode, index)
	}
}

func (dhtNode *DHTNode) findSuccessor(id string) *DHTNode {
	predecessor := dhtNode.findPredecessor(id)
	return predecessor.finger[1].responsibleNode
}

func (dhtNode *DHTNode) findPredecessor(id string) *DHTNode {
	nodeIterator := dhtNode

	for !between(nodeIterator.nodeId, nodeIterator.finger[1].responsibleNode.nodeId, id, true, false) {
		nodeIterator = nodeIterator.closestPrecedingFinger(id)
	}

	return nodeIterator
}

func (dhtNode *DHTNode) closestPrecedingFinger(id string) *DHTNode {
	for i := dhtNode.networkSize; i > 0; i-- {
		if between(dhtNode.nodeId, id, dhtNode.finger[i].responsibleNode.nodeId, false, false) {
			return dhtNode.finger[i].responsibleNode
		}
	}
	return dhtNode
}

func (dhtNode *DHTNode) printFingerTable() {
	fmt.Println("Hi, I'm", dhtNode.nodeId, "and this is jackass")
	for i := 0; i < dhtNode.networkSize+1; i++ {
		fmt.Println("start:", dhtNode.finger[i].start, "end: ", dhtNode.finger[i].end, "responsible node:", dhtNode.finger[i].responsibleNode.nodeId)
	}
}

func (dhtNode *DHTNode) printRing() {
	firstNode := dhtNode.acceleratedLookupUsingFingers("09") // TODO ... srsly?
	currentNode := firstNode

	for currentNode.finger[1].responsibleNode != firstNode {
		fmt.Println(currentNode.nodeId)
		currentNode = currentNode.finger[1].responsibleNode
	}
	fmt.Println(currentNode.nodeId)
}
