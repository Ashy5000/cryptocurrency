use rustc_hash::FxHashMap;
// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
use crate::{buffer::Buffer, vm};
use smartstring::alias::String;

pub(crate) trait MathOperation {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String>;
}

pub(crate) struct Add {}

impl MathOperation for Add {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a.overflowing_add(b).0)
    }
}

pub(crate) struct Subtract {}

impl MathOperation for Subtract {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a.overflowing_sub(b).0)
    }
}

pub(crate) struct Multiply {}

impl MathOperation for Multiply {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a.overflowing_mul(b).0)
    }
}

pub(crate) struct Divide {}

impl MathOperation for Divide {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        if b == 0 {
            return Err("Division by zero".into());
        }
        Ok(a.overflowing_div(b).0)
    }
}

pub(crate) struct Exp {}

impl MathOperation for Exp {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a.overflowing_pow(b as u32).0)
    }
}

pub(crate) struct Modulo;

impl MathOperation for Modulo {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        if b == 0 {
            return Ok(0);
        }
        Ok(a % b)
    }
}

pub(crate) struct And {}

impl MathOperation for And {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a & b)
    }
}

pub(crate) struct Or {}

impl MathOperation for Or {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        Ok(a | b)
    }
}

pub(crate) struct Not {}

impl MathOperation for Not {
    fn execute(&self, a: u64, _b: u64) -> Result<u64, String> {
        if a == 0 {
            Ok(1)
        } else {
            Ok(0)
        }
    }
}

pub(crate) struct Eq {}

impl MathOperation for Eq {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        if a == b {
            Ok(1)
        } else {
            Ok(0)
        }
    }
}

pub(crate) struct Less {}

impl MathOperation for Less {
    fn execute(&self, a: u64, b: u64) -> Result<u64, String> {
        if a < b {
            Ok(1)
        } else {
            Ok(0)
        }
    }
}

pub(crate) fn execute_math_operation(
    operation: impl MathOperation,
    buffers: &mut FxHashMap<Vec<u8>, Buffer>,
    a: &Vec<u8>,
    b: &Vec<u8>,
    res: &Vec<u8>,
    err: &Vec<u8>,
) {
    let status_a = vm::vm_check_buffer_initialization(buffers, &a);
    let mut status_b = true;
    if std::any::type_name_of_val(&operation) != "contracts::math::Not" {
        status_b = vm::vm_check_buffer_initialization(buffers, &b);
        if !status_b {
            vm::vm_throw_local_error(buffers, &err);
        }
    }
    let status_res = vm::vm_check_buffer_initialization(buffers, &res);
    if !status_a || !status_b || !status_res {
        vm::vm_throw_local_error(buffers, &err);
    }
    let buffer_0 = buffers.get(a).unwrap();
    let mut buffer_1 = &Buffer {
        contents: vec![0, 0, 0, 0, 0, 0, 0, 0],
    };
    if std::any::type_name_of_val(&operation) != "contracts::math::Not" {
        buffer_1 = buffers.get(b).unwrap();
    }
    let buffer_0_u64 = buffer_0.as_u64().unwrap();
    let result_u64 = if std::any::type_name_of_val(&operation) != "contracts::math::Not" {
        let buffer_1_u64 = buffer_1.as_u64().unwrap();
        operation.execute(buffer_0_u64, buffer_1_u64)
    } else {
        operation.execute(buffer_0_u64, 0)
    };
    let buffer_result = buffers.get_mut(res).unwrap();
    buffer_result.load_u64(result_u64.expect("Error storing result in buffer"));
}
