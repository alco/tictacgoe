package main

import (
	"net/rpc"
)

func connectToServer(address string) (int, *rpc.Client) {
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	var nums = genNums()
	var serverNums []int
	err = client.Call("BoardRPC.WhoseTurn", nums, &serverNums)
	if err != nil {
		panic(err)
	}

	var firstPlayer = getFirstPlayer(nums, serverNums)
	var serverOK bool
	err = client.Call("BoardRPC.AgreeOnTurn", firstPlayer, &serverOK)
	if err != nil {
		panic(err)
	}

	if !serverOK {
		panic("Could not agree on turn")
	}

	return 3 - firstPlayer, client // First player is from the server's point of view
}

func rpc_makeMove(client *rpc.Client, coords [2]int) (int, error) {
	var err error

	var result int
	err = client.Call("BoardRPC.MakeMove", coords, &result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func rpc_waitForOpponent(client *rpc.Client) (coords [2]int, result int, err error) {
	var data = MoveData{}

	err = client.Call("BoardRPC.MakeOwnMove", 0, &data)
	if err != nil {
		return
	}

	return data.Coords, data.Result, nil
}

func rpc_finishGameWithResult(client *rpc.Client, result int) error {
	err := client.Call("BoardRPC.FinishGame", result, nil)
	if err != nil {
		return err
	}
	return nil
}
