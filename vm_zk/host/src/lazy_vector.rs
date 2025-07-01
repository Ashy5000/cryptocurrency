use serde::Serialize;

pub struct HostVector<T> {
    elements: Vec<T>,
}

impl<T: Clone + Serialize> HostVector<T> {
    pub fn new(elements: Vec<T>) -> Self {
        Self { elements }
    }

    pub fn handle_guest_request(
        &self,
    ) -> impl Fn(risc0_zkvm::Bytes) -> risc0_zkvm::Result<risc0_zkvm::Bytes> + '_ {
        |data| {
            let indices = data.to_vec();
            let data = self.elements[indices[0] as usize].clone();
            let vec = bincode::serde::encode_to_vec(data, bincode::config::standard())?;
            Ok(risc0_zkvm::Bytes::from(vec))
        }
    }
}
