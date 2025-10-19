# 🔧 RECURSION BUG - ROOT CAUSE & FIX

## 🎯 ROOT CAUSE IDENTIFIED:

### Problem:
```go
// value.go has BreakSignal and ContinueSignal:
type BreakSignal struct{}
type ContinueSignal struct{}

// BUT NO ReturnSignal!
```

### Why This Breaks Recursion:

```sky
function count(n)
  if n <= 0
    return n  # ← This executes evalReturnStatement
  end
  return count(n - 1) + 1
end
```

**What Happens:**
1. `count(0)` executes `if n <= 0`
2. `evalIfStatement` calls `evalBlockStatement(consequence)`
3. Inside block: `evalReturnStatement` returns `(0, nil)`
4. `evalIfStatement` returns `(0, nil)` - **just a value, not a signal!**
5. Control continues to next line in parent scope
6. `count(n - 1)` executes with n=0, so `count(-1)`, `count(-2)`... INFINITE!

**The Bug:**
- `evalBlockStatement` line 489 only checks `stmt.(*ast.ReturnStatement)`
- But when return is inside `if`, the statement type is `*ast.IfStatement`
- So return signal is lost!

## ✅ THE FIX:

### Step 1: Add ReturnSignal type
```go
// value.go
type ReturnSignal struct {
    Value Value
}

func (r *ReturnSignal) Error() string {
    return "return"
}
```

### Step 2: evalReturnStatement uses ReturnSignal
```go
func (i *Interpreter) evalReturnStatement(stmt *ast.ReturnStatement) (Value, error) {
    if stmt.ReturnValue != nil {
        val, err := i.evalExpression(stmt.ReturnValue)
        if err != nil {
            return nil, err
        }
        return nil, &ReturnSignal{Value: val}  // ← Signal!
    }
    return nil, &ReturnSignal{Value: &Nil{}}
}
```

### Step 3: Handle ReturnSignal everywhere
```go
// evalBlockStatement
if _, isReturn := err.(*ReturnSignal); isReturn {
    return nil, err  // Propagate signal
}

// evalIfStatement, evalWhileStatement, evalForStatement
// All must check and propagate ReturnSignal

// Function body execution
if retSignal, isReturn := err.(*ReturnSignal); isReturn {
    return retSignal.Value, nil  // Extract value and return normally
}
```

---

## 🧪 Expected Result:

```sky
count(3)
  ↓
count(2)
  ↓
count(1)
  ↓
count(0) → return 0 [ReturnSignal(0)]
  ↑
count(1) receives ReturnSignal → extracts 0, continues: return 0 + 1 = 1
  ↑
count(2) receives ReturnSignal → extracts 1, continues: return 1 + 1 = 2
  ↑
count(3) receives ReturnSignal → extracts 2, continues: return 2 + 1 = 3
  ↑
Result: 3 ✅
```

