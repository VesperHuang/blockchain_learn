package main

import (
	"./lib/bolt"
	"log"
	"fmt"
	"bytes"
	"errors"
	"crypto/ecdsa"
)

//4. 引入區塊鏈
//2. BlockChain結構重寫
//使用數據庫代替陣列

type BlockChain struct {
	//定一個區塊鏈陣列
	//blocks []*Block
	db *bolt.DB

	tail []byte //儲存最後一個區塊的雜湊
}

const blockChainDb = "blockChain.db"
const blockBucket = "blockBucket"

//5. 定義一個區塊鏈
func NewBlockChain(address string) *BlockChain {
	//return &BlockChain{
	//	blocks: []*Block{genesisBlock},
	//}

	//最後一個區塊的雜湊， 從數據庫中讀出來的
	var lastHash []byte

	//1. 打開數據庫
	db, err := bolt.Open(blockChainDb, 0600, nil)
	//defer db.Close()

	if err != nil {
		log.Panic("打開數據庫失敗！")
	}

	//將要運算元據庫（改寫）
	db.Update(func(tx *bolt.Tx) error {
		//2. 找到抽屜bucket(如果沒有，就建立）
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			//沒有抽屜，我們需要建立
			bucket, err = tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				log.Panic("建立bucket(b1)失敗")
			}

			//建立一個創世塊，並作為第一個區塊新增到區塊鏈中
			genesisBlock := GenesisBlock(address)
			//fmt.Printf("genesisBlock :%s\n", genesisBlock)

			//3. 寫數據
			//hash作為key， block的位元組流作為value，尚未實現
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			bucket.Put([]byte("LastHashKey"), genesisBlock.Hash)
			lastHash = genesisBlock.Hash

			////這是爲了讀數據測試，馬上刪掉!
			//blockBytes := bucket.Get(genesisBlock.Hash)
			//block := Deserialize(blockBytes)
			//fmt.Printf("block info : %s\n", block)

		} else {
			lastHash = bucket.Get([]byte("LastHashKey"))
		}

		return nil
	})

	return &BlockChain{db, lastHash}
}

//定義一個創世塊
func GenesisBlock(address string) *Block {
	coinbase := NewCoinbaseTX(address, "Go一期創世塊，老牛逼了！")
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

//5. 新增區塊
func (bc *BlockChain) AddBlock(txs []*Transaction) {

	for _, tx := range txs {
		if !bc.VerifyTransaction(tx) {
			fmt.Printf("礦工發現無效交易!")
			return
		}
	}

	//如何獲取前區塊的雜湊呢？？
	db := bc.db         //區塊鏈數據庫
	lastHash := bc.tail //最後一個區塊的雜湊

	db.Update(func(tx *bolt.Tx) error {

		//完成數據新增
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			log.Panic("bucket 不應該為空，請檢查!")
		}

		//a. 建立新的區塊
		block := NewBlock(txs, lastHash)

		//b. 新增到區塊鏈db中
		//hash作為key， block的位元組流作為value，尚未實現
		bucket.Put(block.Hash, block.Serialize())
		bucket.Put([]byte("LastHashKey"), block.Hash)

		//c. 更新一下記憶體中的區塊鏈，指的是把最後的小尾巴tail更新一下
		bc.tail = block.Hash

		return nil
	})
}

func (bc *BlockChain) Printchain() {

	blockHeight := 0
	bc.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("blockBucket"))

		//從第一個key-> value 進行遍歷，到最後一個固定的key時直接返回
		b.ForEach(func(k, v []byte) error {
			if bytes.Equal(k, []byte("LastHashKey")) {
				return nil
			}

			block := Deserialize(v)
			//fmt.Printf("key=%x, value=%s\n", k, v)
			fmt.Printf("=============== 區塊高度: %d ==============\n", blockHeight)
			blockHeight++
			fmt.Printf("版本號: %d\n", block.Version)
			fmt.Printf("前區塊雜湊值: %x\n", block.PrevHash)
			fmt.Printf("梅克爾根: %x\n", block.MerkelRoot)
			fmt.Printf("時間戳: %d\n", block.TimeStamp)
			fmt.Printf("難度值(隨便寫的）: %d\n", block.Difficulty)
			fmt.Printf("隨機數 : %d\n", block.Nonce)
			fmt.Printf("目前區塊雜湊值: %x\n", block.Hash)
			fmt.Printf("區塊數據 :%s\n", block.Transactions[0].TXInputs[0].PubKey)
			return nil
		})
		return nil
	})
}

//找到指定地址的所有的utxo
func (bc *BlockChain) FindUTXOs(pubKeyHash []byte) []TXOutput {
	var UTXO []TXOutput

	txs := bc.FindUTXOTransactions(pubKeyHash)

	for _, tx := range txs {
		for _, output := range tx.TXOutputs {
			if bytes.Equal(pubKeyHash, output.PubKeyHash) {
				UTXO = append(UTXO, output)
			}
		}
	}

	return UTXO
}

//根據需求找到合理的utxo
func (bc *BlockChain) FindNeedUTXOs(senderPubKeyHash []byte, amount float64) (map[string][]uint64, float64) {
	//找到的合理的utxos集合
	utxos := make(map[string][]uint64)
	var calc float64

	txs := bc.FindUTXOTransactions(senderPubKeyHash)

	for _, tx := range txs {
		for i, output := range tx.TXOutputs {
			//if from == output.PubKeyHash {
			//兩個[]byte的比較
			//直接比較是否相同，返回true或false
			if bytes.Equal(senderPubKeyHash, output.PubKeyHash) {
				//fmt.Printf("222222")
				//UTXO = append(UTXO, output)
				//fmt.Printf("333333 : %f\n", UTXO[0].Value)
				//我們要實現的邏輯就在這裡，找到自己需要的最少的utxo
				//3. 比較一下是否滿足轉賬需求
				//   a. 滿足的話，直接返回 utxos, calc
				//   b. 不滿足繼續統計

				if calc < amount {
					//1. 把utxo加進來，
					//utxos := make(map[string][]uint64)
					//array := utxos[string(tx.TXID)] //確認一下是否可行！！
					//array = append(array, uint64(i))
					utxos[string(tx.TXID)] = append(utxos[string(tx.TXID)], uint64(i))
					//2. 統計一下目前utxo的總額
					//第一次進來: calc =3,  map[3333] = []uint64{0}
					//第二次進來: calc =3 + 2,  map[3333] = []uint64{0, 1}
					//第三次進來：calc = 3 + 2 + 10， map[222] = []uint64{0}
					calc += output.Value

					//加完之後滿足條件了，
					if calc >= amount {
						//break
						fmt.Printf("找到了滿足的金額：%f\n", calc)
						return utxos, calc
					}
				} else {
					fmt.Printf("不滿足轉賬金額,目前總額：%f， 目標金額: %f\n", calc, amount)
				}
			}
		}
	}

	return utxos, calc
}

func (bc *BlockChain) FindUTXOTransactions(senderPubKeyHash []byte) []*Transaction {
	var txs []*Transaction //儲存所有包含utxo交易集合
	//我們定義一個map來儲存消費過的output，key是這個output的交易id，value是這個交易中索引的陣列
	//map[交易id][]int64
	spentOutputs := make(map[string][]int64)

	//建立迭代器
	it := bc.NewIterator()

	for {
		//1.遍歷區塊
		block := it.Next()

		//2. 遍歷交易
		for _, tx := range block.Transactions {
			//fmt.Printf("current txid : %x\n", tx.TXID)

		OUTPUT:
		//3. 遍歷output，找到和自己相關的utxo(在新增output之前檢查一下是否已經消耗過)
		//	i : 0, 1, 2, 3
			for i, output := range tx.TXOutputs {
				//fmt.Printf("current index : %d\n", i)
				//在這裡做一個過濾，將所有消耗過的outputs和目前的所即將新增output對比一下
				//如果相同，則跳過，否則新增
				//如果目前的交易id存在於我們已經表示的map，那麼說明這個交易裡面有消耗過的output

				//map[2222] = []int64{0}
				//map[3333] = []int64{0, 1}
				//這個交易裡面有我們消耗過得output，我們要定位它，然後過濾掉
				if spentOutputs[string(tx.TXID)] != nil {
					for _, j := range spentOutputs[string(tx.TXID)] {
						//[]int64{0, 1} , j : 0, 1
						if int64(i) == j {
							//fmt.Printf("111111")
							//目前準備新增output已經消耗過了，不要再加了
							continue OUTPUT
						}
					}
				}

				//這個output和我們目標的地址相同，滿足條件，加到返回UTXO陣列中
				//if output.PubKeyHash == address {
				if bytes.Equal(output.PubKeyHash, senderPubKeyHash) {
					//fmt.Printf("222222")
					//UTXO = append(UTXO, output)

					//!!!!!重點
					//返回所有包含我的outx的交易的集合
					txs = append(txs, tx)

					//fmt.Printf("333333 : %f\n", UTXO[0].Value)
				} else {
					//fmt.Printf("333333")
				}
			}

			//如果目前交易是挖礦交易的話，那麼不做遍歷，直接跳過

			if !tx.IsCoinbase() {
				//4. 遍歷input，找到自己花費過的utxo的集合(把自己消耗過的標示出來)
				for _, input := range tx.TXInputs {
					//判斷一下目前這個input和目標（李四）是否一致，如果相同，說明這個是李四消耗過的output,就加進來
					//if input.Sig == address {
					//if input.PubKey == senderPubKeyHash  //這是肯定不對的，要做雜湊處理
					pubKeyHash := HashPubKey(input.PubKey)
					if bytes.Equal(pubKeyHash, senderPubKeyHash) {
						//spentOutputs := make(map[string][]int64)
						//indexArray := spentOutputs[string(input.TXid)]
						//indexArray = append(indexArray, input.Index)
						spentOutputs[string(input.TXid)] = append(spentOutputs[string(input.TXid)], input.Index)
						//map[2222] = []int64{0}
						//map[3333] = []int64{0, 1}
					}
				}
			} else {
				//fmt.Printf("這是coinbase，不做input遍歷！")
			}
		}

		if len(block.PrevHash) == 0 {
			break
			fmt.Printf("區塊遍歷完成退出!")
		}
	}

	return txs
}

//根據id查詢交易本身，需要遍歷整個區塊鏈
func (bc *BlockChain) FindTransactionByTXid(id []byte) (Transaction, error) {

	//4. 如果沒找到，返回空Transaction，同時返回錯誤狀態

	fmt.Printf("1111111111 : id%x\n", id)
	it := bc.NewIterator()

	//1. 遍歷區塊鏈
	for {
		block := it.Next()
		//2. 遍歷交易
		for _, tx := range block.Transactions {
			//3. 比較交易，找到了直接退出
			if bytes.Equal(tx.TXID, id) {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			fmt.Printf("區塊鏈遍歷結束!\n")
			break
		}
	}

	return Transaction{}, errors.New("無效的交易id，請檢查!")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
	//簽名，交易建立的最後進行簽名
	prevTXs := make(map[string]Transaction)

	//找到所有引用的交易
	//1. 根據inputs來找，有多少input, 就遍歷多少次
	//2. 找到目標交易，（根據TXid來找）
	//3. 新增到prevTXs裡面
	for _, input := range tx.TXInputs {
		//根據id查詢交易本身，需要遍歷整個區塊鏈
		tx, err := bc.FindTransactionByTXid(input.TXid)

		if err != nil {
			log.Panic(err)
		}

		prevTXs[string(input.TXid)] = tx
		//第一個input查詢之後：prevTXs：
		// map[2222]Transaction222

		//第二個input查詢之後：prevTXs：
		// map[2222]Transaction222
		// map[3333]Transaction333

		//第三個input查詢之後：prevTXs：
		// map[2222]Transaction222
		// map[3333]Transaction333(只不過是重新寫了一次)
	}

	tx.Sign(privateKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {

	if tx.IsCoinbase() {
		return true
	}

	//簽名，交易建立的最後進行簽名
	prevTXs := make(map[string]Transaction)

	//找到所有引用的交易
	//1. 根據inputs來找，有多少input, 就遍歷多少次
	//2. 找到目標交易，（根據TXid來找）
	//3. 新增到prevTXs裡面
	for _, input := range tx.TXInputs {
		//根據id查詢交易本身，需要遍歷整個區塊鏈
		fmt.Printf("2222222 : %x\n", input.TXid)
		tx, err := bc.FindTransactionByTXid(input.TXid)

		if err != nil {
			log.Panic(err)
		}

		prevTXs[string(input.TXid)] = tx

	}

	return tx.Verify(prevTXs)
}
