package mixcoin

import (
	"log"
	"time"
)

func mix(delay int, outAddr string) {
	// TODO: delay * time.Second
	time.Sleep(time.Second)

	requestMixingChunkC <- true
	var outputChunk *Chunk
	outputChunk = <-randMixingChunkC

	err := sendChunk(outputChunk, outAddr)
	if err != nil {
		log.Panicf("error sending chunk: ", err)
	}
}

func generateDelay(returnBy int) int {
	return 1
}
