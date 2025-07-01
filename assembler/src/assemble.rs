use std::str::FromStr;
use bstr::{BString, ByteVec};
use hex::FromHexError;
use crate::instructions::INSTRUCTIONS;

fn decode_int_string(str: &str) -> Result<Vec<u8>, FromHexError> {
    if str.starts_with("0x") {
        hex::decode(&str[2..])
    } else {
        let uint = u64::from_str_radix(str, 10).unwrap();
        let mut res = Vec::new();
        let mut shift_amount = 64 - 8;
        while shift_amount >= 0 {
            let mut piece_u64 = uint;
            piece_u64 >>= shift_amount;
            piece_u64 %= 256;
            let piece = piece_u64 as u8;
            shift_amount -= 8;
            res.push(piece);
        }
        Ok(res)
    }
}


pub(crate) fn assemble(source: String) -> BString {
    let mut res = BString::new(vec![]);
    let lines: Vec<String> = source.lines().map(|x| String::from_str(x).unwrap()).collect();
    for line in lines {
        let mut words: Vec<String> = line.split(" ").map(|x| String::from_str(x).unwrap()).collect();
        res.push_byte(match INSTRUCTIONS.iter().find(|x| x.name == words[0]) {
            None => { res.push_byte(b'\xFF'); res.push_byte(b'\xFF'); continue; }
            Some(x) => {x}
        }.opcode);
        words.remove(0);
        for word in words {
            let raw = decode_int_string(&*word).unwrap();
            let len = raw.len() as u8;
            res.push_byte(len);
            res.push_str(raw.as_slice());
        }
        res.push_byte(b'\xFF');
    }
    res
}