package mixcoin

import (
	"time"
)

func mix(chunk *Chunk) {
	delay := generateDelay(chunk.ReturnBy)
	outAddr := chunk.OutAddr
	time.Sleep(delay * time.Second)

	outputChunk, err := PopRandomMixingChunk()
	if err != nil {
		log.Panicf("error popping random mixing chunk: ", err)
		return
	}
	err = sendChunk(outputChunk, outAddr)
	if err != nil {
		log.Panicf("error sending chunk: ", err)
	}
}

func generateDelay(returnBy int) int {
	return 1
}
