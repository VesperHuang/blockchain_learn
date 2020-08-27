package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"crypto/sha256"
	//"golang.org/x/crypto/ripemd160"
	"./lib/ripemd160"
	//"github.com/btcsuite/btcutil/base58"
	"./lib/base58"
	"fmt"
	"bytes"
)

//這裡的錢包時一結構，每一個錢包儲存了公鑰,私鑰對

type Wallet struct {
	//私鑰
	Private *ecdsa.PrivateKey
	//PubKey *ecdsa.PublicKey
	//約定，這裡的PubKey不儲存原始的公鑰，而是儲存X和Y拼接的字串，在校驗端重新拆分（參考r,s傳遞）
	PubKey []byte //
}

//建立錢包
func NewWallet() *Wallet {
	//建立曲線
	curve := elliptic.P256()
	//產生私鑰
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic()
	}

	//產生公鑰
	pubKeyOrig := privateKey.PublicKey

	//拼接X, Y
	pubKey := append(pubKeyOrig.X.Bytes(), pubKeyOrig.Y.Bytes()...)

	return &Wallet{Private: privateKey, PubKey: pubKey}
}

//產生地址
func (w *Wallet) NewAddress() string {
	pubKey := w.PubKey

	rip160HashValue := HashPubKey(pubKey)
	version := byte(00)
	//拼接version
	payload := append([]byte{version}, rip160HashValue...)

	//checksum
	checkCode := CheckSum(payload)

	//25位元組數據
	payload = append(payload, checkCode...)

	//go語言有一個庫，叫做btcd,這個是go語言實現的比特幣全節點原始碼
	address := base58.Encode(payload)

	return address
}

func HashPubKey(data []byte) []byte {
	hash := sha256.Sum256(data)

	//理解為編碼器
	rip160hasher := ripemd160.New()
	_, err := rip160hasher.Write(hash[:])

	if err != nil {
		log.Panic(err)
	}

	//返回rip160的雜湊結果
	rip160HashValue := rip160hasher.Sum(nil)
	return rip160HashValue
}

func CheckSum(data []byte) []byte {
	//兩次sha256
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(hash1[:])

	//前4位元組校驗碼
	checkCode := hash2[:4]
	return checkCode
}

func IsValidAddress(address string) bool {
	//1. 解碼
	addressByte := base58.Decode(address)

	if len(addressByte) < 4 {
		return false
	}

	//2. 取數據
	payload := addressByte[:len(addressByte)-4]
	checksum1 := addressByte[len(addressByte)-4: ]

	//3. 做checksum函式
	checksum2 := CheckSum(payload)

	fmt.Printf("checksum1 : %x\n", checksum1)
	fmt.Printf("checksum2 : %x\n", checksum2)

	//4. 比較
	return bytes.Equal(checksum1, checksum2)
}
