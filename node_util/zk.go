package node_util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// Overview
// Zero-knowledge proofs are handled by a Rust program using RiscZero.
// The binary is located at `vm_zk/target/release/host`.
// It handles proof creation and verification.
// It accesses the state of the VM and deployed contracts via a file, merkle.txt.
// Then, it creates a merkle tree from the data.
// It accepts a batch of transactions to create a proof.
// The proof includes the merkle root.

// Go utilities

func WriteZkState(state State) {
	var segments [][]byte
	aggregateMerkleTree := Merge(state.ZenData, state.ZenContracts)
	for _, i := range aggregateMerkleTree {
		if i.Data != nil {
			var segment string
			keyLenStr := fmt.Sprintf("%06d", len(i.Key))
			segment += keyLenStr
			segment += i.Key
			valLenStr := fmt.Sprintf("%06d", len(i.Data))
			segment += valLenStr
			segmentBytes := []byte(segment)
			segmentBytes = append(segmentBytes, i.Data...)
			segments = append(segments, segmentBytes)
		}
	}
	var result []byte
	for segment := range segments {
		result = append(result, segments[segment]...)
	}
	err := os.WriteFile("merkle.txt", result, 0644)
	if err != nil {
		panic(err)
	}
}

func GenerateZkArgs(generate bool, hashes []string, gasLimits []float64, senders []PublicKey, merkleRoot string, inputHash string, transitionHash string) string {
	if generate {
		// Generate args for a ZK proof
		contractsArg := "contract.blockasm" // Contracts to be included
		hashesArg := ""                     // Contract hashes
		if len(hashes) != 0 {
			for hash := range hashes {
				hashesArg += hashes[hash] + "%"
			}
			hashesArg = hashesArg[:len(hashesArg)-1]
		}
		gasLimitsArg := "" // Gas limits
		if len(gasLimits) != 0 {
			for gasLimit := range gasLimits {
				gasLimitsArg += fmt.Sprintf("%f", gasLimits[gasLimit]) + "%"
			}
			gasLimitsArg = gasLimitsArg[:len(gasLimitsArg)-1]
		}
		sendersArg := "" // Senders
		if len(senders) != 0 {
			for sender := range senders {
				sendersArg += hex.EncodeToString(senders[sender].Y) + "%"
			}
			sendersArg = sendersArg[:len(sendersArg)-1]
		}
		merkleArg := "merkle.txt"   // State
		recieptArg := "receipt.bin" // Receipt
		args := []string{contractsArg, hashesArg, gasLimitsArg, sendersArg, merkleArg, recieptArg}
		return strings.Join(args, " ")
	} else {
		// Generate args for a ZK verification
		verificationArg := "V"              // Verification
		receiptArg := "receipt.bin"         // Receipt
		merkleRootArg := merkleRoot         // Merkle root
		inputHashArg := inputHash           // Input hash
		transitionHashArg := transitionHash // Transition hash
		args := []string{verificationArg, receiptArg, merkleRootArg, inputHashArg, transitionHashArg}
		return strings.Join(args, " ")
	}
}

func WriteContractsAggregate(contracts []Contract) {
	if len(contracts) == 0 {
		err := os.WriteFile("contract.blockasm", []byte(""), 0644)
		if err != nil {
			panic(err)
		}
		return
	}
	var segments [][]byte
	for _, contract := range contracts {
		segments = append(segments, contract.Contents)
	}
	var res []byte
	for segment := range segments {
		res = append(res, segments[segment]...)
		res = append(res, '*')
	}
	res = res[:len(res)-1]
	err := os.WriteFile("contract.blockasm", []byte(res), 0644)
	if err != nil {
		panic(err)
	}
}

func LoadReceipt() []byte {
	receipt, err := os.ReadFile("receipt.bin")
	if err != nil {
		panic(err)
	}
	return receipt
}

func WriteReceipt(receipt []byte) {
	err := os.WriteFile("receipt.bin", receipt, 0644)
	if err != nil {
		panic(err)
	}
}

func SendZkRequest(args string) (string, error) {
	err := SendString(Conn, args)
	if err != nil {
		return "", err
	}
	return ReceiveString()
}

func ZkProve(contracts []Contract, gasLimits []float64, senders []PublicKey, state State) (string, []byte) {
	// 1. Write contracts to file
	WriteContractsAggregate(contracts)
	// 2. Write state to file
	WriteZkState(state)
	// 3. Generate arguments
	var hashes []string
	for contract := range contracts {
		contractStr := contracts[contract].Contents
		hash := sha256.Sum256([]byte(contractStr))
		hashes = append(hashes, hex.EncodeToString(hash[:]))
	}
	args := GenerateZkArgs(true, hashes, gasLimits, senders, "", "", "")
	// 4. Send request
	res, err := SendZkRequest(args)
	if err != nil {
		panic(err)
	}
	// 5. Read receipt
	receipt := LoadReceipt()
	return res, receipt
}

func ZKVerify(receipt []byte, merkleRoot string, inputHash string, transitionHash string) bool {
	// 1. Write receipt to file
	WriteReceipt(receipt)
	// 2. Generate arguments
	args := GenerateZkArgs(false, nil, nil, nil, merkleRoot, inputHash, transitionHash)
	// 3. Send request
	_, err := SendZkRequest(args)
	return err == nil
}
