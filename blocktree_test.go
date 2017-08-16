package main 

import(
	"testing"
	// "fmt"
)

func TestEqualBTNodes(t *testing.T){
	b1 := BTNode{0, &genesisNode, []byte{0}, "Data", []byte{0}}
	b2 := BTNode{0, &genesisNode, []byte{0}, "Data", []byte{0}}

	eq := equalBTNodes(b1, b2)

	if eq != true {
		t.Error("Identical nodes shoul be equal")
	}
}

func TestCalcBTNodeHash(t *testing.T) {}

func TestIsValidNextBTNode(t *testing.T){
	b1                   := BTNode{0, &genesisNode, []byte{0}, "Data", []byte{0}}
	b1.calcBTNodeHash()
	b2Valid              := BTNode{1, &b1, b1.Hash, "Data", []byte{0}}
	b2InvalidHeight      := BTNode{0, &b1, b1.Hash, "Data", []byte{0}}
	b2InvalidParent      := BTNode{1, &genesisNode, b1.Hash, "Data", []byte{0}}
	b2InvalidParentHash  := BTNode{1, &b1, []byte{0}, "Data", []byte{0}}

	if b1.isValidNextBTNode(&b2Valid) != true {
	 	t.Error("Valid node does not validate")
	}
	if b1.isValidNextBTNode(&b2InvalidHeight) != false {
	 	t.Error("Invalid height validates")
	}
	if b1.isValidNextBTNode(&b2InvalidParent) != false {
	 	t.Error("Invalid parent validates")
	}
	if b1.isValidNextBTNode(&b2InvalidParentHash) != false {
	 	t.Error("Invalid parent hash validates")
	}
}

func TestAddBTNodeIfValid(t *testing.T){
	// create BlockTree with only genesis node
	var bt BlockTree
	var levelZero []*BTNode
	levelZero = append(levelZero, &genesisNode) // just genesis
	bt.Levels = append(bt.Levels, levelZero)

	//  left child of genesis node
	b10Valid := BTNode{Height: 1, Parent: &genesisNode, ParentHash: []byte{0}, Data: "Left", Hash: []byte{0}}
	b10Valid.calcBTNodeHash()

	bt.addBTNodeIfValid(&b10Valid)

	// check if added parent correctly.
	if bt.Levels[1][0].Parent != &genesisNode {
		t.Error("Parent of second level should be genesis node")
	}

	// check if paren hash match
	if testEqByteSlice(bt.Levels[1][0].ParentHash, genesisNode.Hash) == false {
		t.Error("Parent hash of second level does not equal genesis node hash")
	}

	// check heights
	if bt.Levels[1][0].Height != genesisNode.Height + 1{
		t.Error("Heights do not align between genesis node and second level")
	}

	// now to add a right child to the genesis
	b11Valid := BTNode{Height:1, Parent: &genesisNode, ParentHash: genesisNode.Hash, Data: "Right", Hash: []byte{0}}
	b11Valid.calcBTNodeHash()

	// now to test it with addBTNodeIfValid
	bt.addBTNodeIfValid(&b11Valid)

	if bt.Levels[1][1].Parent != &genesisNode {
		t.Error("Parent of second level should be genesis node")
	}

	// check if paren hash match
	if testEqByteSlice(bt.Levels[1][1].ParentHash, genesisNode.Hash) == false {
		t.Error("Parent hash of second level does not equal genesis node hash")
	}

	// check heights
	if bt.Levels[1][1].Height != genesisNode.Height + 1{
		t.Error("Heights do not align between genesis node and second level")
	}

	// // child of right child of genesis node
	// b20Valid := BTNode{Height:2, Parent: &b11Valid, ParentHash: b11Valid.Hash, Data: "Data", Hash: []byte{0}}
	// b20Valid.calcBTNodeHash()


	// var levelOne  []*BTNode
	// var levelTwo  []*BTNode
	// levelOne  = append(levelOne, &b10Valid) // left child
	// levelOne  = append(levelOne, &b11Valid) // right child
	// levelTwo  = append(levelTwo, &b20Valid) //child of b11

	// bt = append(bt, levelOne)
	// bt = append(bt, levelTwo)


}












