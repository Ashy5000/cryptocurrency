//
// Created by ashy5000 on 6/4/24.
//

#ifndef PARSER_H
#define PARSER_H
#include <vector>

#include "Token.h"


class Parser {
public:
    std::vector<Token> parse_tokens(std::string input);
};



#endif //PARSER_H