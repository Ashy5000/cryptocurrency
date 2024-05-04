# Polycash

(Polycash has no association with the Polygon blockchain, despite the similar naming.)

The home of a power-efficient, secure, quantum-resistant, and modern cryptocurrency blockchain designed to resolve the issues presented by the traditional PoW (Proof of Work) incentive mechanism while maintaining decentralization. Written in Golang and Rust for secure algorithm implementations, lightweight networking, memory safety, speed, and reliability.

## Modified PoW protocol
As opposed to the original Proof of Work (PoW) protocol, this blockchain does not calculate block difficulty on a global level. Instead, it is adjusted at a per-miner basis, mostly removing the reward for using more computing power and thus more energy. If a miner has a higher hash rate and thus initially mines blocks faster, they will be provided with higher difficulty blocks and will end up mining blocks at a **nearly constant rate** of *1 min/transaction*. The reason this is not exactly constant is that miners with a higher mining rate do receive slightly lower difficulties than necessary for the constant rate. At infinite computing power, a miner would theoretically have a 50% lower difficulty than expected: this is the lower boundary. A miner with zero computing power would theoretically have a 50% higher difficulty than expected: this is the upper boundary. Keep in mind that the difficulty is still lower for miners with a lower mining rate, but the ratio of difficulty to mining rate is higher with a lower mining rate. This motivates miners to contribute computing power to the network to keep it secure while still maintaining an increase in decentralization. In addition, the block difficulty is retargeted *every* time a miner mines a block, ensuring quick reactions to changes in miner hash rates. To prevent miners from registering multiple private keys in an attempt to avoid the difficulty constraints, there is planned to be a high initial difficulty, or the constant difficulty for the first block a miner mines. In addition, the total number of miners at a time is limited to a maximum that increases as the blockchain's length grows. This prevents miners from using large amounts of computing resources and energy to create a large number of key pairs that have overcome the initial block.

## Time Verification
Using a new verification method, time verification, both the security and the performance of the blockchain are improved as compared to Proof of Work. This protocol can prevent malicious forks from occuring using nodes' verification of new blocks using the curent timestamp, requiring signatures from miner nodes to become valid. Setting rewards and limits for the number of valid signatures in a block strictly enforces the security of this protocol.

## Quantum-Resistant Signatures
This blockchain utilizes the Dilithium2 signature algorithm, a quantum-resistant algorithm chosen as a winner for the NIST Post Quantum Cryptography standardization process.

## Roadmap
The roadmap for this repository looks something like this:

**Development**: Implementing a blockchain and networking system, along with the specific changes to the PoW and networking protocols in order to increase the effectiveness, speed, reliability, and efficiency of the blockchain. Creating a README that documents the design decisions of the blockchain and network and how to use the software. Licensing the software with the GNU General Purpose License v3.0. Performing small scale, local tests on various operating systems and architectures to ensure the software works correctly.

**Testnet**: Running a testing network to verfify the correct operation of the software. Expirementing with its ability to verify large amounts of transactions at once. Making adjustments if neccessary to increase the speed, reliability, and efficiency of the blockchain. Running small-scale tests on the cryptocurrency's financial model.

**Production**: Launching the final product, fixing any issues if and when they arrive.

## Directory structure
The root directory of this project is occupied by the Golang source code that nodes run in order to interact with each other and the decentralized blockchain. In the ```peer_server``` directory, there is Rust code that can be run by servers to maintain a list of peers in the network. Nodes can connect to these servers or maintain their own lists. Using this code is optional, and its only purpose is to make it faster to discover peers. Thirdly, the ```gui_wallet``` directory contains Rust code for a GUI wallet to make transactions easier to make. Note that you will still have to generate a key using the ```keygen``` command in the interactive console (see *To run as a client* below).

## Setup & Usage

For a simpler guide, see the [welcome](docs/welcome.md) page.

*Note: A node must have a publicly accessible IP address in order to join the network. You may have to set up port forwarding on your router.*

Create and clone the repository:

```bash
git clone https://github.com/ashy5000/cryptocurrency
cd cryptocurrency
```

### To run as a client:
To use an interactive console for viewing and adding to the blockchain, run:
```bash
./builds/node/node_linux_x86_64 # replace for your os and architecture
````
Commands:
- `help`: see a list of all commands
- `sync`: update the blockchain and all balances and transactions
- `keygen`: generate a key pair so you can send and receive tokens
- `encrypt`: encrypt the private key so you can store it safely
- `decrypt`: decrypt the private key so you can use it
- `send {recipient} {amount}`: send {amount} tokens to {recipient}
- `balance {key}`: get the balance associated with the public key {key}
- `savestate`: save a backup of the current state of the blockchain to a file
- `loadstate`: load a backup of the current state of the blockchain from a file
- `exit`: exit the console
- `addpeer {ip}`: by default, connects to a peer. If using a centralized peer server, makes yourself known to the network.

To get started, run `keygen` to generate a new key. To get your balance, find your public key in the ```key.json``` file (the long number following ```"Y":```), and run `balance {YOUR KEY HERE}`. To send currency, type `send {RECIPIENT PUBLIC KEY} {AMOUNT}`. You'll have to ask the recipient for their public key. When you're done, type `encrypt` to encrypt your private key and store it safely. You can decrypt it later to use it again with `decrypt`. You must use a passcode that is a multiple of 16 characters long for encryption and decryption. Write it down somewhere safe, as you will not be able to access your private key without it.

### To run a node:
To run the node software, which keeps the blockchain distributed across the p2p network, run:
```bash
./builds/node/node_linux_x86_64 -serve -port 8080  # replace for your os and architecture
# In a new terminal window: (optional, starts a peer server so it is faster to find new nodes)
# This is not at all required.
cd peer_server
cargo run
```


### To run a miner:
To run the mining software, which adds new blocks to the blockchain in exchange for a reward, run:
```bash
./builds/node/node_linux_x86_64 -serve -mine -port 8080
# In a new terminal window: (optional, starts a peer server so it is faster to find new nodes)
# This is not at all required.
cd peer_server
cargo run
```

### To connect to a peer via their peer server:
To connect to a peer that is also running a peer server (ran the commands after `# In a new terminal window:`), run:
```bash
curl http://[PEER IP]:6060 -d 'http://[YOUR IP]:8080'
```

## License
This software is released under the GNU General Public License v3.0.

## Contributing
I'm not accepting pull requests for this repository for now, but you're more than welcome to open an issue if there's something you want improved, added, or fixed.

When the mainnet is fully deployed, this repository will be fully open for contributions. Until then, I'd like to wait until the codebase is fully stable.
