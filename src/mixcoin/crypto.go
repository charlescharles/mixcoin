package mixcoin

import (
	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

func randInt(high int) int {
	bigRet, err := rand.Int(rand.Reader, big.NewInt(int64(high)))
	if err != nil {
		log.Panicf("error generating random int: %v", err)
	}

	return int(bigRet.Int64())
}

func signChunkMessage(chunkMsg *ChunkMessage) error {
	log.Printf("signing chunk message")
	marshaledBytes, _ := json.Marshal(chunkMsg)
	marshaledBuf := bytes.NewBuffer(marshaledBytes)

	cfg := GetConfig()

	keyringFileBuffer, err := os.Open(cfg.PrivRingFile)
	if err != nil {
		log.Panicf("error opening privring file")
		return err
	}

	defer keyringFileBuffer.Close()

	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		fmt.Println(err)
		return err
	}
	entity := entityList[0]
	passphrasebyte := []byte(cfg.Passphrase)
	entity.PrivateKey.Decrypt(passphrasebyte)
	for _, subkey := range entity.Subkeys {
		subkey.PrivateKey.Decrypt(passphrasebyte)
	}
	armoredSigBuf := new(bytes.Buffer)
	err = openpgp.ArmoredDetachSign(armoredSigBuf, entity, marshaledBuf, nil)
	if err != nil {
		fmt.Println(err)
		return err
	}
	armoredSigEnc, err := ioutil.ReadAll(armoredSigBuf)
	if err != nil {
		fmt.Println(err)
		return err
	}
	armoredSig := base64.StdEncoding.EncodeToString(armoredSigEnc)

	chunkMsg.Warrant = armoredSig

	return nil
}
