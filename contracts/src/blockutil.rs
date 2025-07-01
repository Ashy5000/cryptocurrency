use std::process::Command;
use hex::FromHexError;
// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
use crate::sanitization::sanitize_node_console_command;
use smartstring::alias::String;

pub trait BlockUtilInterface {
    fn read_contract(&mut self, location: String) -> Result<Vec<u8>, String>;
    fn get_from_state(&self, property: String) -> (Vec<u8>, bool);
    fn get_blockchain_len(&self) -> u64;
    fn query_oracle(&self, query_type: u64, query_body: Vec<u8>) -> (Vec<u8>, bool);
}

#[derive(Clone)]
pub struct NodeBlockUtilInterface {
    node_executable_path: String,
}

impl Default for NodeBlockUtilInterface {
    fn default() -> Self {
        Self::new()
    }
}

impl NodeBlockUtilInterface {
    pub fn new() -> Self {
        Self {
            node_executable_path: "builds/node/node".into(),
        }
    }
}

impl BlockUtilInterface for NodeBlockUtilInterface {
    fn read_contract(&mut self, location: String) -> Result<Vec<u8>, String> {
        let command = format!("readSmartContract {}", location);
        if !sanitize_node_console_command(&command) {
            println!("Forbidden command");
            return Err("Forbidden command".into());
        }
        let output = Command::new(&*self.node_executable_path.clone())
            .arg("--command")
            .arg(command)
            .output();
        let output = output.expect("Failed to execute node script");
        let mut output = output.stdout;
        output.pop().unwrap();
        Ok(output.iter().skip_while(|x| **x != b'\n').skip(1).skip_while(|x| **x != b'\n').skip(1).map(|x| *x).collect())
    }
    fn get_from_state(&self, property: String) -> (Vec<u8>, bool) {
        let command = format!("sync;getFromState {}", property);
        if !sanitize_node_console_command(&command) {
            println!("Forbidden command");
            return (vec![], false);
        }
        let output = Command::new(&*self.node_executable_path.clone())
            .arg("--command")
            .arg(command)
            .output();
        let output = std::string::String::from_utf8(output.unwrap().stdout)
            .expect("Failed to convert output to string");
        let output = output.trim().to_string();
        let output = output.split("\n").collect::<Vec<&str>>();
        let output = output[output.len() - 1].to_string();
        let data_hex: String = output.chars().skip(6).collect();
        let data = hex::decode(data_hex).unwrap_or_else(|_| Vec::new());
        (data, true)
    }
    fn get_blockchain_len(&self) -> u64 {
        let command = "sync;getBlockchainLen";
        let output = Command::new(&*self.node_executable_path.clone())
            .arg("--command")
            .arg(command)
            .output();
        let output = std::string::String::from_utf8(output.unwrap().stdout)
            .expect("Failed to convert output to string");
        let output = output.trim().to_string();
        let output = output.split("\n").collect::<Vec<&str>>();
        let output = output[output.len() - 1].to_string();
        output.parse::<u64>().unwrap() // Long Live the Turbofish.
    }
    fn query_oracle(&self, query_type: u64, query_body: Vec<u8>) -> (Vec<u8>, bool) {
        let query_body_hex = hex::encode(query_body);
        let command = format!("queryOracle {} {}", query_type, query_body_hex);
        if !sanitize_node_console_command(&command) {
            println!("Forbidden command");
            return (vec![], false);
        }
        let output = Command::new(&*self.node_executable_path.clone())
            .arg("--command")
            .arg(command)
            .output();
        let output = std::string::String::from_utf8(output.unwrap().stdout)
            .expect("Failed to convert output to string");
        let output = output.trim().to_string();
        let output = output.split("\n").collect::<Vec<&str>>();
        let output = output[output.len() - 1].to_string();
        let response_body_hex: String = output.chars().skip(6).collect();
        let response_body: Vec<u8> = hex::decode(response_body_hex).unwrap();
        (response_body, true)
    }
}
