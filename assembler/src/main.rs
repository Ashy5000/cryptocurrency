mod assemble;
mod instructions;

use std::{env, fs};
use crate::assemble::assemble;

fn main() {
    let args: Vec<String> = env::args().collect();
    let input = fs::read_to_string(&args[1]).unwrap();
    let output = assemble(input);
    fs::write(&args[2], output).unwrap();
}
