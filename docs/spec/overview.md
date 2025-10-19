# SKY Programming Language - Overview

## Felsefe ve Tasarım Hedefleri

SKY, modern yazılım geliştirme ihtiyaçlarını karşılamak için tasarlanmış, güvenli ve performanslı bir programlama dilidir.

### Temel Prensipler

1. **Güvenlik Önceliklidir**: Varsayılan olarak güvenli, ancak `unsafe` blokları ile performans-kritik kod yazma esnekliği
2. **Okunabilirlik**: Python-benzeri temiz sözdizimi, girintileme tabanlı bloklar
3. **Performans**: LLVM JIT ve AOT derlemesi ile native hız
4. **İsteğe Bağlı Tipler**: Tip çıkarımı ve isteğe bağlı statik tipler
5. **Modern Eşzamanlılık**: Async/await ve coroutines desteği
6. **Pratiklik**: Kolay FFI ile C kütüphaneleri entegrasyonu

## Çalışma Modeli

### Execution Model

SKY iki çalışma modunu destekler:

#### 1. JIT (Just-In-Time) Compilation
```bash
sky run myprogram.sky
```
- Anlık derleme ve çalıştırma
- Hızlı geliştirme döngüsü
- REPL ve interaktif kullanım için ideal
- LLVM JIT engine kullanır

#### 2. AOT (Ahead-Of-Time) Compilation
```bash
sky build myprogram.sky -o myprogram
```
- Native binary oluşturma
- Maksimum optimizasyon (O2/O3)
- Production deployment için ideal
- Platform-specific binary

### Memory Model

#### Garbage Collection

- **Eşzamanlı GC**: Go-benzeri concurrent garbage collector
- **Düşük Duraksamalı**: Incremental mark-and-sweep
- **Write Barriers**: Tri-color marking algorithm
- **Generational GC**: İlerideki optimizasyon hedefi

#### Unsafe Bloklar

`unsafe` blokları içinde:
- GC devre dışı bırakılır (veya local suspend)
- Ham pointer işlemleri izin verilir
- Adres aritmetiği mümkündür
- Performans-kritik kod için tasarlanmıştır

```sky
function lowLevelOperation
  unsafe
    let ptr = malloc(1024)
    # Ham pointer işlemleri
    free(ptr)
  end
end
```

## Tip Sistemi

### Temel Tipler

- `int` - 64-bit signed integer
- `float` - 64-bit floating point (IEEE 754)
- `string` - UTF-8 encoded string
- `bool` - Boolean (true/false)
- `any` - Dynamic type

### Tip Çıkarımı

SKY güçlü tip çıkarımı sunar:

```sky
let x = 10           # x: int
let y = 3.14         # y: float
let s = "hello"      # s: string
let b = true         # b: bool
```

### İsteğe Bağlı Tip Anotasyonları

Daha açık kod için tipler belirtilebilir:

```sky
let x: int = 10
let y: float = 3.14

function add(a: int, b: int): int
  return a + b
end
```

### Koleksiyon Tipleri

```sky
# List
let numbers: [int] = [1, 2, 3, 4, 5]

# Dictionary
let scores: {string: int} = {
  "Alice": 95,
  "Bob": 87
}
```

### Fonksiyon Tipleri

```sky
# Fonksiyon tipi: (int, int) => int
let add: (int, int) => int = function(a, b) => a + b
```

## Kontrol Yapıları

### If-Elif-Else

```sky
if x < 10
  print("small")
elif x < 100
  print("medium")
else
  print("large")
end
```

### While Döngüsü

```sky
while x > 0
  print(x)
  x = x - 1
end
```

### For Döngüsü

```sky
for item in collection
  print(item)
end

for i in range(10)
  print(i)
end
```

## Fonksiyonlar

### Basit Fonksiyonlar

```sky
function greet(name)
  print("Hello, " + name)
end

function add(a: int, b: int): int
  return a + b
end
```

### Varsayılan Parametreler

```sky
function greet(name, greeting = "Hello")
  print(greeting + ", " + name)
end
```

### Lambda Fonksiyonlar

```sky
let double = function(x) => x * 2
let add = function(a, b) => a + b
```

### Higher-Order Functions

```sky
function map(list, fn)
  let result = []
  for item in list
    result.append(fn(item))
  end
  return result
end

let numbers = [1, 2, 3, 4, 5]
let doubled = map(numbers, function(x) => x * 2)
```

## Async/Await

Modern eşzamansız programlama:

```sky
async function fetchData(url: string): string
  let response = await http_get(url)
  return response.body
end

async function processMultiple
  let data1 = await fetchData("https://api1.example.com")
  let data2 = await fetchData("https://api2.example.com")
  
  print("Data 1: " + data1)
  print("Data 2: " + data2)
end

function main
  await processMultiple()
end
```

## Coroutines

Hafif eşzamanlılık için generators:

```sky
coop function fibonacci(n: int)
  let a = 0
  let b = 1
  
  for i in range(n)
    yield a
    let temp = a
    a = b
    b = temp + b
  end
end

function main
  for num in fibonacci(10)
    print(num)
  end
end
```

## Sınıflar (OOP)

```sky
class Animal
  let name: string
  
  function init(name: string)
    self.name = name
  end
  
  function speak
    print("...")
  end
end

class Dog(Animal)
  function speak
    print(self.name + " says: Woof!")
  end
end

function main
  let dog = Dog("Buddy")
  dog.speak()  # Buddy says: Woof!
end
```

## Modül Sistemi

### Import

```sky
import math
import http
import mymodule as mm

function main
  let result = math.sqrt(16)
  print(result)
end
```

### Private Members

`_` öneki ile private üyeler:

```sky
# mymodule.sky
let _privateVar = 42      # Private
let publicVar = 100       # Public

function _privateFunc     # Private
  print("private")
end

function publicFunc       # Public
  print("public")
end
```

## FFI (Foreign Function Interface)

C kütüphaneleri ile kolay entegrasyon:

```sky
import ffi

function useC
  let libc = ffi.load("libc.so")
  let strlen = libc.symbol("strlen")
  
  let result = strlen("hello world")
  print("Length: " + result)  # 11
end
```

## Unsafe Bloklar

Düşük seviye işlemler için:

```sky
import ffi

function directMemory
  unsafe
    # malloc/free kullanımı
    let ptr = ffi.malloc(1024)
    
    # Ham pointer işlemleri
    let byte_ptr = ptr as *byte
    byte_ptr[0] = 65  # 'A'
    
    # Belleği serbest bırak
    ffi.free(ptr)
  end
end
```

## Error Handling

```sky
function divide(a: int, b: int): int
  if b == 0
    panic("division by zero")
  end
  return a / b
end

function safeDivide(a: int, b: int)
  try
    return divide(a, b)
  catch error
    print("Error: " + error)
    return 0
  end
end
```

## Toolchain

### sky - Ana Derleyici
- `sky run` - JIT execution
- `sky build` - AOT compilation
- `sky test` - Test runner
- `sky repl` - Interactive REPL
- `sky dump` - Diagnostics
- `sky check` - Type checker

### wing - Package Manager
- `wing init` - Yeni proje
- `wing install` - Paket kurulum
- `wing update` - Güncelleme
- `wing build` - Proje build
- `wing publish` - Paket yayınlama

### skyls - Language Server
- LSP protocol implementation
- Editor entegrasyonu
- Auto-completion, go-to-definition, hover, vb.

### skydbg - Debugger
- LLDB/GDB köprüsü
- Breakpoints
- Step debugging
- Variable inspection

## Performance Characteristics

### JIT Compilation
- İlk çalıştırma: ~100ms overhead
- Sonraki çalıştırmalar: Native hız
- Dinamik optimizasyon

### AOT Compilation
- Derleme süresi: Orta
- Çalışma hızı: Native C/C++ seviyesi
- Binary boyutu: Optimize edilmiş

### Memory
- GC overhead: %5-10
- Pause times: < 10ms
- Memory footprint: Düşük

## Karşılaştırma

| Özellik | SKY | Python | Go | Rust |
|---------|-----|--------|----|----- |
| Syntax | ✅ Kolay | ✅ Kolay | ⚠️ Orta | ❌ Zor |
| Performance | ✅ Yüksek | ❌ Düşük | ✅ Yüksek | ✅ Yüksek |
| Safety | ✅ İyi | ⚠️ Orta | ✅ İyi | ✅ Mükemmel |
| Async/Await | ✅ Native | ✅ Native | ❌ Yok | ✅ Native |
| GC | ✅ Concurrent | ✅ Yes | ✅ Concurrent | ❌ No GC |
| FFI | ✅ Kolay | ✅ Kolay | ⚠️ Cgo | ⚠️ Unsafe |
| Learning Curve | ✅ Düşük | ✅ Düşük | ⚠️ Orta | ❌ Yüksek |

## Gelecek Roadmap

### v0.2.0
- [ ] Jenerik tipler
- [ ] Trait/Interface sistemi
- [ ] Pattern matching
- [ ] Improved error handling

### v0.3.0
- [ ] Package registry
- [ ] Standard library genişletme
- [ ] Optimization passes
- [ ] Better debugging tools

### v0.4.0
- [ ] WASM target
- [ ] Cross-compilation
- [ ] Hot reload
- [ ] Makro sistemi

## Topluluk ve Katkı

SKY açık kaynak bir projedir ve katkılara açıktır.

- **GitHub**: https://github.com/mburakmmm/sky-lang
- **Documentation**: https://sky-lang.org/docs
- **Discord**: https://discord.gg/skylang

## Lisans

MIT License - Detaylar için LICENSE dosyasına bakın.

