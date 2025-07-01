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
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/open-quantum-safe/liboqs-go/oqs"
)

func VerifyTransaction(senderKey PublicKey, recipientKey PublicKey, amount string, timestamp time.Time, sig []byte) bool {
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		panic(err)
	}
	transactionString := fmt.Sprintf("%s:%s:%s:%d", senderKey.Y, recipientKey.Y, amount, timestamp.UnixNano())
	hash := sha256.Sum256([]byte(transactionString))
	verifier := oqs.Signature{}
	sigName := "Dilithium3"
	if err := verifier.Init(sigName, nil); err != nil {
		Error("Failed to initialize Dilithium2 verifier", true)
	}
	isValid, err := verifier.Verify(hash[:], sig, senderKey.Y)
	if err != nil {
		panic(err)
	}
	if !isValid {
		Log("Invalid transaction signature detected", true)
		return false
	}
	// Calculate amount spent so far in this block
	var amountSpentInCurrentBlock float64
	for _, transaction := range MiningTransactions {
		if bytes.Equal(transaction.Sender.Y, senderKey.Y) {
			if amountSpentInCurrentBlock+transaction.Amount > amountSpentInCurrentBlock {
				Log("Overflow detected.", true)
				return false
			}
			amountSpentInCurrentBlock += transaction.Amount
			if len(Blockchain) > 50 {
				amountSpentInCurrentBlock += TransactionFee                                  // Base fee
				amountSpentInCurrentBlock += BodyFeePerByte * float64(len(transaction.Body)) // Body fee
				for _, contract := range transaction.Contracts {
					amountSpentInCurrentBlock += GasPrice * contract.GasUsed // Gas fee
				}
			}
		}
	}
	amountSpentInCurrentBlock -= amountFloat
	if GetBalance(senderKey.Y) < amountSpentInCurrentBlock {
		Log("Double spending detected.", true)
		return false
	}
	return true
}

func VerifyMiner(miner PublicKey) bool {
	if Env.Upgrades.Jinan <= len(Blockchain) {
		return true
	}
	if IsNewMiner(miner, len(Blockchain)) && GetMinerCount(len(Blockchain)) >= GetMaxMiners() {
		Log(fmt.Sprintf("Miner count: %d", GetMinerCount(len(Blockchain))), true)
		Log(fmt.Sprintf("Maximum miner count: %d", GetMaxMiners()), true)
		return false
	}
	return true
}

func VerifyTransactions(transactions []Transaction) bool {
	for _, transaction := range transactions {
		if transaction.FromSmartContract {
			return true
		}
		if !VerifyTransaction(transaction.Sender, transaction.Recipient, strconv.FormatFloat(transaction.Amount, 'f', -1, 64), transaction.Timestamp, transaction.SenderSignature.S) {
			Log("Block has invalid transaction/transaction signature. Ignoring block request.", true)
			return false
		}
	}
	return true
}

func DetectFork(block Block) bool {
	for i, b := range Blockchain {
		if i == len(Blockchain)-1 {
			break
		}
		if b.PreviousBlockHash == block.PreviousBlockHash {
			Log("Block creates a fork. Possible reorg necessary.", true)
			SyncBlockchain(len(Blockchain) + BlocksUntilFinality) // Reorg if other chain is more than `BlocksUntilFinality` blocks longer than the current chain
			return true
		}
	}
	return false
}

func DetectDuplicateBlock(hashBytes [64]byte) bool {
	isDuplicate := false
	for i, b := range Blockchain {
		if HashBlock(b, i) == hashBytes {
			isDuplicate = true
		}
	}
	return isDuplicate
}

func VerifySmartContractTransactionsPreZen(block Block) bool {
	// Ensure the transactions created by smart contracts are valid
	// Iterate through smart contracts
	var smartContractCreatedTransactions []Transaction
	var fullTransition = StateTransition{
		LegacyUpdatedData: make(map[string][]byte),
		ZenUpdatedData:    make([]MerkleNode, 0),
	}
	for _, transaction := range ExtractTransactions(block) {
		for _, contract := range transaction.Contracts {
			// Validate the contract
			if !VerifySmartContract(contract) {
				Log("Block has invalid smart contract. Ignoring block request.", true)
				return false
			}
			// Execute the contract
			transactions, transition, gasUsed, err := contract.Execute(GetBalance(transaction.Sender.Y)/GasPrice, transaction.Sender)
			if err != nil {
				continue
			}
			smartContractCreatedTransactions = append(smartContractCreatedTransactions, transactions...)
			// Check gas usage
			if gasUsed != contract.GasUsed {
				Log("Block has invalid smart contract gas usage. Ignoring block request.", true)
				return false
			}
			// Add transition to fullTransition
			for location, value := range transition.LegacyUpdatedData {
				fullTransition.LegacyUpdatedData[location] = value
			}
			fullTransition.ZenUpdatedData = Merge(fullTransition.ZenUpdatedData, transition.ZenUpdatedData)
		}
	}
	if !reflect.DeepEqual(fullTransition.LegacyUpdatedData, block.Transition.LegacyUpdatedData) {
		Log("Block has invalid state transition. Ignoring block request.", true)
		return false
	}
	if len(fullTransition.ZenUpdatedData) != 0 && fullTransition.ZenUpdatedData[0].Hash != block.Transition.ZenUpdatedData[0].Hash {
		Log("Block has invalid merkle root. Ignoring block request.", true)
	}
	// Get the smart contract created transactions in the block
	var smartContractCreatedTransactionsInBlock []Transaction
	for _, transaction := range ExtractTransactions(block) {
		if transaction.FromSmartContract {
			smartContractCreatedTransactionsInBlock = append(smartContractCreatedTransactionsInBlock, transaction)
		}
	}
	// Check if the two lists are the same
	if !reflect.DeepEqual(smartContractCreatedTransactions, smartContractCreatedTransactionsInBlock) {
		Log("Block has invalid smart contract transactions. Ignoring block request.", true)
		return false
	}
	return true
}

func VerifySmartContractTransactions(block Block) bool {
	if len(Blockchain) < Env.Upgrades.Zen && Env.Upgrades.Zen != -1 {
		return VerifySmartContractTransactionsPreZen(block)
	}
	for _, tx := range ExtractTransactions(block) {
		if tx.FromSmartContract {
			// Zen doesn't support smart contract transactions
			return false
		}
	}
	state := CalculateCurrentState()
	var root string
	if len(state.ZenData) > 1 || len(state.ZenContracts) > 1 {
		root = Merge(state.ZenData, state.ZenContracts)[0].Hash
	} else {
		root = ""
	}
	hasher := sha256.New()
	for _, transaction := range ExtractTransactions(block) {
		for _, contract := range transaction.Contracts {
			hasher.Write([]byte(contract.Contents))
		}
	}
	for _, transaction := range ExtractTransactions(block) {
		for _, contract := range transaction.Contracts {
			contractHash := sha256.Sum256([]byte(contract.Contents))
			hasher.Write([]byte(hex.EncodeToString(contractHash[:])))
		}
	}
	for _, transaction := range ExtractTransactions(block) {
		hasher.Write([]byte(strconv.Itoa(int(GetBalance(transaction.Sender.Y) / GasPrice))))
	}
	for _, transaction := range ExtractTransactions(block) {
		hasher.Write([]byte(hex.EncodeToString(transaction.Sender.Y)))
	}
	hasher.Write([]byte(strconv.Itoa(len(Blockchain))))
	hash := hasher.Sum(nil)
	var transitionRoot string
	if len(block.Transition.ZenUpdatedData) > 1 {
		transitionRoot = block.Transition.ZenUpdatedData[0].Hash
	} else {
		transitionRoot = ""
	}
	if !ZKVerify(block.ZenProof, root, hex.EncodeToString(hash), transitionRoot) {
		Warn("Block has invalid ZK proof. Ignoring block request.")
		return false
	}
	return true
}

func VerifyBlock(block Block, blockHeight int) bool {
	isValid := VerifyTransactions(ExtractTransactions(block))
	hashBytes := HashBlock(block, blockHeight)
	hash := binary.BigEndian.Uint64(hashBytes[:]) // Take the last 64 bits-- we won't ever need more than 64 zeroes.
	isValid = hash <= MaximumUint64/block.Difficulty && isValid
	if DetectDuplicateBlock(hashBytes) {
		return false
	}
	isValid = !DetectFork(block) && isValid
	if len(Blockchain) > 0 && block.PreviousBlockHash != HashBlock(Blockchain[len(Blockchain)-1], len(Blockchain)-1) {
		Log("Block has invalid previous block hash. Possible reorg is necessary.", true)
		SyncBlockchain(len(Blockchain) + BlocksUntilFinality) // Reorg if other chain is more than `BlocksUntilFinality` blocks longer than the current chain
		isValid = false
	}
	isValid = VerifyMiner(block.Miner) && isValid
	// Get the correct difficulty for the block
	lastMinedBlock, found := GetLastMinedBlock(block.Miner.Y)
	if !found {
		lastMinedBlock.Difficulty = InitialBlockDifficulty
		lastMinedBlock.MiningTime = time.Minute
	}
	correctDifficulty := GetDifficulty(lastMinedBlock.MiningTime, lastMinedBlock.Difficulty, len(ExtractTransactions(block)), blockHeight)
	if block.Difficulty != correctDifficulty || block.Difficulty < MinimumBlockDifficulty {
		Log("Invalid difficulty detected.", true)
		Log("The node software is designed to prevent difficulty manipulation, so this invalid difficulty will not cause issues for the network.", false)
		Log(fmt.Sprintf("Expected difficulty: %d", correctDifficulty), true)
		Log(fmt.Sprintf("Actual difficulty: %d", block.Difficulty), true)
		isValid = false
	}
	if block.Timestamp.After(time.Now()) {
		Log("Block has invalid timestamp. Ignoring block request.", true)
		Log("Timestamp is in the future.", true)
		isValid = false
	}
	if !VerifyTimeVerifiers(block, block.TimeVerifiers, block.TimeVerifierSignatures, false) || !VerifyTimeVerifiers(block, block.PreMiningTimeVerifiers, block.PreMiningTimeVerifierSignatures, true) {
		Log("Block has invalid time verifiers. Ignoring block request.", true)
		isValid = false
	}
	if !VerifySmartContractTransactions(block) {
		Log("Block has invalid smart contract transactions. Ignoring block request.", true)
		isValid = false
	}
	return isValid
}

func VerifyTimeVerifiers(block Block, verifiers []PublicKey, signatures []Signature, premining bool) bool {
	if len(verifiers) != len(signatures) {
		Log("Signature count does not match verifier count.", true)
		return false
	}
	oqsVerifier := oqs.Signature{}
	sigName := "Dilithium3"
	if err := oqsVerifier.Init(sigName, nil); err != nil {
		Error("Failed to initialize Dilithium2 verifier", true)
	}
	if premining {
		for i, verifier := range verifiers {
			valid, err := oqsVerifier.Verify([]byte(fmt.Sprintf("%d", block.Timestamp.UnixNano())), signatures[i].S, verifier.Y)
			if err != nil {
				panic(err)
			}
			if !valid {
				Log("Invalid time verifier signature detected", true)
				return false
			}
		}
	} else {
		for i, verifier := range verifiers {
			valid, err := oqsVerifier.Verify([]byte(fmt.Sprintf("%d", block.Timestamp.Add(block.MiningTime).UnixNano())), signatures[i].S, verifier.Y)
			if err != nil {
				panic(err)
			}
			if !valid {
				Log("Invalid time verifier signature detected", true)
				return false
			}
		}
	}
	// Ensure all verifiers are unique
	verifierMap := make(map[string]bool)
	for _, verifier := range verifiers {
		if verifierMap[string(verifier.Y)] {
			Log("Time verifier is not unique.", true)
			return false
		}
		verifierMap[string(verifier.Y)] = true
	}
	// Ensure verifiers are miners
	fromLastBlock := 0
	for _, verifier := range verifiers {
		if IsNewMiner(verifier, len(Blockchain)+1) {
			Log("Time verifier is not a miner.", true)
			return false
		}
		for _, exitingVerifier := range Blockchain[len(Blockchain)-1].TimeVerifiers {
			if bytes.Equal(verifier.Y, exitingVerifier.Y) {
				fromLastBlock++
				continue
			}
		}
		for _, exitingVerifier := range Blockchain[len(Blockchain)-1].PreMiningTimeVerifiers {
			if bytes.Equal(verifier.Y, exitingVerifier.Y) {
				fromLastBlock++
				continue
			}
		}
	}
	// Ensure there are enough verifiers
	if len(verifiers) < GetMinVerifiers() {
		Log("Not enough time verifiers.", true)
		return false
	}
	// Ensure enough time verifiers signed the previous block
	if fromLastBlock < int(float64(len(verifiers))*0.75) {
		Log("Not enough existing time verifiers.", true)
		return false
	}
	return true
}

func GetMinVerifiers() int {
	// Get the last block
	lastBlock := Blockchain[len(Blockchain)-1]
	// Get the number of verifiers in the last block
	lastVerifierCount := len(lastBlock.TimeVerifiers)
	// Get the minimum number of verifiers
	minVerifiers := int(float64(lastVerifierCount) * 0.75)
	return minVerifiers
}

func VerifySmartContract(contract Contract) bool {
	contractStr := contract.Contents
	hash := sha256.Sum256([]byte(contractStr))
	fmt.Println(hash)
	for _, party := range contract.Parties {
		verifier := oqs.Signature{}
		sigName := "Dilithium3"
		if err := verifier.Init(sigName, nil); err != nil {
			Error("Failed to initialize Dilithium2 verifier", true)
		}
		isValid, err := verifier.Verify(hash[:], party.Signature.S, party.PublicKey.Y)
		if err != nil {
			panic(err)
		}
		if !isValid {
			Log("Invalid smart contract signature detected.", true)
			return false
		}
	}
	return true
}

func VerifyAuthenticationProof(proof *AuthenticationProof, data []byte) bool {
	// Check that data matches the proof data
	if !bytes.Equal(proof.Data, data) {
		Log("Data does not match proof data.", true)
		return false
	}
	// Hash the data
	hash := sha256.Sum256(data)
	// Verify the proof
	verifier := oqs.Signature{}
	sigName := "Dilithium3"
	if err := verifier.Init(sigName, nil); err != nil {
		Error("Failed to initialize Dilithium2 verifier", true)
	}
	isValid, err := verifier.Verify(hash[:], proof.Signature.S, proof.PublicKey.Y)
	if err != nil {
		panic(err)
	}
	if !isValid {
		Log("Invalid authentication proof signature detected.", true)
	}
	return isValid
}
