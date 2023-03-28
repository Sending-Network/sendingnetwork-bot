package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	ethereumcrypto "github.com/ethereum/go-ethereum/crypto"
	"math/rand"
	"strings"
	"time"
)

func generateWalletAccount() (string, *ecdsa.PrivateKey, error) {
	rand.Seed(time.Now().UnixNano())
	entropyBase := "0000000000000000000000000000000"
	for i := 0; i < 9; i++ {
		entropyBase = fmt.Sprintf("%v%v", entropyBase, rand.Intn(10))
	}
	entropy := []byte(entropyBase)
	reader := bytes.NewReader(entropy)
	privateKey, err := ecdsa.GenerateKey(ethereumcrypto.S256(), reader)
	if err != nil {
		return "", nil, err
	}
	address := strings.ToLower(ethereumcrypto.PubkeyToAddress(*privateKey.Public().(*ecdsa.PublicKey)).String())
	return address, privateKey, nil
}

func main() {
	address, privateKey, err := generateWalletAccount()
	if err != nil {
		panic(err)
	}
	fmt.Printf("WalletAddress: %s\n", address)
	fmt.Printf("PrivateKey: %s\n", hex.EncodeToString(ethereumcrypto.FromECDSA(privateKey)))
}
