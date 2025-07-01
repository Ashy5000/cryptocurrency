// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package node_util

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var Wg sync.WaitGroup

func GetKey(path string) PrivateKey {
	if path == "" {
		path = "key.json"
	}
	keyJson, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var key PrivateKey
	err = json.Unmarshal(keyJson, &key)
	if err != nil {
		panic(err)
	}
	return key
}

// SyncBlockchain synchronizes the blockchain with other peers.
//
// It iterates over each peer, sends a GET request to the peer's blockchain endpoint,
// and receives the response. It then parses the response body into a slice of Block
// structs. It checks the validity of the received blockchain by verifying the previous
// block hash and proof of work. If the blockchain is valid, it updates the longest blockchain
// if it is longer than the current blockchain.
//
// The function takes an integer parameter `finalityBlockHeight` which represents the
// minimum block height required for a blockchain to be considered final.
//
// If the blockchain is successfully synced, it logs a success message. If there is an
// error with any peer, it logs an error message.
//
// Parameters:
//   - finalityBlockHeight: an integer representing the minimum block height required for
//     a blockchain to be considered final.
//
// Return type: None.
func SyncBlockchain(finalityBlockHeight int) {
	longestLength := 0
	var longestBlockchain []Block
	errCount := 0
	for _, peer := range GetPeers() {
		res, err := http.Get(fmt.Sprintf("%s/blockchain", peer))
		if err != nil {
			errCount++
			continue
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		var peerBlockchain []Block
		err = json.Unmarshal(body, &peerBlockchain)
		if err != nil {
			panic(err)
		}
		length := len(peerBlockchain)
		// Check to ensure proof of work is valid
		createsFork := false
		for i, block := range peerBlockchain {
			if i == 0 {
				continue
			}
			previousBlockHash := HashBlock(peerBlockchain[i-1], i-1)
			if !bytes.Equal(block.PreviousBlockHash[:], previousBlockHash[:]) {
				Log("Invalid blockchain received from peer: incorrect previous block hash", true)
				fmt.Println(block.PreviousBlockHash)
				fmt.Println(previousBlockHash)
				goto INVALID
			}
			blockHash := HashBlock(block, i)
			if binary.BigEndian.Uint64(blockHash[:]) > MaximumUint64/block.Difficulty {
				Log("Invalid blockchain received from peer: proof of work", false)
				fmt.Println("Failed at block", i)
				goto INVALID
			}
			if i < len(Blockchain)-1 {
				if blockHash != HashBlock(Blockchain[i], i) {
					createsFork = true
				}
			}
			// Get the correct difficulty for the block
			var lastMinedBlock Block
			lastMinedBlock.Difficulty = MaximumUint64
			j := i - 1
			for j >= 0 {
				prevBlock := peerBlockchain[j]
				if bytes.Equal(prevBlock.Miner.Y, block.Miner.Y) {
					lastMinedBlock = prevBlock
					break
				}
				j--
			}
			var lastTime time.Duration
			var lastDifficulty uint64
			if i == 1 || lastMinedBlock.Difficulty == MaximumUint64 {
				lastTime = time.Minute
				lastDifficulty = MinimumBlockDifficulty
			} else {
				lastTime = lastMinedBlock.MiningTime
				lastDifficulty = lastMinedBlock.Difficulty
			}
			correctDifficulty := GetDifficulty(lastTime, lastDifficulty, len(ExtractTransactions(block)), i)
			if block.Difficulty != correctDifficulty {
				fmt.Println(correctDifficulty)
				fmt.Println(lastTime)
				fmt.Println(lastDifficulty)
				Log("Invalid blockchain received from peer: incorrect difficulty", false)
				goto INVALID
			}
		}
		if createsFork {
			// Require finality
			if length < finalityBlockHeight {
				Log("Ignoring blockchain received from peer due to lack of finality.", false)
				goto INVALID
			}
		}
		if length > longestLength {
			longestLength = length
			longestBlockchain = peerBlockchain
		}
		break
	INVALID:
		errCount++
	}
	if errCount >= len(GetPeers()) {
		Log("Failed to sync blockchain with any peers.", true)
		return
	}
	Log("Blockchain successfully synced!", false)
	Log(fmt.Sprintf("%d out of %d peers responded.", len(GetPeers())-errCount, len(GetPeers())), false)
	if longestLength > len(Blockchain) {
		Blockchain = longestBlockchain
	}
}

func GetBalance(key []byte) float64 {
	total := 0.0
	miningTotal := 0.0
	isGenesis := true
	blocksMined := 0
	for i, block := range Blockchain {
		if isGenesis {
			isGenesis = false
			continue
		}
		for _, transaction := range ExtractTransactions(block) {
			if bytes.Equal(transaction.Sender.Y, key) {
				total -= transaction.Amount
				if i > 50 { // Fees start after 50 blocks
					fee := TransactionFee + (BodyFeePerByte * float64(len(transaction.Body)))
					for _, contract := range transaction.Contracts {
						fee += GasPrice * contract.GasUsed
					}
					total -= fee
				}
			} else if bytes.Equal(transaction.Recipient.Y, key) {
				total += transaction.Amount
			}
		}
		if bytes.Equal(block.Miner.Y, key) {
			lastBlock := Blockchain[i-1]
			miningTotal += float64(len(block.TimeVerifiers)-len(lastBlock.TimeVerifiers)) * 0.1
			if i > 50 { // Fees start after 50 blocks
				fees := 0.0
				for _, transaction := range ExtractTransactions(block) {
					fees += TransactionFee
					fees += BodyFeePerByte * float64(len(transaction.Body))
					for _, contract := range transaction.Contracts {
						fees += GasPrice * contract.GasUsed
					}
				}
				miningTotal += fees
			}
			// Get number of miners at the time of mining
			minerCount := GetMinerCount(i)
			reward := CalculateBlockReward(minerCount, i)
			miningTotal += reward
			blocksMined++
		}
	}
	if blocksMined > BlocksBeforeReward && len(Blockchain) > 50 {
		total += miningTotal - float64(BlocksBeforeReward)
	} else if len(Blockchain) < 50 {
		total += miningTotal
	}
	return total
}

func SendRequest(req *http.Request) {
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		Wg.Done()
		return
	}
	Wg.Done()
}

func Send(receiver string, amount string, transactionBody []byte) {
	key := GetKey("")
	sender := key.PublicKey.Y
	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%d", sender, receiver, amount, timestamp)))
	sigBytes, err := key.X.Sign(hash[:])
	sig := Signature{
		S: sigBytes,
	}
	if err != nil {
		panic(err)
	}
	sigStr, err := json.Marshal(sig)
	if err != nil {
		panic(err)
	}
	senderStr := EncodePublicKey(key.PublicKey)
	receiverStr := EncodePublicKey(PublicKey{Y: []byte(receiver)})
	for _, peer := range GetPeers() {
		Log("Sending transaction to peer: "+peer, false)
		contractsStr, err := json.Marshal(make([]Contract, 0))
		if err != nil {
			panic(err)
		}
		body := strings.NewReader(EncodeMineRequest(senderStr, receiverStr, fmt.Sprintf("%s", amount), string(sigStr), timestamp, string(contractsStr), transactionBody, []Signature{}))
		req, err := http.NewRequest(http.MethodGet, peer+"/mine", body)
		if err != nil {
			panic(err)
		}
		Wg.Add(1)
		go SendRequest(req)
	}
}

func DeploySmartContract(contractPath string, contractLocation string) ([32]byte, error) {
	if contractPath == "" && contractLocation == "" {
		return [32]byte{}, errors.New("must provide contract path or location")
	}
	var contract Contract
	if contractPath != "" {
		contents, err := os.ReadFile(contractPath)
		if err != nil {
			return [32]byte{}, err
		}
		contract = Contract{
			Contents: contents,
			Parties:  make([]ContractParty, 0),
			GasUsed:  0,
			Location: 0,
			Loaded:   true,
		}
	} else {
		contractLocationUint, err := strconv.ParseUint(contractLocation, 10, 64)
		if err != nil {
			return [32]byte{}, err
		}
		contract = Contract{
			Location: contractLocationUint,
		}
	}
	key := GetKey("")
	deployer := GetKey("").PublicKey
	deployerStr := EncodePublicKey(deployer)
	party := ContractParty{
		PublicKey: PublicKey{
			Y: deployer.Y,
		},
	}
	contractHash := sha256.Sum256(contract.Contents)
	fmt.Println(contract.Contents)
	fmt.Println(contractHash)
	partySig, err := key.X.Sign(contractHash[:])
	if err != nil {
		panic(err)
	}
	party.Signature = Signature{
		S: partySig,
	}
	contract.Parties = append(contract.Parties, party)
	contracts := append(make([]Contract, 0), contract)
	contractsStr, err := json.Marshal(contracts)
	if err != nil {
		return [32]byte{}, err
	}
	amount := "0"
	timestamp := time.Now().UnixNano()
	transactionString := fmt.Sprintf("%s:%s:%s:%d", deployer.Y, deployer.Y, amount, timestamp)
	hash := sha256.Sum256([]byte(transactionString))
	sigBytes, err := key.X.Sign(hash[:])
	sig := Signature{
		S: sigBytes,
	}
	if err != nil {
		panic(err)
	}
	sigStr, err := json.Marshal(sig)
	if err != nil {
		panic(err)
	}
	encoded := EncodeMineRequest(deployerStr, deployerStr, amount, string(sigStr), timestamp, string(contractsStr), []byte{}, []Signature{})
	body := strings.NewReader(encoded)
	for _, peer := range GetPeers() {
		Log("Sending smart contract to peer: "+peer, false)
		req, err := http.NewRequest(http.MethodGet, peer+"/mine", body)
		if err != nil {
			return [32]byte{}, err
		}
		Wg.Add(1)
		go SendRequest(req)
	}
	return contractHash, nil
}

func GetLastMinedBlock(key []byte) (Block, bool) {
	for i := len(Blockchain) - 1; i > 0; i-- {
		block := Blockchain[i]
		if bytes.Equal(block.Miner.Y, key) {
			return block, true
		}
	}
	return Block{}, false
}

func IsNewMiner(miner PublicKey, maxBlockPosition int) bool {
	isGenesis := true
	for i, block := range Blockchain {
		if isGenesis {
			isGenesis = false
			continue
		}
		if i > maxBlockPosition {
			break
		}
		if bytes.Equal(block.Miner.Y, miner.Y) {
			return false
		}
	}
	return true
}

func GetMinerCount(maxBlockPosition int) int64 {
	var result int64
	result = 0
	isGenesis := true
	for i, block := range Blockchain {
		if i > maxBlockPosition {
			break
		}
		if isGenesis {
			isGenesis = false
			continue
		}
		if IsNewMiner(block.Miner, i-1) {
			result++
		}
	}
	return result
}

func GetMaxMiners() int64 {
	x := float64(len(Blockchain))
	res := int64(math.Ceil(x / 20.0))
	if res > 0 {
		return res
	}
	return 1
}
