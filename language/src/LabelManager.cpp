//
// Created by ashy5000 on 6/24/24.
//

#include <sstream>
#include "LabelManager.h"

#include <iostream>

LabelManager::LabelManager(const std::string &blockasm) {
    labels = {};
    preLabels = {};
    // Pre-labels can be detected immediately.
    // Labels cannot be detected on LabelManager creation, as the control segment offsets line numbers.
    DetectPreLabels(blockasm);
}

void LabelManager::DetectPreLabels(const std::string& blockasm) {
    std::istringstream iss(blockasm);
    int labelPosition = 0;
    for (std::string line; std::getline(iss, line); ) {
        if(line.empty() || line.at(0) != ';') {
            labelPosition++;
            continue;
        }
        if(line.substr(2, 8) == "PRELABEL") {
            std::string labelIdString = line.substr(11);
            int labelId = stoi(labelIdString);
            preLabels[labelId] = labelPosition;
        }
        labelPosition++;
    }
}

void LabelManager::DetectLabels(const std::string& blockasm) {
    std::istringstream iss(blockasm);
    int labelPosition = 0;
    for (std::string line; std::getline(iss, line); ) {
        if(line.empty() || line.at(0) != ';') {
            labelPosition++;
            continue;
        }
        if(line.substr(2, 5) == "LABEL") {
            std::string labelIdString = line.substr(8);
            int labelId = stoi(labelIdString);
            labels[labelId] = labelPosition;
        }
        labelPosition++;
    }
}


std::string LabelManager::ReplacePreLabels(const std::string& blockasm) {
    // Pre-labels are already detected
    std::istringstream iss(blockasm);
    std::stringstream newBlockasm;
    for(std::string line; std::getline(iss, line); ) {
        if(line.substr(0, 3) != "Jmp") {
            newBlockasm << line << std::endl;
            continue;
        }
        std::istringstream innerIss(line);
        std::string section;
        std::stringstream newLine;
        bool isFirstSection = true;
        while(innerIss >> section) {
            if(section.at(0) != '!') {
                if(isFirstSection) {
                    newLine << section;
                    isFirstSection = false;
                    continue;
                }
                newLine << " " << section;
                continue;
            }
            std::string labelIdString = section.substr(1);
            int labelId = stoi(labelIdString);
            int labelPos = preLabels[labelId];
            if(isFirstSection) {
                newLine << labelPos;
                isFirstSection = false;
                continue;
            }
            newLine << " " << labelPos;
        }
        newBlockasm << newLine.str() << std::endl;
    }
    return newBlockasm.str();
}


std::string LabelManager::ReplaceLabels(const std::string& blockasm) {
    // Labels are not already detected.
    // Detect them now:
    DetectLabels(blockasm);
    std::istringstream iss(blockasm);
    std::stringstream newBlockasm;
    for(std::string line; std::getline(iss, line); ) {
        if(line.substr(0, 3) != "Jmp") {
            newBlockasm << line << std::endl;
            continue;
        }
        std::istringstream innerIss(line);
        std::string section;
        std::stringstream newLine;
        bool isFirstSection = true;
        while(innerIss >> section) {
            if(section.at(0) != '<') {
                if(isFirstSection) {
                    newLine << section;
                    isFirstSection = false;
                    continue;
                }
                newLine << " " << section;
                continue;
            }
            std::string labelIdString = section.substr(1);
            int labelId = stoi(labelIdString);
            int labelPos = labels[labelId] + 1;
            if(isFirstSection) {
                newLine << labelPos;
                isFirstSection = false;
                continue;
            }
            newLine << " " << labelPos;
        }
        newBlockasm << newLine.str() << std::endl;
    }
    return newBlockasm.str();
}

void LabelManager::ReplaceControlLabels(std::string &blockasm) {
    std::istringstream iss(blockasm);
    int control_close = 1;
    for(std::string line; std::getline(iss, line); ) {
        if(line.substr(0, 14) == "; CONTROLCLOSE") {
            break;
        }
        control_close++;
    }
    size_t pos = 0;
    while ((pos = blockasm.find('`', pos)) != std::string::npos) {
        blockasm.replace(pos, 1, std::to_string(control_close));
    }
}

std::string LabelManager::SkipLibs(const std::string& blockasm) {
    std::istringstream iss(blockasm);
    std::stringstream newBlockasm;
    for(std::string line; std::getline(iss, line); ) {
        if(line.substr(0, 3) != "Jmp") {
            newBlockasm << line << std::endl;
            continue;
        }
        std::istringstream innerIss(line);
        std::string section;
        std::stringstream newLine;
        bool isFirstSection = true;
        while(innerIss >> section) {
            if(section.at(0) != '%') {
                if(isFirstSection) {
                    newLine << section;
                    isFirstSection = false;
                    continue;
                }
                newLine << " " << section;
                continue;
            }
            int labelPos = -1;
            int i = 0;
            std::istringstream sourceDetectionIss(blockasm);
            for(std::string innerLine; std::getline(sourceDetectionIss, innerLine); ) {
                if(innerLine.substr(0, 21) == ";^^^^BEGIN_SOURCE^^^^") {
                    labelPos = i + 1;
                    break;
                }
                i++;
            }
            if(labelPos == -1) {
                std::cerr << "Missing source code. Are you trying to skip imports in a library?" << std::endl;
            }
            if(isFirstSection) {
                newLine << labelPos;
                isFirstSection = false;
                continue;
            }
            newLine << " " << labelPos;
        }
        newBlockasm << newLine.str() << std::endl;
    }
    return newBlockasm.str();
}

std::string LabelManager::OffsetCalls(const std::string& blockasm, int offset) {
    // Labels are not already detected.
    // Detect them now:
    DetectLabels(blockasm);
    std::istringstream iss(blockasm);
    std::stringstream newBlockasm;
    for(std::string line; std::getline(iss, line); ) {
        if(line.substr(0, 4) != "Call" && line.substr(0, 3) != "Jmp") {
            newBlockasm << line << std::endl;
            continue;
        }
        std::istringstream innerIss(line);
        std::string section;
        std::stringstream newLine;
        bool isFirstSection = true;
        while(innerIss >> section) {
            if(section.at(0) != '&') {
                if(isFirstSection) {
                    newLine << section;
                    isFirstSection = false;
                    continue;
                }
                newLine << " " << section;
                continue;
            }
            std::string existingLocationString = section.substr(1);
            int existingLocation = stoi(existingLocationString);
            int newLocation = existingLocation + offset;
            newLine << " " << newLocation;
        }
        newBlockasm << newLine.str() << std::endl;
    }
    return newBlockasm.str();
}
