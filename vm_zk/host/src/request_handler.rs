use std::ffi::OsString;
use std::fs;
use std::fs::File;
use std::io::{BufReader, Read};
use contracts::blockutil::{BlockUtilInterface, NodeBlockUtilInterface};
use contracts::merkle::merklize;
use contracts::read_contract::read_contract;
use contracts::vm::ZkInfo;
use risc0_zkvm::Receipt;
use rustc_hash::FxHashMap;
use crate::lazy_vector::HostVector;
use crate::prove::prove;
use crate::socket::Socket;
use crate::verify::verify;

pub(crate) fn handle_request(data: Vec<u8>, socket: &mut Socket) {
    let s = String::from_utf8(data).unwrap();
    let args: Vec<&str> = s.split(" ").collect();
    if args[0] == "V" {
        // Verify
        let receipt_file = fs::read(args[1]).unwrap();
        let receipt: Receipt = rmp_serde::from_slice(&*receipt_file).unwrap();
        let expected_merkle_root = args[2].to_owned();
        let expected_input_hash = args[3].to_owned();
        let expected_transition_hash = args[4].to_owned();
        assert!(verify(receipt, expected_merkle_root, expected_input_hash, expected_transition_hash));
        println!("Verification success!");
        socket.write_message("Verification success!".as_ref()).unwrap();
        return;
    }

    let file = File::open(args[0]).unwrap();
    let mut reader = BufReader::new(file);
    let mut buffer = vec![];
    reader.read_to_end(&mut buffer).unwrap();
    let contract_contents_str = buffer.split(|x| *x == b'%').collect::<Vec<&[u8]>>(); // % marks separation between contracts
    let mut contract_contents = Vec::new();
    for contract in contract_contents_str {
        contract_contents.push(contract.to_vec());
    }
    let contract_hashes_str = args[1].split("%").collect::<Vec<&str>>();
    let mut contract_hashes = Vec::new();
    for hash in contract_hashes_str {
        contract_hashes.push(std::string::String::from(hash));
    }
    let gas_limits_str = args[2].split("%").collect::<Vec<&str>>();
    let mut gas_limits = Vec::new();
    for limit in gas_limits_str {
        gas_limits.push(limit.parse::<f64>().unwrap() as i64);
    }
    let senders_str: Vec<&str> = args[3].split("%").collect::<Vec<&str>>();
    let mut senders: Vec<Vec<u8>> = Vec::new();
    for sender in senders_str {
        senders.push(sender.into());
    }

    // Initialize merkle tree
    let mut data: FxHashMap<String, Vec<u8>> = FxHashMap::default();
    let merkle_raw = fs::read(args[4]).unwrap();
    let mut iter = merkle_raw.into_iter();
    while iter.len() > 0 {
        loop {
            let len_bytes = iter.by_ref().take(6).collect::<Vec<u8>>();
            let len = String::from_utf8(len_bytes).unwrap().parse::<usize>().unwrap();
            let key_bytes = iter.by_ref().take(len).collect::<Vec<u8>>();
            let key = String::from_utf8(key_bytes).unwrap();
            let len_bytes = iter.by_ref().take(6).collect::<Vec<u8>>();
            let len = String::from_utf8(len_bytes).unwrap().parse::<usize>().unwrap();
            let val = iter.by_ref().take(len).collect::<Vec<u8>>();
            data.insert(key, val);
        }
    }
    let mut state_vec = data.iter().collect::<Vec<(&String, &Vec<u8>)>>();
    state_vec.sort_by(|a, b| a.0.cmp(&b.0));
    let tree = merklize(state_vec);
    let lazy_len = tree.len();
    let host_vector = HostVector::new(tree);

    // Create node blockutil for data fetching
    let node_blockutil = NodeBlockUtilInterface::new();

    // Fetch data from node
    let blockchain_len = node_blockutil.get_blockchain_len();

    let run_details = contracts::vm::VmRunDetails {
        contract_contents,
        contract_hash: contract_hashes,
        gas_limits,
        senders,
        lazy_len,
        blockchain_len,
    };

    let receipt = prove(run_details, host_vector, socket);
    let receipt_serialized = rmp_serde::to_vec(&receipt).unwrap();

    let out_file = args[5];
    fs::write(out_file, receipt_serialized).unwrap();
    
    let zk_info: ZkInfo = receipt.journal.decode().unwrap();
    socket.write_message(zk_info.out.as_ref()).unwrap()
}