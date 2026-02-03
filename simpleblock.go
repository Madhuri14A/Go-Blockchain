package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	Hash      string
	//Data        string
	PrevHash     string
	Nonce        int
	Transactions []Transaction
}
type Transaction struct {
	From      string
	To        string
	Amount    int
	Signature []byte
	PublicKey []byte
}

var PendingBlocks = make(chan Transaction)
var chain *BlockChain

type BlockChain struct {
	blocks []*Block
}

func NewKeyPair() (*ecdsa.PrivateKey, []byte) {
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	publicKey := append(
		privateKey.PublicKey.X.Bytes(),
		privateKey.PublicKey.Y.Bytes()...)
	return privateKey, publicKey
}
func (tx *Transaction) Hash() []byte {
	data := fmt.Sprintf("%s%s%d", tx.From, tx.To, tx.Amount)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}
func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey) {
	hash := tx.Hash()
	r, s, _ := ecdsa.Sign(rand.Reader, privKey, hash)
	signature := append(r.Bytes(), s.Bytes()...)
	tx.Signature = signature
}
func VerifyTransaction(tx *Transaction) bool {
	if tx.Signature == nil || tx.PublicKey == nil {
		return false
	}
	x := tx.PublicKey[:len(tx.PublicKey)/2]
	y := tx.PublicKey[len(tx.PublicKey)/2:]

	pubKey := ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(x),
		Y:     new(big.Int).SetBytes(y),
	}
	hash := tx.Hash()

	r := new(big.Int).SetBytes(tx.Signature[:len(tx.Signature)/2])
	s := new(big.Int).SetBytes(tx.Signature[len(tx.Signature)/2:])
	return ecdsa.Verify(&pubKey, hash, r, s)
}
func (b *Block) Mine() {
	target := "0"
	for {
		txBytes, _ := json.Marshal(b.Transactions)
		info := bytes.Join([][]byte{
			txBytes,
			[]byte(b.PrevHash),
			[]byte(fmt.Sprintf("%d", b.Nonce)),
		}, []byte{})
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
	for tx := range PendingBlocks {
		fmt.Println("New transaction received! starting to mine")
		prevBlock := chain.blocks[len(chain.blocks)-1]
		newBlock := &Block{
			Index:        prevBlock.Index + 1,
			Timestamp:    time.Now().String(),
			Transactions: []Transaction{tx},
			PrevHash:     prevBlock.Hash,
			Nonce:        0,
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
		Index:        0,
		Timestamp:    time.Now().String(),
		Transactions: []Transaction{},
		PrevHash:     "",
		Nonce:        0,
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
		privKey, pubKey := NewKeyPair()
		tx := Transaction{
			From:      "madhuri",
			To:        "abc",
			Amount:    1000,
			PublicKey: pubKey,
		}
		tx.Sign(privKey)

		PendingBlocks <- tx
	}()
	fmt.Println("Server is running on http://localhost:8080/chain")
	http.ListenAndServe(":8080", nil)
}
