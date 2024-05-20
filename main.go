package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "os/signal"
    "time"
    "github.com/gorilla/websocket"
)

type JSONParams struct {
    Action string   `json:"action"`
    Subs   []string `json:"subs"`
}

func main() {
    log.SetFlags(0)

    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt)

    apiKey := "YOUR_API_KEY"
    c, _, err := websocket.DefaultDialer.Dial("wss://streamer.cryptocompare.com/v2?api_key=" + apiKey, nil)
    if err != nil {
        log.Fatal("dial:", err)
    }

    jsonObj := JSONParams{Action: "SubAdd", Subs: []string{"0~Coinbase~BTC~USD", "0~Kraken~ADA~USD", "0~Kraken~DOGE~USD", "0~Coinbase~ETH~USD"}}
    s, _ := json.Marshal(jsonObj)
    fmt.Println(string(s))
    err = c.WriteMessage(websocket.TextMessage, []byte(string(s)))
    if err != nil {
        log.Fatal("message:", err)
    }

    defer c.Close()

    done := make(chan struct{})

    go func() {
        defer close(done)
        for {
            _, message, err := c.ReadMessage()
            if err != nil {
                log.Println("read:", err)
                return
            }
            log.Printf("recv: %s", message)
        }
    }()

    for {
        select {
        case <-done:
            return
        case <-interrupt:
            log.Println("interrupt")

            err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
            if err != nil {
                log.Println("write close:", err)
                return
            }
            select {
            case <-done:
            case <-time.After(time.Second):
            }
            return
        }
    }
}
