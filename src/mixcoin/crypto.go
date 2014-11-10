package mixcoin

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"os"
	"strings"

	"code.google.com/p/go.crypto/openpgp"
	"code.google.com/p/go.crypto/openpgp/armor"
)

func randInt(high int) int {
	bigRet, err := rand.Int(rand.Reader, big.NewInt(int64(high)))
	if err != nil {
		log.Panicf("error generating random int: %v", err)
	}

	return int(bigRet.Int64())
}

func serialize(chunkMsg *ChunkMessage) string {
	marshaledBytes, _ := json.Marshal(chunkMsg)
	return string(marshaledBytes)
}

func signText(entity *openpgp.Entity, text string) string {
	b := bytes.NewBuffer(nil)
	w, _ := armor.Encode(b, openpgp.SignatureType, nil)
	err := openpgp.DetachSignText(w, entity, strings.NewReader(text), nil)
	if err != nil {
		panic(err)
	}
	w.Close()
	return b.String()
}

func verifySignature(pubKey, signed, signatureArmor string) bool {
	keyRing, err := openpgp.ReadArmoredKeyRing(strings.NewReader(pubKey))
	if err != nil {
		panic(err)
	}

	signer, err := openpgp.CheckArmoredDetachedSignature(
		keyRing,
		strings.NewReader(signed),
		strings.NewReader(signatureArmor),
	)
	if err != nil {
		panic(err)
	}
	return signer != nil
}

func getPgpEntity() *openpgp.Entity {
	keyringFileBuffer, err := os.Open(cfg.PrivRingFile)
	if err != nil {
		panic(err)
	}

	defer keyringFileBuffer.Close()

	entityList, err := openpgp.ReadKeyRing(keyringFileBuffer)
	if err != nil {
		panic(err)
	}
	entity := entityList[0]
	passphrasebyte := []byte(cfg.Passphrase)
	entity.PrivateKey.Decrypt(passphrasebyte)
	for _, subkey := range entity.Subkeys {
		subkey.PrivateKey.Decrypt(passphrasebyte)
	}
	return entity
}

func signChunkMessage(chunkMsg *ChunkMessage) {
	log.Printf("signing chunk message")

	entity := getPgpEntity()
	serialized := serialize(chunkMsg)
	signature := signText(entity, serialized)
	chunkMsg.Warrant = signature
}

// NOTE: not sure if this works (if setting Warrant to '' removes
// it from the json)
func verifyWarrant(chunkMsg *ChunkMessage, pubKey string) bool {
	warrant := chunkMsg.Warrant
	msgCopy := *chunkMsg
	msgCopy.Warrant = ""
	serialized := serialize(&msgCopy)

	return verifySignature(pubKey, serialized, warrant)
}
