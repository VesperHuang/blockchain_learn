package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"crypto/sha256"
	"fmt"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"crypto/elliptic"
	"strings"
)

const reward = 50

//1. 定義交易結構
type Transaction struct {
	TXID      []byte     //交易ID
	TXInputs  []TXInput  //交易輸入陣列
	TXOutputs []TXOutput //交易輸出的陣列
}

//定義交易輸入
type TXInput struct {
	//引用的交易ID
	TXid []byte
	//引用的output的索引值
	Index int64
	//解鎖指令碼，我們用地址來模擬
	//Sig string

	//真正的數字簽名，由r，s拼成的[]byte
	Signature []byte

	//約定，這裡的PubKey不儲存原始的公鑰，而是儲存X和Y拼接的字串，在校驗端重新拆分（參考r,s傳遞）
	//注意，是公鑰，不是雜湊，也不是地址
	PubKey []byte
}

//定義交易輸出
type TXOutput struct {
	//轉賬金額
	Value float64
	//鎖定指令碼,我們用地址模擬
	//PubKeyHash string

	//收款方的公鑰的雜湊，注意，是雜湊而不是公鑰，也不是地址
	PubKeyHash []byte
}

//由於現在儲存的欄位是地址的公鑰雜湊，所以無法直接建立TXOutput，
//爲了能夠得到公鑰雜湊，我們需要處理一下，寫一個Lock函式
func (output *TXOutput) Lock(address string) {
	//1. 解碼
	//2. 擷取出公鑰雜湊：去除version（1位元組），去除校驗碼（4位元組）

	//真正的鎖定動作！！！！！
	output.PubKeyHash = GetPubKeyFromAddress(address)
}

//給TXOutput提供一個建立的方法，否則無法呼叫Lock
func NewTXOutput(value float64, address string) *TXOutput {
	output := TXOutput{
		Value: value,
	}

	output.Lock(address)
	return &output
}

//設定交易ID
func (tx *Transaction) SetHash() {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	data := buffer.Bytes()
	hash := sha256.Sum256(data)
	tx.TXID = hash[:]
}

//實現一個函式，判斷目前的交易是否為挖礦交易
func (tx *Transaction) IsCoinbase() bool {
	//1. 交易input只有一個
	//if len(tx.TXInputs) == 1  {
	//	input := tx.TXInputs[0]
	//	//2. 交易id為空
	//	//3. 交易的index 為 -1
	//	if !bytes.Equal(input.TXid, []byte{}) || input.Index != -1 {
	//		return false
	//	}
	//}
	//return true

	if len(tx.TXInputs) == 1 && len(tx.TXInputs[0].TXid) == 0 && tx.TXInputs[0].Index == -1 {
		return true
	}

	return false
}

//2. 提供建立交易方法(挖礦交易)
func NewCoinbaseTX(address string, data string) *Transaction {
	//挖礦交易的特點：
	//1. 只有一個input
	//2. 無需引用交易id
	//3. 無需引用index
	//礦工由於挖礦時無需指定簽名，所以這個PubKey欄位可以由礦工自由填寫數據，一般是填寫礦池的名字
	//簽名先填寫為空，後面建立完整交易后，最後做一次簽名即可
	input := TXInput{[]byte{}, -1, nil, []byte(data)}
	//output := TXOutput{reward, address}

	//新的建立方法
	output := NewTXOutput(reward, address)

	//對於挖礦交易來說，只有一個input和一output
	tx := Transaction{[]byte{}, []TXInput{input}, []TXOutput{*output}}
	tx.SetHash()

	return &tx
}

//建立普通的轉賬交易
//3. 建立outputs
//4. 如果有零錢，要找零

func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {

	//1. 建立交易之後要進行數字簽名->所以需要私鑰->打開錢包"NewWallets()"
	ws := NewWallets()

	//2. 找到自己的錢包，根據地址返回自己的wallet
	wallet := ws.WalletsMap[from]
	if wallet == nil {
		fmt.Printf("沒有找到該地址的錢包，交易建立失敗!\n")
		return nil
	}

	//3. 得到對應的公鑰，私鑰
	pubKey := wallet.PubKey
	privateKey := wallet.Private //稍後再用

	//傳遞公鑰的雜湊，而不是傳遞地址
	pubKeyHash := HashPubKey(pubKey)

	//1. 找到最合理UTXO集合 map[string][]uint64
	utxos, resValue := bc.FindNeedUTXOs(pubKeyHash, amount)

	if resValue < amount {
		fmt.Printf("餘額不足，交易失敗!")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//2. 建立交易輸入, 將這些UTXO逐一轉成inputs
	//map[2222] = []int64{0}
	//map[3333] = []int64{0, 1}
	for id, indexArray := range utxos {
		for _, i := range indexArray {
			input := TXInput{[]byte(id), int64(i), nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	//建立交易輸出
	//output := TXOutput{amount, to}
	output := NewTXOutput(amount, to)
	outputs = append(outputs, *output)

	//找零
	if resValue > amount {
		output = NewTXOutput(resValue-amount, from)
		outputs = append(outputs, *output)
	}

	tx := Transaction{[]byte{}, inputs, outputs}
	tx.SetHash()

	bc.SignTransaction(&tx, privateKey)

	return &tx
}

//簽名的具體實現,
// 參數為：私鑰，inputs裡面所有引用的交易的結構map[string]Transaction
//map[2222]Transaction222
//map[3333]Transaction333
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	//1. 建立一個目前交易的副本：txCopy，使用函式： TrimmedCopy：要把Signature和PubKey欄位設定為nil
	txCopy := tx.TrimmedCopy()
	//2. 循環遍歷txCopy的inputs，得到這個input索引的output的公鑰雜湊
	for i, input := range txCopy.TXInputs {
		prevTX := prevTXs[string(input.TXid)]
		if len(prevTX.TXID) == 0 {
			log.Panic("引用的交易無效")
		}

		//不要對input進行賦值，這是一個副本，要對txCopy.TXInputs[xx]進行操作，否則無法把pubKeyHash傳進來
		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash

		//所需要的三個數據都具備了，開始做雜湊處理
		//3. 產生要簽名的數據。要簽名的數據一定是雜湊值
		//a. 我們對每一個input都要簽名一次，簽名的數據是由目前input引用的output的雜湊+目前的outputs（都承載在目前這個txCopy裡面）
		//b. 要對這個拼好的txCopy進行雜湊處理，SetHash得到TXID，這個TXID就是我們要簽名最終數據。
		txCopy.SetHash()

		//還原，以免影響後面input的簽名
		txCopy.TXInputs[i].PubKey = nil
		//signDataHash認為是原始數據
		signDataHash := txCopy.TXID
		//4. 執行簽名動作得到r,s位元組流
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, signDataHash)
		if err != nil {
			log.Panic(err)
		}

		//5. 放到我們所簽名的input的Signature中
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TXInputs[i].Signature = signature
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.TXInputs {
		inputs = append(inputs, TXInput{input.TXid, input.Index, nil, nil})
	}

	for _, output := range tx.TXOutputs {
		outputs = append(outputs, output)
	}

	return Transaction{tx.TXID, inputs, outputs}
}

//分析校驗：
//所需要的數據：公鑰，數據(txCopy，產生雜湊), 簽名
//我們要對每一個簽名過得input進行校驗

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	//1. 得到簽名的數據
	txCopy := tx.TrimmedCopy()

	for i, input := range tx.TXInputs {
		prevTX := prevTXs[string(input.TXid)]
		if len(prevTX.TXID) == 0 {
			log.Panic("引用的交易無效")
		}

		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash
		txCopy.SetHash()
		dataHash := txCopy.TXID
		//2. 得到Signature, 反推會r,s
		signature := input.Signature //拆，r,s
		//3. 拆解PubKey, X, Y 得到原生公鑰
		pubKey := input.PubKey //拆，X, Y


		//1. 定義兩個輔助的big.int
		r := big.Int{}
		s := big.Int{}

		//2. 拆分我們signature，平均分，前半部分給r, 後半部分給s
		r.SetBytes(signature[:len(signature)/2 ])
		s.SetBytes(signature[len(signature)/2:])


		//a. 定義兩個輔助的big.int
		X := big.Int{}
		Y := big.Int{}

		//b. pubKey，平均分，前半部分給X, 後半部分給Y
		X.SetBytes(pubKey[:len(pubKey)/2 ])
		Y.SetBytes(pubKey[len(pubKey)/2:])

		//還原原始的公鑰
		pubKeyOrigin := ecdsa.PublicKey{elliptic.P256(), &X, &Y}

		//4. Verify
		if !ecdsa.Verify(&pubKeyOrigin, dataHash, &r, &s) {
			return false
		}
	}

	return true
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXID))

	for i, input := range tx.TXInputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TXid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.TXOutputs{
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %f", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}


