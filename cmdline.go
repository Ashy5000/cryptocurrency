// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/open-quantum-safe/liboqs-go/oqs"
)

func RunCmd(input string) {
	cmds := strings.Split(input, ";")
	for _, cmd := range cmds {
		fields := strings.Split(cmd, " ")
		action := fields[0]
		if action == "sync" {
			Log("Syncing blockchain...", false)
			SyncBlockchain()
			Log("Blockchain successfully synced!", false)
			Log(fmt.Sprintf("Length: %d", len(blockchain)), false)
		} else if action == "balance" {
			if len(fields) == 1 {
				publicKey := GetKey().PublicKey.Y
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
		} else if action == "send" {
			receiverStrFields := fields[1 : len(fields)-1]
			receiverStr := strings.Join(receiverStrFields, " ")
			var receiver []byte
			err := json.Unmarshal([]byte(receiverStr), &receiver)
			if err != nil {
				panic(err)
			}
			amount := fields[len(fields)-1]
			Send(string(receiver), amount)
			Log("Waiting for all workers to finish", true)
			wg.Wait()
			Log("All workers have finished", true)
		} else if action == "keygen" {
			var privateKey PrivateKey
			sigName := "Dilithium2"
			signer := oqs.Signature{}
			if err := signer.Init(sigName, nil); err != nil {
				Error("Could not initialize Dilithium2 signer", true)
			}
			privateKey.X = signer
			pubKey, err := privateKey.X.GenerateKeyPair()
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
			//			mnemonic0 := GetMnemonic(*privateKey.X)
			//			mnemonic1 := GetMnemonic(*privateKey.PublicKey.Y)
			//			mnemonic2 := GetMnemonic(*privateKey.PublicKey.Parameters.P)
			//			mnemonic3 := GetMnemonic(*privateKey.PublicKey.Parameters.Q)
			//			mnemonic4 := GetMnemonic(*privateKey.PublicKey.Parameters.G)
			//			fmt.Println("Mnemonic:")
			//			fmt.Println("Part 0: " + mnemonic0)
			//			fmt.Println("Part 1: " + mnemonic1)
			//			fmt.Println("Part 2: " + mnemonic2)
			//			fmt.Println("Part 3: " + mnemonic3)
			//			fmt.Println("Part 4: " + mnemonic4)
			//			fmt.Println("Write down the mnemonic and keep it safe, or better yet memorize it. It is the ONLY WAY to recover your private key.")
		} else if action == "encrypt" {
			// Ask the user for a password
			fmt.Print("Enter a password: ")
			inputReader := bufio.NewReader(os.Stdin)
			password, _ := inputReader.ReadString('\n')
			password = password[:len(password)-1]
			// Encrypt the key
			EncryptKey(password)
		} else if action == "decrypt" {
			// Ask the user for a password
			fmt.Print("Enter a password: ")
			inputReader := bufio.NewReader(os.Stdin)
			password, _ := inputReader.ReadString('\n')
			password = password[:len(password)-1]
			// Decrypt the key
			DecryptKey(password)
		} else if action == "savestate" {
			// Save the blockchain to a file
			blockchainJson, err := json.Marshal(blockchain)
			if err != nil {
				panic(err)
			}
			err = os.WriteFile("blockchain.json", blockchainJson, 0644)
			if err != nil {
				panic(err)
			}
		} else if action == "loadstate" {
			// Load the blockchain from a file
			blockchainJson, err := os.ReadFile("blockchain.json")
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(blockchainJson, &blockchain)
			if err != nil {
				panic(err)
			}
		} else if action == "addpeer" {
			if *useLocalPeerList {
				// Add the peer to the local peer list
				AddPeer("http://" + fields[1] + ":8080\n")
			} else {
				// Get the IP address of the peer
				ip := fields[1]
				// Send a request to the peer server to add the peer
				peerServer := "http://192.168.4.87:8080"
				req, err := http.NewRequest(http.MethodGet, peerServer+"/add_peer/"+ip, nil)
				if err != nil {
					panic(err)
				}
				_, err = http.DefaultClient.Do(req)
				if err != nil {
					panic(err)
				}
			}
			fmt.Println("Peer added successfully!")
		} else if action == "bootstrap" {
			Bootstrap()
			fmt.Println("Bootstrap complete!")
		} else if action == "exit" {
			os.Exit(0)
		} else if action == "help" {
			fmt.Println("Commands:")
			fmt.Println("sync - Sync the blockchain with peers")
			fmt.Println("balance <public key> - Get the balance of a public key")
			fmt.Println("send <public key> <amount> - Send an amount to a public key")
			fmt.Println("keygen - Generate a new key")
			fmt.Println("savestate - Save the blockchain to a file")
			fmt.Println("loadstate - Load the blockchain from a file")
			fmt.Println("exit - Exit the console")
		} else if action == "license" {
			license, err := os.ReadFile("COPYING")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(license))
		} else if action == "" {
			return
		} else {
			fmt.Println("Invalid command.")
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
