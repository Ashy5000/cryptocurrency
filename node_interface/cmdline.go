// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package node_interface

import (
	"bufio"
	. "cryptocurrency/analysis"
	. "cryptocurrency/node_util"
	. "cryptocurrency/node_util/oracle"
	. "cryptocurrency/rollup"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/open-quantum-safe/liboqs-go/oqs"
)

var commands = map[string]func([]string){
	"sync":                 SyncCmd,
	"balance":              BalanceCmd,
	"send":                 SendCmd,
	"sendL2":               SendL2Cmd,
	"deploySmartContract":  DeploySmartContractCmd,
	"runSmartContract":     RunSmartContractCmd,
	"keygen":               KeygenCmd,
	"showPublicKey":        ShowPublicKeyCmd,
	"encrypt":              EncryptCmd,
	"decrypt":              DecryptCmd,
	"savestate":            SaveStateCmd,
	"loadstate":            LoadStateCmd,
	"addPeer":              AddPeerCmd,
	"bootstrap":            BootstrapCmd,
	"help":                 HelpCmd,
	"license":              LicenseCmd,
	"getNthBlock":          GetNthBlockCmd,       // Get a property of the nth block
	"getNthTransaction":    GetNthTransactionCmd, // Get a property of the nth transaction in the nth block
	"getFromState":         GetFromStateCmd,
	"startAnalysisConsole": StartAnalysisConsoleCmd,
	"sendWithBody":         SendWithBodyCmd,
	"getBlockchainLen":     GetBlockchainLenCmd,
	"queryOracle":          QueryOracleCmd,
	"readSmartContract":    ReadSmartContractCmd,
}

func SyncCmd([]string) {
	Log("Syncing blockchain...", false)
	SyncBlockchain(-1)
	Log("Blockchain successfully synced!", false)
	Log(fmt.Sprintf("Length: %d", len(Blockchain)), false)
}

func BalanceCmd(fields []string) {
	if len(fields) == 1 {
		publicKey := GetKey("").PublicKey.Y
		balance := GetBalance(publicKey)
		fmt.Println(fmt.Sprintf("Balance: %f", balance))
		return
	}
	keyStrFields := fields[1:]
	keyStr := strings.Join(keyStrFields, " ")
	var key []byte
	err := json.Unmarshal([]byte(keyStr), &key)
	if err != nil {
		panic(err)
	}
	balance := GetBalance(key)
	fmt.Println(fmt.Sprintf("Balance: %f", balance))
}

func SendCmd(fields []string) {
	receiverStrFields := fields[1 : len(fields)-1]
	receiverStr := strings.Join(receiverStrFields, " ")
	var receiver []byte
	err := json.Unmarshal([]byte(receiverStr), &receiver)
	if err != nil {
		panic(err)
	}
	amount := fields[len(fields)-1]
	var transactionBody []byte
	Send(string(receiver), amount, transactionBody)
	Log("Waiting for all workers to finish", true)
	Wg.Wait()
	Log("All workers have finished", true)
}

func SendWithBodyCmd(fields []string) {
	receiverStrFields := fields[2 : len(fields)-1]
	receiverStr := strings.Join(receiverStrFields, " ")
	var receiver []byte
	err := json.Unmarshal([]byte(receiverStr), &receiver)
	if err != nil {
		panic(err)
	}
	amount := fields[len(fields)-1]
	transactionBody := []byte(fields[1])
	Send(string(receiver), amount, transactionBody)
	Log("Waiting for all workers to finish", true)
	Wg.Wait()
	Log("All workers have finished", true)
}

func SendL2Cmd(fields []string) {
	receiverStrFields := fields[1 : len(fields)-1]
	receiverStr := strings.Join(receiverStrFields, " ")
	var receiver []byte
	err := json.Unmarshal([]byte(receiverStr), &receiver)
	if err != nil {
		panic(err)
	}
	recieverPubKey := PublicKey{
		Y: receiver,
	}
	amountStr := fields[len(fields)-1]
	amountFloat, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		panic(err)
	}
	amount := uint64(amountFloat * 1000000)
	SendL2Transaction(GetKey("").PublicKey, recieverPubKey, amount)
}

func DeploySmartContractCmd(fields []string) {
	path := fields[1]
	hash, err := DeploySmartContract(path, "")
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(hash[:]))
	fmt.Println("Smart contract deployed successfully!")
}

func RunSmartContractCmd(fields []string) {
	location := fields[1]
	_, err := DeploySmartContract("", location)
	if err != nil {
		panic(err)
	}
}

func KeygenCmd([]string) {
	var privateKey PrivateKey
	sigName := "Dilithium3"
	signer := oqs.Signature{}
	if err := signer.Init(sigName, nil); err != nil {
		Error("Could not initialize Dilithium2 signer", true)
	}
	privateKey.X = signer
	pubKey, err := privateKey.X.GenerateKeyPair()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(privateKey.X.ExportSecretKey()))
	privateKey.PublicKey = PublicKey{
		Y: pubKey,
	}
	keyJson, err := json.Marshal(privateKey)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("key.json", keyJson, 0644)
	if err != nil {
		panic(err)
	}
	// TODO: Implement mnemonics for Dilithium2
}

func ShowPublicKeyCmd([]string) {
	// Show the public key in the key.json file
	publicKey := GetKey("").PublicKey
	publicKeyJson, err := json.Marshal(publicKey.Y)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(publicKeyJson))
}

func EncryptCmd([]string) {
	// Ask the user for a password
	fmt.Print("Enter a password: ")
	inputReader := bufio.NewReader(os.Stdin)
	password, _ := inputReader.ReadString('\n')
	password = password[:len(password)-1]
	// Encrypt the key
	EncryptKey(password)
}

func DecryptCmd([]string) {
	// Ask the user for a password
	fmt.Print("Enter a password: ")
	inputReader := bufio.NewReader(os.Stdin)
	password, _ := inputReader.ReadString('\n')
	password = password[:len(password)-1]
	// Decrypt the key
	DecryptKey(password)
}

func SaveStateCmd([]string) {
	blockchainJson, err := json.Marshal(Blockchain)
	// Save the blockchain to a file
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("blockchain.json", blockchainJson, 0644)
	if err != nil {
		panic(err)
	}
}

// LoadStateCmd loads the blockchain from a file.
//
// It reads the content of the "blockchain.json" file and unmarshals it into the Blockchain variable.
// If any error occurs during the process, it panics.
//
// Parameters:
// - fields: a slice of strings representing the command line fields.
//
// Return type: none.
func LoadStateCmd([]string) {
	// Load the blockchain from a file
	blockchainJson, err := os.ReadFile("blockchain.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(blockchainJson, &Blockchain)
	if err != nil {
		panic(err)
	}
}

func AddPeerCmd(fields []string) {
	// Add the peer to the local peer list
	AddPeer("http://" + fields[1] + ":8080\n")
	// Add the peer to the peer's peer list
	peerServer := fields[1]
	localIp := fields[2]
	body := strings.NewReader(localIp)
	req, err := http.NewRequest(http.MethodGet, peerServer+"/addPeer", body)
	if err != nil {
		panic(err)
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println("Peer added successfully!")
}

func BootstrapCmd([]string) {
	Bootstrap()
	fmt.Println("Bootstrap complete!")
}

func GetFromStateCmd(fields []string) {
	address := fields[1]
	state := CalculateCurrentState()
	// Zen
	dataBytes, ok := GetValue(state.ZenData, address)
	if ok {
		dataHex := hex.EncodeToString(dataBytes)
		fmt.Println("Data:", dataHex)
		return
	}
	// Legacy (fallback)
	dataBytes = state.LegacyData[address]
	dataHex := hex.EncodeToString(dataBytes)
	fmt.Println("Data:", dataHex)
}

func HelpCmd([]string) {
	fmt.Println("Commands:")
	fmt.Println("help - Display this help menu")
	fmt.Println("license - Display this software's license (GNU GPL v3)")
	fmt.Println("sync - Sync the blockchain with peers")
	fmt.Println("keygen - Generate a new key")
	fmt.Println("showPublicKey - Print your public key")
	fmt.Println("encrypt - Encrypt your keys for extra security")
	fmt.Println("decrypt - Decrypt your keys so you can use them")
	fmt.Println("send <public key> <amount> - Send an amount to a public key")
	fmt.Println("sendL2 <public key> <amount> - Send an amount to a public key via L2 rollups (alpha)")
	fmt.Println("balance <public key> - Get the balance of a public key")
	fmt.Println("savestate - Save the blockchain to a file")
	fmt.Println("loadstate - Load the blockchain from a file")
	fmt.Println("deploySmartContract <blockasm path> - Deploy a smart contract to the blockchain")
	fmt.Println("addPeer <ip> - Connect to a peer")
	fmt.Println("startAnalysisConsole - Start a specialized console for analyzing the blockchain and network")
	fmt.Println("bootstrap - Connect to more peers")
	fmt.Println("exit - Exit the console")
}

func LicenseCmd([]string) {
	license, err := os.ReadFile("COPYING")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(license))
}

func GetNthBlockCmd(fields []string) {
	n64, err := strconv.ParseInt(fields[1]+".0", 10, 32)
	if err == nil {
		panic("Invalid block number " + fields[1])
	}
	n := int(n64)
	if len(Blockchain)-1 < n || n < 0 {
		panic("Block out of range")
	}
	block := Blockchain[n]
	property := fields[2]
	switch property {
	case "hash":
		hash := HashBlock(block, n)
		fmt.Println(hex.EncodeToString(hash[:]))
	case "prev_hash":
		fmt.Println(hex.EncodeToString(block.PreviousBlockHash[:]))
	case "transaction_count":
		fmt.Println(len(ExtractTransactions(block)))
	default:
		fmt.Println("Invalid property", property)
	}
}

func GetNthTransactionCmd(fields []string) {
	blockPos, err := strconv.ParseInt(fields[1], 10, 32)
	if err != nil {
		panic(err)
	}
	txPos, err := strconv.ParseInt(fields[2], 10, 32)
	if err != nil {
		panic(err)
	}
	block := Blockchain[blockPos]
	tx := ExtractTransactions(block)[txPos]
	property := fields[3]
	switch property {
	case "sender":
		fmt.Println(hex.EncodeToString(tx.Sender.Y[:]))
	case "recipient":
		fmt.Println(hex.EncodeToString(tx.Recipient.Y[:]))
	case "amount":
		fmt.Println(tx.Amount)
	case "body":
		fmt.Println(hex.EncodeToString(tx.Body))
	}
}

func StartAnalysisConsoleCmd([]string) {
	StartAnalysisCmdline()
}

func GetBlockchainLenCmd([]string) {
	fmt.Println(len(Blockchain))
}

func QueryOracleCmd(fields []string) {
	queryTypeInt, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		panic(err)
	}
	var queryBody []byte
	_, err = hex.Decode(queryBody, []byte(fields[1]))
	if err != nil {
		return
	}
	query := OracleQuery{
		Body: queryBody,
		Type: OracleQueryType(queryTypeInt),
	}
	res := CalculateOracleResponse(query)
	fmt.Println(string(res.Body))
}

func ReadSmartContractCmd(fields []string) {
	state := CalculateCurrentState()
	// Zen
	contract, ok := GetValue(state.ZenContracts, fields[1])
	if ok {
		fmt.Println(string(contract))
		return
	}
	// Legacy (fallback)
	loc, err := strconv.ParseUint(fields[1], 16, 64)
	if err != nil {
		panic(err)
	}
	for l, c := range state.LegacyContracts {
		if l == loc {
			fmt.Println(c.Contents)
			break
		}
	}
}

func RunCmd(input string) {
	cmds := strings.Split(input, ";")
	for _, cmd := range cmds {
		fields := strings.Split(cmd, " ")
		action := fields[0]
		switch action {
		case "exit":
			os.Exit(0)
		case "":
			return
		default:
			fn, ok := commands[action]
			if ok {
				fn(fields)
			} else {
				fmt.Println("Invalid command")
			}
		}
	}
}

func StartCmdLine() {
	fmt.Println("Copyright (C) 2024 Asher Wrobel")
	fmt.Println("This program comes with ABSOLUTELY NO WARRANTY. This is free software, and you are welcome to redistribute it under certain conditions.")
	fmt.Println("To see the license, type `license`.")
	for {
		inputReader := bufio.NewReader(os.Stdin)
		fmt.Printf("BlockCMD console (encrypted: %t): ", IsKeyEncrypted())
		cmd, _ := inputReader.ReadString('\n')
		cmd = cmd[:len(cmd)-1]
		RunCmd(cmd)
	}
}
