// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
mod blockutil;
mod buffer;
mod math;
mod read_contract;
mod sanitization;
mod stack;
mod state;
mod syntax_tree;
mod vm;

mod msgpack;
mod conversion;

use crate::blockutil::NodeBlockUtilInterface;
use crate::msgpack::{decode_pending_state, PendingState};
use crate::state::{CachedState, OnchainState, StateManager};
use crate::vm::run_vm;
use std::env;
use std::process::ExitCode;

fn main() -> ExitCode {
    let contract_contents = read_contract::read_contract();
    let args: Vec<String> = env::args().collect();
    let contract_hash = &args[2];
    let gas_limit_f64: f64 = args[3].parse().unwrap();
    let gas_limit = gas_limit_f64 as i64;
    let sender: Vec<u8> = args[4].clone().into();
    let pending_state = decode_pending_state();
    let mut interface = NodeBlockUtilInterface::new();
    let mut state_manager = StateManager::<CachedState, OnchainState, PendingState>::new(
        interface.clone(),
        contract_hash.clone(),
        pending_state,
    );
    let (exit_code, gas_used, _out) = run_vm(
        contract_contents,
        contract_hash.parse().unwrap(),
        gas_limit,
        sender,
        &mut state_manager,
        &mut interface,
    );
    println!("Gas used: {}.0", gas_used);
    ExitCode::from(exit_code as u8)
}
