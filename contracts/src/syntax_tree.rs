// Copyright 2024, Asher Wrobel
/*
This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
use smartstring::alias::String;
#[derive(Debug, Clone)]
pub struct Line {
    pub command: u8,
    pub args: Vec<Vec<u8>>,
}

pub(crate) fn build_line() -> Line {
    Line {
        command: b'\xFF',
        args: Vec::new(),
    }
}

#[derive(Debug)]
pub struct SyntaxTree {
    pub lines: Vec<Line>,
}

impl SyntaxTree {
    pub fn create(&mut self, contract_contents: &Vec<u8>) {
        let mut iter = contract_contents.iter();
        while iter.len() > 0 {
            let mut line = build_line();
            line.command = *iter.next().unwrap();
            if line.command == b'\n' {
                continue;
            }
            loop {
                let len = *iter.next().unwrap();
                if len == b'\xFF' {
                    break;
                }
                let mut arg = vec![];
                for _ in 0..len {
                    arg.push(*iter.next().unwrap());
                }
                line.args.push(arg);
            }
            self.lines.push(line);
        }
    }
}

pub fn build_syntax_tree() -> SyntaxTree {
    SyntaxTree { lines: Vec::new() }
}
