# 📚 SKY Standard Library Roadmap

## ✅ ZATEN MEVCUT (Skip)

### Built-in Functions (43 adet)
- ✅ `print()`, `len()`, `type()`, `int()`, `float()`, `bool()`, `str()`
- ✅ `list()`, `dict()`, `sum()`, `input()`, `isinstance()`
- ✅ `map()`, `filter()`, `any()`, `all()`

### String Methods
- ✅ `upper()`, `lower()`, `split()`, `join()`, `strip()`, `lstrip()`, `rstrip()`
- ✅ `replace()`, `find()`, `count()`, `startswith()`, `endswith()`
- ✅ `contains()`, `index()`, `slice()`

### List Methods
- ✅ `append()`, `extend()`, `insert()`, `pop()`, `remove()`, `clear()`
- ✅ `reverse()`, `sort()`, `copy()`, `index()`, `count()`

### Dict Methods
- ✅ `keys()`, `values()`, `items()`, `get()`, `update()`, `clear()`, `copy()`

### Runtime Infrastructure
- ✅ GC (concurrent mark-and-sweep)
- ✅ Channels, Actors, Select
- ✅ Async/Await, Event Loop
- ✅ FFI (libffi integration)

---

## 🎯 TODO: STDLIB IMPLEMENTATION

### Strateji: %80 Sky + %20 Go
- **Sky**: High-level algorithms, business logic, user-facing APIs
- **Go**: System calls, performance-critical, security-critical

---

# 📦 TODO LIST (Öncelik Sırasına Göre)

## 🌑 PHASE 1: CORE ESSENTIALS (2-3 hafta)

### 1.1 core.types (Sky %100)
**Yer**: `std/core/types.sky`

```sky
# Option[T] implementation
enum Option[T]
  Some(T)
  None
end

# Result[T, E] implementation  
enum Result[T, E]
  Ok(T)
  Err(E)
end
```

**TODO**:
- [ ] Option[T] enum tanımı
- [ ] Result[T, E] enum tanımı
- [ ] Helper functions (is_some, is_none, unwrap, unwrap_or)
- [ ] Tests

---

### 1.2 core.error (Go %100)
**Yer**: `internal/runtime/skylib/error.go`

```go
// Error types ve runtime exception handling
type Error interface {
    Error() string
    Stacktrace() []string
}

type IOError struct { ... }
type ValueError struct { ... }
```

**TODO**:
- [ ] Error interface
- [ ] IOError, ValueError, KeyError, TypeError, IndexError, RuntimeError
- [ ] Stacktrace capture
- [ ] Bridge to Sky

---

### 1.3 collections.list (Sky %90 + Go %10)
**Yer**: `std/collections/list.sky` + `internal/runtime/skylib/list_core.go`

**Sky'da** (High-level algorithms):
```sky
function unique(list: [any]): [any]
function group_by(list: [any], fn: function): {any: [any]}
function zip(list1: [any], list2: [any]): [[any]]
function enumerate(list: [any]): [[int, any]]
function reduce(list: [any], fn: function, initial: any): any
function flatten(list: [[any]]): [any]
```

**Go'da** (Performance-critical):
```go
// Fast sort algorithms
func SortInPlace(list []interface{}, compareFn func(a, b interface{}) int)
```

**TODO**:
- [ ] unique() - Sky
- [ ] group_by() - Sky
- [ ] zip() - Sky
- [ ] enumerate() - Sky  
- [ ] reduce() - Sky
- [ ] flatten() - Sky
- [ ] Fast sort backend - Go

---

### 1.4 collections.set (Sky %100)
**Yer**: `std/collections/set.sky`

```sky
class Set
  function add(x)
  function remove(x)
  function union(other)
  function intersection(other)
  function difference(other)
end
```

**TODO**:
- [ ] Set class implementation
- [ ] All set operations
- [ ] Tests

---

### 1.5 collections.dict (Sky %80 + Go %20)
**Yer**: `std/collections/dict.sky`

**Sky'da**:
```sky
function merge(d1: {any: any}, d2: {any: any}, prefer: string): {any: any}
function invert(dict: {any: any}): {any: any}
function map_values(dict: {any: any}, fn: function): {any: any}
function filter_keys(dict: {any: any}, fn: function): {any: any}
```

**TODO**:
- [ ] merge() - Sky
- [ ] invert() - Sky
- [ ] map_values() - Sky
- [ ] filter_keys() - Sky

---

### 1.6 string (Sky %70 + Go %30)
**Yer**: `std/string/string.sky` + `internal/runtime/skylib/string_core.go`

**Sky'da** (Already mostly done, add missing):
```sky
function capitalize(s: string): string
function title(s: string): string
function swapcase(s: string): string
function is_alpha(s: string): bool
function is_digit(s: string): bool
function is_space(s: string): bool
function reverse(s: string): string
function repeat(s: string, n: int): string
```

**Go'da** (Unicode, regex):
```go
func IsAlpha(s string) bool
func IsDigit(s string) bool
func Normalize(s string, form string) string
```

**TODO**:
- [ ] capitalize() - Sky
- [ ] title() - Sky
- [ ] swapcase() - Sky
- [ ] is_alpha/digit/space() - Go (Unicode-aware)
- [ ] reverse() - Sky
- [ ] repeat() - Sky

---

### 1.7 iter (Sky %100)
**Yer**: `std/iter/iter.sky`

```sky
class Iter
  function take(n: int): Iter
  function drop(n: int): Iter
  function chain(other: Iter): Iter
  function cycle(): Iter
end
```

**TODO**:
- [ ] Lazy iterator class
- [ ] take(), drop(), chain(), cycle()
- [ ] Generator protocol
- [ ] Tests

---

## 🧮 PHASE 2: MATH & UTILITIES (1-2 hafta)

### 2.1 math (Sky %60 + Go %40)
**Yer**: `std/math/math.sky` + `internal/runtime/skylib/math_core.go`

**Sky'da** (Wrappers):
```sky
const PI = 3.141592653589793
const E = 2.718281828459045
const TAU = 6.283185307179586

function clamp(x: float, min: float, max: float): float
function sign(x: float): int
```

**Go'da** (Native math):
```go
func Sin(x float64) float64
func Cos(x float64) float64
func Sqrt(x float64) float64
func Pow(x, y float64) float64
```

**TODO**:
- [ ] Constants - Sky
- [ ] Wrappers for math.h - Go
- [ ] clamp, sign - Sky
- [ ] Tests

---

### 2.2 rand (Go %100)
**Yer**: `internal/runtime/skylib/rand.go`

```go
func Seed(seed int64)
func IntN(max int) int
func Float() float64
func Choice(seq []interface{}) interface{}
func Shuffle(list []interface{})
func UUID() string
```

**TODO**:
- [ ] All rand functions - Go
- [ ] Crypto-safe random option
- [ ] UUID generation
- [ ] Bridge to Sky

---

### 2.3 time & datetime (Go %80 + Sky %20)
**Yer**: `internal/runtime/skylib/time.go` + `std/datetime/datetime.sky`

**Go'da**:
```go
func Now() int64
func Sleep(ms int)
func Measure(fn func()) time.Duration
```

**Sky'da** (Formatting):
```sky
function format(timestamp: int, fmt: string): string
function parse(s: string, fmt: string): int
```

**TODO**:
- [ ] Time primitives - Go
- [ ] DateTime formatting - Sky
- [ ] Timezone support - Go
- [ ] Duration arithmetic - Sky

---

## 🖥️ PHASE 3: SYSTEM & I/O (2-3 hafta)

### 3.1 os (Go %100)
**Yer**: `internal/runtime/skylib/os.go`

```go
func GetEnv(key string) string
func SetEnv(key, value string)
func Getcwd() string
func Chdir(path string) error
func CPUCount() int
func Platform() string
func Exec(cmd string, args []string) (string, error)
```

**TODO**:
- [ ] Environment variables
- [ ] Process info (PID, etc)
- [ ] Platform detection
- [ ] Process execution
- [ ] Exit codes

---

### 3.2 path (Sky %100)
**Yer**: `std/path/path.sky`

```sky
function join(parts: [string]): string
function basename(path: string): string
function dirname(path: string): string
function extname(path: string): string
function normalize(path: string): string
function is_abs(path: string): bool
```

**TODO**:
- [ ] All path operations - Sky (string manipulation)
- [ ] Platform-aware path separators
- [ ] Tests (Unix + Windows paths)

---

### 3.3 fs (Go %100)
**Yer**: `internal/runtime/skylib/fs.go`

```go
func Exists(path string) bool
func IsFile(path string) bool
func IsDir(path string) bool
func ReadText(path string) (string, error)
func WriteText(path, data string) error
func ReadBytes(path string) ([]byte, error)
func WriteBytes(path string, data []byte) error
func Mkdir(path string, recursive bool) error
func Remove(path string, recursive bool) error
func Rename(src, dst string) error
func ListDir(path string) ([]string, error)
```

**TODO**:
- [ ] All fs operations - Go
- [ ] Permission handling
- [ ] Symlink support
- [ ] Bridge to Sky

---

### 3.4 io (Go %80 + Sky %20)
**Yer**: `internal/runtime/skylib/io.go`

```go
type Reader interface { Read([]byte) (int, error) }
type Writer interface { Write([]byte) (int, error) }
type BufReader struct { ... }
type BufWriter struct { ... }
```

**TODO**:
- [ ] Reader/Writer interfaces - Go
- [ ] Buffered I/O - Go
- [ ] stdin/stdout/stderr - Go
- [ ] Sky wrappers

---

## 🌐 PHASE 4: NETWORKING (3-4 hafta)

### 4.1 net (Go %100)
**Yer**: `internal/runtime/skylib/net.go`

```go
func Resolve(host string) ([]string, error)
func TCPConnect(addr string) (net.Conn, error)
func TCPListen(addr string) (net.Listener, error)
type UDPSocket struct { ... }
```

**TODO**:
- [ ] DNS resolution
- [ ] TCP client/server
- [ ] UDP sockets
- [ ] Connection pooling

---

### 4.2 http (Go %50 + Sky %50)
**Yer**: `internal/runtime/skylib/http_core.go` + `std/http/http.sky`

**Go'da** (Core):
```go
func HTTPGet(url string, headers map[string]string) (*Response, error)
func HTTPPost(url string, body []byte, headers map[string]string) (*Response, error)
func StartServer(addr string, handler func(*Request) *Response) error
```

**Sky'da** (API):
```sky
class Client
  async function get(url: string, headers: {string: string}?): Response
  async function post(url: string, data: any, headers: {string: string}?): Response
end

class Server
  function route(path: string, handler: function)
  async function listen(port: int)
end
```

**TODO**:
- [ ] HTTP client - Go core + Sky wrapper
- [ ] HTTP server - Go core + Sky router
- [ ] Middleware system - Sky
- [ ] Cookies, headers, status codes
- [ ] Streaming support

---

### 4.3 socket (Go %100)
**Yer**: `internal/runtime/skylib/socket.go`

```go
type Socket struct { ... }
func (s *Socket) Bind(addr string) error
func (s *Socket) Listen(backlog int) error
func (s *Socket) Accept() (*Socket, error)
func (s *Socket) Connect(addr string) error
func (s *Socket) Send(data []byte) (int, error)
func (s *Socket) Recv(n int) ([]byte, error)
```

**TODO**:
- [ ] Low-level socket API - Go
- [ ] Unix sockets support
- [ ] Socket options (SO_REUSEADDR, etc)

---

## ⚙️ PHASE 5: ASYNC ECOSYSTEM (1-2 hafta)

### 5.1 async utilities (Sky %100)
**Yer**: `std/async/async.sky`

```sky
async function gather(tasks: [Task]): [any]
async function race(tasks: [Task]): any
async function with_timeout(ms: int, fn: function): any
async function retry(fn: function, max_attempts: int, delay: int): any
```

**TODO**:
- [ ] gather() - Wait for all
- [ ] race() - First to complete
- [ ] with_timeout() - Timeout wrapper
- [ ] retry() - Retry with backoff
- [ ] Tests

---

### 5.2 task management (Sky %80 + Go %20)
**Yer**: `std/task/task.sky`

**Already have**: CancellationToken, TaskTree (Go)

**Add** (Sky):
```sky
function current_task(): Task
function cancel_all()
function wait_any(tasks: [Task]): any
```

**TODO**:
- [ ] current_task() - Sky
- [ ] Task metadata - Sky
- [ ] Integration with existing cancellation

---

## 🔐 PHASE 6: SECURITY (2-3 hafta)

### 6.1 crypto.hash (Go %100)
**Yer**: `internal/runtime/skylib/crypto_hash.go`

```go
func MD5(data []byte) string
func SHA256(data []byte) string
func SHA512(data []byte) string
func HMAC(hash string, key, msg []byte) []byte
```

**TODO**:
- [ ] All hash functions - Go
- [ ] HMAC - Go
- [ ] Streaming hashing
- [ ] Bridge to Sky

---

### 6.2 crypto.enc (Go %100)
**Yer**: `internal/runtime/skylib/crypto_enc.go`

```go
func AESGCMEncrypt(key, plaintext []byte) ([]byte, error)
func AESGCMDecrypt(key, ciphertext []byte) ([]byte, error)
func ChaCha20Encrypt(key, nonce, plaintext []byte) []byte
func RandBytes(n int) []byte
```

**TODO**:
- [ ] AES-GCM - Go
- [ ] ChaCha20-Poly1305 - Go
- [ ] Crypto-safe random - Go
- [ ] Key derivation (PBKDF2, Argon2)

---

## 🧩 PHASE 7: ENCODING & COMPRESSION (2 hafta)

### 7.1 encoding.json (Go %70 + Sky %30)
**Yer**: `internal/runtime/skylib/json_core.go` + `std/json/json.sky`

**Go'da** (Parser):
```go
func ParseJSON(data []byte) (interface{}, error)
func StringifyJSON(obj interface{}) (string, error)
func ParseJSONPretty(data []byte) (interface{}, error)
```

**Sky'da** (Convenience):
```sky
function parse_file(path: string): any
function dump_file(path: string, obj: any, pretty: bool)
```

**TODO**:
- [ ] Fast JSON parser - Go
- [ ] Pretty printing - Go
- [ ] File helpers - Sky
- [ ] Streaming JSON - Go

---

### 7.2 encoding.yaml/toml/csv (Go %80 + Sky %20)
**Yer**: `internal/runtime/skylib/encoding.go`

**TODO**:
- [ ] YAML parser - Go
- [ ] TOML parser - Go
- [ ] CSV parser - Go
- [ ] Sky wrappers

---

### 7.3 compression (Go %100)
**Yer**: `internal/runtime/skylib/compression.go`

```go
func GzipCompress(data []byte) ([]byte, error)
func GzipDecompress(data []byte) ([]byte, error)
func ZstdCompress(data []byte, level int) ([]byte, error)
func ZstdDecompress(data []byte) ([]byte, error)
```

**TODO**:
- [ ] Gzip - Go
- [ ] Zstd - Go
- [ ] Zip (archive) - Go
- [ ] Bridge to Sky

---

## 🧠 PHASE 8: DEVELOPER TOOLS (2 hafta)

### 8.1 testing (Sky %100)
**Yer**: `std/testing/testing.sky`

```sky
function assert_eq(a: any, b: any, msg: string?)
function assert_ne(a: any, b: any, msg: string?)
function assert_true(cond: bool, msg: string?)
function assert_raises(fn: function, error_type: type, msg: string?)
function bench(name: string, fn: function)
```

**TODO**:
- [ ] Assertion framework - Sky
- [ ] Test discovery - Sky
- [ ] Benchmark harness - Sky
- [ ] Test reporter - Sky

---

### 8.2 log (Go %60 + Sky %40)
**Yer**: `internal/runtime/skylib/log.go` + `std/log/log.sky`

**Go'da**:
```go
type Logger struct { ... }
func (l *Logger) Info(msg string, fields map[string]interface{})
func (l *Logger) Error(msg string, fields map[string]interface{})
```

**Sky'da**:
```sky
function info(msg: string, **kwargs)
function error(msg: string, **kwargs)
function with_fields(fields: {string: any}): Logger
```

**TODO**:
- [ ] Structured logging - Go
- [ ] Log levels - Go
- [ ] File rotation - Go
- [ ] Sky API wrappers

---

### 8.3 fmt (Sky %100)
**Yer**: `std/fmt/fmt.sky`

```sky
function format(template: string, **kwargs): string
function sprintf(fmt: string, *args): string
```

**TODO**:
- [ ] String interpolation - Sky
- [ ] printf-style formatting - Sky
- [ ] Tests

---

## 🧬 PHASE 9: META & REFLECTION (2-3 hafta)

### 9.1 reflect (Go %100)
**Yer**: `internal/runtime/skylib/reflect.go`

```go
func TypeName(obj interface{}) string
func Fields(obj interface{}) []string
func Methods(obj interface{}) []string
func GetAttr(obj interface{}, name string) (interface{}, error)
func SetAttr(obj interface{}, name string, value interface{}) error
func Invoke(obj interface{}, method string, args []interface{}) (interface{}, error)
```

**TODO**:
- [ ] Runtime type inspection - Go
- [ ] Dynamic method invocation - Go
- [ ] Bridge to Sky

---

### 9.2 unicode (Go %100)
**Yer**: `internal/runtime/skylib/unicode.go`

```go
func Normalize(s string, form string) string
func IsLetter(r rune) bool
func IsDigit(r rune) bool
func Graphemes(s string) []string
func Width(s string) int
```

**TODO**:
- [ ] Unicode normalization - Go
- [ ] Character classification - Go
- [ ] Grapheme clusters - Go
- [ ] East Asian Width - Go

---

## 💡 PHASE 10: EXTENDED FEATURES (Future)

### 10.1 db.sqlite (Go %80 + Sky %20)
**TODO** (Lower priority):
- [ ] SQLite binding - Go
- [ ] Query builder - Sky

### 10.2 cli (Sky %100)
**TODO**:
- [ ] Argument parsing - Sky
- [ ] Flag parsing - Sky
- [ ] Interactive prompts - Sky

### 10.3 regex (Go %100)
**TODO**:
- [ ] Regex engine - Go (use Go's regexp)
- [ ] Pattern compilation
- [ ] Match, find, replace

---

# 📊 SUMMARY

## Phase Breakdown

| Phase | Modules | Sky% | Go% | Duration | Priority |
|-------|---------|------|-----|----------|----------|
| **P1** | Core | 85% | 15% | 2-3w | 🔴 Critical |
| **P2** | Math | 60% | 40% | 1-2w | 🔴 Critical |
| **P3** | System | 20% | 80% | 2-3w | 🔴 Critical |
| **P4** | Network | 40% | 60% | 3-4w | 🟡 High |
| **P5** | Async | 90% | 10% | 1-2w | 🟡 High |
| **P6** | Security | 0% | 100% | 2-3w | 🟡 High |
| **P7** | Encoding | 30% | 70% | 2w | 🟢 Medium |
| **P8** | DevTools | 80% | 20% | 2w | 🟢 Medium |
| **P9** | Meta | 10% | 90% | 2-3w | 🔵 Low |
| **P10** | Extended | 60% | 40% | TBD | 🔵 Future |

## Overall Distribution

- **Sky Implementation**: ~78% (mostly algorithms, APIs, wrappers)
- **Go Implementation**: ~22% (system calls, performance, security)
- **Total Estimated**: 20-30 weeks for Phases 1-9
- **Total Modules**: ~50 modules

## Immediate Next Steps (Week 1)

1. ✅ Option[T] and Result[T, E] enums (with existing enum parser)
2. ✅ collections.set class
3. ✅ iter module (lazy evaluation)
4. ✅ path module (pure Sky)
5. ✅ testing framework

## Dependencies

- **Enum/Match interpreter integration** must be completed first for Option/Result
- **Go runtime bindings** needed for fs, net, crypto
- **Bridge mechanism** for Go↔Sky type conversions

---

**Status**: Ready to implement Phase 1 🚀

