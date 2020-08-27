package main

import (
	"time"
	"bytes"
	"encoding/binary"
	"log"
	"encoding/gob"
	"crypto/sha256"
)

//0. 定義結構
type Block struct {
	//1.版本號
	Version uint64
	//2. 前區塊雜湊
	PrevHash []byte
	//3. Merkel根
	MerkelRoot []byte
	//4. 時間戳
	TimeStamp uint64
	//5. 難度值
	Difficulty uint64
	//6. 隨機數，也就是挖礦要找的數據
	Nonce uint64

	//a. 目前區塊雜湊,正常比特幣區塊中沒有當前區塊的雜湊，我們爲了是方便做了簡化！
	Hash []byte
	//b. 數據
	//Data []byte
	//真實的交易陣列
	Transactions []*Transaction
}

//1. 補充區塊欄位
//2. 更新計算雜湊函式
//3. 優化程式碼

//實現一個輔助函式，功能是將uint64轉成[]byte
func Uint64ToByte(num uint64) []byte {
	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

//2. 建立區塊
func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := Block{
		Version:    00,
		PrevHash:   prevBlockHash,
		MerkelRoot: []byte{},
		TimeStamp:  uint64(time.Now().Unix()),
		Difficulty: 0, //隨便填寫的無效值
		Nonce:      0, //同上
		Hash:       []byte{},
		//Data:       []byte(data),
		Transactions: txs,
	}

	block.MerkelRoot = block.MakeMerkelRoot()

	//block.SetHash()
	//建立一個pow對像
	pow := NewProofOfWork(&block)
	//查詢隨機數，不停的進行雜湊運算
	hash, nonce := pow.Run()

	//根據挖礦結果對區塊數據進行更新（補充）
	block.Hash = hash
	block.Nonce = nonce

	return &block
}

//序列化
func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer

	//- 使用gob進行序列化（編碼）得到位元組流
	//1. 定義一個編碼器
	//2. 使用編碼器進行編碼
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(&block)
	if err != nil {
		log.Panic("編碼出錯!")
	}

	return buffer.Bytes()
}

//反序列化
func Deserialize(data []byte) Block {

	decoder := gob.NewDecoder(bytes.NewReader(data))

	var block Block
	//2. 使用解碼器進行解碼
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic("解碼出錯!", err)
	}

	return block
}

/*
//3. 產生雜湊
func (block *Block) SetHash() {
	//var blockInfo []byte
	//1. 拼裝數據
	/*
	blockInfo = append(blockInfo, Uint64ToByte(block.Version)...)
	blockInfo = append(blockInfo, block.PrevHash...)
	blockInfo = append(blockInfo, block.MerkelRoot...)
	blockInfo = append(blockInfo, Uint64ToByte(block.TimeStamp)...)
	blockInfo = append(blockInfo, Uint64ToByte(block.Difficulty)...)
	blockInfo = append(blockInfo, Uint64ToByte(block.Nonce)...)
	blockInfo = append(blockInfo, block.Data...)
	*/
/*
tmp := [][]byte{
	Uint64ToByte(block.Version),
	block.PrevHash,
	block.MerkelRoot,
	Uint64ToByte(block.TimeStamp),
	Uint64ToByte(block.Difficulty),
	Uint64ToByte(block.Nonce),
	block.Data,
}

//將二維的切片陣列鏈接起來，返回一個一維的切片
blockInfo := bytes.Join(tmp, []byte{})

//2. sha256
//func Sum256(data []byte) [Size]byte {
hash := sha256.Sum256(blockInfo)
block.Hash = hash[:]
}
*/

//模擬梅克爾根，只是對交易的數據做簡單的拼接，而不做二叉樹處理！
func (block *Block) MakeMerkelRoot() []byte {
	var info []byte
	//var finalInfo [][]byte
	for _, tx := range block.Transactions {
		//將交易的雜湊值拼接起來，再整體做雜湊處理
		info = append(info, tx.TXID...)
		//finalInfo = [][]byte{tx.TXID}
	}

	hash := sha256.Sum256(info)
	return hash[:]
}