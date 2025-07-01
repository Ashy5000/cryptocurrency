//
// Created by ashy5000 on 6/13/24.
//

#ifndef SYSTEMFUNCTIONS_H
#define SYSTEMFUNCTIONS_H
#include <iomanip>
#include <iterator>
#include <sstream>

#include "ParamsParser.h"
#include "Signature.h"
#include "SystemFunction.h"
#include "Linker.h"

const std::vector SYSTEM_FUNCTIONS = {
    SystemFunction(
        [](const std::vector<Token>& params, int &nextAllocatedLocation, std::vector<Variable>& vars, std::stringstream &blockasm, Linker &l) {
            auto sig = Signature({Type::uint64}, Type::type_placeholder);
            auto pp = ParamsParser(params, {sig});
            const std::tuple<std::vector<int>, Signature> parsingResult = pp.ParseParams(nextAllocatedLocation, vars, blockasm, l);
            const std::vector<int> locations = std::get<0>(parsingResult);
            const int exitCodeLocation = locations[0];
            blockasm << "CpyBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << exitCodeLocation << " 0x~ 0x00000000" << std::endl;
            blockasm << "Jmp `" << std::endl;
            if(exitCodeLocation >= nextAllocatedLocation) {
                nextAllocatedLocation = exitCodeLocation + 1;
            }
        },
        "contract",
        "exit"
    ),
    SystemFunction(
        [](const std::vector<Token>& params, int &nextAllocatedLocation, std::vector<Variable>& vars, std::stringstream &blockasm, Linker &l) {
            if(params[0].type != TokenType::expr) {
                std::cerr << "Can't return a non-expression" << std::endl;
                exit(EXIT_FAILURE);
            }
            const std::tuple exprTuple = ExpressionBlockasmGenerator::GenerateBlockasmFromExpression(params[0], nextAllocatedLocation, vars, blockasm, l);
            const int exprLoc = std::get<0>(exprTuple);
            if(exprLoc >= nextAllocatedLocation) {
                nextAllocatedLocation = exprLoc + 1;
            }
            int locLoc = nextAllocatedLocation++;
            int valLoc = nextAllocatedLocation++;
            blockasm << "InitBfr 0x" << std::setw(8) << std::hex << locLoc << " 0x00000000" << std::endl;
            blockasm << "SetCnst 0x" << std::setw(8) << std::hex << locLoc << " 0x0000000001fff0 0x00000000" << std::endl;
            blockasm << "InitBfr 0x" << std::setw(8) << std::hex << valLoc << " 0x00000000" << std::endl;
            blockasm << "UpdateState 0x" << std::setfill('0') << std::setw(8) << std::hex << locLoc << " 0x";
            blockasm << std::setw(8) << std::hex << exprLoc << " 0x00000000" << std::endl;
        },
        "contract",
        "return"
    ),
    SystemFunction(
            [](const std::vector<Token>& params, int &nextAllocatedLocation, std::vector<Variable>& vars, std::stringstream &blockasm, Linker &) {
            const std::string typeString = params[1].children[0].value;
            auto type = Type::type_placeholder;
            if(typeString == "uint64") {
                type = Type::uint64;
            }
            if(type == Type::type_placeholder) {
                std::cerr << "Unknown type " << typeString << std::endl;
                exit(EXIT_FAILURE);
            }
            auto var = Variable(params[0].children[0].value, nextAllocatedLocation, type);
            vars.emplace_back(var);
            blockasm << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << nextAllocatedLocation++ << " 0x00000000" << std::endl;
        },
        "memory",
        "alloc"
    ),
    SystemFunction(
            [](const std::vector<Token>& params, int &, std::vector<Variable>& vars, std::stringstream &blockasm, Linker &) {
            int indexToRemove = -1;
            for(int j = 0; j < vars.size(); j++) {
                if(const Variable& var = vars[j]; var.name == params[0].children[0].value) {
                    indexToRemove = j;
                    break;
                }
            }
            if(indexToRemove == -1) {
                std::cerr << "Cannot free undefined variable." << std::endl;
                exit(EXIT_FAILURE);
            }
            blockasm << "Free 0x" << std::setfill('0') << std::setw(8) << std::hex << vars[indexToRemove].location << " 0x00000000" << std::endl;
            vars.erase(vars.begin() + indexToRemove);
        },
        "memory",
        "free"
    ),
    SystemFunction(
            [](const std::vector<Token>& params, int &, const std::vector<Variable>& vars, std::stringstream &blockasm, Linker &) {
            int indexToRename = -1;
            for(int j = 0; j < vars.size(); j++) {
                if(const Variable& var = vars[j]; var.name == params[0].children[0].value) {
                    indexToRename = j;
                    break;
                }
            }
            if(indexToRename == -1) {
                std::cerr << "Cannot set undefined variable." << std::endl;
                exit(EXIT_FAILURE);
            }
            char* end;
            const int val = static_cast<int>(std::strtol(params[1].children[0].value.c_str(), &end, 10));
            if(errno == ERANGE) {
                std::cerr << "Expected integer as value" << std::endl;
            }
            blockasm << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << vars[indexToRename].location << " 0x";
            blockasm << std::setfill('0') << std::setw(8) << std::hex << val << " 0x00000000" << std::endl;
        },
        "memory",
        "set"
    ),
    SystemFunction(
            [](const std::vector<Token>& params, int &nextAllocatedLocation, std::vector<Variable>& vars, std::stringstream &blockasm, Linker &l) {
            auto uint64Sig = Signature({Type::uint64}, Type::type_placeholder);
            auto stringSig = Signature({Type::string}, Type::type_placeholder);
            auto pp = ParamsParser(params, {uint64Sig, stringSig});
            const std::tuple<std::vector<int>, Signature> parsingResult = pp.ParseParams(nextAllocatedLocation, vars, blockasm, l);
            const std::vector<int> locations = std::get<0>(parsingResult);
            const Signature sig = std::get<1>(parsingResult);
            const int dataLocation = locations[0];
            if(sig.expectedTypes[0] == Type::uint64) {
                blockasm << "Stdout 0x" << std::setfill('0') << std::setw(8) << std::hex << dataLocation << " 0x00000000" << std::endl;
            } else {
                blockasm << "PrintStr 0x" << std::setfill('0') << std::setw(8) << std::hex << dataLocation << " 0x00000000" << std::endl;
            }
        },
        "io",
        "print"
    ),
    SystemFunction(
            [](const std::vector<Token>& params, int &nextAllocatedLocation, std::vector<Variable>& vars, std::stringstream &blockasm, Linker &l) {
            auto sig = Signature({Type::uint64}, Type::type_placeholder);
            auto pp = ParamsParser(params, {sig});
            const std::tuple<std::vector<int>, Signature> parsingResult = pp.ParseParams(nextAllocatedLocation, vars, blockasm, l);
            const std::vector<int> locations = std::get<0>(parsingResult);
            const int dataLocation = locations[0];
            blockasm << "Stderr 0x" << std::setfill('0') << std::setw(8) << std::hex << dataLocation << " 0x00000000" << std::endl;
        },
        "io",
        "err"
    )
};

#endif //SYSTEMFUNCTIONS_H
