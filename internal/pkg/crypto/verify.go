package crypto

import (
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func VerifyETHSignature(address string, message string, signature string) error {
	// Hash the unsigned message using EIP-191
	hashedMessage := []byte("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(message)) + message)
	hash := crypto.Keccak256Hash(hashedMessage)

	decodedMessage, err := hexutil.Decode(signature)
	if err != nil {
		return errors.Wrap(err, "signature invalid")
	}
	if len(decodedMessage) != 65 {
		return errors.New("signature invalid")
	}
	// Handles cases where EIP-115 is not implemented (most wallets don't implement it)
	if decodedMessage[64] == 27 || decodedMessage[64] == 28 {
		decodedMessage[64] -= 27
	}

	// Recover a public key from the signed message
	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), decodedMessage)
	if sigPublicKeyECDSA == nil {
		err = errors.New("could not get a public get from the message signature")
	}
	if err != nil {
		return err
	}
	expected := strings.ToLower(crypto.PubkeyToAddress(*sigPublicKeyECDSA).String())
	get := strings.ToLower(address)
	if expected != get {
		return errors.Errorf("address and signature not match, expected %s but get %s", expected, get)
	}
	return nil
}
