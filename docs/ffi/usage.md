# FFI Usage Guide

## Foreign Function Interface

SKY provides seamless C library integration through libffi.

## Basic Usage

### Loading a Library

```sky
import ffi

function main
  # Load libc
  let libc = ffi.load("libc.so.6")  # Linux
  # let libc = ffi.load("libc.dylib")  # macOS
  
  libc.close()
end
```

### Calling C Functions

```sky
import ffi

function callStrlen
  let libc = ffi.load("libc.so.6")
  
  # Get symbol
  let strlen = libc.symbol("strlen")
  
  # Set function signature
  strlen.setSignature(ffi.IntType, ffi.StringType)
  
  # Call
  let result = strlen.call("Hello, World!")
  print(result)  # Output: 13
  
  libc.close()
end
```

## Type System

### FFI Types

| SKY Type | FFI Type | C Type |
|----------|----------|--------|
| int | ffi.IntType | int64_t |
| float | ffi.FloatType | double |
| string | ffi.StringType | char* |
| pointer | ffi.PointerType | void* |
| void | ffi.VoidType | void |

### Type Marshalling

Automatic conversion:
- SKY int ↔ C int64_t
- SKY float ↔ C double
- SKY string ↔ C char* (null-terminated)
- SKY pointer ↔ C void*

## Advanced Usage

### Math Functions

```sky
import ffi

function mathDemo
  let libm = ffi.load("libm.so.6")
  
  # sqrt
  let sqrt = libm.symbol("sqrt")
  sqrt.setSignature(ffi.FloatType, ffi.FloatType)
  print("sqrt(16) = " + sqrt.call(16.0))
  
  # pow
  let pow = libm.symbol("pow")
  pow.setSignature(ffi.FloatType, ffi.FloatType, ffi.FloatType)
  print("pow(2, 10) = " + pow.call(2.0, 10.0))
  
  libm.close()
end
```

### Memory Management

```sky
import ffi

function memoryExample
  unsafe
    # Allocate
    let ptr = ffi.malloc(1024)
    
    # Use memory
    # ...
    
    # Free
    ffi.free(ptr)
  end
end
```

### Custom Structures

```sky
import ffi

function structExample
  unsafe
    # Allocate struct
    let structSize = 16  # 2 x int64
    let ptr = ffi.malloc(structSize)
    
    # Write fields
    # (requires pointer manipulation)
    
    ffi.free(ptr)
  end
end
```

## Safety

### Unsafe Blocks

FFI operations must be in `unsafe` blocks:

```sky
function safeFFI
  unsafe
    let libc = ffi.load("libc.so.6")
    let strlen = libc.symbol("strlen")
    strlen.setSignature(ffi.IntType, ffi.StringType)
    let result = strlen.call("test")
    print(result)
    libc.close()
  end
end
```

### Error Handling

```sky
function robustFFI
  try
    unsafe
      let lib = ffi.load("mylib.so")
      # Use library
      lib.close()
    end
  catch error
    print("FFI error: " + error)
  end
end
```

## Platform Differences

### Linux

```sky
let libc = ffi.load("libc.so.6")
let libm = ffi.load("libm.so.6")
let libpthread = ffi.load("libpthread.so.0")
```

### macOS

```sky
let libc = ffi.load("libc.dylib")
let libSystem = ffi.load("/usr/lib/libSystem.dylib")
```

### Windows

```sky
let msvcrt = ffi.load("msvcrt.dll")
let kernel32 = ffi.load("kernel32.dll")
```

## Common Libraries

### libc Functions

```sky
# String functions
strlen(str)
strcmp(str1, str2)
strcpy(dest, src)

# Memory functions  
malloc(size)
free(ptr)
memcpy(dest, src, n)

# I/O functions
printf(format, ...)
fopen(path, mode)
fclose(file)
```

### libm Functions

```sky
# Math functions
sqrt(x)
pow(x, y)
sin(x)
cos(x)
log(x)
exp(x)
```

## Performance

- Function call overhead: ~20ns
- Type marshalling: ~10ns
- Pointer operations: Native speed

## Best Practices

1. **Always close libraries**
   ```sky
   let lib = ffi.load("mylib.so")
   # ... use library ...
   lib.close()  # Important!
   ```

2. **Check errors**
   ```sky
   try
     let lib = ffi.load("lib.so")
   catch error
     print("Failed to load: " + error)
   end
   ```

3. **Use unsafe blocks**
   ```sky
   unsafe
     # FFI operations here
   end
   ```

4. **Free allocated memory**
   ```sky
   unsafe
     let ptr = ffi.malloc(size)
     # ... use ...
     ffi.free(ptr)  # Always free!
   end
   ```

## Debugging

### FFI Errors

```bash
$ sky run program.sky
FFI error: symbol not found strlen: undefined symbol
```

Check:
- Library path correct?
- Symbol name correct?
- Library loaded?

### Memory Leaks

Use valgrind or sanitizers:

```bash
valgrind --leak-check=full sky run program.sky
```

## Examples

See `examples/ffi/` for more examples:
- `strlen.sky` - String length
- `math.sky` - Math functions
- (more to come)

