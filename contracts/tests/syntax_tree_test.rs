extern crate contracts;

#[cfg(test)]
mod syntax_tree_test {
    #[test]
    fn test_syntax_tree() {
        let tree = contracts::syntax_tree::SyntaxTree {
            lines: vec![contracts::syntax_tree::Line {
                command: "InitBfr".parse().unwrap(),
                args: vec![vec![0, 0, 0, 1], vec![0, 0, 0, 0]],
            }],
        };
        assert_eq!(tree.lines[0].command, "InitBfr".to_owned());
        assert_eq!(tree.lines[0].args[0], "00000001".to_owned());
        assert_eq!(tree.lines[0].args[1], "00000000".to_owned());
    }
    #[test]
    fn test_syntax_tree_parse() {
        let mut tree = contracts::syntax_tree::build_syntax_tree();
        let asm = "InitBfr 0x00000001 0x00000000 ; Initialize a buffer".to_owned();
        tree.create(asm.into());
        assert_eq!(tree.lines.len(), 1);
        assert_eq!(tree.lines[0].args[0], "00000001".to_owned());
        assert_eq!(tree.lines[0].args[1], "00000000".to_owned());
    }
}
