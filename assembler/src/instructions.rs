pub(crate) struct Instruction {
    pub(crate) name: &'static str,
    pub(crate) opcode: u8
}

pub(crate) const INSTRUCTIONS: [Instruction; 41] = [
    Instruction{
        name: "ExitBfr",
        opcode: b'\x01'
    },
    Instruction{
        name: "Exit",
        opcode: b'\x00'
    },
    Instruction{
        name: "InitBfr",
        opcode: b'\x02'
    },
    Instruction{
        name: "CpyBfr",
        opcode: b'\x03'
    },
    Instruction{
        name: "FreeBfr",
        opcode: b'\x04'
    },
    Instruction{
        name: "BfrStat",
        opcode: b'\x05'
    },
    Instruction{
        name: "BfrLen",
        opcode: b'\x06'
    },
    Instruction {
        name: "Add",
        opcode: b'\x07'
    },
    Instruction {
        name: "Sub",
        opcode: b'\x08'
    },
    Instruction {
        name: "Mul",
        opcode: b'\x09'
    },
    Instruction{
        name: "Div",
        opcode: b'\x0A'
    },
    Instruction{
        name: "Exp",
        opcode: b'\x0B'
    },
    Instruction{
        name: "Mod",
        opcode: b'\x0C'
    },
    Instruction{
        name: "Eq",
        opcode: b'\x0D'
    },
    Instruction{
        name: "Less",
        opcode: b'\x0E'
    },
    Instruction{
        name: "And",
        opcode: b'\x0F'
    },
    Instruction{
        name: "Or",
        opcode: b'\x10'
    },
    Instruction{
        name: "Not",
        opcode: b'\x11'
    },
    Instruction{
        name: "App",
        opcode: b'\x12'
    },
    Instruction{
        name: "Slice",
        opcode: b'\x13'
    },
    Instruction{
        name: "Shiftl",
        opcode: b'\x14'
    },
    Instruction{
        name: "Shiftr",
        opcode: b'\x15'
    },
    Instruction{
        name: "JmpCond",
        opcode: b'\x17'
    },
    Instruction{
        name: "Jmp",
        opcode: b'\x16'
    },
    Instruction{
        name: "Call",
        opcode: b'\x18'
    },
    Instruction{
        name: "Ret",
        opcode: b'\x19'
    },
    Instruction{
        name: "Stdout",
        opcode: b'\x1A'
    },
    Instruction{
        name: "PrintStr",
        opcode: b'\x1B'
    },
    Instruction{
        name: "Stderr",
        opcode: b'\x1C'
    },
    Instruction{
        name: "SetCnst",
        opcode: b'\x1D'
    },
    Instruction{
        name: "Tx",
        opcode: b'\x1E'
    },
    Instruction{
        name: "ChainLen",
        opcode: b'\x1F'
    },
    Instruction{
        name: "UpdateStateExternal",
        opcode: b'\x21'
    },
    Instruction{
        name: "UpdateState",
        opcode: b'\x20'
    },
    Instruction{
        name: "GetFromStateExternalSync",
        opcode: b'\x25'
    },
    Instruction{
        name: "GetFromStateExternal",
        opcode: b'\x23'
    },
    Instruction{
        name: "GetFromStateSync",
        opcode: b'\x24'
    },
    Instruction{
        name: "GetFromState",
        opcode: b'\x22'
    },
    Instruction{
        name: "QueryOracle",
        opcode: b'\x26'
    },
    Instruction{
        name: "Invoke",
        opcode: b'\x27'
    },
    Instruction{
        name: "GetSender",
        opcode: b'\x28'
    }
];