# SKY - FULL FEATURE TEST PLAN
# .cursorrules'daki TÜM özelliklerin sıkı testi

## 📋 ÖZELLİK LİSTESİ (.cursorrules'dan)

### TEMEL DİL ÖZELLİKLERİ
1. ✅ Variables: let, const
2. ✅ Functions: function ... end
3. ✅ Control Flow: if/elif/else/end
4. ✅ Loops: while ... end
5. ⚠️ Loops: for ... in ... end
6. ✅ Return statements
7. ✅ Recursion
8. ⚠️ Classes: class ... end
9. ⚠️ OOP: self, super
10. ⚠️ Import: import module

### ASYNC & CONCURRENCY
11. ❌ async functions
12. ❌ await expressions
13. ❌ coop functions (coroutines)
14. ❌ yield statements

### DÜŞÜK SEVİYE
15. ❌ unsafe blocks
16. ❌ Raw pointers (unsafe içinde)

### TİP SİSTEMİ
17. ✅ Basic types: int, float, string, bool, any
18. ✅ Type annotations: let x: int
19. ✅ Type inference: let x = 10

### OPERATÖRLER
20. ✅ Arithmetic: +, -, *, /, %
21. ✅ Comparison: ==, !=, <, <=, >, >=
22. ✅ Logical: &&, ||, !
23. ✅ Assignment: =, +=, -=, *=, /=, %=

### BUILT-IN FUNCTIONS
24. ✅ print()
25. ✅ len()
26. ✅ range()

### EXECUTION MODES
27. ✅ JIT (LLVM) - framework exists
28. ✅ Interpreter - working
29. ✅ Bytecode VM - working
30. ⚠️ AOT compilation - framework only

### TOOLS
31. ✅ sky run
32. ✅ sky run --vm
33. ✅ sky build (framework)
34. ✅ sky test
35. ✅ sky repl
36. ✅ sky dump --tokens/--ast/--bytecode
37. ✅ sky check
38. ✅ wing (package manager)
39. ✅ skyls (LSP server)
40. ✅ skydbg (debugger framework)

---

## 🔬 TEST PLANI

### Phase 1: ÇALIŞANLAR (18/40)
- Variables, functions, control flow
- Recursion, loops (while)
- Arithmetic, comparisons, logic
- Built-ins, interpreter, VM
- CLI tools

### Phase 2: EKSİKLER (22/40)
- async/await (0/4)
- Classes/OOP (0/3)
- unsafe (0/2)
- for loops (partial)
- Import system (0/1)
- Advanced features (0/12)

