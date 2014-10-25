package mixcoin

import (
	"bytes"
	"code.google.com/p/go.crypto/openpgp"
	"encoding/base64"
	"encoding/json"
	"ioutil"
	"log"
	"os"
)

func signChunk(chunk *Chunk) error {
	log.Printf("signing chunk")
	marshaledBytes, _ := json.Marshal(chunk)
	marshaledBuf := bytes.NewBuffer(marshaledBytes)

	cfg = GetConfig()

	keyringFileBuffer, err := os.Open(cfg.PrivRingFile)
	if err != nil {
		log.Panicf(err)
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

	chunk.Warrant = armoredSig

	return nil
}
