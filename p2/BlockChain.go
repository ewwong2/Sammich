package p2

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"../p1"
	"golang.org/x/crypto/sha3"
)

// Block is a struct that contains information for a block in the blockchain.
type Block struct {
	Header      Header                `json:"header"`
	AcceptValue p1.MerklePatriciaTrie `json:"acceptance"`
	ApplyValue  p1.MerklePatriciaTrie `json:"application"`
}

// Header contains all the header information in a block.
type Header struct {
	Hash       string `json:"hash"`
	Timestamp  int64  `json:"timeStamp"`
	Height     int32  `json:"height"`
	ParentHash string `json:"parentHash"`
	// Size is the size of the two mpt added together
	Size int32 `json:"size"`
}

// BlockChain contains the highest length of the BlockChain and the Chain of the blockchain.
type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

// NewBlockChain returns a new blockchain
func NewBlockChain() BlockChain {
	return BlockChain{Chain: make(map[int32][]Block), Length: 0}
}

// Initial is the constructor for Block. The timestamp is taken at creation time. It is assumed that proper care
// will be taken to match the parentHash to the corresponding parent.
func (blk *Block) Initial(height int32, parentHash string, acceptValue p1.MerklePatriciaTrie,
	applyValue p1.MerklePatriciaTrie) error {
	timeStamp := time.Now().Unix()
	size := int32(len([]byte(fmt.Sprint(acceptValue))) + len([]byte(fmt.Sprint(applyValue))))
	blk.Header = Header{hashString(string(height) + string(timeStamp) + parentHash + acceptValue.Root +
		applyValue.Root + string(size)), timeStamp, height, parentHash, size}
	blk.AcceptValue = acceptValue
	blk.ApplyValue = applyValue
	return nil
}

// NewBlock is a special constructor for Block that allows for a manual input for the timestamp.
// This is useful for test applications.
func (blk *Block) NewBlock(height int32, timeStamp int64, parentHash string, acceptValue p1.MerklePatriciaTrie,
	applyValue p1.MerklePatriciaTrie) error {
	size := int32(len([]byte(fmt.Sprint(acceptValue))) + len([]byte(fmt.Sprint(applyValue))))
	blk.Header = Header{hashString(string(height) + string(timeStamp) + parentHash + acceptValue.Root +
		applyValue.Root + string(size)), timeStamp, height, parentHash, size}
	blk.AcceptValue = acceptValue
	blk.ApplyValue = applyValue
	return nil
}

// DecodeFromJson decodes a JSON string into blk.
// An error is returned if json.Unmarshal is unable to decode the string.
func (blk *Block) DecodeFromJson(jsonString string) error {
	err := json.Unmarshal([]byte(jsonString), &blk)
	return err
}

// EncodeToJson encodes a block to a JSON string. This string is returned.
// If the value could not be encoded, an error will be thrown.
func (blk *Block) EncodeToJson() (string, error) {
	res, err := json.Marshal(blk)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// Get returns the list of Blocks at height level in the blockchain.
func (bc *BlockChain) Get(height int32) []Block {
	if bc.Chain == nil || bc.Length < height || height < 0 {
		return nil
	}
	return bc.Chain[height-1]
}

// Insert inserts block into the BlockChain.
// FIXME: In the current state, there is no input validation for the inserted Blocks to verify if it does create a valid
//        blockchain.
func (bc *BlockChain) Insert(block Block) error {
	if bc.Chain == nil {
		bc.Chain = make(map[int32][]Block)
		bc.Length = 0
	} else if block.Header.Height < 0 {
		return errors.New("height out of range")
	}

	// FIXME: This intuitively makes more sense to include this, otherwise there may be
	//        an improper block that does not connect to a parent hash.
	// TODO: This is if we want to restrict the blockchain to not take blocks that
	//       have a height larger than the length of the blockchain
	//if bc.Length+1 < block.Header.Height {
	//	return errors.New("invalid height")
	//}
	//fmt.Printf("Trying to add %s to %+v\n", block.Header.Hash, bc.Chain[block.Header.Height-1])
	for _, v := range bc.Chain[block.Header.Height-1] {
		if v.Header.Hash == block.Header.Hash {
			//fmt.Printf("Unable to add %s to %+v\n", block.Header.Hash, bc.Chain[block.Header.Height-1])
			return errors.New("duplicate block")
		}
	}
	//fmt.Printf("Able to add %s to %+v\n", block.Header.Hash, bc.Chain[block.Header.Height-1])
	bc.Chain[block.Header.Height-1] = append(bc.Chain[block.Header.Height-1], block)
	if block.Header.Height > bc.Length {
		bc.Length = block.Header.Height
	}
	return nil
}

// DecodeFromJson decodes jsonString into the bc BlockChain.
// An error is thrown if json.UnMarshal could not decode the string.
func (bc *BlockChain) DecodeFromJson(jsonString string) error {
	err := json.Unmarshal([]byte(jsonString), &bc)
	return err
}

// EncodeToJson encodes the blockchain bc to a JSON string. This string is returned.
// An error is thrown if blockchain could not be encoded.
func (bc *BlockChain) EncodeToJson() (string, error) {
	resp, err := json.Marshal(bc)
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (bc *BlockChain) ShowAcceptances() map[string]int32 {
	acc := make(map[string]int32)

	for _, v := range bc.Chain {
		blk := v[0]
		for k2, v2 := range blk.AcceptValue.Values.Db {
			uid, err := strconv.Atoi(v2)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not show %s accept with %s\n", k2, v2)
				continue
			}
			acc[k2] = int32(uid)
		}
	}

	return acc
}

func (bc *BlockChain) ShowApplications() []string {
	var merits []string

	for _, v := range bc.Chain {
		blk := v[0]
		for _, v2 := range blk.ApplyValue.Values.Db {
			merits = append(merits, v2)
		}
	}

	return merits
}

// constructMpt takes a map of string, string and inserts each value into a MerklePatriciaTrie.
// This MPT is returned.
func constructMpt(mptMap map[string]string) p1.MerklePatriciaTrie {
	mpt := p1.MerklePatriciaTrie{}
	for k, v := range mptMap {
		mpt.Insert(k, v)
	}
	return mpt
}

// hashString hashes the given string. The hashed string is returned.
func hashString(str string) string {
	sum := sha3.Sum256([]byte(str))
	return hex.EncodeToString(sum[:])
}

// GenBlock generates the next block at the next height
func (bc *BlockChain) GenBlock(acceptMpt p1.MerklePatriciaTrie, applyMpt p1.MerklePatriciaTrie) (Block, error) {
	if bc.Length == 0 || len(bc.Chain[bc.Length-1]) == 0 {
		return Block{}, errors.New("missing parent")
	}
	block := Block{}
	block.Initial(bc.Length+1, bc.Chain[bc.Length-1][0].Header.Hash, acceptMpt, applyMpt)
	//fmt.Printf("Able to add %s to %+v\n", block.Header.Hash, bc.Chain[block.Header.Height-1])
	bc.Chain[bc.Length] = append(bc.Chain[bc.Length], block)
	bc.Length++
	return block, nil
}

// GetHighest returns the list of blocks at the highest height
func (bc *BlockChain) GetHighest() ([]Block, error) {
	fmt.Println(bc.Length)
	if bc.Length > 0 {
		return bc.Chain[bc.Length-1], nil
	} else {
		return []Block{}, errors.New("empty blockchain")
	}
}

// CheckParentHash adds the block to the blockchain if the parent exists,
// otherwise false is returned.
func (bc *BlockChain) CheckParentHash(insertBlock Block) bool {
	if bc.Length == 0 || insertBlock.Header.Height < 2 || len(bc.Chain[insertBlock.Header.Height-2]) == 0 {
		return false
	}
	for _, v := range bc.Chain[insertBlock.Header.Height-2] {
		if v.Header.Hash == insertBlock.Header.ParentHash {
			err := bc.Insert(insertBlock)
			return err == nil
		}
	}
	return false
}

// Show returns a string representation of the blockchain
func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.Header.Hash+"<="+block.Header.ParentHash)
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}
