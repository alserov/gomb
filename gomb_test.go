package gomb

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

func TestGOMB(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(19)

	p, err := NewProducer(Params{Addr: "127.0.0.1:6336", Topic: "Test", ID: "1"})
	require.NoError(t, err)
	require.NotEmpty(t, p)
	defer p.Close()

	for i := range 10 {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			err = p.Produce(&Message{
				Key:    fmt.Sprintf("%d", i),
				Val:    []byte(fmt.Sprintf("value №%d", i)),
				SentAt: time.Now(),
			})
			require.NoError(t, err)
		}(&wg)
	}

	go func(wg *sync.WaitGroup) {
		c, err := NewConsumer(Params{Addr: "127.0.0.1:6336", Topic: "Test"})
		require.NoError(t, err)
		require.NotEmpty(t, c)

		chMsgs, chErr := c.Consume()

		go func() {
			for err = range chErr {
				require.NoError(t, err)
			}
		}()

		for msg := range chMsgs {
			wg.Done()
			fmt.Println("received message: ", msg)
		}
	}(&wg)

	wg.Wait()
}
