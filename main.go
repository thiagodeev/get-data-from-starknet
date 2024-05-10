package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/NethermindEth/starknet.go/rpc"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	provider, err := rpc.NewProvider(os.Getenv("RPC_URL"))
	if err != nil {
		log.Fatal(err)
	}

	latestBlockNumber, rpcErr := provider.BlockNumber(context.Background())
	if rpcErr != nil {
		log.Fatal(rpcErr.Data)
	}

	blockList := make(chan *rpc.Block, 10)
	var wg sync.WaitGroup

	var i uint64 = 1
	for i <= 10 {
		wg.Add(1)

		go func(i uint64) {
			defer wg.Done()
			getBlock(blockList, provider, i)
		}(i)
		i++
	}

	wg.Wait()

	var blockNumber uint64 = 1
	for blockNumber <= latestBlockNumber {
		block := <-blockList
		go getBlock(blockList, provider, i)
		fmt.Printf("Searching at Block number: %d \n", block.BlockNumber)
		if result := findV0Tsx(block); result {
			break
		}

		blockNumber++
		i++
	}

}

func getBlock(blockList chan *rpc.Block, provider *rpc.Provider, blockNumber uint64) error {
	fmt.Printf("Fetching Block number: %d \n", blockNumber)
	var block *rpc.Block

	result, err := provider.BlockWithTxs(context.Background(), rpc.BlockID{Number: &blockNumber})
	if err != nil {
		return err
	}

	block = result.(*rpc.Block)
	blockList <- block

	return nil
}

func findV0Tsx(block *rpc.Block) bool {
	for _, blockTransaction := range block.Transactions {
		switch tsx := blockTransaction.(type) {
		case rpc.BlockInvokeTxnV0:
			fmt.Println("v0 Transaction founded. Hash:")
			fmt.Println(tsx.TransactionHash.String())
			return true
		default:
			continue
		}
	}

	return false
}
