use crate::lazy_vector::LazyVector;
use contracts::merkle::{get_from_merkle, MerkleContainer, MerkleNode};
use contracts::msgpack::PendingState;
use contracts::state::State;
use rustc_hash::FxHashMap;

impl MerkleContainer<Vec<u8>> for LazyVector<MerkleNode<Vec<u8>>> {
    fn get_wrapper(&mut self, index: usize) -> Option<MerkleNode<Vec<u8>>> {
        self.get(index)
    }
}

#[derive(Clone)]
pub(crate) struct MerkleState {
    pub(crate) contents: LazyVector<MerkleNode<Vec<u8>>>,
    pub(crate) prefix: String,
    pub(crate) pending_state: *mut PendingState,
    pub(crate) transition: FxHashMap<String, Vec<u8>>
}

impl MerkleState {
    pub(crate) fn new(
        contents: LazyVector<MerkleNode<Vec<u8>>>,
        prefix: String,
        pending_state: *mut PendingState,
    ) -> Self {
        MerkleState {
            contents,
            prefix,
            pending_state,
            transition: FxHashMap::default()
        }
    }
}

impl State for MerkleState {
    fn write(&mut self, location: String, contents: Vec<u8>, out: &mut String) {
        // Same write functionality as OnchainState
        if location.as_bytes().starts_with(&self.prefix.as_bytes()) {
            let string = format!(
                "State change: {}|{}",
                location,
                hex::encode(contents.clone())
            );
            println!("{}\n", string);
            self.transition.insert(location.clone(), contents.clone());
            out.push_str(&string);
        } else {
            // Ensure value is set to the ESWV
            let existing_value_vec_u8 = self.get(location.clone()).unwrap();
            let existing_value = String::from_utf8_lossy(&*existing_value_vec_u8);
            if existing_value == "ExternalStateWriteableValue" {
                let string = format!(
                    "State change: {}|{}",
                    location,
                    hex::encode(contents.clone())
                );
                println!("{}\n", string);
                self.transition.insert(location.clone(), contents.clone());
                out.push_str(&string);
            }
        }
        unsafe {
            (*self.pending_state).write(location, contents, out);
        }
    }

    fn get(&mut self, location: String) -> Result<Vec<u8>, String> {
        Ok(get_from_merkle::<LazyVector<MerkleNode<Vec<u8>>>>(
            &mut self.contents,
            location,
        ))
    }

    fn dump(&mut self) -> FxHashMap<String, Vec<u8>> {
        self.transition.clone()
    }

    fn update_prefix(&mut self, prefix: String) {
        self.prefix = prefix;
    }
}
