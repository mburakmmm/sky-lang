# Sky Coroutines - Coop/Yield Semantiği

## Genel Bakış

Sky, coroutine'leri `coop function` ve `yield` anahtar kelimeleri ile destekler. Coroutine'ler, işbirlikçi çoklu görev (cooperative multitasking) sağlar ve `yield` noktalarında durup devam edebilir.

## Coroutine Temelleri

### Coop Function Tanımı
```sky
coop function ad(param: TYPE)
  # gövde
  yield expr
  return expr
```

### Coroutine Özellikleri
- **Cooperative**: Coroutine kendi isteği ile `yield` eder
- **Suspended**: `yield` sonrası durur ve çağırana döner
- **Resumable**: `resume()` ile kaldığı yerden devam eder
- **Stateful**: Durumu korur (local variables, IP, stack)

## Coroutine API

### Coroutine Oluşturma
```sky
coop function say(n: int)
  int i = 0
  while i < n
    yield i
    i = i + 1
  return n

var c = say(3)  # Coroutine objesi döner, hemen çalışmaz
```

### Coroutine Kontrolü
```sky
# Coroutine'i devam ettir
var değer = c.resume()  # yield'den dönen değer

# Coroutine'in bitip bitmediğini kontrol et
bool bitti = c.is_done()

# Coroutine'in durumunu kontrol et
string durum = c.status()  # "Suspended", "Running", "Done"
```

## Coroutine Durumları

### 1. Suspended (Duraklatılmış)
- `yield` sonrası durum
- `resume()` ile devam ettirilebilir
- Local state korunur

### 2. Running (Çalışıyor)
- `resume()` çağrıldıktan sonra
- `yield` veya `return` sonrası tekrar Suspended/Done olur

### 3. Done (Tamamlandı)
- `return` sonrası veya fonksiyon sonu
- `resume()` çağrılırsa hata verir
- `is_done()` true döndürür

## Coroutine Örnekleri

### Basit Counter
```sky
coop function counter(n: int)
  int i = 0
  while i < n
    yield i
    i = i + 1
  return n

var c = counter(3)
print(c.resume())  # 0
print(c.resume())  # 1
print(c.resume())  # 2
print(c.is_done())  # true
```

### Fibonacci Generator
```sky
coop function fibonacci(n: int)
  int a = 0
  int b = 1
  int i = 0
  while i < n
    yield a
    int temp = a + b
    a = b
    b = temp
    i = i + 1
  return n

var fib = fibonacci(10)
while not fib.is_done()
  print(fib.resume())
```

### Ping-Pong Coroutine
```sky
coop function ping(n: int)
  int i = 0
  while i < n
    yield "ping " + i
    i = i + 1
  return "ping done"

coop function pong(n: int)
  int i = 0
  while i < n
    yield "pong " + i
    i = i + 1
  return "pong done"

var ping_c = ping(3)
var pong_c = pong(3)

while not ping_c.is_done() or not pong_c.is_done()
  if not ping_c.is_done()
    print(ping_c.resume())
  if not pong_c.is_done()
    print(pong_c.resume())
```

## Coroutine ile Async Karşılaştırması

### Async Function
```sky
async function indir(url: string)
  var data = await http.get(url)
  return data

var future = indir("https://example.com")
var sonuç = await future  # Blocking wait
```

### Coop Function
```sky
coop function indir(url: string)
  var data = await http.get(url)
  yield data
  return "Tamamlandı"

var coroutine = indir("https://example.com")
var data = coroutine.resume()  # Non-blocking
var sonuç = coroutine.resume()  # Devam ettir
```

## Coroutine Hata Yönetimi

### Done Coroutine Resume Hatası
```sky
coop function test()
  yield 1
  return 2

var c = test()
print(c.resume())  # 1
print(c.resume())  # 2
print(c.resume())  # E2001: Coroutine already finished
```

### Coroutine İçinde Hata
```sky
coop function hata_ver()
  yield 1
  int x = 1 / 0  # Runtime error
  yield 2

var c = hata_ver()
print(c.resume())  # 1
print(c.resume())  # Runtime error, coroutine Done olur
```

## Coroutine Best Practices

### 1. Yield Noktaları
```sky
coop function iyi_örnek()
  int i = 0
  while i < 1000
    if i % 100 == 0
      yield i  # Düzenli yield noktaları
    i = i + 1
  return i
```

### 2. Resource Cleanup
```sky
coop function kaynak_yönetimi()
  var dosya = io.open("test.txt")
  try
    while not dosya.eof()
      var satır = dosya.read_line()
      yield satır
  finally
    dosya.close()  # Kaynak temizliği
  return "Tamamlandı"
```

### 3. Coroutine Composition
```sky
coop function birleştir(c1: Coroutine, c2: Coroutine)
  while not c1.is_done() or not c2.is_done()
    if not c1.is_done()
      yield c1.resume()
    if not c2.is_done()
      yield c2.resume()
  return "Her ikisi de tamamlandı"
```

## Coroutine Implementasyon Detayları

### VM Seviyesinde
- **Frame Snapshot**: Coroutine durumu frame olarak saklanır
- **Stack Management**: Local variables ve call stack korunur
- **Instruction Pointer**: Kaldığı yerden devam eder

### Bytecode Seviyesinde
- **YIELD**: Mevcut frame'i snapshot'la, çağırana dön
- **COOP_RESUME**: Hedef frame'i aktif et, kaldığı yerden çalıştır
- **COOP_IS_DONE**: Coroutine durumunu kontrol et

### Memory Management
- **GC Roots**: Suspended coroutine'ler GC kökü olarak işaretlenir
- **Frame Cleanup**: Done coroutine'lerin frame'leri temizlenir

## Coroutine Sınırlamaları

### MVP Sınırlamaları
- **Single-threaded**: Coroutine'ler aynı thread'de çalışır
- **No preemption**: Coroutine kendi isteği ile yield etmeli
- **No coroutine communication**: Coroutine'ler arası direct iletişim yok

### Gelecek Özellikler
- **Coroutine channels**: Coroutine'ler arası iletişim
- **Coroutine pools**: Çoklu coroutine yönetimi
- **Coroutine scheduling**: Otomatik scheduling

## Coroutine Test Örnekleri

### Unit Tests
```sky
# Coroutine oluşturma testi
coop function test_create()
  yield 1
  return 2

var c = test_create()
assert not c.is_done()
assert c.status() == "Suspended"

# Resume testi
var değer = c.resume()
assert değer == 1
assert not c.is_done()

# Done testi
var sonuç = c.resume()
assert sonuç == 2
assert c.is_done()
assert c.status() == "Done"

# Hata testi
var hata = c.resume()  # E2001 beklenir
```

### Integration Tests
```sky
# Çoklu coroutine testi
coop function a()
  yield 1
  yield 2
  return 3

coop function b()
  yield "x"
  yield "y"
  return "z"

var c1 = a()
var c2 = b()

var sonuçlar = []
while not c1.is_done() or not c2.is_done()
  if not c1.is_done()
    sonuçlar.append(c1.resume())
  if not c2.is_done()
    sonuçlar.append(c2.resume())

assert sonuçlar == [1, "x", 2, "y", 3, "z"]
```

## Coroutine Performans

### Memory Usage
- **Frame Size**: Her coroutine için frame boyutu
- **Stack Depth**: Maksimum call stack derinliği
- **GC Pressure**: Suspended coroutine'ler GC basıncı

### Execution Speed
- **Yield Overhead**: Yield işlemi maliyeti
- **Resume Overhead**: Resume işlemi maliyeti
- **Context Switching**: Frame değiştirme maliyeti

Coroutine'ler, Sky dilinin güçlü özelliklerinden biridir ve işbirlikçi çoklu görev için ideal bir çözüm sağlar.
