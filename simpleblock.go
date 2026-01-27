package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
}

type BlockChain struct {
	blocks []*Block
}

func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash}

	block.DeriveHash()
	return block
}

func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.blocks = append(chain.blocks, new)
}

var chain *BlockChain

func getChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chain.blocks)
}
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
func (chain *BlockChain) IsValid() bool {

	for i := 1; i < len(chain.blocks); i++ {
		currentBlock := chain.blocks[i]
		prevBlock := chain.blocks[i-1]

		if string(currentBlock.PrevHash) != string(prevBlock.Hash) {
			return false
		}
	}
	return true
}

func main() {
	chain = InitBlockChain()
	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block")
	chain.AddBlock("Third Block")
	for _, block := range chain.blocks {
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("hash: %x\n", block.Hash)
		fmt.Printf("Is the blockchain valid? %v\n", chain.IsValid())

		// hacking
		chain.blocks[1].Data = []byte("First Block after Genesiss")
		chain.blocks[1].DeriveHash() // Recalculate hash for block 1

		fmt.Printf("Is it valid after the hack? %v\n", chain.IsValid())
		http.HandleFunc("/chain", getChain)

		fmt.Println("Server is running on http://localhost:8080/chain")
		http.ListenAndServe(":8080", nil)
	}

}
