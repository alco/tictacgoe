package main

import (
    "net/rpc"
)

func main() {
    client, err := rpc.Dial("tcp", "localhost:8888")
    if err != nil {
        panic(err)
    }
    var reply int
    err = client.Call("BoardRPC.MakeMove", [2]int{1,2}, &reply)
    if err != nil {
        panic(err)
    }
    println(reply)
}
