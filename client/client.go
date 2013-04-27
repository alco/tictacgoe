package main

import (
	"fmt"
	"math/rand"
    "net/rpc"
)

type ArbitrageData struct {
	Nums   []int
	Player int
	Index  int
}

func main() {
    client, err := rpc.Dial("tcp", "localhost:8888")
    if err != nil {
        panic(err)
    }
    var reply ArbitrageData
	var nums = make([]int, 10)
	for i := range nums {
		nums[i] = 1+rand.Int() % 2  // 1 - my turn; 2 - his turn
	}
    err = client.Call("BoardRPC.WhoseTurn", nums, &reply)
    if err != nil {
        panic(err)
    }

	var firstPlayer, index int
	for i := range nums {
		var myNum, hisNum = nums[i], reply.Nums[i]
		if myNum == hisNum {
			firstPlayer = myNum
			index = i
			break
		}
	}
	fmt.Printf("Nums = %v\n", nums)
	fmt.Printf("player = %v\n", firstPlayer)
	fmt.Printf("index = %v\n", index)

	var serverOK bool
    err = client.Call("BoardRPC.AgreeOnTurn", firstPlayer, &serverOK)
    if err != nil {
        panic(err)
    }

	if !serverOK {
		panic("Could not agree on turn")
	}

	println("*** Game started ***")

	/*if firstPlayer != reply.Player || index != reply.Index {*/
		/*panic("Server is a CHEATER!!!")*/
	/*}*/
    /*println("First player: ", firstPlayer)*/
}
