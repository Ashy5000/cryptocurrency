use crate::{buffer::Buffer, vm};

use std::collections::HashMap;

pub(crate) trait MathOperation {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String>;
}

pub(crate) struct Add {}

impl MathOperation for Add {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a + b)
    }
}

pub(crate) struct Subtract {}

impl MathOperation for Subtract {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a - b)
    }
}

pub(crate) struct Multiply {}

impl MathOperation for Multiply {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a * b)
    }
}

pub(crate) struct Divide {}

impl MathOperation for Divide {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        if b == 0 {
            return Err("Division by zero".to_string());
        }
        Ok(a / b)
    }
}

pub(crate) fn execute_math_operation(
    operation: impl MathOperation,
    buffers: &mut HashMap<String, Buffer>,
    a: String,
    b: String,
    res: String,
    err: String,
) {
    let status_a = vm::vm_check_buffer_initialization(buffers, a.clone());
    let status_b = vm::vm_check_buffer_initialization(buffers, b.clone());
    let status_res = vm::vm_check_buffer_initialization(buffers, res.clone());
    if !status_a || !status_b || !status_res {
        vm::vm_throw_local_error(buffers, err);
    }
    let buffer_0 = buffers.get(&a).unwrap();
    let buffer_1 = buffers.get(&b).unwrap();
    let buffer_0_u64 = buffer_0.as_u64().unwrap();
    let buffer_1_u64 = buffer_1.as_u64().unwrap();
    let result_u64 = operation.execute(buffer_0_u64, buffer_1_u64);
    let buffer_result = buffers.get_mut(&res).unwrap();
    buffer_result.load_u64(result_u64.expect("Error storing result in buffer"));
}
