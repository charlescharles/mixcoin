package mixcoin

import (
	"github.com/stretchr/testify/assert"
	"log"
	"strconv"
	"testing"
)

var (
	chunkTemplate = &ChunkMessage{
		Val:      100,
		SendBy:   300,
		ReturnBy: 300,
		OutAddr:  "charles",
		Fee:      2,
		Nonce:    123,
		Confirm:  1,

		MixAddr: "whatevs",
		Warrant: "i signed this",
	}
)

func TestMix(t *testing.T) {
	var chunks []*ChunkMessage
	var addrs []string
	for i := 0; i < 10; i++ {
		chunk := *chunkTemplate
		addr := "addr" + strconv.Itoa(i+1)
		chunk.OutAddr = addr
		addrs = append(addrs, addr)
		chunks = append(chunks, &chunk)
	}

	ch := make(chan string, 10)
	mix := NewMix(ch)

	for _, chunk := range chunks {
		mix.Put(chunk)
	}

	var sent []string
	for i := 0; i < 10; i++ {
		sent = append(sent, <-ch)
	}

	select {
	case <-ch:
		assert.Fail(t, "mix shouldn't have any more addresses")
	default:
		log.Printf("mix is empty as expected")
	}

	log.Printf("want: %v", addrs)
	log.Printf("got: %v", sent)
	assert.True(t, setEquals(sent, addrs), "mix should send the same addrs it got")
}
