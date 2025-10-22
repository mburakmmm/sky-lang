#include <iostream>
#include <string>
#include <vector>
#include <sstream>
#include <chrono>
#include <algorithm>

std::string toUpper(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), ::toupper);
    return result;
}

std::string toLower(const std::string& str) {
    std::string result = str;
    std::transform(result.begin(), result.end(), result.begin(), ::tolower);
    return result;
}

std::string replaceAll(std::string str, const std::string& from, const std::string& to) {
    size_t start_pos = 0;
    while((start_pos = str.find(from, start_pos)) != std::string::npos) {
        str.replace(start_pos, from.length(), to);
        start_pos += to.length();
    }
    return str;
}

std::string join(const std::vector<std::string>& vec, const std::string& delimiter) {
    std::string result;
    for (size_t i = 0; i < vec.size(); ++i) {
        if (i != 0) result += delimiter;
        result += vec[i];
    }
    return result;
}

std::vector<std::string> split(const std::string& str, char delimiter) {
    std::vector<std::string> result;
    std::stringstream ss(str);
    std::string item;
    while (std::getline(ss, item, delimiter)) {
        result.push_back(item);
    }
    return result;
}

void stringOperations() {
    std::string text = "Hello World from SKY Programming Language";
    int iterations = 100000;
    
    auto start = std::chrono::high_resolution_clock::now();
    
    for (int i = 0; i < iterations; i++) {
        std::string upper = toUpper(text);
        std::string lower = toLower(text);
        std::string replaced = replaceAll(text, "SKY", "GO");
        std::string joined = join({"a", "b", "c"}, "-");
        std::vector<std::string> splitResult = split(text, ' ');
    }
    
    auto end = std::chrono::high_resolution_clock::now();
    auto duration = std::chrono::duration_cast<std::chrono::milliseconds>(end - start);
    
    std::cout << "C++ String operations (" << iterations << " iterations)" << std::endl;
    std::cout << "Duration: " << duration.count() << " ms" << std::endl;
}

int main() {
    stringOperations();
    return 0;
}
