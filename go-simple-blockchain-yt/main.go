package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Block struct {
	Pos			int
	Data		Bookcheckout
	TimeStamp	string
	Hash		string
	PrevHash	string	
}

type Bookcheckout struct {
	BookID			string	`json:"book_id"`
	User 			string	`json:"user"`
	CheckoutDate 	string	`json:"checkout_date"`
	IsGenesis 		bool	`json:"is_genesis"`
}

type Book struct {
	ID			string	`json:"id"`
	Title		string	`json:"title"`
	Author		string	`json:"author"` 
	PublishData	string	`json:"publish_data"`
	ISBN		string	`json:"isbn"`
}

type Blockchain struct {
	blocks []*Block
}

var BlockChain *Blockchain

func (b *Block) generateHash() {
	bytes, _ := json.Marshal(b.Data)

	data := fmt.Sprint(b.Pos) + b.TimeStamp + string(bytes) + b.PrevHash
	
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
	fmt.Printf("b.generateHash: %v\n", b)
}

func CreateBlock(prevBlock *Block, checkouitem Bookcheckout) *Block {
	block := &Block{}
	block.Pos = prevBlock.Pos + 1
	block.TimeStamp = time.Now().String()
	block.PrevHash = prevBlock.Hash
	block.Data = checkouitem
	block.generateHash()
	fmt.Printf("new block: %v\n", block)

	return block
}

func (b *Block) validateHash(hash string)  bool {
	b.generateHash()
	if b.Hash != hash {
		return false
	}
	return true
}

func validBlock(block, prevBlock *Block) bool {
	fmt.Printf("valid block, block %v\n", block)
	fmt.Printf("valid block, prevBlock %v\n", prevBlock)
	if prevBlock.Hash != block.PrevHash {
		return false
	}

	if !block.validateHash(block.Hash) {
		return false
	}

	if prevBlock.Pos + 1 != block.Pos {
		return false
	}

	return true
}

func (bc *Blockchain) AddBlock(data Bookcheckout) {
	prevBlock := bc.blocks[len(bc.blocks)-1]

	block := CreateBlock(prevBlock, data)
	fmt.Printf("created block %v\n",block)
	fmt.Printf("before append block, blocks: %v",bc.blocks)


	if validBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
		////
		fmt.Printf("after append block, blocks: %v\n",bc.blocks)
	}
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkoutitem Bookcheckout

	if err := json.NewDecoder(r.Body).Decode(&checkoutitem); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not write block: %v", err)
		w.Write([]byte("could not write block"))
	}

	BlockChain.AddBlock(checkoutitem)
	resp, err := json.MarshalIndent(checkoutitem, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("failed to marshalindent checkoutitem: %v", err)
		w.Write([]byte("failed to marshalindent checkoutitem"))
	}
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(resp))

	go func() {
		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", "  ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
	}()
}

func newBook(w http.ResponseWriter, r *http.Request) {
	var book Book

	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not create: %v", err)
		w.Write([]byte("could not create new book"))
		return
	}

	h := md5.New()
	io.WriteString(h, book.Title+book.Author+book.ISBN+book.PublishData)
	book.ID = fmt.Sprintf("%x", h.Sum(nil))

	resp, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not save book data"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func GenesisBlock() *Block {
	return CreateBlock(&Block{}, Bookcheckout{IsGenesis: true})
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.blocks, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	// io.WriteString(w, string(jbytes))
	w.WriteHeader(http.StatusOK)
	w.Write(jbytes)
}

func main() {
	BlockChain = NewBlockchain()

	r := mux.NewRouter()
	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newBook).Methods("POST")

	go func() {
		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", "  ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
	}()

	log.Println("Listening on port 3000")

	log.Fatal(http.ListenAndServe(":3000", r))
}
