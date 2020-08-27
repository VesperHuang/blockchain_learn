package main

import (
	"os"
	"fmt"
	"strconv"
)

//這是一個用來接收命令列參數並且控制區塊鏈操作的檔案

type CLI struct {
	bc *BlockChain
}

const Usage = `
	printChain               "正向列印區塊鏈"
	printChainR              "反向列印區塊鏈"
	getBalance --address ADDRESS "獲取指定地址的餘額"
	send FROM TO AMOUNT MINER DATA "由FROM轉AMOUNT給TO，由MINER挖礦，同時寫入DATA"
	newWallet   "建立一個新的錢包(私鑰公鑰對)"
	listAddresses "列舉所有的錢包地址"
`

//接受參數的動作，我們放到一個函式中

func (cli *CLI) Run() {

	//./block printChain
	//./block addBlock --data "HelloWorld"
	//1. 得到所有的命令
	args := os.Args
	if len(args) < 2 {
		fmt.Printf(Usage)
		return
	}

	//2. 分析命令
	cmd := args[1]
	switch cmd {
	case "printChain":
		fmt.Printf("正向列印區塊\n")
		cli.PrinBlockChain()
	case "printChainR":
		fmt.Printf("反向列印區塊\n")
		cli.PrinBlockChainReverse()
	case "getBalance":
		fmt.Printf("獲取餘額\n")
		if len(args) == 4 && args[2] == "--address" {
			address := args[3]
			cli.GetBalance(address)
		}
	case "send":
		fmt.Printf("轉賬開始...\n")
		if len(args) != 7 {
			fmt.Printf("參數個數錯誤，請檢查！\n")
			fmt.Printf(Usage)
			return
		}
		//./block send FROM TO AMOUNT MINER DATA "由FROM轉AMOUNT給TO，由MINER挖礦，同時寫入DATA"
		from := args[2]
		to := args[3]
		amount, _ := strconv.ParseFloat(args[4], 64) //知識點，請注意
		miner := args[5]
		data := args[6]
		cli.Send(from, to, amount, miner, data)
	case "newWallet":
		fmt.Printf("建立新的錢包...\n")
		cli.NewWallet()
	case "listAddresses":
		fmt.Printf("列舉所有地址...\n")
		cli.ListAddresses()
	default:
		fmt.Printf("無效的命令，請檢查!\n")
		fmt.Printf(Usage)
	}
}
