use rustc_hash::FxHashMap;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};

#[derive(Clone, Serialize, Deserialize)]
pub struct MerkleNode<T> {
    pub hash: String,
    pub value: T,
    pub children: FxHashMap<char, usize>,
    pub parent: usize,
}

fn hash_node<A: MerkleContainer<Vec<u8>> + Clone>(tree: &mut A, node_index: usize) -> String {
    let mut hasher = Sha256::new();
    hasher.update(tree.get_wrapper(node_index).unwrap().value.clone());
    let mut indices = vec![];
    for (_, index) in tree.get_wrapper(node_index).unwrap().children.iter() {
        indices.push(*index);
    }
    indices.sort();
    for index in indices.iter() {
        hasher.update(tree.get_wrapper(index.clone()).unwrap().hash.clone());
    }
    let hash = hasher.finalize();
    hex::encode(hash)
}

fn fill_vec_node_hash(tree: &mut Vec<MerkleNode<Vec<u8>>>, node_index: usize) {
    tree[node_index].hash = hash_node(&mut tree.clone(), node_index);
}

fn hash_vec_tree(tree: &mut Vec<MerkleNode<Vec<u8>>>, index: usize) {
    if tree.len() == 1 {
        tree[0].hash = "".to_owned();
        return
    }
    if tree[index].children.len() == 0 {
        // Base case
        fill_vec_node_hash(tree, index);
    } else {
        // Recursive case
        for (_, child_index) in tree[index].children.clone().iter() {
            hash_vec_tree(tree, *child_index);
            fill_vec_node_hash(tree, index);
        }
    }
}

fn verify_node<A: MerkleContainer<Vec<u8>> + Clone>(tree: &mut A, node_index: usize) {
    assert_eq!(
        hash_node(tree, node_index),
        tree.get_wrapper(node_index).unwrap().hash,
        "Assertion failed on node with index {}",
        node_index
    );
    if node_index != 0 {
        verify_node(tree, tree.clone().get_wrapper(node_index).unwrap().parent);
    }
}

pub fn merklize(state: Vec<(&std::string::String, &Vec<u8>)>) -> Vec<MerkleNode<Vec<u8>>> {
    let mut tree: Vec<MerkleNode<Vec<u8>>> = vec![MerkleNode {
        hash: String::new(),
        value: Vec::new(),
        children: FxHashMap::default(),
        parent: 0,
    }];
    for (loc, val) in state.iter() {
        let mut active: usize = 0;
        for c in loc.chars() {
            if tree[active].children.contains_key(&c) {
                active = tree[active].children[&c];
                continue;
            }
            tree.push(MerkleNode {
                hash: String::new(),
                value: Vec::new(),
                children: FxHashMap::default(),
                parent: active,
            });
            let new_index = tree.len() - 1;
            tree[active].children.insert(c, new_index);
            active = new_index;
        }
        tree[active].value = val.clone().clone();
    }
    hash_vec_tree(&mut tree, 0);
    tree
}

pub trait MerkleContainer<T> {
    fn get_wrapper(&mut self, index: usize) -> Option<MerkleNode<T>>;
}

impl MerkleContainer<Vec<u8>> for Vec<MerkleNode<Vec<u8>>> {
    fn get_wrapper(&mut self, index: usize) -> Option<MerkleNode<Vec<u8>>> {
        self.get(index).cloned()
    }
}

pub fn get_from_merkle<A: MerkleContainer<Vec<u8>> + Clone>(tree: &mut A, loc: String) -> Vec<u8> {
    let mut active: usize = 0;
    for c in loc.chars() {
        let node = tree.get_wrapper(active).unwrap();
        match node.children.get(&c) {
            Some(x) => {
                active = *x;
            }
            None => {
                return Vec::new();
            }
        }
    }
    verify_node::<A>(tree, active);
    tree.get_wrapper(active).unwrap().value.clone()
}
