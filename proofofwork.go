package main

import (
	"math/big"
	"bytes"
	"crypto/sha256"
	"fmt"
)

//定義一個工作量證明的結構ProofOfWork
type ProofOfWork struct {
	//a. block
	block *Block
	//b. 目標值
	//一個非常大數，它有很豐富的方法：比較，賦值方法
	target *big.Int
}

//2. 提供建立POW的函式
//
//- NewProofOfWork(參數)
func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}

	//我們指定的難度值，現在是一個string型別，需要進行轉換
	targetStr := "0000100000000000000000000000000000000000000000000000000000000000"
	//
	//引入的輔助變數，目的是將上面的難度值轉成big.int
	tmpInt := big.Int{}
	//將難度值賦值給big.int，指定16進位制的格式
	tmpInt.SetString(targetStr, 16)

	pow.target = &tmpInt
	return &pow
}

//3. 提供計算不斷計算hash的哈數
//- Run()

func (pow *ProofOfWork) Run() ([]byte, uint64) {
	//1. 拼裝數據（區塊的數據，還有不斷變化的隨機數）
	//2. 做雜湊運算
	//3. 與pow中的target進行比較
	//a. 找到了，退出返回
	//b. 沒找到，繼續找，隨機數加1

	var nonce uint64
	block := pow.block
	var hash [32]byte

	fmt.Println("開始挖礦...")
	for {
		//1. 拼裝數據（區塊的數據，還有不斷變化的隨機數）
		tmp := [][]byte{
			Uint64ToByte(block.Version),
			block.PrevHash,
			block.MerkelRoot,
			Uint64ToByte(block.TimeStamp),
			Uint64ToByte(block.Difficulty),
			Uint64ToByte(nonce),
			//只對區塊頭做雜湊值，區塊體通過MerkelRoot產生影響
			//block.Data,
		}

		//將二維的切片陣列鏈接起來，返回一個一維的切片
		blockInfo := bytes.Join(tmp, []byte{})

		//2. 做雜湊運算
		//func Sum256(data []byte) [Size]byte {
		hash = sha256.Sum256(blockInfo)
		//3. 與pow中的target進行比較
		tmpInt := big.Int{}
		//將我們得到hash陣列轉換成一個big.int
		tmpInt.SetBytes(hash[:])

		//比較目前的雜湊與目標雜湊值，如果目前的雜湊值小於目標的雜湊值，就說明找到了，否則繼續找

		//   -1 if x <  y
		//    0 if x == y
		//   +1 if x >  y

		//func (x *Int) Cmp(y *Int) (r int) {
		if tmpInt.Cmp(pow.target) == -1 {
			//a. 找到了，退出返回
			fmt.Printf("挖礦成功！hash : %x, nonce : %d\n", hash, nonce)
			//break
			return hash[:], nonce
		} else {
			//b. 沒找到，繼續找，隨機數加1
			nonce++
		}

	}

	//return []byte("HelloWorld"), 10
}

//
//4. 提供一個校驗函式
//
//- IsValid()
