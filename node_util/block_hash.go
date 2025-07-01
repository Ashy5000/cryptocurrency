// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package node_util

import (
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/sha3"
)

type OldTransition struct {
	UpdatedData map[string][]byte
}

type OldContract struct {
	Contents []byte
	Parties  []ContractParty
	GasUsed  float64
}

type OldTransaction struct {
	Sender            PublicKey
	Recipient         PublicKey
	Amount            float64
	SenderSignature   Signature
	Timestamp         time.Time
	Contracts         []OldContract
	FromSmartContract bool
	Body              []byte
	BodySignatures    []Signature
}

type OldBlock struct {
	Transactions                    []OldTransaction `json:"transactions"`
	Miner                           PublicKey        `json:"miner"`
	Nonce                           int64            `json:"nonce"`
	MiningTime                      time.Duration    `json:"miningTime"`
	Difficulty                      uint64           `json:"difficulty"`
	PreviousBlockHash               [64]byte         `json:"previousBlockHash"`
	Timestamp                       time.Time        `json:"timestamp"`
	PreMiningTimeVerifierSignatures []Signature      `json:"preMiningTimeVerifierSignatures"`
	PreMiningTimeVerifiers          []PublicKey      `json:"preMiningTimeVerifiers"`
	TimeVerifierSignatures          []Signature      `json:"timeVerifierSignature"`
	TimeVerifiers                   []PublicKey      `json:"timeVerifiers"`
	Transition                      OldTransition    `json:"transition"`
}

type PreZenTransition struct {
	UpdatedData  map[string][]byte
	NewContracts map[uint64]Contract
}

type PreZenTransaction struct {
	Sender            PublicKey
	Recipient         PublicKey
	Amount            float64
	SenderSignature   Signature
	Timestamp         time.Time
	Contracts         []Contract
	FromSmartContract bool
	Body              []byte
	BodySignatures    []Signature
}

type PreZenBlock struct {
	Transactions                    []PreZenTransaction `json:"transactions"`
	Miner                           PublicKey           `json:"miner"`
	Nonce                           int64               `json:"nonce"`
	MiningTime                      time.Duration       `json:"miningTime"`
	Difficulty                      uint64              `json:"difficulty"`
	PreviousBlockHash               [64]byte            `json:"previousBlockHash"`
	Timestamp                       time.Time           `json:"timestamp"`
	PreMiningTimeVerifierSignatures []Signature         `json:"preMiningTimeVerifierSignatures"`
	PreMiningTimeVerifiers          []PublicKey         `json:"preMiningTimeVerifiers"`
	TimeVerifierSignatures          []Signature         `json:"timeVerifierSignature"`
	TimeVerifiers                   []PublicKey         `json:"timeVerifiers"`
	Transition                      PreZenTransition    `json:"transition"`
}

func HashBlock(block Block, blockHeight int) [64]byte {
	// Automatically downgrades to older block formats if necessary
	if Env.Upgrades.Washington < blockHeight {
		if Env.Upgrades.Zen < blockHeight && Env.Upgrades.Zen != -1 {
			var blockCpy Block
			marshaled, err := json.Marshal(block)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(marshaled, &blockCpy)
			if err != nil {
				panic(err)
			}
			blockCpy.MiningTime = time.Minute
			blockCpy.TimeVerifiers = []PublicKey{}
			blockCpy.TimeVerifierSignatures = []Signature{}
			blockCpy.Timestamp = time.Time{}
			blockCpy.ZenTransactions = []MerkleNode{}
			for i := range block.LegacyTransactions {
				blockCpy.LegacyTransactions[i].Timestamp = time.Time{}
				blockCpy.LegacyTransactions[i].Body = []byte{}
			}
			for _, node := range block.ZenTransactions {
				if node.Data == nil || len(node.Data) == 0 {
					continue
				}
				var tx Transaction
				err = json.Unmarshal(node.Data, &tx)
				if err != nil {
					panic(err)
				}
				tx.Timestamp = time.Time{}
				tx.Body = []byte{}
				serialized, err := json.Marshal(tx)
				if err != nil {
					panic(err)
				}
				blockCpy.ZenTransactions = InsertValue(blockCpy.ZenTransactions, node.Key, serialized)
			}
			// Just include merkle roots when hashing
			if len(blockCpy.Transition.ZenUpdatedData) != 0 {
				blockCpy.Transition.ZenUpdatedData = []MerkleNode{
					blockCpy.Transition.ZenUpdatedData[0],
				}
			} else {
				blockCpy.Transition.ZenUpdatedData = []MerkleNode{}
			}
			if len(blockCpy.Transition.ZenNewContracts) != 0 {
				blockCpy.Transition.ZenNewContracts = []MerkleNode{
					blockCpy.Transition.ZenNewContracts[0],
				}
			} else {
				blockCpy.Transition.ZenNewContracts = []MerkleNode{}
			}
			if len(blockCpy.ZenTransactions) != 0 {
				blockCpy.ZenTransactions = []MerkleNode{
					blockCpy.ZenTransactions[0],
				}
			}
			// Remove ZK proof (already proven via ZK logic)
			blockCpy.ZenProof = nil
			blockBytes := []byte(fmt.Sprintf("%v", blockCpy))
			sum := sha3.Sum512(blockBytes)
			return sum
		}
		var preZenBlock PreZenBlock
		preZenTransactions := make([]PreZenTransaction, 0)
		for _, transaction := range block.LegacyTransactions {
			preZenTransaction := PreZenTransaction{}
			preZenTransaction.Sender = transaction.Sender
			preZenTransaction.Recipient = transaction.Recipient
			preZenTransaction.Amount = transaction.Amount
			preZenTransaction.SenderSignature = transaction.SenderSignature
			preZenTransaction.Timestamp = time.Time{}
			preZenTransaction.FromSmartContract = transaction.FromSmartContract
			preZenTransaction.Body = []byte{}
			preZenTransaction.BodySignatures = transaction.BodySignatures
			preZenTransaction.Contracts = transaction.Contracts
			preZenTransactions = append(preZenTransactions, preZenTransaction)
		}
		preZenBlock.Transactions = preZenTransactions
		preZenBlock.Miner = block.Miner
		preZenBlock.Nonce = block.Nonce
		preZenBlock.MiningTime = time.Minute
		preZenBlock.Difficulty = block.Difficulty
		preZenBlock.PreviousBlockHash = block.PreviousBlockHash
		preZenBlock.PreMiningTimeVerifierSignatures = block.PreMiningTimeVerifierSignatures
		preZenBlock.PreMiningTimeVerifiers = block.PreMiningTimeVerifiers
		preZenBlock.TimeVerifierSignatures = block.TimeVerifierSignatures
		preZenBlock.TimeVerifiers = block.TimeVerifiers
		preZenTransition := PreZenTransition{}
		preZenTransition.UpdatedData = block.Transition.LegacyUpdatedData
		preZenTransition.NewContracts = block.Transition.LegacyNewContracts
		preZenBlock.Transition = preZenTransition
		blockBytes := []byte(fmt.Sprintf("%v", preZenBlock))
		sum := sha3.Sum512(blockBytes)
		return sum
	}
	oldBlock := OldBlock{}
	oldTransactions := make([]OldTransaction, 0)
	for _, transaction := range block.LegacyTransactions {
		oldTransaction := OldTransaction{}
		oldTransaction.Sender = transaction.Sender
		oldTransaction.Recipient = transaction.Recipient
		oldTransaction.Amount = transaction.Amount
		oldTransaction.SenderSignature = transaction.SenderSignature
		oldTransaction.Timestamp = time.Time{}
		oldTransaction.FromSmartContract = transaction.FromSmartContract
		oldTransaction.Body = []byte{}
		oldTransaction.BodySignatures = transaction.BodySignatures
		oldContracts := make([]OldContract, 0)
		for _, contract := range transaction.Contracts {
			oldContract := OldContract{}
			oldContract.Contents = contract.Contents
			oldContract.Parties = contract.Parties
			oldContract.GasUsed = contract.GasUsed
			oldContracts = append(oldContracts, oldContract)
		}
		oldTransaction.Contracts = oldContracts
		oldTransactions = append(oldTransactions, oldTransaction)
	}
	oldBlock.Transactions = oldTransactions
	oldBlock.Miner = block.Miner
	oldBlock.Nonce = block.Nonce
	oldBlock.MiningTime = time.Minute
	oldBlock.Difficulty = block.Difficulty
	oldBlock.PreviousBlockHash = block.PreviousBlockHash
	oldBlock.PreMiningTimeVerifierSignatures = block.PreMiningTimeVerifierSignatures
	oldBlock.PreMiningTimeVerifiers = block.PreMiningTimeVerifiers
	oldBlock.TimeVerifierSignatures = []Signature{}
	oldBlock.TimeVerifiers = []PublicKey{}
	oldBlock.Timestamp = time.Time{}
	oldTransition := OldTransition{}
	oldTransition.UpdatedData = block.Transition.LegacyUpdatedData
	oldBlock.Transition = oldTransition
	blockBytes := []byte(fmt.Sprintf("%v", oldBlock))
	sum := sha3.Sum512(blockBytes)
	return sum
}
