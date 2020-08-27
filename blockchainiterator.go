package main

import (
	"./lib/bolt"
	"log"
)

type BlockChainIterator struct {
	db *bolt.DB
	//遊標，用於不斷索引
	currentHashPointer []byte
}

//func NewIterator(bc *BlockChain)  {
//
//}

func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{
		bc.db,
		//最初指向區塊鏈的最後一個區塊，隨著Next的呼叫，不斷變化
		bc.tail,
	}
}

//迭代器是屬於區塊鏈的
//Next方式是屬於迭代器的
//1. 返回目前的區塊
//2. 指針前移
func (it *BlockChainIterator) Next() *Block {
	var block Block
	it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			log.Panic("迭代器遍歷時bucket不應該為空，請檢查!")
		}

		blockTmp := bucket.Get(it.currentHashPointer)
		//解碼動作
		block = Deserialize(blockTmp)
		//遊標雜湊左移
		it.currentHashPointer = block.PrevHash

		return nil
	})

	return &block
}

