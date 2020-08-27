package main

import (
	"fmt"
	//"time"
)

//正向列印
func (cli *CLI) PrinBlockChain() {
	cli.bc.Printchain()
	fmt.Printf("列印區塊鏈完成\n")
}

//反向列印
func (cli *CLI) PrinBlockChainReverse() {
	bc := cli.bc
	//建立迭代器
	it := bc.NewIterator()

	//呼叫迭代器，返回我們的每一個區塊數據
	for {
		//返回區塊，左移
		block := it.Next()

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		/*
		fmt.Printf("===========================\n\n")
		fmt.Printf("版本號: %d\n", block.Version)
		fmt.Printf("前區塊雜湊值: %x\n", block.PrevHash)
		fmt.Printf("梅克爾根: %x\n", block.MerkelRoot)
		timeFormat := time.Unix(int64(block.TimeStamp), 0).Format("2006-01-02 15:04:05")
		fmt.Printf("時間戳: %s\n", timeFormat)
		fmt.Printf("難度值(隨便寫的）: %d\n", block.Difficulty)
		fmt.Printf("隨機數 : %d\n", block.Nonce)
		fmt.Printf("目前區塊雜湊值: %x\n", block.Hash)
		fmt.Printf("區塊數據 :%s\n", block.Transactions[0].TXInputs[0].PubKey)
		*/

		if len(block.PrevHash) == 0 {
			fmt.Printf("區塊鏈遍歷結束！")
			break
		}
	}
}

func (cli *CLI) GetBalance(address string) {

	//1. 校驗地址
	if !IsValidAddress(address) {
		fmt.Printf("地址無效 : %s\n", address)
		return
	}

	//2. 產生公鑰雜湊
	pubKeyHash := GetPubKeyFromAddress(address)

	utxos := cli.bc.FindUTXOs(pubKeyHash)

	total := 0.0
	for _, utxo := range utxos {
		total += utxo.Value
	}

	fmt.Printf("\"%s\"的餘額為：%f\n", address, total)
}

func (cli *CLI) Send(from, to string, amount float64, miner, data string) {
	//fmt.Printf("from : %s\n", from)
	//fmt.Printf("to : %s\n", to)
	//fmt.Printf("amount : %f\n", amount)
	//fmt.Printf("miner : %s\n", miner)
	//fmt.Printf("data : %s\n", data)
	//1. 校驗地址
	if !IsValidAddress(from) {
		fmt.Printf("地址無效 from: %s\n", from)
		return
	}
	if !IsValidAddress(to) {
		fmt.Printf("地址無效 to: %s\n", to)
		return
	}
	if !IsValidAddress(miner) {
		fmt.Printf("地址無效 miner: %s\n", miner)
		return
	}

	//1. 建立挖礦交易
	coinbase := NewCoinbaseTX(miner, data)
	//2. 建立一個普通交易
	tx := NewTransaction(from, to, amount, cli.bc)
	if tx == nil {
		//fmt.Printf("無效的交易")
		return
	}
	//3. 新增到區塊

	cli.bc.AddBlock([]*Transaction{coinbase, tx})
	fmt.Printf("轉賬成功！")
}

func (cli *CLI) NewWallet() {
	ws := NewWallets()
	address := ws.CreateWallet()
	fmt.Printf("地址：%s\n", address)

}

func (cli *CLI) ListAddresses() {
	ws := NewWallets()
	addresses := ws.ListAllAddresses()
	for _, address := range addresses {
		fmt.Printf("地址：%s\n", address)
	}
}
