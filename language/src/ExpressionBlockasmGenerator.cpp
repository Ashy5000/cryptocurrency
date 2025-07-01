//
// Created by ashy5000 on 6/5/24.
//

#include "ExpressionBlockasmGenerator.h"

#include <iomanip>
#include <iostream>
#include <iterator>
#include <sstream>
#include <tuple>

#include "Operator.h"
#include "OperatorType.h"
#include "Variable.h"

bool ExpressionBlockasmGenerator::IsAcceptable(TokenType type) {
    switch (type) {
        case TokenType::type_placeholder:
        case TokenType::system_at:
        case TokenType::open_paren:
        case TokenType::close_paren:
        case TokenType::semi:
        case TokenType::newline:
        case TokenType::concat:
        case TokenType::add:
        case TokenType::sub:
        case TokenType::mul:
        case TokenType::exp:
        case TokenType::comma:
        case TokenType::eq:
        case TokenType::excl:
        case TokenType::open_curly:
        case TokenType::close_curly:
        case TokenType::greater:
        case TokenType::block:
        case TokenType::div:
            return false;
        case TokenType::identifier:
        case TokenType::expr:
        case TokenType::int_lit:
        case TokenType::string_lit:
            return true;
        default:
            return false;
    };
}


std::tuple<int, Type> ExpressionBlockasmGenerator::GenerateBlockasmFromExpression(Token expression, int nextAllocatedLocation, std::vector<Variable>& vars, std::stringstream &blockasm, Linker &l) {
    if(!IsAcceptable(expression.type)) {
        std::cerr << "Expected valid expression when generating Blockasm." << std::endl;
        exit(EXIT_FAILURE);
    }
    if(expression.children.empty()) {
        Token newExpression(TokenType::expr, "");
        newExpression.children.push_back(expression);
        expression = newExpression;
    }
    if(expression.children[0].type == TokenType::open_paren) {
        auto first = expression.children.begin() + 1;
        auto last = expression.children.end() - 1;
        expression.children = std::vector(first, last);
    }
    if(expression.children.size() == 1) {
        if (expression.children[0].type == TokenType::int_lit) {
            blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation
                     << " 0x00000000" << std::endl;
            int val = stoi(expression.children[0].value);
            blockasm << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation << " 0x";
            blockasm << std::setfill('0') << std::setw(14) << std::hex << val << " 0x00000000" << std::endl;
            return std::make_tuple(nextAllocatedLocation, Type::uint64);
        }
        if (expression.children[0].type == TokenType::string_lit) {
            std::stringstream buffer(expression.children[0].value);
            blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation
                     << " 0x00000000" << std::endl;
            blockasm << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation << " 0x";
            std::istreambuf_iterator it(buffer.rdbuf());
            std::istreambuf_iterator<char> end;
            std::stringstream out;
            out << std::hex;
            std::copy(it, end, std::ostream_iterator<int>(out));
            blockasm << out.str();
            blockasm << " 0x00000000" << std::endl;
            return std::make_tuple(nextAllocatedLocation, Type::string);
        }
        if (expression.children[0].type == TokenType::identifier) {
            if (expression.children[0].value.at(0) == '\'') {
                int location = -1;
                auto type = Type::type_placeholder;
                for (const Variable &var: vars) {
                    if (var.name == expression.children[0].value) {
                        location = var.location;
                        type = var.type;
                        break;
                    }
                }
                if (location == -1) {
                    std::cerr << "Undefined variable " << expression.children[0].value << "." << std::endl;
                    exit(EXIT_FAILURE);
                }
                int locationLoc = nextAllocatedLocation++;
                blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << locationLoc << " 0x00000000" << std::endl;
                blockasm << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << locationLoc << " 0x";
                blockasm << std::setfill('0') << std::setw(14) << std::hex << location << " 0x00000000" << std::endl;
                int resultLoc = nextAllocatedLocation;
                blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << resultLoc << " 0x00000000" << std::endl;
                blockasm << "GetFromState 0x" << std::setfill('0') << std::setw(8) << std::hex << locationLoc << " 0x";
                blockasm << std::setfill('0') << std::setw(8) << std::hex << resultLoc << " 0x00000000" << std::endl;
                return std::make_tuple(resultLoc, type);
            }
            auto referencedVar = Variable("", 0, Type::type_placeholder);
            for (const Variable &var: vars) {
                if (var.name == expression.children[0].value) {
                    referencedVar = var;
                    break;
                }
            }
            return std::make_tuple(referencedVar.location, referencedVar.type);
        }
        if(IsAcceptable(expression.children[0].type)) {
            std::tuple exprTuple = GenerateBlockasmFromExpression(expression.children[0], nextAllocatedLocation, vars, blockasm, l);
            return exprTuple;
        }
        std::cerr << "Unknown expression." << std::endl;
        exit(EXIT_FAILURE);
    }
    if(expression.children[0].type == TokenType::excl) {
        auto tokens = expression.children;
        if(tokens[1].type != TokenType::identifier) {
            std::cerr << "Expected identifier." << std::endl;
            exit(EXIT_FAILURE);
        }
        if(tokens[2].type != TokenType::open_paren) {
            std::cerr << "Expected '('" << std::endl;
            exit(EXIT_FAILURE);
        }
        std::vector<Token> params;
        std::string functionName = tokens[1].value;
        for(auto & token : tokens[3].children) {
            if(token.type == TokenType::expr) {
                params.emplace_back(token);
            }
            if(token.type == TokenType::newline) {
                break;
            }
        }
        std::vector<int> paramLocs;
        for(const Token& param : params) {
            std::tuple exprTuple = GenerateBlockasmFromExpression(param, nextAllocatedLocation, vars, blockasm, l);
            int exprLoc = std::get<0>(exprTuple);
            if(exprLoc >= nextAllocatedLocation) {
                nextAllocatedLocation = exprLoc + 1;
            }
            paramLocs.emplace_back(exprLoc);
        }
        std::tuple functionCallTuple = l.CallFunction(functionName, paramLocs, vars);
        std::string functionCallBlockasm = std::get<0>(functionCallTuple);
        blockasm << functionCallBlockasm;
        Type t = std::get<1>(functionCallTuple);
        return std::make_tuple(0x00000001, t);
    }
    auto type = OperatorType::type_placeholder;
    int operatorPos = 0;
    for(;operatorPos < expression.children.size(); operatorPos++) {
        Token t = expression.children[operatorPos];
        type = OperatorTypeFromToken(t);
        if(type != OperatorType::type_placeholder) {
            break;
        }
    }
    std::vector preOperatorTokens(expression.children.begin(), expression.children.begin() + operatorPos);
    std::vector postOperatorTokens(expression.children.begin() + operatorPos + 1, expression.children.end());
    auto preOperatorExpr = Token({TokenType::expr, {}});
    preOperatorExpr.children = preOperatorTokens;
    auto postOperatorExpr = Token({TokenType::expr, {}});
    postOperatorExpr.children = postOperatorTokens;
    std::tuple exprATuple = GenerateBlockasmFromExpression(preOperatorExpr, nextAllocatedLocation, vars, blockasm, l);
    int exprALoc = std::get<0>(exprATuple);
    if(exprALoc >= nextAllocatedLocation) {
        nextAllocatedLocation = exprALoc + 1;
    }
    std::tuple exprBTuple = GenerateBlockasmFromExpression(postOperatorExpr, nextAllocatedLocation, vars, blockasm, l);
    int exprBLoc = std::get<0>(exprBTuple);
    if(exprALoc >= nextAllocatedLocation) {
        nextAllocatedLocation = exprALoc + 1;
    }
    blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation + 1 << " 0x00000000" << std::endl;
    std::string operatorString = OperatorToString(Operator{type});
    blockasm << operatorString << " 0x" << std::setfill('0') << std::setw(8) << std::hex << exprALoc << " 0x";
    blockasm << std::setfill('0') << std::setw(8) << std::hex << exprBLoc << " 0x";
    blockasm << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation + 1 << " 0x00000000" << std::endl;
    Type returnType = OperatorToType(Operator{type});
    return std::make_tuple(nextAllocatedLocation + 1, returnType);
}
