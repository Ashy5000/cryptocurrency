// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"crypto/dsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"
)

// transactionHashes is a map of transaction hashes to their current status. 0 means the transaction is unmined, 1 means the transaction is being mined, and 2 means the transaction has been mined.
var transactionHashes = make(map[[32]byte]int)
var miningTransactions []Transaction

func RequestTimeVerification(block Block) ([]Signature, []dsa.PublicKey) {
  fmt.Println("Requesting time verification")
  var signatures []Signature
  var publicKeys []dsa.PublicKey
	// Convert the block to a string (JSON)
	bodyChars, err := json.Marshal(&block)
	if err != nil {
		panic(err)
	}
	for _, peer := range GetPeers() {
    if int64(len(block.TimeVerifiers)) >= GetMinerCount(len(blockchain))/5 {
			break
		}
		// Verify that the peer has mined a block (only miners can be time verifiers)
		req, err := http.NewRequest(http.MethodGet, peer+"/identify", nil)
		if err != nil {
			panic(err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Peer down.")
			continue
		}
		// Get the response body
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		// Convert the response body to a string
		bodyString := string(bodyBytes)
		// Convert the response body to a big.Int
		peerY, ok := new(big.Int).SetString(bodyString, 10)
		if !ok {
			fmt.Println("Could not convert peer Y to big.Int")
			continue
		}
		// Create a dsa.PublicKey from the big.Int
		peerKey := dsa.PublicKey{
			Y: peerY,
		}
		// Verify that the peer has mined a block
		if IsNewMiner(peerKey, len(blockchain) + 1) {
			fmt.Println("Peer has not mined a block.")
			continue
		}
		// Ask to verify the time
		body := strings.NewReader(string(bodyChars))
		req, err = http.NewRequest(http.MethodGet, peer+"/verifyTime", body)
		if err != nil {
			panic(err)
		}
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Peer down.")
			continue
		}
		// Get the response body
		bodyBytes, err = io.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		if string(bodyBytes) == "invalid" {
			fmt.Println("Time verifier believes block is invalid.")
			continue
		}
		// Split the response body into the signature and the public key
		split := strings.Split(string(bodyBytes), "%")
		// Unmarshal the signature
		var signature Signature
		err = json.Unmarshal([]byte(split[0]), &signature)
		if err != nil {
			panic(err)
		}
		// Unmarshal the public key
		var publicKey dsa.PublicKey
		err = json.Unmarshal([]byte(split[1]), &publicKey)
		if err != nil {
			panic(err)
		}
		// Add the time verifier to the block
		publicKeys = append(publicKeys, publicKey)
		// Add the time verifier signature to the block
		signatures = append(signatures, signature)
    fmt.Println("Got verification.")
	}
  return signatures, publicKeys
}

func CreateBlock() (Block, error) {
	if len(miningTransactions) == 0 {
		return Block{}, errors.New("pool dry")
	}
	start := time.Now()
	previousBlock, previousBlockFound := GetLastMinedBlock()
	if !previousBlockFound {
		previousBlock.Difficulty = initialBlockDifficulty
		previousBlock.MiningTime = time.Minute
	}
	block := Block{
		Miner:                  GetKey().PublicKey,
		Transactions:           miningTransactions,
		Nonce:                  0,
		Difficulty:             GetDifficulty(previousBlock.MiningTime, previousBlock.Difficulty),
		Timestamp:              time.Now(),
		TimeVerifierSignatures: []Signature{},
		TimeVerifiers:          []dsa.PublicKey{},
    MiningTime:             0,
	}
	if len(blockchain) > 0 {
		block.PreviousBlockHash = HashBlock(blockchain[len(blockchain)-1])
	} else {
		block.PreviousBlockHash = [32]byte{}
	}
	hashBytes := HashBlock(block)
	hash := binary.BigEndian.Uint64(hashBytes[:]) // Take the last 64 bits-- we won't ever need more than 64 zeroes.
  // Request time verifiers
  block.PreMiningTimeVerifierSignatures, block.PreMiningTimeVerifiers = RequestTimeVerification(block)
	fmt.Printf("Mining block with difficulty %d\n", block.Difficulty)
	for hash > maximumUint64/block.Difficulty {
		for i, transaction := range miningTransactions {
			transactionString := fmt.Sprintf("%s:%s:%f:%d", EncodePublicKey(transaction.Sender), EncodePublicKey(transaction.Recipient), transaction.Amount, transaction.Timestamp.UnixNano())
			transactionBytes := []byte(transactionString)
			hash := sha256.Sum256(transactionBytes)
			if transactionHashes[hash] > 1 {
				miningTransactions[i] = miningTransactions[len(miningTransactions)-1]
				miningTransactions = miningTransactions[:len(miningTransactions)-1]
				i--
			}
		}
		if len(miningTransactions) > 0 {
			previousBlock, previousBlockFound = GetLastMinedBlock()
			if !previousBlockFound {
				previousBlock.Difficulty = initialBlockDifficulty
				previousBlock.MiningTime = time.Minute
			}
			if len(blockchain) > 0 {
				block.PreviousBlockHash = HashBlock(blockchain[len(blockchain)-1])
			} else {
				block.PreviousBlockHash = [32]byte{}
			}
			block.Difficulty = GetDifficulty(previousBlock.MiningTime, previousBlock.Difficulty)
      block.Transactions = miningTransactions
			block.Nonce++
			hashBytes = HashBlock(block)
			hash = binary.BigEndian.Uint64(hashBytes[:])
		} else {
			fmt.Println("Pool dry.")
			return Block{}, errors.New("pool dry")
		}
	}
	block.MiningTime = time.Since(start)
	// Ask for time verifiers
  block.TimeVerifierSignatures, block.TimeVerifiers = RequestTimeVerification(block)
	if int64(len(block.TimeVerifiers)) < GetMinerCount(len(blockchain))/5 {
		fmt.Println("Not enough time verifiers.")
		return Block{}, errors.New("lost block")
	}
	miningTransactions = []Transaction{}
	return block, nil
}
