// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package node_util

import (
	"time"
)

var Blockchain []Block

func GenesisBlock() Block {
	return Block{
		Transactions:           nil,
		Miner:                  PublicKey{},
		Nonce:                  0,
		MiningTime:             0,
		Difficulty:             0,
		PreviousBlockHash:      [64]byte{},
		Timestamp:              time.Time{},
		TimeVerifierSignatures: []Signature{},
		TimeVerifiers:          []PublicKey{},
	}
}

func Append(block Block) {
	Blockchain = append(Blockchain, block)
}
