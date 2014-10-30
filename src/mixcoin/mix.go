package mixcoin

import (
	"log"
	"time"
)

func mix(delay int, outAddr string) {
	time.Sleep(time.Duration(delay) * 10 * time.Minute)

	requestMixingChunkC <- true
	outputChunk := <-randMixingChunkC

	err := sendChunk(outputChunk, outAddr)
	if err != nil {
		log.Panicf("error sending chunk: ", err)
	}
}

func generateDelay(returnBy int) int {
	cfg := GetConfig()
	currHeight, err := getBlockchainHeight()
	if err != nil {
		log.Panicf("error getting blockchain height: %v", err)
	}
	rand := randInt(returnBy - 1 - currHeight)
	return currHeight + rand
}
