#!/usr/bin/env node

function stringOperations() {
    const text = "Hello World from SKY Programming Language";
    const iterations = 100000;
    
    const start = Date.now();
    
    for (let i = 0; i < iterations; i++) {
        const upper = text.toUpperCase();
        const lower = text.toLowerCase();
        const replaced = text.replace("SKY", "GO");
        const joined = ["a", "b", "c"].join("-");
        const splitResult = text.split(" ");
    }
    
    const end = Date.now();
    const duration = end - start;
    
    console.log(`JavaScript String operations (${iterations} iterations)`);
    console.log(`Duration: ${duration} ms`);
}

function main() {
    stringOperations();
}

main();
