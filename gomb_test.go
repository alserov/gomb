package gomb

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const messagesAmount = 10

func TestGOMB(t *testing.T) {
	p, err := NewProducer(Params{Addr: "127.0.0.1:6336", Topic: "topic1", ID: "1"})
	defer func() {
		require.NoError(t, p.Close())
	}()
	require.NoError(t, err)
	require.NotEmpty(t, p)

	c, err := NewConsumer(Params{Addr: "127.0.0.1:6336", Topic: "topic1", ID: "2"})
	defer func() {
		require.NoError(t, c.Close())
	}()
	require.NoError(t, err)
	require.NotEmpty(t, c)

	wg := sync.WaitGroup{}

	go func(wg *sync.WaitGroup) {
		for i := range messagesAmount {
			wg.Add(1)
			go func() {
				err = p.Produce(&Message{
					Key: fmt.Sprintf("%d", i),
					Val: []byte(fmt.Sprintf("value №%d", i)),
				})
				require.NoError(t, err)
			}()
		}
	}(&wg)

	msgs := c.Consume()
	defer close(msgs)

	count := 0

	go func(wg *sync.WaitGroup) {
		for msg := range msgs {
			fmt.Println(msg)
			count++
			wg.Done()
		}
	}(&wg)

	time.Sleep(time.Millisecond * 10)
	wg.Wait()

	require.Equal(t, messagesAmount, count)
}

func TestMultipleConsumers(t *testing.T) {
	p, err := NewProducer(Params{Addr: "127.0.0.1:6336", Topic: "topic1", ID: "1"})
	defer func() {
		require.NoError(t, p.Close())
	}()
	require.NoError(t, err)
	require.NotEmpty(t, p)

	c, err := NewConsumer(Params{Addr: "127.0.0.1:6336", Topic: "topic1", ID: "2"})
	defer func() {
		require.NoError(t, c.Close())
	}()
	require.NoError(t, err)
	require.NotEmpty(t, c)

	c1, err := NewConsumer(Params{Addr: "127.0.0.1:6336", Topic: "topic1", ID: "3"})
	defer func() {
		require.NoError(t, c1.Close())
	}()
	require.NoError(t, err)
	require.NotEmpty(t, c)

	wg := sync.WaitGroup{}

	go func(wg *sync.WaitGroup) {
		for i := range messagesAmount {
			wg.Add(2)
			go func() {
				err = p.Produce(&Message{
					Key: fmt.Sprintf("%d", i),
					Val: []byte(fmt.Sprintf("value №%d", i)),
				})
				require.NoError(t, err)
			}()
		}
	}(&wg)

	msgs := c.Consume()
	defer close(msgs)

	msgs1 := c1.Consume()
	defer close(msgs1)

	count := atomic.Int32{}

	go func(wg *sync.WaitGroup) {
		for msg := range msgs {
			fmt.Println(msg)
			count.Add(1)
			wg.Done()
		}

	}(&wg)
	go func(wg *sync.WaitGroup) {
		for msg := range msgs1 {
			fmt.Println(msg)
			count.Add(1)
			wg.Done()
		}
	}(&wg)

	time.Sleep(time.Millisecond * 10)
	wg.Wait()

	require.Equal(t, messagesAmount*2, int(count.Load()))
}
