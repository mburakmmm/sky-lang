# Sky Bridges - Python ve JS Köprüleri

## Genel Bakış

Sky, Python ve JavaScript ile köprü (bridge) sistemi sağlar. Bu köprüler, Sky kodundan Python ve JS kütüphanelerini kullanmayı mümkün kılar.

## Python Bridge

### Python Bridge Kurulumu
Python bridge, `pyo3` crate'i ile CPython'u gömülü olarak kullanır.

### Python Import
```sky
var math = python.import("math")
var os = python.import("os")
var sys = python.import("sys")
```

### Python Fonksiyon Çağrıları
```sky
var math = python.import("math")
float kök = math.sqrt(16)
float sin = math.sin(3.14159 / 2)
float log = math.log(10)
```

### Python Obje Erişimi
```sky
var sys = python.import("sys")
string versiyon = sys.version
list yollar = sys.path
```

### Python Bridge Örnekleri

#### Matematik İşlemleri
```sky
var math = python.import("math")

# Temel matematik
float pi = math.pi
float e = math.e
float kök = math.sqrt(25)
float kuvvet = math.pow(2, 8)

# Trigonometri
float sin = math.sin(math.pi / 2)
float cos = math.cos(0)
float tan = math.tan(math.pi / 4)

# Logaritma
float log = math.log(10)
float log10 = math.log10(100)
float log2 = math.log2(8)
```

#### Dosya İşlemleri
```sky
var os = python.import("os")

# Dosya sistemi
string cwd = os.getcwd()
list dosyalar = os.listdir(".")
bool var_mı = os.path.exists("test.txt")

# Environment variables
string home = os.environ.get("HOME")
string path = os.environ.get("PATH")
```

#### JSON İşlemleri
```sky
var json = python.import("json")

# JSON parse
string json_str = '{"ad": "Sky", "versiyon": 1}'
map obj = json.loads(json_str)

# JSON stringify
map data = {"ad": "Sky", "versiyon": 1}
string json_result = json.dumps(data)
```

### Python Bridge Tip Dönüşümleri

| Python Tipi | Sky Tipi | Notlar |
|-------------|----------|---------|
| `int` | `int` | Direct mapping |
| `float` | `float` | Direct mapping |
| `bool` | `bool` | Direct mapping |
| `str` | `string` | Direct mapping |
| `list` | `list` | Recursive conversion |
| `dict` | `map` | Recursive conversion |
| `None` | `null` | Direct mapping |
| `tuple` | `list` | Tuple to list conversion |
| `set` | `list` | Set to list conversion |

### Python Bridge Hata Yönetimi
```sky
var math = python.import("math")

try
  float kök = math.sqrt(-1)  # ValueError
  print(kök)
except PythonError as e
  print("Python hatası: " + e.message)
```

## JavaScript Bridge

### JavaScript Bridge Kurulumu
JavaScript bridge, `quickjs-rs` crate'i ile QuickJS'yi gömülü olarak kullanır.

### JavaScript Eval
```sky
var jsfn = js.eval("(x) => x * 2")
var js_obj = js.eval("{name: 'Sky', version: 1}")
var js_array = js.eval("[1, 2, 3, 4, 5]")
```

### JavaScript Fonksiyon Çağrıları
```sky
var jsfn = js.eval("(x) => x + 1")
int sonuç = jsfn(41)  # 42

var math_fn = js.eval("Math.sqrt")
float kök = math_fn(16)  # 4.0
```

### JavaScript Obje Erişimi
```sky
var obj = js.eval("{name: 'Sky', version: 1}")
string ad = obj.name
int versiyon = obj.version
```

### JavaScript Bridge Örnekleri

#### Matematik İşlemleri
```sky
# Math objesi
var sqrt = js.eval("Math.sqrt")
var pow = js.eval("Math.pow")
var sin = js.eval("Math.sin")
var cos = js.eval("Math.cos")

float kök = sqrt(25)
float kuvvet = pow(2, 8)
float sin_değer = sin(3.14159 / 2)
float cos_değer = cos(0)
```

#### Array İşlemleri
```sky
var array = js.eval("[1, 2, 3, 4, 5]")
var map_fn = js.eval("(arr) => arr.map(x => x * 2)")
var filter_fn = js.eval("(arr) => arr.filter(x => x > 2)")

list çiftler = map_fn(array)
list büyükler = filter_fn(array)
```

#### String İşlemleri
```sky
var upper = js.eval("(str) => str.toUpperCase()")
var lower = js.eval("(str) => str.toLowerCase()")
var split = js.eval("(str, sep) => str.split(sep)")

string büyük = upper("sky")
string küçük = lower("SKY")
list parçalar = split("a,b,c", ",")
```

#### Object İşlemleri
```sky
var obj = js.eval("{name: 'Sky', version: 1, features: ['async', 'coop']}")
var keys_fn = js.eval("Object.keys")
var values_fn = js.eval("Object.values")

list anahtarlar = keys_fn(obj)
list değerler = values_fn(obj)
```

### JavaScript Bridge Tip Dönüşümleri

| JavaScript Tipi | Sky Tipi | Notlar |
|----------------|----------|---------|
| `number` | `int`/`float` | Integer vs float detection |
| `boolean` | `bool` | Direct mapping |
| `string` | `string` | Direct mapping |
| `Array` | `list` | Recursive conversion |
| `Object` | `map` | Recursive conversion |
| `null` | `null` | Direct mapping |
| `undefined` | `null` | Undefined to null |
| `Function` | `NativeFn` | Function wrapper |

### JavaScript Bridge Hata Yönetimi
```sky
var jsfn = js.eval("(x) => x.nonexistent.property")

try
  var sonuç = jsfn(42)  # TypeError
  print(sonuç)
except JSError as e
  print("JS hatası: " + e.message)
```

## Bridge Performans

### Python Bridge
- **GIL**: Global Interpreter Lock etkisi
- **Memory**: Python runtime memory usage
- **Serialization**: Object conversion overhead
- **Error Handling**: Exception propagation

### JavaScript Bridge
- **Context**: QuickJS context creation
- **Memory**: JavaScript runtime memory usage
- **Serialization**: Value conversion overhead
- **Error Handling**: Error object conversion

## Bridge Güvenlik

### Python Bridge
- **Sandboxing**: Python code execution limits
- **Import Restrictions**: Allowed modules whitelist
- **Resource Limits**: Memory and CPU limits
- **Error Isolation**: Python errors don't crash Sky

### JavaScript Bridge
- **Sandboxing**: JavaScript execution limits
- **Global Access**: Limited global object access
- **Resource Limits**: Memory and CPU limits
- **Error Isolation**: JS errors don't crash Sky

## Bridge Best Practices

### 1. Error Handling
```sky
# Python bridge
try
  var result = python.import("nonexistent")
except PythonError as e
  print("Python import hatası: " + e.message)

# JS bridge
try
  var result = js.eval("invalid syntax")
except JSError as e
  print("JS eval hatası: " + e.message)
```

### 2. Resource Management
```sky
# Python bridge - context manager
var context = python.context()
try
  var result = context.eval("expensive_operation()")
finally
  context.cleanup()

# JS bridge - context management
var js_context = js.context()
try
  var result = js_context.eval("heavy_computation()")
finally
  js_context.cleanup()
```

### 3. Type Safety
```sky
# Tip kontrolü
var math = python.import("math")
var result = math.sqrt(16)

if result.is_float()
  float kök = result
  print("Kök: " + kök)
else
  print("Beklenmeyen tip: " + result.type())
```

## Bridge Test Örnekleri

### Python Bridge Tests
```sky
# Import testi
var math = python.import("math")
assert math != null

# Function call testi
float pi = math.pi
assert pi > 3.14

# Error handling testi
try
  var result = math.sqrt(-1)
  assert false  # Buraya gelmemeli
except PythonError as e
  assert e.message.contains("math domain error")
```

### JavaScript Bridge Tests
```sky
# Eval testi
var fn = js.eval("(x) => x * 2")
assert fn != null

# Function call testi
int result = fn(21)
assert result == 42

# Error handling testi
try
  var result = js.eval("undefined_variable")
  assert false  # Buraya gelmemeli
except JSError as e
  assert e.message.contains("undefined_variable")
```

## Bridge Sınırlamaları

### MVP Sınırlamaları
- **Limited Types**: Sadece temel tip dönüşümleri
- **No Async**: Bridge'ler async değil
- **No Coroutines**: Bridge'ler coroutine desteklemiyor
- **Memory Leaks**: Bridge cleanup gerekli

### Gelecek Özellikler
- **Async Bridge**: Async Python/JS functions
- **Coroutine Bridge**: Coroutine support
- **Advanced Types**: Complex object conversion
- **Performance Optimization**: Faster serialization

## Bridge Kullanım Senaryoları

### Python Bridge
- **Scientific Computing**: NumPy, SciPy, Pandas
- **Machine Learning**: TensorFlow, PyTorch, Scikit-learn
- **Data Processing**: JSON, CSV, XML parsing
- **System Integration**: OS, file system, network

### JavaScript Bridge
- **Frontend Integration**: DOM manipulation, Canvas
- **Data Processing**: Array operations, Object manipulation
- **String Processing**: Regex, text manipulation
- **Web APIs**: Fetch, WebSocket, LocalStorage

Bridge sistemi, Sky dilinin Python ve JavaScript ekosistemlerine erişim sağlamasını mümkün kılar.
