// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	. "cryptocurrency/node_util"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHashBlock(t *testing.T) {
	t.Run("It returns a checksum of the block", func(t *testing.T) {
		// Arrange
		key := PublicKey{
			Y: []byte("123"),
		}
		block := Block{
			ZenTransactions: []MerkleNode{
				{
					Data:     []byte{},
					Children: map[byte]int{},
					Parent:   -1,
					Hash:     "",
					Key:      "",
				},
			},
			Miner:                           key,
			Nonce:                           0,
			MiningTime:                      0,
			Difficulty:                      0,
			PreviousBlockHash:               [64]byte{},
			Timestamp:                       time.Time{},
			PreMiningTimeVerifierSignatures: []Signature{},
			PreMiningTimeVerifiers:          []PublicKey{},
			TimeVerifierSignatures:          []Signature{},
			TimeVerifiers:                   []PublicKey{},
			Transition: StateTransition{
				LegacyUpdatedData: map[string][]byte{},
				LegacyNewContracts: map[uint64]Contract{
					0: {
						Contents: "",
						Parties:  nil,
						GasUsed:  0,
						Location: 0,
						Loaded:   false,
					},
				},
			},
		}
		var emptyHash [32]byte
		// Act
		hash := HashBlock(block, 10)
		// Assert
		assert.NotEqual(t, emptyHash, hash)
	})
}
