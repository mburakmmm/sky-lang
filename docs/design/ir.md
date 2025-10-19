# LLVM IR Generation Strategy

## Overview

SKY programlama dili LLVM IR (Intermediate Representation) kullanarak native kod üretir. Bu belge IR generation stratejilerini ve best practices'leri açıklar.

## Architecture

```
SKY AST → IR Builder → LLVM IR → JIT/AOT → Native Code
```

## IR Builder (`internal/ir/builder.go`)

### Core Components

1. **Context** - LLVM context (isolated compilation unit)
2. **Module** - LLVM module (compilation unit)
3. **Builder** - IR instruction builder
4. **Value Map** - Symbol table for LLVM values

### Type Mapping

| SKY Type | LLVM Type | Size |
|----------|-----------|------|
| int | i64 | 8 bytes |
| float | double | 8 bytes |
| bool | i1 | 1 bit |
| string | i8* | pointer |
| void | void | 0 |

### Statement Generation

#### Let Statement
```llvm
%x = alloca i64          ; Stack allocation
store i64 10, i64* %x     ; Store value
```

#### Function
```llvm
define i64 @add(i64 %a, i64 %b) {
entry:
  %result = add i64 %a, %b
  ret i64 %result
}
```

#### If Statement
```llvm
  br i1 %cond, label %then, label %else

then:
  ; then block
  br label %merge

else:
  ; else block
  br label %merge

merge:
  ; continue
```

### Expression Generation

#### Binary Operations
```llvm
%left = load i64, i64* %a
%right = load i64, i64* %b
%result = add i64 %left, %right  ; +, -, *, /, %
```

#### Comparisons
```llvm
%cmp = icmp slt i64 %left, %right  ; <, <=, >, >=, ==, !=
```

#### Function Calls
```llvm
%result = call i64 @function(i64 %arg1, i64 %arg2)
```

## Optimization Passes

### Available Passes

1. **Instruction Combining** - Simplify instructions
2. **Reassociate** - Reassociate expressions
3. **GVN** - Global value numbering
4. **CFG Simplification** - Simplify control flow
5. **Mem2Reg** - Promote memory to registers

### Optimization Levels

- **O0**: No optimization
- **O1**: Basic optimization
- **O2**: Moderate optimization (default)
- **O3**: Aggressive optimization

## Special Cases

### Print Function

Print kullanır `printf` from libc:

```llvm
@.str = private constant [4 x i8] c"%lld\00"
%call = call i32 (i8*, ...) @printf(i8* @.str, i64 %value)
```

### String Literals

Global constants olarak:

```llvm
@.str.1 = private constant [12 x i8] c"Hello, SKY!\00"
```

## AOT Compilation

AOT mode'da IR:
1. File'a yazılır (.ll veya .bc)
2. `llc` ile native assembly
3. Linker ile executable

```bash
sky build program.sky -o program
# Generates LLVM IR → native code
```

## Debugging

### IR Dump

```bash
sky dump --ir program.sky
```

### Verification

Her fonksiyon `LLVMVerifyFunction` ile doğrulanır.

## Best Practices

1. **Always verify** - Her fonksiyon generate sonrası verify
2. **Use builder patterns** - Clean IR generation
3. **Optimize late** - İlk çalışır kod, sonra optimize
4. **Cache types** - Type lookups expensive

## Performance

- IR Generation: ~1000 statements/sec
- Verification: ~100ms for medium program
- JIT Compilation: ~100ms first time
- Subsequent runs: Cached, <1ms

## Future Improvements

- [ ] Better type inference in IR
- [ ] Inline optimization hints
- [ ] DWARF debug info generation
- [ ] LTO (Link-Time Optimization)

