# 🧹 Examples Cleanup Plan

## 📊 Current Status:
- **Total Files**: ~80+ files
- **Duplicates**: ~50+ duplicate test files
- **Bypass Files**: ~10+ _bypass.sky files
- **Debug Files**: ~15+ debug/test variations

---

## ✅ KEEP (Clean, Working Examples):

### 1. **Getting Started**
- `smoke/hello.sky` - Hello World
- `mvp/arith.sky` - Basic arithmetic
- `mvp/if.sky` - If statement

### 2. **Async**
- `async/simple_async.sky` - Basic async/await
- `async/multiple_async.sky` - Multiple async ops

### 3. **OOP**
- `oop/class_basic.sky` - Class with init, methods

### 4. **Enum/Match**
- `enum/basic_enum.sky` - Simple enum
- `enum/enum_with_payload.sky` - Enum with data

### 5. **Collections**
- `collections/list_demo.sky` - List operations
- `collections/dict_demo.sky` - Dict operations

### 6. **Control Flow**
- `control/break_continue.sky` - Loop control

### 7. **Stdlib**
- `stdlib/test_all_stdlib.sky` - Full stdlib demo

### 8. **Recursion**
- `recursion/test_factorial.sky` - Recursion demo

### 9. **Unsafe**
- `unsafe/unsafe_test.sky` - Low-level ops

---

## ❌ DELETE (Duplicates, Debug Files):

### enum/ (DELETE 15 files):
- ❌ test_simple_payload.sky
- ❌ test_comments_fixed.sky
- ❌ test_comments.sky
- ❌ test_payload_debug.sky
- ❌ comment_test.sky
- ❌ debug_enum.sky
- ❌ enum_no_match.sky
- ❌ err_only.sky
- ❌ match_only.sky
- ❌ match_test.sky
- ❌ minimal.sky
- ❌ no_match_order.sky
- ❌ order_test.sky
- ❌ payload_match.sky
- ❌ payload_minimal.sky
- ❌ simple_match.sky
- ❌ simple.sky
- ❌ ultra_simple.sky
- ❌ with_import.sky
- ❌ WORKING_ENUM_TEST.sky
- ❌ ENUM_STATUS.md

### advanced/ (DELETE 4 files):
- ❌ complex_test.sky
- ❌ simple_test.sky
- ❌ final_test.sky
- ❌ stress_test.sky
- KEEP: comprehensive.sky

### comprehensive/ (DELETE 3 files):
- ❌ test_no_recursion.sky
- ❌ test_all_features.sky
- KEEP: test_final.sky

### recursion/ (DELETE 1 file):
- ❌ test_simple.sky
- KEEP: test_factorial.sky

### stdlib/ (DELETE 3 files):
- ❌ test_native_fs.sky
- ❌ test_option.sky
- ❌ result_option.sky
- ❌ test_set.sky
- KEEP: test_all_stdlib.sky

### All _bypass files (DELETE ~8 files):
- ❌ _bypass.sky (everywhere)
- ❌ _enum_bypass.sky
- ❌ _native_bypass.sky
- ❌ _dummy.sky
- ❌ _builtins.sky

---

## 📁 NEW STRUCTURE:

```
examples/
├── 01_hello/
│   └── hello.sky                    # Hello World
├── 02_basics/
│   ├── variables.sky                # Variables & types
│   ├── operators.sky                # Arithmetic, logic
│   └── control_flow.sky             # if/for/while/break
├── 03_functions/
│   ├── simple_function.sky          # Basic functions
│   └── recursion.sky                # Factorial example
├── 04_collections/
│   ├── lists.sky                    # List operations
│   └── dicts.sky                    # Dict operations
├── 05_oop/
│   ├── classes.sky                  # Class basics
│   └── inheritance.sky              # Inheritance
├── 06_async/
│   ├── simple_async.sky             # Basic async/await
│   └── multiple_async.sky           # Multiple operations
├── 07_enums/
│   ├── basic_enum.sky               # Simple enum
│   └── enum_with_payload.sky        # ADT with data
├── 08_stdlib/
│   ├── file_operations.sky          # FS demo
│   ├── crypto_demo.sky              # Crypto demo
│   └── full_stdlib.sky              # All features
├── 09_advanced/
│   ├── pattern_matching.sky         # Match examples
│   ├── unsafe_blocks.sky            # Low-level
│   └── comprehensive.sky            # All features
└── README.md                        # Examples guide
```

---

## 🎯 Goal:
- Reduce from 80+ files to 20 clean examples
- Remove all duplicates
- Clear organization
- Each example teaches one concept

