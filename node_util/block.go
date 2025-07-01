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
	"strconv"
	"strings"
	"time"
)

type Signature struct {
	S []byte
}

func (i Signature) MarshalJSON() ([]byte, error) {
	sigBytes, err := json.Marshal(i.S)
	if err != nil {
		return []byte(""), err
	}
	sigBytes = []byte(strings.Replace(string(sigBytes), `"`, "-", -1))
	return []byte(`"` + string(sigBytes) + `"`), nil
}

func (i *Signature) UnmarshalJSON(data []byte) error {
	// Convert data to string
	str := string(data)
	// Remove double quotes
	str = strings.Replace(str, `"`, "", -1)
	// Convert dashes to double quotes
	str = strings.Replace(str, "-", `"`, -1)
	// Convert string to byte array
	err := json.Unmarshal([]byte(str), &i.S)
	if err != nil {
		return err
	}
	return nil
}

type Transaction struct {
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

func (i Transaction) MarshalJSON() ([]byte, error) {
	var res string
	senderStr := EncodePublicKey(i.Sender)
	senderLen := len(senderStr)
	senderLenStr := fmt.Sprintf("%06d", senderLen)
	res = senderLenStr
	res += senderStr
	recipientStr := EncodePublicKey(i.Recipient)
	recipientLen := len(recipientStr)
	recipientLenStr := fmt.Sprintf("%06d", recipientLen)
	res += recipientLenStr
	res += recipientStr
	amountStr := fmt.Sprintf("%f", i.Amount)
	amountLen := len(amountStr)
	amountLenStr := fmt.Sprintf("%06d", amountLen)
	res += amountLenStr
	res += amountStr
	signatureBytes, err := json.Marshal(i.SenderSignature)
	if err != nil {
		return nil, err
	}
	signatureStr := string(signatureBytes)
	signatureLen := len(signatureStr)
	signatureLenStr := fmt.Sprintf("%06d", signatureLen)
	res += signatureLenStr
	res += signatureStr
	timestampStr := strconv.FormatInt(i.Timestamp.UnixNano(), 10)
	timestampLen := len(timestampStr)
	timestampLenStr := fmt.Sprintf("%06d", timestampLen)
	res += timestampLenStr
	res += timestampStr
	contractsBytes, err := json.Marshal(i.Contracts)
	if err != nil {
		return nil, err
	}
	contractsStr := string(contractsBytes)
	contractsLen := len(contractsStr)
	contractsLenStr := fmt.Sprintf("%06d", contractsLen)
	res += contractsLenStr
	res += contractsStr
	if i.FromSmartContract {
		res += "T"
	} else {
		res += "F"
	}
	bodyStr := string(i.Body)
	bodyLen := len(bodyStr)
	bodyLenStr := fmt.Sprintf("%06d", bodyLen)
	res += bodyLenStr
	res += bodyStr
	bodySigsBytes, err := json.Marshal(i.BodySignatures)
	if err != nil {
		return nil, err
	}
	bodySigsStr := string(bodySigsBytes)
	bodySigsLen := len(bodySigsStr)
	bodySigsLenStr := fmt.Sprintf("%06d", bodySigsLen)
	res += bodySigsLenStr
	res += bodySigsStr
	res = strings.Replace(res, "\"", "\\\"", -1)
	return []byte(`"` + res + `"`), nil
}

func (i *Transaction) UnmarshalJSON(data []byte) error {
	data = []byte(strings.Replace(string(data), "\\\"", "\"", -1))
	data = data[1 : len(data)-2]
	data, senderBytes := Consume(data)
	i.Sender = DecodePublicKey(string(senderBytes))
	data, recipientBytes := Consume(data)
	i.Recipient = DecodePublicKey(string(recipientBytes))
	data, amountBytes := Consume(data)
	amount, err := strconv.ParseFloat(string(amountBytes), 64)
	if err != nil {
		return err
	}
	i.Amount = amount
	data, signatureBytes := Consume(data)
	var signature Signature
	err = json.Unmarshal(signatureBytes, &signature)
	if err != nil {
		return err
	}
	i.SenderSignature = signature
	data, timestampBytes := Consume(data)
	timestampInt, err := strconv.ParseInt(string(timestampBytes), 10, 64)
	if err != nil {
		return err
	}
	i.Timestamp = time.Unix(0, timestampInt)
	data, contractsBytes := Consume(data)
	var contracts []Contract
	err = json.Unmarshal(contractsBytes, &contracts)
	if err != nil {
		return err
	}
	i.Contracts = contracts
	fromSmartContractByte := data[0]
	data = data[1:]
	if fromSmartContractByte == 'T' {
		i.FromSmartContract = true
	} else {
		i.FromSmartContract = false
	}
	data, bodyBytes := Consume(data)
	i.Body = bodyBytes
	data, bodySigsBytes := Consume(data)
	var sigs []Signature
	err = json.Unmarshal(bodySigsBytes, &sigs)
	if err != nil {
		return err
	}
	return nil
}

type Block struct {
	LegacyTransactions              []Transaction   `json:"transactions"`
	ZenTransactions                 []MerkleNode    `json:"zenTransactions"`
	Miner                           PublicKey       `json:"miner"`
	Nonce                           int64           `json:"nonce"`
	MiningTime                      time.Duration   `json:"miningTime"`
	Difficulty                      uint64          `json:"difficulty"`
	PreviousBlockHash               [64]byte        `json:"previousBlockHash"`
	Timestamp                       time.Time       `json:"timestamp"`
	PreMiningTimeVerifierSignatures []Signature     `json:"preMiningTimeVerifierSignatures"`
	PreMiningTimeVerifiers          []PublicKey     `json:"preMiningTimeVerifiers"`
	TimeVerifierSignatures          []Signature     `json:"timeVerifierSignature"`
	TimeVerifiers                   []PublicKey     `json:"timeVerifiers"`
	Transition                      StateTransition `json:"transition"`
	ZenProof                        []byte          `json:"zenProof"`
}

func ExtractTransactions(block Block) []Transaction {
	txs := block.LegacyTransactions
	for _, node := range block.ZenTransactions {
		var tx Transaction
		if node.Data == nil {
			continue
		}
		err := json.Unmarshal(node.Data, &tx)
		if err != nil {
			panic(err)
		}
		txs = append(txs, tx)
	}
	return txs
}
