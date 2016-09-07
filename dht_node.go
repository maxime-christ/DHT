package dht

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type Contact struct {
	ip   string
	port string
}

type DHTNode struct {
	nodeId      string
	successor   *DHTNode
	predecessor *DHTNode
	contact     Contact
}

func makeDHTNode(nodeId *string, ip string, port string) *DHTNode {
	dhtNode := new(DHTNode)
	dhtNode.contact.ip = ip
	dhtNode.contact.port = port

	if nodeId == nil {
		genNodeId := generateNodeId()
		dhtNode.nodeId = genNodeId
	} else {
		dhtNode.nodeId = *nodeId
	}

	dhtNode.successor = nil
	dhtNode.predecessor = nil

	return dhtNode
}

func (dhtNode *DHTNode) addToRing(newDHTNode *DHTNode) {
	if dhtNode.isAlone() {
		// There's only one node in the ring
		dhtNode.predecessor = newDHTNode
		dhtNode.successor = newDHTNode
		newDHTNode.predecessor = dhtNode
		newDHTNode.successor = dhtNode
	} else {
		// Several nodes already in the ring
		targetNextNode := dhtNode.acceleratedLookupUsingFingers(newDHTNode.nodeId)
		targetPreviousNode := targetNextNode.predecessor

		newDHTNode.successor = targetNextNode
		newDHTNode.predecessor = targetPreviousNode
		targetPreviousNode.successor = newDHTNode
		targetNextNode.predecessor = newDHTNode
	}
}

func (dhtNode *DHTNode) lookup(key string) (responsibleNode *DHTNode) {
	fmt.Println("lookup call, biatch", dhtNode.nodeId)
	if dhtNode.responsible(key) {
		responsibleNode = dhtNode
	} else {
		nodeIdBytes, _ := hex.DecodeString(dhtNode.nodeId)
		keyBytes, error := hex.DecodeString(key)
		if error != nil {
			fmt.Println("The key you are looking for is not readable!", key)
		}

		if bytes.Compare(keyBytes, nodeIdBytes) == 1 {
			responsibleNode = dhtNode.successor.lookup(key)
		} else {
			responsibleNode = dhtNode.predecessor.lookup(key)
		}
	}
	return responsibleNode
}

func (dhtNode *DHTNode) acceleratedLookupUsingFingers(key string) *DHTNode {
	// TODO
	return dhtNode.lookup(key) // XXX This is not correct obviously
}

func (dhtNode *DHTNode) responsible(key string) bool {
	if dhtNode.isAlone() {
		// if it is alone, dhtNode.predecessor next will fail
		return true
	}

	keyBytes, error := hex.DecodeString(key)
	if error != nil {
		fmt.Println("The key you are looking for is not readable!", key)
	}

	predecessorIdBytes, error := hex.DecodeString(dhtNode.predecessor.nodeId)
	nodeIdBytes, _ := hex.DecodeString(dhtNode.nodeId)

	if between(predecessorIdBytes, nodeIdBytes, keyBytes) {
		return true
	}

	return false
}

func (dhtNode *DHTNode) printRing() {
	firstNode := dhtNode.acceleratedLookupUsingFingers("09")
	currentNode := firstNode

	for currentNode.successor != firstNode {
		fmt.Println(currentNode.nodeId)
		currentNode = currentNode.successor
	}
	fmt.Println(currentNode.nodeId)
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
	return dhtNode.successor == nil && dhtNode.predecessor == nil
}
