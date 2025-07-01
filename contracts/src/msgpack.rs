use crate::state::State;
use rmpv::decode::read_value;
use rustc_hash::FxHashMap;
use serde::{Deserialize, Serialize};
use std::fs::File;
use std::io::Read;

#[derive(Serialize, Deserialize, Clone, Debug)]
pub struct PendingState {
    pub data: FxHashMap<String, Vec<u8>>,
}

impl PendingState {
    pub fn new() -> PendingState {
        PendingState {
            data: FxHashMap::default(),
        }
    }
}

impl State for PendingState {
    fn write(&mut self, location: String, contents: Vec<u8>, _out: &mut String) {
        self.data.insert(location, contents);
    }

    fn get(&mut self, location: String) -> Result<Vec<u8>, String> {
        Ok(self.data[&location].clone())
    }
    fn dump(&mut self) -> FxHashMap<String, Vec<u8>> {
        self.data.clone()
    }

    fn update_prefix(&mut self, prefix: String) {
        panic!("PendingState does not use a prefix.");
    }
}

pub fn decode_pending_state() -> PendingState {
    let mut file = File::open("pending_state.msgpack").unwrap();
    let mut vec: Vec<u8> = Vec::new();
    file.read_to_end(&mut vec).unwrap();
    let mut reader = &*vec;
    let value = read_value(&mut reader).expect("Failed to decode pending state");
    assert!(value.is_map());
    let value_map = value.as_map().expect("Failed to decode pending state");
    let mut res = PendingState {
        data: FxHashMap::default(),
    };
    for pair in value_map {
        let key = &pair.0;
        assert!(key.is_str());
        let key_string = String::from(key.as_str().unwrap());
        let value = &pair.1;
        assert!(value.is_str());
        let value_bytes = value.as_slice().unwrap();
        res.data.insert(key_string, Vec::from(value_bytes));
    }
    res
}
