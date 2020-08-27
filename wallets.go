package main

import (
	"io/ioutil"
	"bytes"
	"encoding/gob"
	"log"
	"crypto/elliptic"
	"os"
	"go一期/lib/base58"
)

const walletFile = "wallet.dat"

//定一個 Wallets結構，它儲存所有的wallet以及它的地址
type Wallets struct {
	//map[地址]錢包
	WalletsMap map[string]*Wallet
}

//建立方法，返回目前所有錢包的實例
func NewWallets() *Wallets {
	var ws Wallets
	ws.WalletsMap = make(map[string]*Wallet)
	ws.loadFile()
	return &ws
}

func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := wallet.NewAddress()

	ws.WalletsMap[address] = wallet

	ws.saveToFile()
	return address
}

//儲存方法，把新建的wallet新增進去
func (ws *Wallets) saveToFile() {

	var buffer bytes.Buffer

	//panic: gob: type not registered for interface: elliptic.p256Curve
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(ws)
	//一定要注意校驗！！！
	if err != nil {
		log.Panic(err)
	}

	ioutil.WriteFile(walletFile, buffer.Bytes(), 0600)
}

//讀取檔案方法，把所有的wallet讀出來
func (ws *Wallets) loadFile() {
	//在讀取之前，要先確認檔案是否在，如果不存在，直接退出
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		//ws.WalletsMap = make(map[string]*Wallet)
		return
	}

	//讀取內容
	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	//解碼
	//panic: gob: type not registered for interface: elliptic.p256Curve
	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(content))

	var wsLocal Wallets

	err = decoder.Decode(&wsLocal)
	if err != nil {
		log.Panic(err)
	}

	//ws = &wsLocal
	//對於結構來說，裡面有map的，要指定賦值，不要再最外層直接賦值
	ws.WalletsMap = wsLocal.WalletsMap
}

func (ws *Wallets) ListAllAddresses() []string {
	var addresses []string
	//遍歷錢包，將所有的key取出來返回
	for address := range ws.WalletsMap {
		addresses = append(addresses, address)
	}

	return addresses
}

//通過地址返回公鑰的雜湊值
func GetPubKeyFromAddress(address string) []byte {
	//1. 解碼
	//2. 擷取出公鑰雜湊：去除version（1位元組），去除校驗碼（4位元組）
	addressByte := base58.Decode(address) //25位元組
	len := len(addressByte)

	pubKeyHash := addressByte[1:len-4]

	return pubKeyHash
}
