//
// Created by ashy5000 on 10/6/24.
//

#include "ControlModule.hpp"

#include <iomanip>
#include <iostream>
#include <sstream>

RegisteredFunctionInfo ControlModule::registerFunction(std::default_random_engine rnd) {
  const int random_int = static_cast<int>(rnd());
  RegisteredFunctionInfo info {
    random_int,
    static_cast<int>(registeredFunctionInfos.size())
  };
  info.id += info.preLabelId;
  registeredFunctionInfos.emplace_back(info);
  return info;
}

std::string ControlModule::compile(int &nextAllocatedLocation) {
  std::stringstream result;
  // Create buffer for return value
  result << "InitBfr 0x~ 0x00000000" << std::endl;
  // Set selector location
  selectorLocLocation = nextAllocatedLocation++;
  result << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << selectorLocLocation << " 0x00000000" << std::endl;
  result << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << selectorLocLocation << " 0x0000000001000000 0x00000000" << std::endl;
  // Get selected function ID
  const int selectorLocation = nextAllocatedLocation++;
  result << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << selectorLocation << " 0x00000000" << std::endl;
  result << "GetFromState 0x" << std::setfill('0') << std::setw(8) << std::hex << selectorLocLocation << " 0x";
  result << std::setfill('0') << std::setw(8) << std::hex << selectorLocation << " 0x00000000" << std::endl;
  // Create buffer to hold comparison result
  const int cmpResLocation = nextAllocatedLocation++;
  result << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << cmpResLocation << " 0x00000000" << std::endl;
  // Create buffer to hold current ID
  const int idLocation = nextAllocatedLocation++;
  result << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << idLocation << " 0x00000000" << std::endl;
  for(auto [id, preLabelId] : registeredFunctionInfos) {
    // Set current ID
    result << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << idLocation << " 0x";
    result << std::setfill('0') << std::setw(8) << std::hex << id << " 0x00000000" << std::endl;
    // Compare
    result << "Eq 0x" << std::setfill('0') << std::setw(8) << std::hex << idLocation << " 0x";
    result << std::setfill('0') << std::setw(8) << std::hex << selectorLocation << " 0x";
    result << std::setfill('0') << std::setw(8) << std::hex << cmpResLocation << " 0x00000000" << std::endl;
    // Jump
    result << "JmpCond 0x" << std::setfill('0') << std::setw(8) << std::hex << cmpResLocation << " !";
    result << preLabelId << " 0x00000000" << std::endl;
  }
  return result.str();
}

std::string ControlModule::close(int &nextAllocatedLocation) {
  if(selectorLocLocation == -1) {
    std::cerr << "Control module not compiled." << std::endl;
    exit(EXIT_FAILURE);
  }
  std::stringstream result;
  const int eswvLocation = nextAllocatedLocation++;
  result << "InitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << eswvLocation << " 0x00000000" << std::endl;
  result << "SetCnst 0x" << std::setfill('0') << std::setw(8) << std::hex << eswvLocation << " 0x45787465726E616C5374617465577269746561626C6556616C7565 0x00000000" << std::endl;
  result << "UpdateState 0x" << std::setfill('0') << std::setw(8) << std::hex << selectorLocLocation << " 0x";
  result << std::setfill('0') << std::setw(8) << std::hex << eswvLocation << " 0x00000000" << std::endl;
  retLocation = nextAllocatedLocation++;
  result << "ExitBfr 0x" << std::setfill('0') << std::setw(8) << std::hex << retLocation << " 0x00000000" << std::endl;
  return result.str();
}

int ControlModule::getRetLocation() const {
  return retLocation;
}