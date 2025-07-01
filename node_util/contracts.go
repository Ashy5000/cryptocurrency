// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package node_util

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

type ContractParty struct {
	Signature Signature
	PublicKey PublicKey
}

type Contract struct {
	Contents []byte
	Parties  []ContractParty
	GasUsed  float64
	Location uint64
	Loaded   bool
}

type TransitionKVPair struct {
	Key   string
	Value []byte
}

var ExternalStateWriteableValue = []byte("ExternalStateWriteableValue")

func (c Contract) IsNewContract() bool {
	return c.Location == 0
}

func (c Contract) LoadContract() {
	state := CalculateCurrentState()
	// Try with merkle
	contract, ok := GetValue(state.ZenContracts, strconv.FormatUint(c.Location, 10))
	if ok {
		c.Contents = contract
		c.Parties = nil // Zen drops parties from specification (can be replaced by new VM features)
		// By only storing contracts in the merkle tree, the root hash will match between the consensus client and the VM
		// This way, the merkle tree doesn't have to be rebuilt for each transaction
		c.GasUsed = 0
		c.Loaded = true
		return
	}
	// Fallback to legacy
	for _, contract := range state.LegacyContracts {
		if contract.Location == c.Location {
			c.Contents = contract.Contents
			c.Parties = contract.Parties
			c.GasUsed = 0
			c.Loaded = true
			break
		}
	}
}

var executionLocked = false

func (c Contract) Execute(maxGas float64, sender PublicKey) ([]Transaction, StateTransition, float64, error) {
	for executionLocked {
	}
	executionLocked = true
	if !c.Loaded {
		c.LoadContract()
	}
	if !VerifySmartContract(c) {
		Warn("Invalid contract detected.")
		executionLocked = false
		return make([]Transaction, 0), StateTransition{}, 0, nil
	}
	if err := os.WriteFile("contract.blockasm", []byte(c.Contents), 0666); err != nil {
		executionLocked = false
		return nil, StateTransition{}, 0, err
	}
	contractStr := c.Contents
	hash := sha256.Sum256([]byte(contractStr))
	pendingState := GetPendingState()
	pendingStateData := pendingState.LegacyData
	for _, node := range pendingState.ZenData {
		if node.Data != nil {
			pendingStateData[node.Key] = node.Data
		}
	}
	if pendingStateData == nil {
		pendingStateData = make(map[string][]byte)
	}
	pendingStateSerialized, err := msgpack.Marshal(pendingStateData)
	err = os.WriteFile("pending_state.msgpack", pendingStateSerialized, 0644)
	if err != nil {
		executionLocked = false
		return nil, StateTransition{}, 0, err
	}
	fmt.Println("./contracts/target/release/contracts", "contract.blockasm", hex.EncodeToString(hash[:]), fmt.Sprintf("%f", maxGas), hex.EncodeToString(sender.Y))
	out, err := exec.Command("./contracts/target/release/contracts", "contract.blockasm", hex.EncodeToString(hash[:]), fmt.Sprintf("%f", maxGas), hex.EncodeToString(sender.Y)).Output()
	if err != nil {
		fmt.Println("Errored with output:", string(out))
		fmt.Println("Error: ", err)
		fmt.Println("Contract hash:", hex.EncodeToString(hash[:]))
		executionLocked = false
		return nil, StateTransition{}, 0, err
	}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	transactions := make([]Transaction, 0)
	gasUsed := 0.0
	transition := StateTransition{
		LegacyUpdatedData:  make(map[string][]byte),
		LegacyNewContracts: make(map[uint64]Contract),
		ZenUpdatedData:     make([]MerkleNode, 0),
		ZenNewContracts:    make([]MerkleNode, 0),
	}
	stateChangeLines := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		if line[:2] != "TX" {
			if len(line) < 10 {
				continue
			}
			if line[:9] != "Gas used:" {
				if line[:14] == "State change: " {
					stateChangeLines = append(stateChangeLines, line)
				}
				continue
			}
			gasUsed, err = strconv.ParseFloat(line[10:], 64)
			if err != nil {
				executionLocked = false
				return nil, StateTransition{}, 0, err
			}
			continue
		}
		words := strings.Split(line, " ")
		var senderY []byte
		err = json.Unmarshal([]byte(words[1]), &senderY)
		if err != nil {
			executionLocked = false
			return nil, StateTransition{}, 0, err
		}
		senderIsParty := false
		for _, party := range c.Parties {
			if bytes.Equal(party.PublicKey.Y, senderY) {
				senderIsParty = true
				break
			}
		}
		if !senderIsParty {
			Warn("Invalid sender detected.")
			executionLocked = false
			return nil, StateTransition{}, 0, nil
		}
		sender := PublicKey{
			Y: senderY,
		}
		var receiverY []byte
		err = json.Unmarshal([]byte(words[2]), &receiverY)
		if err != nil {
			executionLocked = false
			return nil, StateTransition{}, 0, err
		}
		receiver := PublicKey{
			Y: receiverY,
		}
		subdividedAmount, err := strconv.Atoi(words[3])
		if err != nil {
			executionLocked = false
			return nil, StateTransition{}, 0, err
		}
		amount := float64(subdividedAmount * 1000000)
		transaction := Transaction{
			Sender:            sender,
			Recipient:         receiver,
			Amount:            amount,
			FromSmartContract: true,
		}
		transactions = append(transactions, transaction)
	}
	zenTransitionSlice := make([]TransitionKVPair, 0)
	for _, line := range stateChangeLines {
		stateChangeString := line[14:]
		parts := strings.Split(stateChangeString, "|")
		address := parts[0]
		valueHex := parts[1]
		if valueHex == "" {
			transition.ZenUpdatedData = InsertValue(transition.ZenUpdatedData, address, []byte{})
		}
		valueBytes, err := hex.DecodeString(valueHex)
		if err != nil {
			Warn("Error decoding state change:")
		}
		fmt.Println("Applying state change:", address, valueBytes)
		if Env.Upgrades.Zen <= len(Blockchain) {
			// Zen insert
			zenTransitionSlice = append(zenTransitionSlice, TransitionKVPair{
				Key:   address,
				Value: valueBytes,
			})
		} else {
			// Legacy insert
			transition.LegacyUpdatedData[address] = valueBytes
		}
	}
	sort.Slice(zenTransitionSlice, func(i, j int) bool {
		return zenTransitionSlice[i].Key < zenTransitionSlice[j].Key
	})
	for _, pair := range zenTransitionSlice {
		transition.ZenUpdatedData = InsertValue(transition.ZenUpdatedData, pair.Key, pair.Value)
	}
	if c.IsNewContract() {
		if Env.Upgrades.Zen <= len(Blockchain) && Env.Upgrades.Zen != -1 {
			// Zen update
			transition.ZenNewContracts = InsertValue(transition.ZenNewContracts, hex.EncodeToString(hash[:]), []byte(c.Contents))
		} else {
			// Legacy update
			transition.LegacyNewContracts = map[uint64]Contract{
				c.Location: c,
			}
		}
	}
	executionLocked = false
	return transactions, transition, gasUsed, nil
}
