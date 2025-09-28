# Sky Programlama Dili

Sky, Python-benzeri girintili sözdizimi ve basit tasarımı olan bir programlama dilidir. Dinamik & güçlü tipli çalışma zamanı, async/await, coroutines, VM bytecode ve Python/JS köprüleri destekler.

## Özellikler

- **Girintili sözdizimi**: Python-benzeri, basit ve okunaklı tasarım
- **Tip beyanı zorunlu**: `var`, `int`, `float`, `bool`, `string`, `list`, `map`
- **Dinamik & güçlü tip**: Runtime kontrolleri ile tip güvenliği
- **String interpolation**: Dart (`$var`) ve Python f-string (`f"{var}"`) desteği
- **Async/await**: Event loop tabanlı asenkron programlama
- **Coroutines**: İşbirlikçi çoklu görev (`coop function`, `yield`)
- **VM bytecode**: Stack-based bytecode interpreter
- **Garbage collection**: Mark-sweep GC
- **Python/JS köprüleri**: Python ve JavaScript kütüphanelerini kullanma
- **Unicode desteği**: Türkçe karakterli değişken adları

## Kurulum

### Gereksinimler

- Rust 1.79+
- Python 3.8+ (Python bridge için)
- Node.js (JavaScript bridge için)

### Derleme

```bash
git clone https://github.com/sky-lang/sky.git
cd sky
cargo build --release
```

### Kurulum

```bash
cargo install --path .
```

## Hızlı Başlangıç

### Hello World

```sky
print("Merhaba Sky!")
```

### Değişken Tanımlama

```sky
int sayı = 42
float pi = 3.14159
bool doğru = true
string mesaj = "Merhaba Sky"
list sayılar = [1, 2, 3]
map sözlük = {"ad": "Sky", "versiyon": 1}
var dinamik = 42  # var tipi her tür kabul eder

# String interpolation
string ad = "Sky"
string mesaj1 = "Merhaba $ad"  # Dart tarzı
string mesaj2 = f"Merhaba {ad}"  # Python f-string tarzı
```

### Fonksiyon Tanımlama

```sky
function selam(isim: string)
  return "Merhaba " + isim

async function indir(url: string)
  var data = await http.get(url)
  return data

coop function say(n: int)
  int i = 0
  while i < n
    yield i
    i = i + 1
  return n
```

### Kontrol Akışı

```sky
if x > 0
  print("Pozitif")
elif x < 0
  print("Negatif")
else
  print("Sıfır")

for elem: var in liste
  print(elem)

while x > 0
  x = x - 1
  print(x)
```

## CLI Kullanımı

### Dosya Çalıştırma

```bash
sky run program.sky
```

### REPL

```bash
sky repl
```

### Kod Biçimlendirme

```bash
sky fmt program.sky
```

### Syntax Kontrolü

```bash
sky check program.sky
```

## Örnekler

### Basit Program

```sky
int sayı = 42
string mesaj = "Merhaba Sky"
print(mesaj + " " + sayı)
```

### Async Programlama

```sky
async function indir(url: string)
  var data = await http.get(url)
  return data

var sonuç = indir("https://example.com")
print(await sonuç)
```

### Coroutines

```sky
coop function say(n: int)
  int i = 0
  while i < n
    yield i
    i = i + 1
  return n

var c = say(3)
print(c.resume())  # 0
print(c.resume())  # 1
print(c.resume())  # 2
print(c.is_done())  # true
```

### Python Bridge

```sky
var math = python.import("math")
float kök = math.sqrt(16)
print(kök)
```

### JavaScript Bridge

```sky
var jsfn = js.eval("(x)=>x*2")
int iki_kat = jsfn(21)
print(iki_kat)  # 42
```

### String Interpolation

```sky
string ad = "Sky"
int yaş = 1

# Dart tarzı interpolation
string mesaj1 = "Merhaba $ad, yaşın $yaş"
string mesaj2 = "Toplam: ${10 + 20}"

# Python f-string tarzı
string mesaj3 = f"Merhaba {ad}, yaşın {yaş + 1}"
string mesaj4 = f"Pi yaklaşık {3.14159} değerindedir"
```

## Dil Özellikleri

### Tip Sistemi

Sky'da **tip beyanı zorunludur**:

```sky
int x = 42        # Tamsayı
float y = 3.14    # Ondalıklı sayı
bool z = true     # Boolean
string s = "test" # Metin
list l = [1,2,3]  # Liste
map m = {"a": 1}  # Sözlük
var v = 42        # Dinamik tip
```

**Hata örnekleri:**
```sky
x = 3  # E0001: Missing type annotation
```

### Girinti Kuralları

- **Tab yasak**: Sadece boşluk kullanılır
- **Tutarlı girinti**: Aynı seviyede aynı sayıda boşluk
- **Blok belirleme**: Girinti ile bloklar belirlenir

```sky
function test()
  int x = 1
  if x > 0
    print("Pozitif")
    if x > 10
      print("Çok büyük")
  print("Bitti")
```

### Async/Await

```sky
async function indir(url: string)
  var data = await http.get(url)
  return data

var future = indir("https://example.com")
var sonuç = await future
```

### Coroutines

```sky
coop function say(n: int)
  int i = 0
  while i < n
    yield i
    i = i + 1
  return n

var c = say(3)
while not c.is_done()
  print(c.resume())
```

## Hata Kodları

- **E0001**: Missing type annotation
- **E0101**: Invalid indentation
- **E0201**: await outside async function
- **E0202**: yield outside coop function
- **E1001**: Type mismatch
- **E2001**: Coroutine already finished
- **E3001**: Python bridge error
- **E3002**: JS bridge error

## Dokümantasyon

- [Grammar](sky/docs/spec/grammar.md) - Dil yazımı ve sözdizimi
- [Coroutines](sky/docs/spec/coroutines.md) - Coop/yield semantiği
- [Bridges](sky/docs/spec/bridges.md) - Python/JS köprüleri
- [Strings](sky/docs/spec/strings.md) - String interpolation

## Örnekler

- [hello.sky](sky/examples/hello.sky) - Basit örnek
- [async.sky](sky/examples/async.sky) - Async/await örnekleri
- [coop_basic.sky](sky/examples/coop_basic.sky) - Coroutine örnekleri
- [py_bridge.sky](sky/examples/py_bridge.sky) - Python bridge örnekleri
- [js_bridge.sky](sky/examples/js_bridge.sky) - JavaScript bridge örnekleri
- [strings.sky](sky/examples/strings.sky) - String interpolation örnekleri

## Geliştirme

### Test Çalıştırma

```bash
cargo test
```

### Format

```bash
cargo fmt
```

### Lint

```bash
cargo clippy
```

## Lisans

MIT License - Detaylar için [LICENSE](LICENSE) dosyasına bakın.

## Katkıda Bulunma

1. Fork yapın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Commit yapın (`git commit -m 'Add amazing feature'`)
4. Push yapın (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

## Roadmap

- [x] Lexer ve Parser
- [x] VM ve Bytecode
- [x] Async/Await
- [x] Coroutines
- [x] Python/JS Bridges
- [x] CLI ve Formatter
- [x] VSCode Support
- [ ] JIT Compiler
- [ ] Multithreading
- [ ] Macro System
- [ ] Package Manager

## İletişim

- GitHub Issues: [sky-lang/sky/issues](https://github.com/sky-lang/sky/issues)
- Discord: [Sky Language Community](https://discord.gg/sky-lang)
- Twitter: [@SkyLanguage](https://twitter.com/SkyLanguage)

---

**Sky** - Python-benzeri girintili sözdizimi, tip beyanı zorunlu, dinamik & güçlü tip, string interpolation, async/await, coroutines, VM bytecode, Python/JS köprüleri.
