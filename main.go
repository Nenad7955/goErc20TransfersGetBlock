package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

var activity map[string]int
var mutex = &sync.Mutex{}

func main() {
	activity = make(map[string]int)

	test, err := GetEntity[BlockNumberResponse](Request{
		Jsonrpc: "2.0",
		Method:  "eth_blockNumber",
		ID:      "getblock.io",
	})

	latestBlock, err := strconv.ParseInt(test.Result[2:], 16, 64)
	if err != nil {
		return
	}
	startBlock := latestBlock - 100

	//latestBlock := 19724332 //for testing purposes
	//startBlock := latestBlock - 1

	wg := sync.WaitGroup{}
	for blockNumber := startBlock; blockNumber < latestBlock; blockNumber++ {
		wg.Add(1)
		go func(blockNumber string) {
			//should maybe add sleep here
			params := []string{"0x" + blockNumber}
			txCount, err := GetEntity[BlockTxCountResponse](Request{
				Jsonrpc: "2.0",
				Method:  "eth_getBlockTransactionCountByNumber",
				Params:  params,
				ID:      "getblock.io",
			})

			for err != nil { // retrying on fail / rate limiter
				txCount, err = GetEntity[BlockTxCountResponse](Request{
					Jsonrpc: "2.0",
					Method:  "eth_getBlockTransactionCountByNumber",
					Params:  params,
					ID:      "getblock.io",
				})
			}

			count, _ := strconv.ParseInt(txCount.Result[2:], 16, 64)

			for i := 0; i < int(count); i++ {
				wg.Add(1)
				go func(blockNumber string, index int) {
					params := []string{"0x" + blockNumber, "0x" + strconv.FormatInt(int64(index), 16)}
					tx, err := GetEntity[TransactionResponse](Request{
						Jsonrpc: "2.0",
						Method:  "eth_getTransactionByBlockNumberAndIndex",
						Params:  params,
						ID:      "getblock.io",
					})

					for err != nil { // retrying on fail / rate limiter
						//should maybe add sleep here
						tx, err = GetEntity[TransactionResponse](Request{
							Jsonrpc: "2.0",
							Method:  "eth_getTransactionByBlockNumberAndIndex",
							Params:  params,
							ID:      "getblock.io",
						})
					}

					if len(tx.Result.Input) > 2 {
						signature := tx.Result.Input[:10]

						from := tx.Result.From
						to := "0x" + tx.Result.Input[34:74]

						//maybe switch to a channel for synced storing of data, but what`s the fun in that
						if signature == "0xa9059cbb" {
							mutex.Lock()
							activity[from]++
							activity[to]++
							mutex.Unlock()
						}
					}

					wg.Done()
				}(blockNumber, i)
			}

			wg.Done()
		}(strconv.FormatInt(int64(blockNumber), 16))
	}

	wg.Wait()

	fmt.Println("top 5:")
	printTop5()
}

func printTop5() {
	keys := make([]string, 0, len(activity))

	for key := range activity {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return activity[keys[i]] > activity[keys[j]]
	})

	for idx, key := range keys {
		fmt.Printf("%s %d\n", key, activity[key])
		if idx == 5 {
			break
		}
	}
}
