package main 

import(
	"crypto/sha256"
	"encoding/binary"
	"fmt"
)

type BTNode struct {
	Height	   uint32
	Parent     *BTNode
	ParentHash []byte
	Data       string
	Hash 	   []byte
}

type BlockWrapper struct {
	Block  *BTNode
	Sender string
}

type BlockTree struct {
	Levels [][]*BTNode
	Top        *BTNode
}

var genesisNode = BTNode{Height: 0, Parent: nil, ParentHash: []byte{0}, Data: "Genesis", Hash: []byte{0}}

func emptyBlock() BTNode {
	return BTNode{Height: 0, Parent: nil, ParentHash: nil, Data: "", Hash: nil}
}

/* 
addBTNodeIfValid takes a proposed block and an existing block tree
and checks if the parent of this node exists on the blockchain. If
it does, it adds it to the appropirate level of the blocktree and
returns true, otherwise returns false. 
*/
func (bt *BlockTree) addBTNodeIfValid(newBTNode *BTNode) bool {
	parentHeight         := newBTNode.Height - 1
	if uint32(len(bt.Levels)) <= parentHeight  {
		// does not have the parent, so appears as nil.
		return false
	}
	nodesAtLevelOfParent := bt.Levels[parentHeight]

	for _ , oldBTNode := range nodesAtLevelOfParent {
		if oldBTNode.isValidNextBTNode(newBTNode) {
			// append to parent level
			if !bt.hasLevelAtHeight(newBTNode.Height) { // we define genesis block at height 0 // does not have height
				// this is now the longest chain, append a new level
				var newLevel []*BTNode
				newLevel = append(newLevel, newBTNode) // new level containing the only block high enough
				bt.Levels = append(bt.Levels, newLevel) // should automatically be at correct height
				bt.Top = newBTNode
				fmt.Println("just added new top of blockchain")
			} else {
				// not the longest chain, directly inject into height at newBTNode.height		
				bt.Levels[newBTNode.Height] = append(bt.Levels[newBTNode.Height], newBTNode)
			}
			return true
		} else {
			fmt.Println("No matching node found")
		}
	}
	return false
} // should check to see which is now the longest

func (oldBTNode *BTNode) isValidNextBTNode(newBTNode *BTNode) bool {
	heightValid    := oldBTNode.Height + 1 == newBTNode.Height
	// fmt.Printf("Height valid: %v\n", heightValid)
	var parentValid bool
	if newBTNode.Parent != nil{
		parentValid = equalBTNodes(*oldBTNode, *newBTNode.Parent)		
	} else {
		parentValid = false
	}
	// fmt.Printf("Parent valid: %v\n", parentValid)
	parentHashValid := testEqByteSlice(oldBTNode.Hash, newBTNode.ParentHash)
	// fmt.Printf("Parent hash valid: %v\n", parentHashValid)

	/* 
	need to include hash valid that checks if the hash of this
	block (the thing added by calcBTNodeHash()) is correct.

	evenutally I will need to expand the calcBTNodeHash() function to
	include trandactions for the cryptocoin branch of this project
	*/

	return heightValid && parentValid && parentHashValid
}

func equalBTNodes(b1, b2 BTNode) bool {
	heightEq     := b1.Height == b2.Height
	parentHashEq := testEqByteSlice(b1.ParentHash, b2.ParentHash)
	dataEq       := b1.Data == b2.Data
	hashEq 		 := testEqByteSlice(b1.Hash, b2.Hash)

	return heightEq && parentHashEq && dataEq && hashEq
}

func (b *BTNode) calcBTNodeHash(){
	height := make([]byte, 4)
	binary.LittleEndian.PutUint32(height, b.Height)
	data := []byte(b.Data)

	h := sha256.New()
	h.Write(height)
	h.Write(data)
	h.Write(b.ParentHash)

	b.Hash = h.Sum(nil)
}

// if youve sent a block that the other doesn't have, they'll request
// the full chain that it is derived from so they can validate
func (bt *BlockTree) deriveChainToBlock(topBlock *BTNode) []*BTNode {
	var treeLevelOfNode []*BTNode
	empty := []*BTNode{}

	if bt.hasLevelAtHeight(topBlock.Height){ // have a level corresponding to this height
		treeLevelOfNode = bt.Levels[topBlock.Height]
	} else{
		return empty //none exists
	}

	for _ , block := range treeLevelOfNode{
		if equalBTNodes(*block, *topBlock) {
			fmt.Println("Found the block they want")
			chain := block.constructChain()
			return chain
		}
	}
	return empty
}
// constructs a blockchain from given tip 
func (b *BTNode) constructChain() (chain []*BTNode) {
	block := b
	for block.Parent != nil {
		chain = append(chain, block.Parent)
		block = block.Parent
	}
	return chain
}

/*
Realizing now that some of this logic might be unnecessary,
and we could just pipeline every block in the chain to the 
blockchannel, regardless of if we have them or not.

I'll do that instead...ugh.
*/
func (bt *BlockTree) addMissingBlocks(blockchainSubset []*BTNode, blockChannel chan *BlockWrapper){
	for _ , b :=range blockchainSubset {
		blockWrapper := BlockWrapper{b, ""}
		blockChannel <- &blockWrapper //send it to the block channel to be processed as normal
	}
}

func (bt *BlockTree) findMissingBlocks(blockchain []*BTNode) []*BTNode{
	// loop through blocks starting at top going toward genesis
	for i, block := range blockchain{
		if bt.hasLevelAtHeight(block.Height){ // if our blockchain has that height it is a candidate
			level := bt.Levels[block.Height] // get the level to examine
			if levelHasBTNode(level, block){
				fmt.Println("Found the first common block")
				return blockchain[:i] // get the last block up to the one I have
			} else {
				fmt.Println("despite having the correct blockheight, no blocks in the level matched, trying next block..")
			}
		} else {
			fmt.Println("You do not have the blocktree level associated with this block")
		}
	}
	return []*BTNode{}
	// return slice between what I have and the proposed node
}

func (bt *BlockTree) hasLevelAtHeight(height uint32) bool {
	blockTreeHeight := uint32(len(bt.Levels) - 1)
	hasLevel := blockTreeHeight >= height
	return hasLevel
}

func levelHasBTNode(level []*BTNode, block *BTNode) bool{
	for _ , b := range level {
		if equalBTNodes(*b, *block){
			return true
		}
	}
	return false
}

// func getBTNodeFromLevel(level []*BTNode, block *BTNode) *BTNode{
// 	for _ , b := range level {
// 		if equalBTNodes(*b, *block){
// 			return b
// 		}
// 	}
// 	fmt.Println("No similar block was found in level, returing empty block")
// 	return &BTNode{} // if we loop through and find nothing
// }

















