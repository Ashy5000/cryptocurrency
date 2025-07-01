mod lazy_vector;
mod merkle_state;
mod ptr_wrapper_state;
mod zk_blockutil;

use crate::lazy_vector::LazyVector;
use crate::merkle_state::MerkleState;
use crate::ptr_wrapper_state::PtrWrapperState;
use crate::zk_blockutil::ZkBlockutilInterface;
use contracts::merkle::{merklize, MerkleNode};
use contracts::msgpack::PendingState;
use contracts::state::{CachedState, State, StateManager};
use contracts::vm::{run_vm, VmRunDetails, ZkContractResult, ZkInfo};
use risc0_zkvm::guest::env;
use risc0_zkvm::sha::rust_crypto::{Digest, Sha256};
use smartstring::alias::String;

fn main() {
    // Read the input
    let run_details: VmRunDetails = env::read();

    // Extract details
    let contract_contents = run_details.contract_contents.clone();
    let contract_hashes = run_details.contract_hash.clone();
    let gas_limits = run_details.gas_limits.clone();
    let senders = run_details.senders.clone();
    let blockchain_len = run_details.blockchain_len.clone();

    // Setup lazy vector
    let lazy_len = run_details.lazy_len;
    let mut lazy_vec: LazyVector<MerkleNode<Vec<u8>>> = LazyVector::new(lazy_len);
    let merkle_root = lazy_vec.get(0).unwrap().hash;

    // Setup blockutil interface
    let mut interface = ZkBlockutilInterface::new(blockchain_len, lazy_vec.clone());

    // Setup pending state
    let mut pending_state = PendingState::new();
    let pending_state_ptr = &mut pending_state as *mut PendingState;

    // Setup merkle state
    let mut merkle_state =
        MerkleState::new(lazy_vec, std::string::String::new(), pending_state_ptr);
    let merkle_state_ptr = &mut merkle_state as *mut MerkleState;

    // Initialize results
    let mut results: Vec<ZkContractResult> = Vec::new();

    // Initialize string output
    let mut out_final = String::new();

    for i in 0..contract_contents.len() {
        // Setup individual state
        merkle_state.prefix = contract_hashes[i].clone().parse().unwrap();
        let pending_wrapper_state = PtrWrapperState::new(pending_state_ptr);
        let merkle_wrapper_state = PtrWrapperState::<MerkleState>::new(merkle_state_ptr);
        let mut state_manager = StateManager {
            cached_state: CachedState::new(),
            onchain_state: merkle_wrapper_state,
            pending_state: pending_wrapper_state,
        };

        // Run VM
        let (exit_code, gas_used, out) = run_vm(
            contract_contents[i].clone(),
            contract_hashes[i].clone().parse().unwrap(),
            gas_limits[i],
            senders[i].clone(),
            &mut state_manager,
            &mut interface,
        );

        // Store result
        let result = ZkContractResult {
            exit_code,
            gas_used,
        };
        results.push(result);
        out_final.push_str(&out);
    }

    let mut hasher = Sha256::new();
    for contract in contract_contents {
        hasher.update(contract);
    }
    for hash in contract_hashes {
        hasher.update(hash);
    }
    for limit in gas_limits {
        hasher.update(limit.to_string());
    }
    for sender in senders {
        hasher.update(sender);
    }
    hasher.update(blockchain_len.to_string());
    let input_hash_output = hasher.finalize();
    let input_hash = hex::encode(input_hash_output);
    
    // Calculate state transition hash
    let state_transition_dump = merkle_state.dump();
    let mut state_vec = state_transition_dump.iter().collect::<Vec<(&std::string::String, &Vec<u8>)>>();
    state_vec.sort_by(|a, b| a.0.cmp(&b.0));
    let state_transition_merkle = merklize(state_vec);
    let state_transition_root = state_transition_merkle[0].hash.clone();
    println!("{}", state_transition_root);

    // Format output
    let output = ZkInfo {
        results,
        out: std::string::String::from(out_final),
        merkle_root,
        state_transition_root,
        input_hash
    };

    // Write public output to the journal
    env::commit(&output);
}
