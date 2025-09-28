# Sky Dil Yazımı - Grammar

## Genel Bakış

Sky, Python-benzeri girintili sözdizimi ve basit tasarımı olan bir programlama dilidir. Dinamik & güçlü tipli çalışma zamanı, async/await, coroutines, VM bytecode ve Python/JS köprüleri destekler.

## Temel Sözdizimi

### Dosya Yapısı
```
FILE := STATEMENT*
```

### Yorumlar
```
COMMENT := "#" [^"\n"]*
```

### Tanımlayıcılar (Identifiers)
```
IDENTIFIER := [a-zA-Z_] [a-zA-Z0-9_çğıöşüÇĞIÖŞÜ]*
```

Sky, Türkçe karakterleri destekler: ç, ğ, ı, ö, ş, ü (büyük/küçük)

### Anahtar Kelimeler
```
KEYWORDS := var | int | float | bool | string | list | map
          | function | async | coop | return | if | elif | else
          | for | while | break | continue | await | yield
          | import | true | false | null
```

### Literaller
```
INTEGER := 0 | [1-9][0-9]*
FLOAT := [0-9]+"."[0-9]+
STRING := "\"" [^"\"]* "\""
BOOLEAN := true | false
NULL := null

LITERAL := INTEGER | FLOAT | STRING | BOOLEAN | NULL | LIST_LITERAL | MAP_LITERAL
```

### Liste ve Sözlük Literalleri
```
LIST_LITERAL := "[" [EXPRESSION ("," EXPRESSION)*] "]"
MAP_LITERAL := "{" [STRING ":" EXPRESSION ("," STRING ":" EXPRESSION)*] "}"
```

## Tip Bildirimleri

Sky'da **tip beyanı zorunludur**. Tüm değişken tanımlamaları tip ile başlamalıdır.

### Tip Türleri
```
TYPE := var | int | float | bool | string | list | map
```

- `var`: Dinamik tip (Any) - her tür kabul eder
- `int`: Tamsayı tipi
- `float`: Ondalıklı sayı tipi
- `bool`: Boolean tipi
- `string`: Metin tipi
- `list`: Liste tipi
- `map`: Sözlük tipi

### Değişken Bildirimi
```
VAR_DECL := TYPE IDENTIFIER "=" EXPRESSION
```

**Örnekler:**
```sky
int sayı = 42
float pi = 3.14159
bool doğru = true
string mesaj = "Merhaba Sky"
list sayılar = [1, 2, 3]
map sözlük = {"ad": "Sky", "versiyon": 1}
var dinamik = 42  # var tipi her tür kabul eder
```

**Hata Örnekleri:**
```sky
x = 3  # E0001: Missing type annotation
sayı = 42  # E0001: Missing type annotation
```

## İfadeler (Expressions)

### Aritmetik İfadeler
```
EXPRESSION := TERM (("+" | "-") TERM)*
TERM := FACTOR (("*" | "/" | "%") FACTOR)*
FACTOR := UNARY | PRIMARY
UNARY := ("-" | "!") FACTOR
PRIMARY := LITERAL | IDENTIFIER | CALL | ATTRIBUTE | INDEX | "(" EXPRESSION ")"
```

### Karşılaştırma İfadeleri
```
COMPARISON := EXPRESSION (("==" | "!=" | "<" | "<=" | ">" | ">=") EXPRESSION)*
```

### Mantıksal İfadeler
```
LOGICAL := COMPARISON (("and" | "or") COMPARISON)*
```

### Fonksiyon Çağrıları
```
CALL := IDENTIFIER "(" [EXPRESSION ("," EXPRESSION)*] ")"
```

### Üye Erişimi
```
ATTRIBUTE := EXPRESSION "." IDENTIFIER
INDEX := EXPRESSION "[" EXPRESSION "]"
```

### Özel İfadeler
```
AWAIT := "await" EXPRESSION  # Sadece async function içinde
YIELD := "yield" [EXPRESSION]  # Sadece coop function içinde
```

## Statement'lar

### Değişken Bildirimi
```
VAR_STMT := TYPE IDENTIFIER "=" EXPRESSION
```

### Fonksiyon Tanımları
```
FUNCTION := [ASYNC | COOP] "function" IDENTIFIER "(" PARAM_LIST ")" BLOCK
ASYNC := "async"
COOP := "coop"
PARAM_LIST := [PARAM ("," PARAM)*]
PARAM := IDENTIFIER ":" TYPE
BLOCK := INDENTED_STATEMENTS
```

**Örnekler:**
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

#### If Statement
```
IF := "if" EXPRESSION BLOCK [ELIF_BLOCK]* [ELSE_BLOCK]?
ELIF_BLOCK := "elif" EXPRESSION BLOCK
ELSE_BLOCK := "else" BLOCK
```

#### For Loop
```
FOR := "for" IDENTIFIER ":" TYPE "in" EXPRESSION BLOCK
```

#### While Loop
```
WHILE := "while" EXPRESSION BLOCK
```

**Örnekler:**
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

### Diğer Statement'lar
```
RETURN := "return" [EXPRESSION]?
BREAK := "break"
CONTINUE := "continue"
IMPORT := "import" IDENTIFIER ("." IDENTIFIER)*
EXPR_STMT := EXPRESSION
```

## Girinti Kuralları

Sky, Python-benzeri girinti kullanır. Bloklar girinti ile belirlenir.

### Girinti Kuralları
1. **Tab yasak**: Sadece boşluk kullanılır
2. **Tutarlı girinti**: Aynı seviyede aynı sayıda boşluk
3. **Girinti artışı**: Blok başlangıcında girinti artar
4. **Girinti azalması**: Blok bitiminde girinti azalır

### Girinti Örnekleri
```sky
function test()
  int x = 1
  if x > 0
    print("Pozitif")
    if x > 10
      print("Çok büyük")
  print("Bitti")
```

## Hata Kodları

### Syntax Hataları
- **E0001**: Missing type annotation
- **E0101**: Invalid indentation (tabs or inconsistent indent)
- **E0201**: await outside async function
- **E0202**: yield outside coop function

### Runtime Hataları
- **E1001**: Type mismatch (expected X, found Y)
- **E2001**: Coroutine already finished

### Bridge Hataları
- **E3001**: Python bridge error
- **E3002**: JS bridge error

## Örnekler

### Basit Program
```sky
int sayı = 42
string mesaj = "Merhaba Sky"
print(mesaj + " " + sayı)
```

### Fonksiyon Örneği
```sky
function faktöriyel(n: int)
  if n <= 1
    return 1
  else
    return n * faktöriyel(n - 1)

int sonuç = faktöriyel(5)
print(sonuç)
```

### Async Örneği
```sky
async function indir(url: string)
  var data = await http.get(url)
  return data

var sonuç = indir("https://example.com")
print(await sonuç)
```

### Coroutine Örneği
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

### Python Bridge Örneği
```sky
var math = python.import("math")
float kök = math.sqrt(16)
print(kök)
```

### JS Bridge Örneği
```sky
var jsfn = js.eval("(x)=>x*2")
int iki_kat = jsfn(21)
print(iki_kat)  # 42
```

## Öncelik Kuralları

1. **Üst seviye**: `await`, `yield`
2. **Unary**: `-`, `!`
3. **Multiplicative**: `*`, `/`, `%`
4. **Additive**: `+`, `-`
5. **Comparison**: `==`, `!=`, `<`, `<=`, `>`, `>=`
6. **Logical**: `and`, `or`

## Özel Kurallar

### Tip Zorunluluğu
- Tüm değişken tanımlamaları tip ile başlamalıdır
- Fonksiyon parametreleri tip belirtmelidir
- `var` tipi dinamik tip olarak kullanılır

### Async/Await
- `await` sadece `async function` içinde kullanılabilir
- `async function` çağrıldığında Future döndürür

### Coroutines
- `yield` sadece `coop function` içinde kullanılabilir
- `coop function` çağrıldığında Coroutine döndürür
- Coroutine'ler `resume()` ile devam ettirilir

### Girinti
- Tab karakterleri yasaktır
- Tutarlı girinti kullanılmalıdır
- Bloklar girinti ile belirlenir

Bu grammar, Sky dilinin tüm özelliklerini kapsar ve implementasyon için referans olarak kullanılabilir.
