package crypto

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func GetPrivateKey() (*ecdsa.PrivateKey, error) {
	joined, err := os.ReadFile("accounts.txt")
	if err != nil {
		fmt.Printf("Error reading private key from file error %s", err)
		return nil, err
	}
	parts := strings.Split(string(joined), "|")
	// 去掉前缀'0x'
	parts[1] = strings.TrimPrefix(parts[1], "0x")
	privateKeyBytes := parts[1]
	privateKey, err := ethcrypto.HexToECDSA(privateKeyBytes)
	if err != nil {
		fmt.Printf("Error converting private key hex to ECDSA error %s", err)
		return nil, err
	}
	return privateKey, nil
}

func TestVerifySignature(t *testing.T) {
	message := "hello world"
	nonce := "5hYj6byf9LY"
	// Add more test cases as needed
	privateKey, err := GetPrivateKey()
	if err != nil {
		t.Errorf("Error generating private key: %v", err)
		return
	}
	message = message + nonce
	fmt.Printf("Message: %v\n", message)
	msg := "\x19Ethereum Signed Message:\n" + fmt.Sprint(len(message)) + message
	messageBytes := []byte(msg)
	// message add eth format prefix
	signature, err := ethcrypto.Sign(ethcrypto.Keccak256(messageBytes), privateKey)
	if err != nil {
		t.Errorf("Error signing message: %v", err)
		return
	}
	signatureHex := common.Bytes2Hex(signature)
	signatureHex = "0x" + signatureHex
	fmt.Printf("Signature: %v\n", signatureHex)

	// Test case 4: Valid signature with generated private key
	walletAddr := ethcrypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	fmt.Printf("Wallet address: %v\n", walletAddr)
	err = VerifyETHSignature(walletAddr, message, signatureHex)
	if err != nil {
		t.Errorf("Error verifying signature: %v", err)
		return
	}
}
