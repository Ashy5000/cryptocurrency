//
// Created by ashy5000 on 6/5/24.
//

#include "BlockasmGenerator.h"

#include <cstdlib>
#include <iomanip>
#include <iostream>
#include <random>
#include <sstream>
#include <tuple>
#include <utility>

#include "ControlModule.hpp"
#include "ExpressionBlockasmGenerator.h"
#include "Linker.h"
#include "RegisteredFunctionInfo.h"
#include "SystemFunctions.h"
#include "Variable.h"

BlockasmGenerator::BlockasmGenerator(std::vector<Token> tokens_p, const int nextAllocatedLocation_p, std::vector<Variable> vars_p, const bool useLinker_p, const std::default_random_engine rnd_p) {
    tokens = std::move(tokens_p);
    nextAllocatedLocation = nextAllocatedLocation_p;
    nextAllocatedStateLocation = 0;
    blockasm = {};
    useLinker = useLinker_p;
    rnd = rnd_p;
    if(useLinker) {
        blockasm << ";^^^^BEGIN_SOURCE^^^^" << std::endl;
    }
    vars = std::move(vars_p);
}


std::string BlockasmGenerator::GenerateBlockasm(ControlModule &cm) {
    int nextLabel = 0;
    auto l = Linker({"string.blockasm", "format.blockasm"});
    for(int i = 0; i < tokens.size(); i++) {
        if(const Token token = tokens[i]; token.type == TokenType::system_at) {
            std::tuple tuple = GenerateSystemFunctionBlockasm(i, nextAllocatedLocation, vars, l);
            std::vector<Variable> newVars = std::get<0>(tuple);
            vars.insert(vars.end(), newVars.begin(), newVars.end());
            const int tokensConsumed = std::get<1>(tuple);
            i += tokensConsumed;
        } else if(token.type == TokenType::semi) {
            if(tokens[i + 1].type == TokenType::concat && tokens[i + 3].type == TokenType::sub && tokens[i + 4].type == TokenType::greater) {
                std::string functionName = tokens[i + 2].value;
                std::string returnTypeString = tokens[i + 5].value;
                std::vector<std::string> params;
                for(int j = i + 1; j < tokens.size(); j++) {
                    if(tokens[j].type == TokenType::newline) {
                        break;
                    }
                    if(tokens[j].type == TokenType::div) {
                        params.push_back(tokens[j + 1].value);
                        j++;
                    }
                }
                std::stringstream metaString;
                metaString << "; ";
                metaString << "FN " << functionName << " ";
                metaString << "RET " << returnTypeString << " ";
                for(const std::string& param : params) {
                    metaString << "PARAM " << param << " ";
                }
                blockasm << metaString.str() << std::endl;
                auto [id, preLabelId] = cm.registerFunction(rnd);
                blockasm << "; PRELABEL " << preLabelId << std::endl;
            }
        } else if(token.type == TokenType::identifier) {
            if(token.value == "load") {
                if(tokens[i + 1].type == TokenType::concat) {
                    if(tokens[i + 2].type == TokenType::identifier) {
                        std::string varName = tokens[i + 2].value;
                        if(tokens[i + 2].value.at(0) != '\'') {
                            std::cerr << "Non-state variables cannot be loaded" << std::endl;
                            exit(EXIT_FAILURE);
                        }
                        auto var = Variable(varName, nextAllocatedStateLocation++, Type::loaded);
                        vars.push_back(var);
                    }
                }
            }
            if(tokens[i + 1].type == TokenType::eq) {
                if(tokens[i + 2].type == TokenType::eq) {
                    // e.g. newVar == 3
                    std::string varName = token.value;
                    std::tuple exprTuple = ExpressionBlockasmGenerator::GenerateBlockasmFromExpression(tokens[i + 3], nextAllocatedLocation, vars, blockasm, l);
                    int exprLoc = std::get<0>(exprTuple);
                    if(exprLoc >= nextAllocatedLocation) {
                        nextAllocatedLocation = exprLoc + 1;
                    }
                    Type type = std::get<1>(exprTuple);
                    if(varName.at(0) == '\'') {
                        int location = nextAllocatedStateLocation;
                        bool newVar = true;
                        for(const Variable& var : vars) {
                            if(var.name == varName) {
                                location = var.location;
                                newVar = false;
                            }
                        }
                        blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation << " 0x00000000" << std::endl;
                        blockasm << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation << " 0x";
                        blockasm << std::setfill('0') << std::setw(14) << std::hex << location << " 0x00000000" << std::endl;
                        nextAllocatedLocation++;
                        blockasm << "UpdateState 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation - 1 << " 0x";
                        blockasm << std::setfill('0') << std::setw(8) << std::hex << exprLoc << " 0x00000000" << std::endl;
                        if(newVar) {
                            auto var = Variable(varName, location, type);
                            vars.emplace_back(var);
                            nextAllocatedStateLocation++;
                        }
                    } else {
                        auto var = Variable(varName, exprLoc, type);
                        nextAllocatedLocation++;
                        vars.emplace_back(var);
                        i += 6;
                    }
                } else {
                    // e.g. existingVar = 5
                    std::string varName = token.value;
                    std::tuple exprTuple = ExpressionBlockasmGenerator::GenerateBlockasmFromExpression(tokens[i + 2], nextAllocatedLocation, vars, blockasm, l);
                    int exprLoc = std::get<0>(exprTuple);
                    if(exprLoc >= nextAllocatedLocation) {
                        nextAllocatedLocation = exprLoc + 1;
                    }
                    Type type = std::get<1>(exprTuple);
                    bool varFound = false;
                    for(Variable &var : vars) {
                        if(var.name == varName) {
                            if(var.type != type && var.type != Type::loaded) {
                                std::cerr << "Expression type does not match variable type." << std::endl;
                                exit(EXIT_FAILURE);
                            }
                            blockasm << "Free 0x" << std::setfill('0') << std::setw(8) << std::hex << var.location << " 0x00000000" << std::endl;
                            var.location = exprLoc;
                            varFound = true;
                            break;
                        }
                    }
                    if(!varFound) {
                        std::cerr << "Unknown variable " << varName << std::endl;
                        exit(EXIT_FAILURE);
                    }
                    i += 5;
                }
            } else if(token.value == "if") {
                std::tuple exprTuple = ExpressionBlockasmGenerator::GenerateBlockasmFromExpression(tokens[i + 1], nextAllocatedLocation, vars, blockasm, l);
                int exprLoc = std::get<0>(exprTuple);
                if(exprLoc >= nextAllocatedLocation) {
                    nextAllocatedLocation = exprLoc + 1;
                }
                if(Type type = std::get<1>(exprTuple); type != Type::boolean) {
                    std::cerr << "Expected bool in if statement";
                    exit(EXIT_FAILURE);
                }
                blockasm << "Not 0x" << std::setfill('0') << std::setw(8) << std::hex << exprLoc << " 0x";
                blockasm << std::setfill('0') << std::setw(8) << std::hex << exprLoc << " 0x00000000" << std::endl;
                blockasm << "JmpCond 0x" << std::setfill('0') << std::setw(8) << std::hex << exprLoc << " ";
                blockasm << "<" << nextLabel << " 0x00000000" << std::endl;
                auto subGenerator = BlockasmGenerator(tokens[i + 3].children, nextAllocatedLocation, vars, false, rnd);
                blockasm << subGenerator.GenerateBlockasm(cm);
                if(int subGeneratorNextAllocatedLocation = subGenerator.GetNextAllocatedLocation(); subGeneratorNextAllocatedLocation > nextAllocatedLocation) {
                    nextAllocatedLocation = subGeneratorNextAllocatedLocation + 1;
                }
                blockasm << "; LABEL " << nextLabel++ << std::endl;
          } else if(token.value == "for") {
              // for(i (0) (100)) {}
              // for(IDENTIFIER (EXPR) (EXPR)) {BLOCK}
              std::string varName = tokens[i + 2].children[0].value; // i: IDENTIFIER
              int varLoc = -1;
              for(const Variable& var : vars) {
                 if(var.name == varName) {
                     varLoc = var.location;
                     break;
                 }
              }
              if(varLoc == -1) {
                  std::cerr << "Unknown variable " << varName << std::endl;
                  exit(EXIT_FAILURE);
              }
              Token beginExprToken = tokens[i + 2].children[2]; // 0: EXPR
              std::tuple beginExprTuple = ExpressionBlockasmGenerator::GenerateBlockasmFromExpression(beginExprToken, nextAllocatedLocation, vars, blockasm, l);
              if(int beginExprLoc = std::get<0>(beginExprTuple); beginExprLoc >= nextAllocatedLocation) {
                  nextAllocatedLocation = beginExprLoc + 1;
              }
              if(Type beginExprType = std::get<1>(beginExprTuple); beginExprType != Type::uint64) {
                  std::cerr << "Begin expression of for loop has incorrect type" << std::endl;
                  exit(EXIT_FAILURE);
              }
              Token endExprToken = tokens[i + 2].children[5]; // 100: EXPR
              std::tuple endExprTuple = ExpressionBlockasmGenerator::GenerateBlockasmFromExpression(endExprToken, nextAllocatedLocation, vars, blockasm, l);
              int endExprLoc = std::get<0>(endExprTuple);
              if(endExprLoc >= nextAllocatedLocation) {
                  nextAllocatedLocation = endExprLoc + 1;
              }
              if(Type endExprType = std::get<1>(endExprTuple); endExprType != Type::uint64) {
                std::cerr << "End expression of for loop has incorrect type" << std::endl;
                exit(EXIT_FAILURE);
              }
              int labelId = nextLabel++;
              int oneLoc = nextAllocatedLocation++;
              blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << oneLoc << " 0x00000000" << std::endl;
              blockasm << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << oneLoc << " 0x0000000000000001 0x00000000" << std::endl;
              int cmpLoc = nextAllocatedLocation++;
              blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << cmpLoc << " 0x00000000" << std::endl;
              blockasm << "; LABEL " << labelId << std::endl;
              blockasm << "Add 0x" << std::setfill('0') << std::setw(8) << std::hex << varLoc << " 0x";
              blockasm << std::setfill('0') << std::setw(8) << std::hex << oneLoc << " 0x";
              blockasm << std::setfill('0') << std::setw(8) << std::hex << varLoc << " 0x00000000" << std::endl;
              Token blockToken = tokens[i + 5];
              auto subGenerator = BlockasmGenerator(blockToken.children, nextAllocatedLocation, vars, false, rnd);
              blockasm << subGenerator.GenerateBlockasm(cm);
              if(int subGeneratorNextAllocatedLocation = subGenerator.nextAllocatedLocation; subGeneratorNextAllocatedLocation > nextAllocatedLocation) {
                  nextAllocatedLocation = subGeneratorNextAllocatedLocation + 1;
              }
              blockasm << "Eq 0x" << std::setfill('0') << std::setw(8) << std::hex << varLoc << " 0x";
              blockasm << std::setfill('0') << std::setw(8) << std::hex << endExprLoc << " 0x";
              blockasm << std::setfill('0') << std::setw(8) << std::hex << cmpLoc << " 0x00000000" << std::endl;
              blockasm << "Not 0x" << std::setfill('0') << std::setw(8) << std::hex << cmpLoc << " 0x";
              blockasm << std::setfill('0') << std::setw(8) << std::hex << cmpLoc << " 0x00000000" << std::endl;
              blockasm << "JmpCond 0x" << std::setfill('0') << std::setw(8) << std::hex << cmpLoc << " ";
              blockasm << "<" << labelId << " 0x00000000" << std::endl;
        }
        } else if(token.type == TokenType::div) {
            if(tokens[i + 1].type == TokenType::identifier) {
                if(!useLinker) {
                    std::cerr << "Imports are only allowed in root of file" << std::endl;
                    exit(EXIT_FAILURE);
                }
                l.InjectIfNotPresent(tokens[i + 1].value, blockasm);
            }
        } else if(token.type == TokenType::excl) {
            if(tokens[i + 1].type == TokenType::identifier) {
                std::string functionName = tokens[i + 1].value;
                if(tokens[i + 2].type != TokenType::open_paren) {
                    std::cerr << "Expected '('" << std::endl;
                    exit(EXIT_FAILURE);
                }
                std::vector<Token> params;
                for(int j = i + 3; j < tokens.size(); j++) {
                    if(tokens[j].type == TokenType::expr) {
                        params.emplace_back(tokens[j]);
                    }
                    if(tokens[j].type == TokenType::newline) {
                        break;
                    }
                }
                std::vector<int> paramLocs;
                for(const Token& param : params) {
                    std::tuple exprTuple = ExpressionBlockasmGenerator::GenerateBlockasmFromExpression(param, nextAllocatedLocation, vars, blockasm, l);
                    int exprLoc = std::get<0>(exprTuple);
                    if(exprLoc >= nextAllocatedLocation) {
                        nextAllocatedLocation = exprLoc + 1;
                    }
                    paramLocs.emplace_back(exprLoc);
                }
                std::tuple functionCallTuple = l.CallFunction(functionName, paramLocs, vars);
                std::string functionCallBlockasm = std::get<0>(functionCallTuple);
                blockasm << functionCallBlockasm;
            }
        }
    }
    if(useLinker) {
        Linker::SkipLibs(blockasm);
    }
    std::string blockasmStr = blockasm.str();
    return blockasmStr;
}

std::tuple<std::vector<Variable>, int> BlockasmGenerator::GenerateSystemFunctionBlockasm(const int i, int &nextAllocatedLocation, std::vector<Variable> vars, Linker l) {
    Token identifier = tokens[i + 1];
    if(identifier.type != TokenType::identifier) {
        std::cerr << "System at (@) must be followed by an identifier." << std::endl;
        exit(EXIT_FAILURE);
    }
    std::vector<Token> params;
    std::vector<Token> currentExprTokens;
    for(int j = 0; j < tokens[i + 2].children.size(); j++) {
        Token t = tokens[i + 2].children[j];
        if(t.type == TokenType::comma) {
            auto expr = Token(TokenType::expr, {});
            for(const Token& exprT : currentExprTokens) {
                expr.children.emplace_back(exprT);
            }
            params.emplace_back(expr);
            currentExprTokens.clear();
            continue;
        }
        currentExprTokens.emplace_back(t);
        if(j == tokens[i + 2].children.size() - 1) {
            auto expr = Token(TokenType::expr, {});
            for(const Token& exprT : currentExprTokens) {
                expr.children.emplace_back(exprT);
            }
            params.emplace_back(expr);
            currentExprTokens.clear();
            break;
        }
    }
    if(Token semiToken = tokens[i + 3]; semiToken.type != TokenType::semi) {
        std::cerr << "Expected semicolon." << std::endl;
        exit(EXIT_FAILURE);
    }
    if(Token newlineToken = tokens[i + 4]; newlineToken.type != TokenType::newline) {
        std::cerr << "Unexpected token after semicolon." << std::endl;
        exit(EXIT_FAILURE);
    }
    std::string delimiter = "::";
    auto delimiterPos = identifier.value.find(delimiter);
    if(delimiterPos == std::string::npos) {
        std::cerr << "Invalid system function format." << std::endl;
        exit(EXIT_FAILURE);
    }
    std::string module = identifier.value.substr(0, delimiterPos);
    std::string function = identifier.value.substr(delimiterPos + 2);
    for(const SystemFunction& func : SYSTEM_FUNCTIONS) {
        if(func.module == module && func.name == function) {
            func.generateBlockasm(params, nextAllocatedLocation, vars, blockasm, l);
            return std::make_tuple(vars, 3);
        }
    }
    std::cerr << "Unknown module." << std::endl;
    exit(EXIT_FAILURE);
}

int BlockasmGenerator::GetNextAllocatedLocation() const {
    return nextAllocatedLocation;
}
