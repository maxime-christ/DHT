package dht

import (
	"fmt"
	// "github.com/nu7hatch/gouuid"
	"encoding/hex"
	"math/big"
	"testing"
)

func TestFindPredecessor(t *testing.T) {
	id0 := "00"
	id2 := "02"
	id5 := "05"

	node0 := makeDHTNode(&id0, "localhost", "1111", 3)
	node2 := makeDHTNode(&id2, "localhost", "1112", 3)
	node5 := makeDHTNode(&id5, "localhost", "1115", 3)

	node5.addToRing(node2)
	node2.addToRing(node0)

	for i := big.NewInt(0); i.Cmp(big.NewInt(7)) == -1; i.Add(i, big.NewInt(1)) {
		fmt.Println("predecessor of", i, "is", node0.findPredecessor(hex.EncodeToString(i.Bytes())).nodeId)
	}
}

func TestDHT1(t *testing.T) {
	id0 := "00"
	id1 := "01"
	id2 := "02"
	id3 := "03"
	id4 := "04"
	id5 := "05"
	id6 := "06"
	id7 := "07"

	node0b := makeDHTNode(&id0, "localhost", "1111", 3)
	node1b := makeDHTNode(&id1, "localhost", "1112", 3)
	node2b := makeDHTNode(&id2, "localhost", "1113", 3)
	node3b := makeDHTNode(&id3, "localhost", "1114", 3)
	node4b := makeDHTNode(&id4, "localhost", "1115", 3)
	node5b := makeDHTNode(&id5, "localhost", "1116", 3)
	node6b := makeDHTNode(&id6, "localhost", "1117", 3)
	node7b := makeDHTNode(&id7, "localhost", "1118", 3)

	// var fakeid *string
	// node0b := makeDHTNode(fakeid, "localhost", "1111")
	// node1b := makeDHTNode(fakeid, "localhost", "1112")
	// node2b := makeDHTNode(fakeid, "localhost", "1113")
	// node3b := makeDHTNode(fakeid, "localhost", "1114")
	// node4b := makeDHTNode(fakeid, "localhost", "1115")
	// node5b := makeDHTNode(fakeid, "localhost", "1116")
	// node6b := makeDHTNode(fakeid, "localhost", "1117")
	// node7b := makeDHTNode(fakeid, "localhost", "1118")

	node0b.addToRing(node1b)
	node1b.addToRing(node3b)
	node3b.addToRing(node6b)
	node1b.addToRing(node4b)
	node4b.addToRing(node5b)
	node3b.addToRing(node7b)
	node1b.addToRing(node2b)

	fmt.Println("-> ring structure")
	node1b.printRing()

	node1b.printFingerTable()
	node2b.printFingerTable()
	node3b.printFingerTable()
	node4b.printFingerTable()
	node5b.printFingerTable()
	node6b.printFingerTable()
	node7b.printFingerTable()
}

func TestDHT2(t *testing.T) {
	id := "009621131631ec5cebdd70f93aaecc2ca9c7f380"
	node1 := makeDHTNode(&id, "localhost", "1111", 160)
	node2 := makeDHTNode(nil, "localhost", "1112", 160)
	node3 := makeDHTNode(nil, "localhost", "1113", 160)
	node4 := makeDHTNode(nil, "localhost", "1114", 160)
	node5 := makeDHTNode(nil, "localhost", "1115", 160)
	node6 := makeDHTNode(nil, "localhost", "1116", 160)
	node7 := makeDHTNode(nil, "localhost", "1117", 160)
	node8 := makeDHTNode(nil, "localhost", "1118", 160)
	node9 := makeDHTNode(nil, "localhost", "1119", 160)

	key1 := "2b230fe12d1c9c60a8e489d028417ac89de57635"
	key2 := "87adb987ebbd55db2c5309fd4b23203450ab0083"
	key3 := "74475501523a71c34f945ae4e87d571c2c57f6f3"

	fmt.Println("TEST: " + node1.findSuccessor(key1).nodeId + " is responsible for " + key1)
	fmt.Println("TEST: " + node1.findSuccessor(key2).nodeId + " is responsible for " + key2)
	fmt.Println("TEST: " + node1.findSuccessor(key3).nodeId + " is responsible for " + key3)

	fmt.Println("adding Node2")
	node1.addToRing(node2)
	fmt.Println("adding Node3")
	node1.addToRing(node3)
	fmt.Println("adding Node4")
	node1.addToRing(node4)
	fmt.Println("adding Node5")
	node4.addToRing(node5)
	fmt.Println("adding Node6")
	node3.addToRing(node6)
	fmt.Println("adding Node7")
	node3.addToRing(node7)
	fmt.Println("adding Node8")
	node3.addToRing(node8)
	fmt.Println("adding Node9")
	node7.addToRing(node9)

	node1.printFingerTable()
	node2.printFingerTable()
	node3.printFingerTable()
	node4.printFingerTable()
	node5.printFingerTable()
	node6.printFingerTable()
	node7.printFingerTable()
	node8.printFingerTable()
	node9.printFingerTable()

	fmt.Println("-> ring structure")
	node1.printRing()

	nodeForKey1 := node1.findSuccessor(key1)
	fmt.Println("dht node " + nodeForKey1.nodeId + " running at " + nodeForKey1.contact.ip + ":" + nodeForKey1.contact.port + " is responsible for " + key1)

	nodeForKey2 := node1.findSuccessor(key2)
	fmt.Println("dht node " + nodeForKey2.nodeId + " running at " + nodeForKey2.contact.ip + ":" + nodeForKey2.contact.port + " is responsible for " + key2)

	nodeForKey3 := node1.findSuccessor(key3)
	fmt.Println("dht node " + nodeForKey3.nodeId + " running at " + nodeForKey3.contact.ip + ":" + nodeForKey3.contact.port + " is responsible for " + key3)
}
