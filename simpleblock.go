package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	Hash      string
	Data      string
	PrevHash  string
	Nonce     int
}

type BlockChain struct {
	blocks []*Block
}

var PendingBlocks = make(chan string)
var chain *BlockChain

func (b *Block) Mine() {
	target := "0"
	for {

		info := bytes.Join([][]byte{[]byte(b.Data), []byte(b.PrevHash), []byte(fmt.Sprintf("%d", b.Nonce))}, []byte{})
		hash := sha256.Sum256(info)

		b.Hash = fmt.Sprintf("%x", hash)

		if b.Hash[:1] == target {
			fmt.Printf("A new Block is Mined! %s\n", b.Hash)
			break
		}
		b.Nonce++
		time.Sleep(1 * time.Millisecond)
	}
}
func BlockMiner(chain *BlockChain) {
	for data := range PendingBlocks {
		fmt.Println("New data received! starting to mine")
		prevBlock := chain.blocks[len(chain.blocks)-1]
		newBlock := &Block{
			Index:     prevBlock.Index + 1,
			Timestamp: time.Now().String(),
			Data:      data,
			PrevHash:  prevBlock.Hash,
			Nonce:     0,
		}
		newBlock.Mine()
		chain.blocks = append(chain.blocks, newBlock)
	}
}

func getChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(chain.blocks)
}

func InitBlockChain() *BlockChain {
	genesis := &Block{
		Index:     0,
		Timestamp: time.Now().String(),
		Data:      "Genesis Block",
		PrevHash:  "",
		Nonce:     0,
	}

	fmt.Println("Mining Genesis Block...")
	genesis.Mine()

	return &BlockChain{blocks: []*Block{genesis}}
}
func (chain *BlockChain) IsValid() bool {

	for i := 1; i < len(chain.blocks); i++ {
		currentBlock := chain.blocks[i]
		prevBlock := chain.blocks[i-1]

		if currentBlock.PrevHash != prevBlock.Hash {
			return false
		}
	}
	return true
}

func main() {
	chain = InitBlockChain()

	go BlockMiner(chain)

	http.HandleFunc("/chain", getChain)
	go func() {
		PendingBlocks <- "First Block after Genesis"
		PendingBlocks <- "Second Block"
	}()
	fmt.Println("Server is running on http://localhost:8080/chain")
	http.ListenAndServe(":8080", nil)
}
