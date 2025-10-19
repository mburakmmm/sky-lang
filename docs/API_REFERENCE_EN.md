# 📚 SKY Programming Language - API Reference (English)

## 🌟 Introduction

SKY is a modern, type-safe programming language written in Go. It combines Python's simplicity with Go's performance.

---

## 📖 Table of Contents

1. [Basic Syntax](#basic-syntax)
2. [Data Types](#data-types)
3. [Functions](#functions)
4. [Control Structures](#control-structures)
5. [Classes and OOP](#classes-and-oop)
6. [Async/Await](#asyncawait)
7. [Pattern Matching](#pattern-matching)
8. [Standard Library](#standard-library)
9. [Package Management](#package-management)

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
print(names.upper())  # Convert list to uppercase
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

### Multiple Async Operations

```sky
async function getUser(id)
  return {"id": str(id), "name": "User" + str(id)}
end

async function main
  let user1 = await getUser(1)
  let user2 = await getUser(2)
  
  print("User 1:", user1)
  print("User 2:", user2)
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

## 📚 Standard Library

### File Operations (FS)

```sky
# Write file
fs_write_text("test.txt", "Hello World")

# Read file
let content = fs_read_text("test.txt")
print(content)

# Check if file exists
if fs_exists("test.txt")
  print("File exists")
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

### Cryptography (Crypto)

```sky
# MD5 hash
let hash_md5 = crypto_md5("password123")
print("MD5:", hash_md5)

# SHA256 hash
let hash_sha256 = crypto_sha256("password123")
print("SHA256:", hash_sha256)
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

```sky
# Current timestamp
let now = time_now()
print("Timestamp:", now)

# Sleep (milliseconds)
time_sleep(1000)  # Sleep for 1 second
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

