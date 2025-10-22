# 📚 SKY Programming Language - API Reference (English)

## 🌟 Introduction

SKY is a modern, type-safe programming language written in Go. It combines Python's simplicity with Go's performance.

---

## 📖 Table of Contents

1. [Error Handling](#error-handling) ⚠️
2. [Basic Syntax](#basic-syntax)
3. [Data Types](#data-types)
4. [Functions](#functions)
5. [Control Structures](#control-structures)
6. [Classes and OOP](#classes-and-oop)
7. [Async/Await](#asyncawait)
8. [Pattern Matching](#pattern-matching)
9. [Built-in Functions](#built-in-functions)
10. [Standard Library](#standard-library)
11. [Module System](#module-system)
12. [Package Management](#package-management)

---

## ⚠️ Error Handling

SKY uses **try/catch/finally** model for error handling. This is the common exception-based error handling approach in modern languages.

### Try-Catch-Finally

```sky
# Basic try-catch usage
try
  let result = fs_read_text("file.txt")
  print("File read: " + result)
catch error
  print("Error occurred: " + error)
end

# Finally block for cleanup
try
  let file = fs_open("data.txt")
  # File operations...
catch error
  print("File error: " + error)
finally
  # Always runs
  print("Cleaning up...")
end
```

### Throw (Error Throwing)

```sky
function divide(a: int, b: int): int
  if b == 0
    throw "Division by zero error!"
  end
  return a / b
end

# Usage
try
  let result = divide(10, 0)
  print("Result: " + result)
catch error
  print("Error caught: " + error)
end
```

### Error Types

SKY has two types of errors:

1. **Runtime Errors**: Errors that occur during program execution
   - `fs_read_text("nonexistent.txt")` → File not found error
   - `list[10]` (5-element list) → Index out of range error
   - `dict["nonexistent_key"]` → Key not found error

2. **Throw Errors**: Errors thrown by the programmer
   - `throw "Custom error message"`
   - `throw error_object`

### Error Handling Best Practices

```sky
# 1. Specific error catching
try
  let data = http_get("https://api.example.com/data")
  process_data(data)
catch error
  if error.contains("network")
    print("Network error, retrying...")
  else
    print("Unknown error: " + error)
  end
end

# 2. Error logging
try
  risky_operation()
catch error
  log_error("Risky operation failed: " + error)
  # Error can be re-thrown
  throw error
end

# 3. Resource cleanup
try
  let connection = db_connect()
  # Database operations...
catch error
  print("DB error: " + error)
finally
  if connection != null
    connection.close()
  end
end
```

---

## 🎯 Basic Syntax

### Variables

```sky
# Variable declaration
let x = 10
let name = "Sky"
let pi = 3.14
let active = true

# Constant declaration
const MAX_SIZE = 100

# Type annotation (optional)
let age: int = 25
let price: float = 19.99
```

### Comments

```sky
# Single-line comment

# For multiple lines
# use # on each line
```

---

## 📦 Data Types

### Primitive Types

```sky
# Integer
let count = 42
let negative = -10

# Float
let temperature = 36.6
let epsilon = 0.001

# String
let greeting = "Hello World"
let empty = ""

# Boolean
let is_valid = true
let is_empty = false

# Nil (null value)
let nothing = nil
```

### Collection Types

#### List

```sky
# Create list
let numbers = [1, 2, 3, 4, 5]
let names = ["Alice", "Bob", "Charlie"]
let mixed = [1, "hello", true]

# Access element
print(numbers[0])  # 1

# List methods
let length = len(numbers)
print("List length:", length)
```

##### List Methods

| Method | Description | Example |
|--------|-------------|---------|
| `.append(item)` | Add element | `list.append(6)` |
| `.pop([index])` | Remove element | `list.pop()` or `list.pop(0)` |
| `.remove(item)` | Remove by value | `list.remove("Alice")` |
| `.insert(index, item)` | Insert at position | `list.insert(1, "David")` |
| `.sort()` | Sort | `list.sort()` |
| `.reverse()` | Reverse | `list.reverse()` |
| `.clear()` | Clear | `list.clear()` |
| `.count(item)` | Count elements | `list.count("Alice")` |
| `.index(item)` | Get position | `list.index("Alice")` |
| `.copy()` | Copy | `let new_list = list.copy()` |

```sky
let fruits = ["apple", "banana", "cherry"]

# Add element
fruits.append("orange")
print(fruits)  # ["apple", "banana", "cherry", "orange"]

# Remove element
let last = fruits.pop()
print(last)    # "orange"
print(fruits)  # ["apple", "banana", "cherry"]

# Sort
fruits.sort()
print(fruits)  # ["apple", "banana", "cherry"]

# Reverse
fruits.reverse()
print(fruits)  # ["cherry", "banana", "apple"]
```

#### Dictionary (Dict)

```sky
# Create dictionary
let person = {
  "name": "John",
  "age": "30",
  "city": "New York"
}

# Access element
print(person["name"])  # John

# Add element
person["email"] = "john@example.com"
```

##### Dictionary Methods

| Method | Description | Example |
|--------|-------------|---------|
| `.keys()` | Get all keys | `dict.keys()` |
| `.values()` | Get all values | `dict.values()` |
| `.items()` | Get key-value pairs | `dict.items()` |
| `.get(key, default)` | Safe access | `dict.get("name", "Unknown")` |
| `.set(key, value)` | Set value | `dict.set("age", 31)` |
| `.has_key(key)` | Check if key exists | `dict.has_key("name")` |
| `.delete(key)` | Delete key | `dict.delete("age")` |
| `.clear()` | Clear | `dict.clear()` |
| `.copy()` | Copy | `let new_dict = dict.copy()` |
| `.update(other)` | Update with other dict | `dict.update(other_dict)` |

```sky
let person = {
  "name": "John",
  "age": 30,
  "city": "New York"
}

# Get keys
let keys = person.keys()
print(keys)  # ["name", "age", "city"]

# Get values
let values = person.values()
print(values)  # ["John", 30, "New York"]

# Safe access
let name = person.get("name", "Unknown")
let phone = person.get("phone", "None")  # Returns "None"

# Check if key exists
if person.has_key("age")
  print("Age information available")
end

# Delete element
person.delete("age")
print(person)  # {"name": "John", "city": "New York"}

# Iteration
for key, value in person.items()
  print(key + ": " + value)
end
```

---

## 🔧 Functions

### Basic Function

```sky
function add(a, b)
  return a + b
end

let result = add(5, 3)  # 8
```

### Typed Function

```sky
function multiply(x: int, y: int): int
  return x * y
end
```

### Default Parameters

```sky
function greet(name = "Guest")
  print("Hello", name)
end

greet()          # Hello Guest
greet("Alice")   # Hello Alice
```

### Recursive Functions

```sky
function factorial(n)
  if n <= 1
    return 1
  end
  return n * factorial(n - 1)
end

print(factorial(5))  # 120
```

### Function Type Annotation

In SKY, function types are specified using the `(parameter_types) => return_type` syntax.

```sky
# Function type annotation examples
function test_callback(callback: (int, string) => bool): void
  let result = callback(42, "test")
  print("Callback result:", result)
end

# Use 'any' for empty parameter lists
function test_empty_callback(callback: any): void
  callback()
end

# Usage with lambda expressions
test_callback(function(x: int, s: string): bool
  print("Callback called: x =", x, "s =", s)
  return true
end)

test_empty_callback(function(): void
  print("Empty callback called")
end)
```

**Note:** Empty parameter list syntax `() => void` is not yet supported. Use `any` type instead.

---

## 🎮 Control Structures

### If-Else

```sky
let age = 18

if age >= 18
  print("Adult")
else
  print("Minor")
end

# With elif
let grade = 85

if grade >= 90
  print("A")
elif grade >= 80
  print("B")
elif grade >= 70
  print("C")
else
  print("Pass")
end
```

### While Loop

```sky
let counter = 0

while counter < 5
  print(counter)
  counter = counter + 1
end
```

### For Loop

```sky
# Iterate over list
let fruits = ["apple", "banana", "orange"]

for fruit in fruits
  print(fruit)
end

# Iterate over range
for i in range(10)
  print(i)  # 0 to 9
end
```

### Break and Continue

```sky
for i in range(10)
  if i == 3
    continue  # Skip 3
  end
  
  if i == 7
    break  # Stop at 7
  end
  
  print(i)
end
```

---

## 🏛️ Classes and OOP

### Class Definition

```sky
class Person
  function init(name, age)
    self.name = name
    self.age = age
  end
  
  function greet()
    print("Hello, I'm", self.name)
  end
  
  function info()
    print(self.name, "is", self.age, "years old")
  end
end

# Usage
let john = Person("John", 25)
john.greet()  # Hello, I'm John
john.info()   # John is 25 years old
```

### Inheritance

```sky
class Animal
  function init(name)
    self.name = name
  end
  
  function make_sound()
    print(self.name, "makes a sound")
  end
end

class Cat
  function init(name, breed)
    super.init(name)
    self.breed = breed
  end
  
  function make_sound()
    print(self.name, "meows")
  end
end

let cat = Cat("Fluffy", "Persian")
cat.make_sound()  # Fluffy meows
```

### Multiple Inheritance

SKY supports multiple inheritance using the `:` operator:

```sky
class Flying
  function fly(): void
    print("Flying...")
  end
end

class Swimming
  function swim(): void
    print("Swimming...")
  end
end

class Duck : Flying, Swimming
  function init(name: string)
    self.name = name
  end
  
  # Override parent methods
  function fly(): void
    print(self.name, "is flying...")
  end
  
  function swim(): void
    print(self.name, "is swimming...")
  end
  
  # Own methods
  function search_food(): void
    print(self.name, "is searching for food...")
  end
end

# Usage
let duck = Duck("Donald")
duck.fly()        # Donald is flying...
duck.swim()       # Donald is swimming...
duck.search_food() # Donald is searching for food...
```

#### Multiple Inheritance Rules

1. **Method Conflicts**: If two parent classes have methods with the same name, the child class must override them
2. **Diamond Problem**: SKY automatically resolves the diamond problem in multiple inheritance
3. **Constructor Chain**: Parent constructors are called automatically

```sky
class A
  function init()
    print("A constructor")
  end
end

class B
  function init()
    print("B constructor")
  end
end

class C : A, B
  function init()
    super.init()  # Call parent constructors
    print("C constructor")
  end
end

let c = C()  # A constructor, B constructor, C constructor
```

### Abstract Classes

```sky
# Abstract class (behaves like interface)
abstract class Shape
  function init()
    # Abstract class constructor
  end
  
  # Abstract methods (must be implemented)
  abstract function calculate_area(): float
  abstract function calculate_perimeter(): float
  
  # Concrete method
  function print_info(): void
    print("Area:", self.calculate_area())
    print("Perimeter:", self.calculate_perimeter())
  end
end

class Square : Shape
  function init(side: float)
    self.side = side
  end
  
  # Implement abstract methods
  function calculate_area(): float
    return self.side * self.side
  end
  
  function calculate_perimeter(): float
    return 4 * self.side
  end
end

# Usage
let square = Square(5.0)
square.print_info()  # Area: 25, Perimeter: 20
```

---

## ⚡ Async/Await

### Asynchronous Functions

```sky
async function fetchData(id)
  print("Fetching data:", id)
  # Async operation simulation
  return "Data-" + str(id)
end

async function main
  let result = await fetchData(42)
  print("Result:", result)
end
```

### Parallel Async Operations

#### Sequential Processing

```sky
async function getUser(id)
  # Simulated API call
  await sleep(100)  # Wait 100ms
  return {"id": str(id), "name": "User" + str(id)}
end

async function main
  # Sequential - takes 200ms total
  let user1 = await getUser(1)
  let user2 = await getUser(2)
  
  print("User 1:", user1)
  print("User 2:", user2)
end
```

#### Parallel Processing

```sky
async function main
  # Parallel - takes 100ms total
  let promise1 = getUser(1)  # Returns promise
  let promise2 = getUser(2)  # Returns promise
  
  # Wait for both
  let user1 = await promise1
  let user2 = await promise2
  
  print("User 1:", user1)
  print("User 2:", user2)
end
```

#### Built-in Promise.all() Function

SKY provides built-in `Promise.all()` for parallel operations:

```sky
async function main
  let ids = [1, 2, 3, 4, 5]
  
  # Promise.all() for parallel processing
  let promises = []
  for id in ids
    promises.append(getUser(id))
  end
  
  # Wait for all promises in parallel
  let results = await Promise.all(promises)
  
  print("All users:", results)
end
```

#### Promise.all() Features

| Feature | Description |
|---------|-------------|
| **Parallel Execution** | All promises start simultaneously |
| **Fast Failure** | If one promise fails, immediately returns error |
| **Ordered Results** | Results are returned in promise order |
| **Type Safety** | All promises must return the same type |

#### Error Handling with Promise.all()

```sky
async function safeGetUser(id)
  try
    return await getUser(id)
  catch error
    return {"id": id, "error": "User not found"}
  end
end

async function main
  let ids = [1, 2, 3, 999]  # 999 is invalid ID
  
  let promises = []
  for id in ids
    promises.append(safeGetUser(id))
  end
  
  # Error handling with Promise.all()
  try
    let results = await Promise.all(promises)
    print("Results:", results)
  catch error
    print("General error:", error)
  end
end
```

#### Promise.allSettled() Alternative

```sky
# Promise.allSettled() similar function
async function Promise_allSettled(promises: list): list
  let results = []
  
  for promise in promises
    try
      let result = await promise
      results.append({"status": "fulfilled", "value": result})
    catch error
      results.append({"status": "rejected", "reason": error})
    end
  end
  
  return results
end

# Usage
let results = await Promise_allSettled(promises)
for result in results
  if result["status"] == "fulfilled"
    print("Success:", result["value"])
  else
    print("Error:", result["reason"])
  end
end
```

---

## 🎯 Pattern Matching

### Enum Definition

```sky
enum Result
  Success(int)
  Error(string)
end
```

### Match Expression

```sky
let operation = Success(42)

match operation
  Success(value) => print("Success:", value)
  Error(message) => print("Error:", message)
end
```

### Complex Example

```sky
enum Status
  Pending
  Processing
  Completed(string)
  Failed(int, string)
end

let status = Completed("File saved")

match status
  Pending => print("Waiting...")
  Processing => print("Processing...")
  Completed(msg) => print("Completed:", msg)
  Failed(code, desc) => print("Error", code, ":", desc)
end
```

---

## 🛠️ Built-in Functions

SKY's always available basic functions:

### Output Functions

| Function | Description | Example |
|----------|-------------|---------|
| `print(...)` | Print values | `print("Hello", 42)` |
| `println(...)` | Print values + newline | `println("Hello")` |

### Type Conversions

| Function | Description | Example |
|----------|-------------|---------|
| `int(value)` | Convert string/float to int | `int("42")` → `42` |
| `float(value)` | Convert string/int to float | `float("3.14")` → `3.14` |
| `str(value)` | Convert any value to string | `str(42)` → `"42"` |
| `bool(value)` | Convert value to boolean | `bool(1)` → `true` |

### Collection Functions

| Function | Description | Example |
|----------|-------------|---------|
| `len(collection)` | Get length | `len([1,2,3])` → `3` |
| `join(separator, list)` | Convert list to string | `join("-", ["a","b"])` → `"a-b"` |
| `range(start, end)` | Number range | `range(1, 5)` → `[1,2,3,4]` |

### Type Checking

| Function | Description | Example |
|----------|-------------|---------|
| `type(value)` | Get value type | `type(42)` → `"int"` |
| `is_int(value)` | Is int? | `is_int(42)` → `true` |
| `is_string(value)` | Is string? | `is_string("hello")` → `true` |
| `is_list(value)` | Is list? | `is_list([1,2])` → `true` |
| `is_dict(value)` | Is dict? | `is_dict({})` → `true` |

### Time Functions

| Function | Description | Example |
|----------|-------------|---------|
| `time_now()` | Current time (ms) | `time_now()` → `1698000000000` |
| `sleep(ms)` | Wait (milliseconds) | `sleep(1000)` |

### Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `null` | `nil` | Null value |
| `nil` | `nil` | Null value (alias) |
| `true` | `true` | True |
| `false` | `false` | False |

### Example Usage

```sky
# Basic output
print("Hello", "World")
println("With newline")

# Type conversions
let num_str = "42"
let num = int(num_str)
let pi_str = str(3.14159)

# Collection operations
let fruits = ["apple", "banana", "cherry"]
let joined = join(", ", fruits)
print(joined)  # "apple, banana, cherry"

# Type checking
if is_string(value)
  print("This is a string: " + value)
end

# Time operations
let start = time_now()
# Do work...
let end_time = time_now()
let duration = end_time - start
print("Operation took " + duration + " ms")
```

---

## 📚 Standard Library

### File Operations (FS)

#### Basic File Operations

| Function | Description | Example |
|----------|-------------|---------|
| `fs_read_text(path)` | Read file | `fs_read_text("data.txt")` |
| `fs_write_text(path, content)` | Write file | `fs_write_text("out.txt", "data")` |
| `fs_exists(path)` | Check if file exists | `fs_exists("file.txt")` |
| `fs_read_bytes(path)` | Binary read | `fs_read_bytes("image.png")` |
| `fs_write_bytes(path, data)` | Binary write | `fs_write_bytes("out.bin", bytes)` |

#### Directory Operations

| Function | Description | Example |
|----------|-------------|---------|
| `fs_mkdir(path)` | Create directory | `fs_mkdir("new_folder")` |
| `fs_rmdir(path)` | Remove directory | `fs_rmdir("old_folder")` |
| `fs_list_dir(path)` | List directory | `fs_list_dir(".")` |
| `fs_delete(path)` | Delete file/directory | `fs_delete("file.txt")` |

#### Example Usage

```sky
# Write and read file
fs_write_text("test.txt", "Hello World")
let content = fs_read_text("test.txt")
print(content)

# Check if file exists
if fs_exists("test.txt")
  print("File exists")
end

# Directory operations
fs_mkdir("new_folder")
let files = fs_list_dir(".")
for file in files
  print("File:", file)
end

# Error handling with file reading
try
  let data = fs_read_text("nonexistent.txt")
  print(data)
catch error
  print("File read error:", error)
end
```

### Operating System (OS)

```sky
# Platform information
let platform = os_platform()
print("Platform:", platform)

# Current working directory
let cwd = os_getcwd()
print("Directory:", cwd)

# Environment variable
let home = os_getenv("HOME")
print("Home:", home)
```

### HTTP Operations

#### Basic HTTP Methods

| Function | Description | Example |
|----------|-------------|---------|
| `http_get(url)` | GET request | `http_get("https://api.example.com")` |
| `http_post(url, data)` | POST request | `http_post("https://api.example.com", data)` |
| `http_put(url, data)` | PUT request | `http_put("https://api.example.com/1", data)` |
| `http_delete(url)` | DELETE request | `http_delete("https://api.example.com/1")` |

#### HTTP Response Properties

| Property | Description | Example |
|----------|-------------|---------|
| `response.status_code` | HTTP status code | `200`, `404`, `500` |
| `response.body` | Response body | `"{\"name\":\"John\"}"` |
| `response.headers` | Response headers | `{"Content-Type": "application/json"}` |

#### Example Usage

```sky
# GET request
let response = http_get("https://api.github.com/users/octocat")
print("Status:", response.status_code)
print("Body:", response.body)

# POST request
let data = {"name": "John", "age": 30}
let response = http_post("https://api.example.com/users", data)
if response.status_code == 201
  print("User created")
end

# Error handling with HTTP
try
  let response = http_get("https://api.example.com/data")
  if response.status_code == 200
    print("Success:", response.body)
  else
    print("Error code:", response.status_code)
  end
catch error
  print("Network error:", error)
end
```

### Cryptography (Crypto)

#### Hash Functions

| Function | Description | Example |
|----------|-------------|---------|
| `crypto_md5(data)` | MD5 hash | `crypto_md5("password123")` |
| `crypto_sha1(data)` | SHA1 hash | `crypto_sha1("password123")` |
| `crypto_sha256(data)` | SHA256 hash | `crypto_sha256("password123")` |
| `crypto_sha512(data)` | SHA512 hash | `crypto_sha512("password123")` |

#### Encryption Functions

| Function | Description | Example |
|----------|-------------|---------|
| `crypto_aes_encrypt(data, key)` | AES encryption | `crypto_aes_encrypt("data", "key")` |
| `crypto_aes_decrypt(data, key)` | AES decryption | `crypto_aes_decrypt(encrypted, "key")` |
| `crypto_hmac(data, key)` | HMAC signature | `crypto_hmac("data", "secret")` |

#### Example Usage

```sky
# Hash operations
let hash_md5 = crypto_md5("password123")
let hash_sha256 = crypto_sha256("password123")
print("MD5:", hash_md5)
print("SHA256:", hash_sha256)

# Encryption
let data = "Secret data"
let key = "secret_key"
let encrypted = crypto_aes_encrypt(data, key)
let decrypted = crypto_aes_decrypt(encrypted, key)
print("Encrypted:", encrypted)
print("Decrypted:", decrypted)
```

### JSON Operations

```sky
# JSON encode
let data = {"name": "John", "age": "25"}
let json_str = json_encode(data)
print(json_str)  # {"name":"John","age":"25"}

# JSON decode
let parsed = json_decode(json_str)
print(parsed["name"])  # John
```

### Time and Date

#### Basic Time Functions

| Function | Description | Example |
|----------|-------------|---------|
| `time_now()` | Current time (ms) | `time_now()` → `1698000000000` |
| `sleep(ms)` | Wait (milliseconds) | `sleep(1000)` |
| `time_format(timestamp, format)` | Format time | `time_format(now, "%Y-%m-%d %H:%M")` |
| `time_parse(date_string, format)` | Parse string to time | `time_parse("2023-10-22", "%Y-%m-%d")` |
| `time_add(timestamp, duration)` | Add time | `time_add(now, "1h30m")` |
| `time_diff(timestamp1, timestamp2)` | Time difference | `time_diff(end, start)` |

#### Time Formats

| Format | Description | Example |
|--------|-------------|---------|
| `%Y` | Year (4 digits) | `2023` |
| `%m` | Month (01-12) | `10` |
| `%d` | Day (01-31) | `22` |
| `%H` | Hour (00-23) | `14` |
| `%M` | Minute (00-59) | `30` |
| `%S` | Second (00-59) | `45` |
| `%A` | Weekday | `Sunday` |
| `%B` | Month name | `October` |

#### Example Usage

```sky
# Get current time
let now = time_now()
print("Timestamp:", now)

# Format time
let formatted = time_format(now, "%Y-%m-%d %H:%M:%S")
print("Formatted:", formatted)  # 2023-10-22 14:30:45

# Parse string to time
let parsed = time_parse("2023-10-22 14:30:00", "%Y-%m-%d %H:%M:%S")
print("Parsed:", parsed)

# Add time
let future = time_add(now, "2h30m")  # Add 2 hours 30 minutes
let future_formatted = time_format(future, "%Y-%m-%d %H:%M")
print("Future:", future_formatted)

# Time difference
let start = time_now()
sleep(2000)  # Wait 2 seconds
let end_time = time_now()
let duration = time_diff(end_time, start)
print("Duration:", duration, "ms")

# Date comparison
let today = time_now()
let tomorrow = time_add(today, "24h")
if tomorrow > today
  print("Tomorrow is after today")
end
```

#### Duration Formats

| Format | Description | Example |
|--------|-------------|---------|
| `s` | Seconds | `30s` |
| `m` | Minutes | `5m` |
| `h` | Hours | `2h` |
| `d` | Days | `7d` |
| `w` | Weeks | `2w` |

```sky
# Duration examples
let short_duration = "30s"      # 30 seconds
let medium_duration = "2h30m"    # 2 hours 30 minutes
let long_duration = "1w3d12h"    # 1 week 3 days 12 hours
```

### Random Numbers

```sky
# Random integer (0-99)
let number = rand_int(100)
print("Random:", number)

# Generate UUID
let uuid = rand_uuid()
print("UUID:", uuid)
```

### String Methods

```sky
let text = "hello world"

print(text.upper())        # HELLO WORLD
print(text.lower())        # hello world
print(text.split())        # [hello, world]
print("  test  ".strip())  # test
```

### Type Conversions

```sky
# String to number
let number = int("42")
let decimal = float("3.14")

# Number to string
let text = str(123)

# To boolean
let true_val = bool(1)
let false_val = bool(0)

# Type check
print(type(42))        # int
print(type("hello"))   # string
```

---

## 📁 Module System

SKY's module system allows you to organize your code and create reusable components.

### Module Importing

#### Basic Import

```sky
# Import module completely
import math

# Usage
let result = math.add(5, 3)
```

#### Import with Alias

```sky
# Import module with different name
import math as mathematics

# Usage
let result = mathematics.add(5, 3)
```

#### Selective Import

```sky
# Import only specific functions
import math { add, subtract }

# Usage
let result = add(5, 3)
let diff = subtract(10, 4)
```

### Module Creation

#### Simple Module (math.sky)

```sky
# math.sky
function add(a: int, b: int): int
  return a + b
end

function subtract(a: int, b: int): int
  return a - b
end

function multiply(a: int, b: int): int
  return a * b
end

# Private function (not exported)
function _internal_calc(x: int): int
  return x * 2
end
```

#### Usage

```sky
# main.sky
import math

function main: void
  let sum = math.add(10, 20)
  let product = math.multiply(5, 6)
  print("Sum:", sum)
  print("Product:", product)
end
```

### Module Structure

```
project/
├── main.sky
├── utils/
│   ├── math.sky
│   ├── string.sky
│   └── io.sky
└── models/
    ├── user.sky
    └── product.sky
```

### Circular Dependencies

SKY automatically detects circular dependencies and throws an error:

```sky
# A.sky
import B
# ...

# B.sky  
import A  # ❌ Error: Circular dependency!
```

**Solution**: Move common code to a separate module:

```sky
# common.sky
function shared_function()
  # Common code
end

# A.sky
import common
# ...

# B.sky
import common
# ...
```

### Module Search Path

SKY searches for modules in this order:

1. **Relative path**: `./utils/math.sky`
2. **Project root**: `./math.sky`
3. **Standard library**: `math` (built-in)
4. **Wing packages**: Packages installed with `wing install`

### Module Examples

#### HTTP Module (http.sky)

```sky
# http.sky
function get(url: string): dict
  # HTTP GET implementation
  return {"status": 200, "body": "data"}
end

function post(url: string, data: dict): dict
  # HTTP POST implementation
  return {"status": 201, "body": "created"}
end
```

#### Usage

```sky
import http

let response = http.get("https://api.example.com")
print("Status:", response["status"])
```

---

## 📦 Package Management (Wing)

### Create New Project

```bash
wing init
```

### Install Package

```bash
wing install http
wing install json@1.0.0  # Specific version
```

### Update Packages

```bash
wing update           # All packages
wing update http      # Specific package
```

### Build Project

```bash
wing build
```

### Publish Package

```bash
wing publish
```

---

## 🔒 Unsafe Blocks

For low-level operations:

```sky
unsafe
  let pointer = 0xDEADBEEF
  # Raw memory operations
end
```

⚠️ **Warning**: Use unsafe blocks with caution!

---

## 💡 Best Practices

### 1. Naming Conventions

```sky
# Variables: snake_case
let user_name = "john"
let total_price = 100

# Functions: snake_case
function calculate_total()
  # ...
end

# Classes: PascalCase
class UserManager
  # ...
end

# Constants: UPPER_CASE
const MAX_ATTEMPTS = 3
```

### 2. Error Handling

```sky
enum Result
  Success(string)
  Error(string)
end

function read_file(path)
  if fs_exists(path)
    let content = fs_read_text(path)
    return Success(content)
  else
    return Error("File not found")
  end
end
```

### 3. Documentation

```sky
# Get user information
#
# Parameters:
#   id: User ID
#
# Returns:
#   User information or nil
function get_user(id)
  # ...
end
```

---

## 🚀 Example Project

```sky
# main.sky - Simple web scraper

import http
import json

async function fetch_page(url)
  let response = await http.get(url)
  return response.body
end

async function process_data(html)
  # HTML processing
  let data = {"title": "Example", "content": html}
  return data
end

async function save(data)
  let json_str = json_encode(data)
  fs_write_text("result.json", json_str)
  print("Saved!")
end

async function main
  print("Fetching page...")
  let html = await fetch_page("https://example.com")
  
  print("Processing data...")
  let data = await process_data(html)
  
  print("Saving...")
  await save(data)
  
  print("Done!")
end
```

---

## 📖 More Information

- **GitHub**: https://github.com/mburakmmm/sky-lang
- **Examples**: `examples/` directory
- **Tests**: `tests/` directory

---

**SKY Programming Language** - Fast, Safe, Easy 🚀

