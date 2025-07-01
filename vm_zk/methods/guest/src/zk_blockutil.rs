use crate::lazy_vector::LazyVector;
use contracts::blockutil::BlockUtilInterface;
use contracts::merkle::{get_from_merkle, MerkleNode};
use smartstring::alias::String;

#[derive(Clone)]
pub(crate) struct ZkBlockutilInterface {
    blockchain_len: u64,
    lazy_vec: LazyVector<MerkleNode<Vec<u8>>>,
}

impl ZkBlockutilInterface {
    pub fn new(blockchain_len: u64, lazy_vec: LazyVector<MerkleNode<Vec<u8>>>) -> Self {
        Self {
            blockchain_len,
            lazy_vec,
        }
    }
}

impl BlockUtilInterface for ZkBlockutilInterface {
    fn read_contract(&mut self, location: String) -> Result<Vec<u8>, String> {
        let res = get_from_merkle::<LazyVector<MerkleNode<Vec<u8>>>>(
                &mut self.lazy_vec,
                location.parse().unwrap());
        Ok(res)
    }

    fn get_from_state(&self, _property: String) -> (Vec<u8>, bool) {
        // Implemented through an alternative state manager, not through the blockutil interface.
        // Never used, and thus left unimplemented.
        unimplemented!()
    }

    fn get_blockchain_len(&self) -> u64 {
        self.blockchain_len
    }

    fn query_oracle(&self, _query_type: u64, _query_body: Vec<u8>) -> (Vec<u8>, bool) {
        // Oracle queries are impossible using a ZK version of the PVM.
        // They will be left unimplemented out of necessity.
        unimplemented!()
    }
}
