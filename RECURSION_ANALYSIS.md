# 🔍 RECURSION BUG - DETAILED ANALYSIS

## 🐛 Problem:

```sky
function count(n)
  if n <= 0
    return n
  end
  return count(n - 1) + 1
end

count(3)  # Expected: 3, Actual: INFINITE LOOP
```

**Observed Behavior:**
```
count(3) -> count(2) -> count(1) -> count(0) [return 0]
BUT THEN:
count(-1) -> count(-2) -> count(-3) ... [INFINITE!]
```

## 🔎 Root Cause Analysis:

### Hypothesis 1: Return statement not propagating
- `return n` executes in count(0)
- But control flow continues to next line somehow?
- Or return value lost?

### Hypothesis 2: Condition check failing
- `n <= 0` might not be evaluating correctly
- Type coercion issue?

### Hypothesis 3: Environment/Scope issue
- Variable `n` might be getting corrupted
- Parameter not binding correctly

## 🧪 Debug Plan:

1. Add detailed logging to evalReturnStatement
2. Check evalBlockStatement return handling
3. Verify evalCallExpression return propagation
4. Test with simpler case (no arithmetic)

---

## 📝 Code Investigation:

