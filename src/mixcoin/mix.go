package mixcoin

import (
	"log"
	"time"
)

func mix(delay int, outAddr string) {
	log.Printf("mixing chunk to address: %s", outAddr)
	log.Printf("waiting %d blocks", delay)
	time.Sleep(time.Duration(delay) * 10 * time.Minute)

	outputChunk, err := GetPool().GetRandomChunk(Mixing)

	log.Printf("sending output chunk: %v", outputChunk)

	err := sendChunkWithFee(outputChunk, outAddr)
	if err != nil {
		log.Panicf("error sending chunk: ", err)
	}
}

func generateDelay(returnBy int) int {
	currHeight, err := getBlockchainHeight()
	if err != nil {
		log.Panicf("error getting blockchain height: %v", err)
	}
	log.Printf("generating delay with returnby %d and currheight %d", returnBy, currHeight)
	rand := randInt(returnBy - 1 - currHeight)
	log.Printf("generated delay %v", rand)
	return 0
	//return currHeight + rand
}
