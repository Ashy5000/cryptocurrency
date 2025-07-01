pub(crate) fn to_u64(data: &Vec<u8>) -> Result<u64, &'static str> {
    let mut result: u64 = 0;
    let mut shift_amount = (data.len() * 8) - 8;
    for piece in data {
        result += u64::from(*piece) << shift_amount;
        if shift_amount == 0 {
            break;
        }
        shift_amount -= 8;
    }
    Ok(result)
}

pub(crate) fn to_vec(data: u64) -> Result<Vec<u8>, &'static str> {
    let mut res = Vec::new();
    let mut shift_amount = 64 - 8;
    while shift_amount >= 0 {
        let mut piece_u64 = data;
        piece_u64 >>= shift_amount;
        piece_u64 %= 256;
        let piece = piece_u64 as u8;
        shift_amount -= 8;
        res.push(piece);
    }
    Ok(res)
}