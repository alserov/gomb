# GOMB - message broker <img src="./gombsvg.svg" alt="drawing" style="width:60px;"/>

## Producer example

*By default, producer will create a new topic if it is first initialisation, if it is not so, it will produce to existing one*

```go
package main

import (
	"fmt"
	"github.com/alserov/gomb"
	"time"
)

func main() {
	p, err := gomb.NewProducer(gomb.Params{Addr: "127.0.0.1:6336", Topic: "topic1", ID: "1"})
	defer p.Close()
	if err != nil {
		...
	}

	err = p.Produce(&gomb.Message{
		Key:    fmt.Sprintf("key"),
		Val:    []byte("it's message from gomb"),
	})
	if err != nil {
		...
    }
}

```


## Consumer example

*Requires topic to already exists before consuming, predefine it in env parameters of a server or by creating a producer*

```go
package main

import (
	"fmt"
	"github.com/alserov/gomb"
)

func main() {
	c, err := gomb.NewConsumer(gomb.Params{Addr: "127.0.0.1:6336", Topic: "topic1"})
	defer c.Close()
	if err != nil {
		...
	}

	chMsgs, chErr := c.Consume()
	go func() {
		for err = range chErr {
			...
		}
	}()

	for msg := range chMsgs {
		fmt.Println("received message: ", msg)
	}
}

```


## Env parameters for server

```dotenv
GOMB_ADDR="0.0.0.0:6336" // address where server will be serving
GOMB_TOPICS="topic1,topic2" // predefined topics
```

