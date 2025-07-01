package node_util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func ConsumeLength(data []byte) ([]byte, int) {
	toConsume, err := strconv.Atoi(string(data[:6]))
	if err != nil {
		panic(err)
	}
	return data[6:], toConsume
}

func ConsumeBody(data []byte, toConsume int) ([]byte, []byte) {
	body := data[:toConsume]
	if toConsume < len(data) {
		return data[toConsume:], body
	} else {
		return []byte{}, body
	}
}

func Consume(data []byte) ([]byte, []byte) {
	data, toConsume := ConsumeLength(data)
	data, body := ConsumeBody(data, toConsume)
	return data, body
}

func EncodeMineRequest(sender string, recipient string, amount string, sig string, timestamp int64, contracts string, txBody []byte, bodySigs []Signature) string {
	res := fmt.Sprintf("%06d", len(sender))
	res += sender
	res += fmt.Sprintf("%06d", len(recipient))
	res += recipient
	res += fmt.Sprintf("%06d", len(amount))
	res += amount
	res += fmt.Sprintf("%06d", len(sig))
	res += sig
	timestampStr := fmt.Sprintf("%d", timestamp)
	res += fmt.Sprintf("%06d", len(timestampStr))
	res += timestampStr
	res += fmt.Sprintf("%06d", len(contracts))
	res += contracts
	res += fmt.Sprintf("%06d", len(txBody))
	res += string(txBody)
	sigsBytes, err := json.Marshal(bodySigs)
	if err != nil {
		panic(err)
	}
	res += fmt.Sprintf("%06d", len(sigsBytes))
	res += string(sigsBytes)
	return res
}

func DecodeMineRequest(body []byte) (PublicKey, PublicKey, float64, Signature, time.Time, []Contract, []byte, []Signature) {
	body, senderBytes := Consume(body)
	sender := DecodePublicKey(string(senderBytes))
	body, recipientBytes := Consume(body)
	recipient := DecodePublicKey(string(recipientBytes))
	body, amountBytes := Consume(body)
	amount, err := strconv.ParseFloat(string(amountBytes), 64)
	if err != nil {
		panic(err)
	}
	body, sigBytes := Consume(body)
	var sig Signature
	err = json.Unmarshal(sigBytes, &sig)
	if err != nil {
		panic(err)
	}
	body, timestampBytes := Consume(body)
	timestampInt, err := strconv.ParseInt(string(timestampBytes), 10, 64)
	if err != nil {
		panic(err)
	}
	timestamp := time.Unix(0, timestampInt)
	body, contractsBytes := Consume(body)
	var contracts []Contract
	err = json.Unmarshal(contractsBytes, &contracts)
	body, txBody := Consume(body)
	body, bodySigsBytes := Consume(body)
	var bodySigs []Signature
	err = json.Unmarshal(bodySigsBytes, &bodySigs)
	if err != nil {
		panic(err)
	}
	return sender, recipient, amount, sig, timestamp, contracts, txBody, bodySigs
}
