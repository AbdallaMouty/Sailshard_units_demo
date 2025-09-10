package main

import (
    "bufio"
    "encoding/json"
 //   "fmt"
    "log"
    "net"
    "os"
    "sync"
    "time"
)

type TickMsg struct {
    Type string `json:"type"`
    Tick uint64 `json:"tick"`
    Ts   int64  `json:"ts"`
}

type EchoMsg struct {
    Type string `json:"type"`
    From string `json:"from"`
    Body string `json:"body"`
    Ts   int64  `json:"ts"`
}

func main() {
    addr := ":27015"
    ln, err := net.Listen("tcp", addr)
    if err != nil { log.Fatal(err) }
    defer ln.Close()
    log.Printf("Sailshard server running on %s", addr)

    var mu sync.Mutex
    clients := make(map[net.Conn]struct{})

    go func() {
        ticker := time.NewTicker(time.Second/20)
        defer ticker.Stop()
        var tick uint64
        for range ticker.C {
            tick++
            msg := TickMsg{"tick", tick, time.Now().UnixMilli()}
            data, _ := json.Marshal(msg)
            data = append(data, '\n')
            mu.Lock()
            for c := range clients { c.Write(data) }
            mu.Unlock()
        }
    }()

    go func() {
        for {
            conn, err := ln.Accept()
            if err != nil { continue }
            mu.Lock(); clients[conn]=struct{}{}; mu.Unlock()
            go handle(conn, func(){ mu.Lock(); delete(clients,conn); mu.Unlock() })
        }
    }()

    sc := bufio.NewScanner(os.Stdin)
    for sc.Scan() {
        line := sc.Text()
        msg := EchoMsg{"server_line","server",line,time.Now().UnixMilli()}
        data, _ := json.Marshal(msg); data = append(data,'\n')
        mu.Lock(); for c:=range clients{ c.Write(data) }; mu.Unlock()
    }
}

func handle(c net.Conn,onClose func()){defer func(){onClose();c.Close()}()
    peer:=c.RemoteAddr().String()
    sc:=bufio.NewScanner(c)
    for sc.Scan(){
        body:=sc.Text()
        resp:=EchoMsg{"echo",peer,body,time.Now().UnixMilli()}
        data,_:=json.Marshal(resp);data=append(data,'\n')
        c.Write(data)
    }}
