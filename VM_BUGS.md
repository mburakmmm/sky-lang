# VM BUGS - Detaylı Analiz

## 🐛 BULUNAN KRITIK HATALAR

### 1. Return Values Yanlış (EN ÖNEMLİ!)
**Semptom**:
- factorial(5) = 12 (beklenen: 120) ❌
- factorial(10) = 14400 (beklenen: 3628800) ❌
- fib(5) = 2 (beklenen: 5) ❌
- fib(10) = 6 (beklenen: 55) ❌

**Sebep**: OpReturn sonrası stack/frame cleanup yanlış

**Konum**: `internal/vm/vm.go:282` OpReturn case

### 2. FOR LOOPS Çalışmıyor
**Semptom**:
```
runtime error: unsupported operands for +: int64 and []interface {}
```

**Sebep**: range() array döndürüyor ama iterator logic yok

**Konum**: `internal/vm/compiler.go:268` compileForStatement

### 3. String Concatenation Partial
**Semptom**: "text" + number bazen çalışıyor bazen değil

**Konum**: `internal/vm/vm.go:393` binaryOp

---

## 📋 DÜZELTME PLANI

### Öncelik 1: Return Values (CRITICAL)
- [ ] OpReturn frame cleanup
- [ ] OpCall return value handling
- [ ] Local variable stack management

### Öncelik 2: FOR LOOPS  
- [ ] Iterator protocol
- [ ] range() handling
- [ ] Loop variable binding

### Öncelik 3: String Ops
- [ ] Full type coercion in all ops
- [ ] print() multi-arg support

---

## 🔬 DEBUG TRACE

### Factorial Bug:
```
factorial(5) should be:
5 * factorial(4) = 5 * 24 = 120
But getting: 12

Hypothesis: Return değeri 10x küçük, muhtemelen
stack slot yanlış hesaplama
```

### Fibonacci Bug:
```
fib(5) should be:
fib(4) + fib(3) = 3 + 2 = 5
But getting: 2

Same issue as factorial
```

