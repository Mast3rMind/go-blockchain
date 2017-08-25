package main 

import(
	"crypto/sha256"
	"io/ioutil"
	"fmt"

	"github.com/nvonpentz/go-hashable-keys"
)

type Packet struct {
	Hash      []byte
	Signature []byte
	Owner     []byte
}

func readDocument(filePath string) []byte{
	document, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}

	return document
}

func hashDocument(document []byte) []byte{
	h := sha256.New()
	h.Write(document)
	return h.Sum(nil)
}


func signHash(hash []byte, keys hashkeys.Keypair) []byte{
	signature, err := keys.Sign(hash) //sign the hash of the transaction
	if err !=nil {
		fmt.Println(err)
	}

	return signature
}

func createPacket(filepath string, keys hashkeys.Keypair) Packet {
	document := readDocument(filepath)
	documentHash := hashDocument(document)
	signature := signHash(documentHash, keys)

	return Packet{Hash: documentHash, Signature: signature, Owner: keys.Public}
}



func verifyPacketSignature(packet Packet) bool {
	return hashkeys.SignatureVerify(packet.Owner, packet.Signature, packet.Hash)
}



