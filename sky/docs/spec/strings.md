# Sky String Interpolation

Sky, Dart ve Python tarzı string interpolation desteği sağlar. Bu özellik, string'lerin içinde değişken ve ifadeleri gömmenizi mümkün kılar.

## Genel Bakış

Sky'da iki tür string interpolation vardır:

1. **Dart tarzı interpolation** (varsayılan): `$ident` ve `${expr}`
2. **Python f-string tarzı**: `f"..."` ile başlayan string'ler

## Dart Tarzı Interpolation

### Temel Kullanım

```sky
string ad = "Sky"
int yaş = 1

# Basit değişken interpolation
string mesaj1 = "Merhaba $ad"  # "Merhaba Sky"

# Expression interpolation
string mesaj2 = "Yaş: ${yaş + 1}"  # "Yaş: 2"

# Karmaşık ifadeler
string mesaj3 = "Toplam: ${10 + 20}"  # "Toplam: 30"
```

### Syntax Kuralları

- **`$ident`**: Tek değişken için
- **`${expr}`**: Karmaşık ifadeler için
- **İç içe interpolation yasak**: `${ 1 + ${2} }` geçersiz
- **Normal parantez kullanın**: `${ 1 + (2) }` geçerli

### Örnekler

```sky
# Değişken interpolation
int x = 42
string s1 = "x = $x"  # "x = 42"

# Expression interpolation
string s2 = "Kare: ${x * x}"  # "Kare: 1764"

# String birleştirme
string ad = "Sky"
string s3 = "Merhaba $ad, hoş geldin!"  # "Merhaba Sky, hoş geldin!"

# Fonksiyon çağrıları
function selam(isim: string)
  return "Merhaba " + isim

string s4 = "Mesaj: ${selam("Sky")}"  # "Mesaj: Merhaba Sky"

# List ve Map interpolation
list sayılar = [1, 2, 3]
string s5 = "Sayılar: $sayılar"  # "Sayılar: [1, 2, 3]"

map bilgi = {"ad": "Sky", "versiyon": 1}
string s6 = "Bilgi: $bilgi"  # "Bilgi: {ad: Sky, versiyon: 1}"
```

## Python F-String Tarzı

### Temel Kullanım

```sky
string ad = "Sky"
int yaş = 1

# f-string ile başlar
string mesaj = f"Merhaba {ad}, yaşın {yaş + 1}"  # "Merhaba Sky, yaşın 2"
```

### Syntax Kuralları

- **`f"..."` ile başlar**: f-string belirteci
- **`{expr}`**: Expression interpolation
- **`{{` ve `}}`**: Literal süslü parantezler
- **Normal string'lerde `{` özel değil**: `"{x}"` düz metin

### Örnekler

```sky
# Basit f-string
int x = 42
string s1 = f"x = {x}"  # "x = 42"

# Karmaşık ifadeler
string s2 = f"Kare: {x * x}"  # "Kare: 1764"

# Süslü parantez kaçışı
string s3 = f"{{literal}} {x}"  # "{literal} 42"

# Fonksiyon çağrıları
function topla(a: int, b: int)
  return a + b

string s4 = f"Toplam: {topla(10, 20)}"  # "Toplam: 30"

# Async fonksiyonlar (async context'te)
async function indir(url: string)
  return "Veri: " + url

# async function içinde
string s5 = f"Sonuç: {await indir("test")}"  # "Sonuç: Veri: test"
```

## Kaçış Kuralları

### Normal String'ler

```sky
# Kaçış karakterleri
string s1 = "Tab: \t, Newline: \n, Quote: \", Backslash: \\"

# $ karakteri kaçışı
string s2 = "Dolar işareti: \\$"  # "Dolar işareti: $"

# {} karakterleri (f-string değilse düz metin)
string s3 = "Süslü parantez: {x}"  # "Süslü parantez: {x}"
```

### F-String'ler

```sky
# Süslü parantez kaçışı
string s1 = f"{{literal}} {x}"  # "{literal} 42"
string s2 = f"{x} {{literal}}"  # "42 {literal}"

# Diğer kaçışlar normal
string s3 = f"Tab: \t, Newline: \n"  # "Tab: \t, Newline: \n"
```

## Stringify Kuralları

Interpolation'da kullanılan değerler otomatik olarak string'e dönüştürülür:

### Temel Tipler

```sky
# Null
string s1 = "Değer: ${null}"  # "Değer: null"

# Boolean
string s2 = "Doğru: ${true}"  # "Doğru: true"

# Sayılar
string s3 = "Sayı: ${42}"  # "Sayı: 42"
string s4 = "Float: ${3.14}"  # "Float: 3.14"

# String (değişmez)
string s5 = "Metin: ${"Sky"}"  # "Metin: Sky"
```

### Karmaşık Tipler

```sky
# List
list items = [1, 2, 3]
string s1 = "Liste: ${items}"  # "Liste: [1, 2, 3]"

# Map
map data = {"a": 1, "b": 2}
string s2 = "Map: ${data}"  # "Map: {a: 1, b: 2}"

# Function
function test()
  return 42

string s3 = "Fonksiyon: ${test}"  # "Fonksiyon: <function test>"

# Coroutine
coop function say()
  yield 1
  return 2

string s4 = "Coroutine: ${say()}"  # "Coroutine: <coroutine 1>"
```

## Hata Durumları

### Lexer Hataları

```sky
# E4101: Unterminated interpolation
string s1 = "${x"  # Kapatıcı } yok

# E4102: Invalid identifier
string s2 = "$"  # Boş identifier
string s3 = "${}"  # Boş expression

# E4103: Nested interpolation
string s4 = "${ 1 + ${2} }"  # İç içe interpolation yasak

# E4104: F-string stray braces
string s5 = f"{{stray} {x}"  # Kaçışsız { veya }
```

### Runtime Hataları

```sky
# Type error (otomatik stringify ile çözülür)
int x = 42
string s = "x = $x"  # Çalışır: "x = 42"

# Undefined variable
string s2 = "y = $y"  # E1001: Undefined variable 'y'
```

## Performans

### Optimizasyonlar

```sky
# Tek parça string (optimize edilir)
string s1 = "Merhaba"  # CONST_STRING

# Tek expression (optimize edilir)
string s2 = "${x}"  # TO_STRING

# Çoklu parça (CONCAT kullanır)
string s3 = "Merhaba ${ad}"  # CONST_STRING + TO_STRING + CONCAT
```

### Bytecode

```sky
# "Merhaba ${x}" için bytecode:
# 1. CONST_STRING "Merhaba "
# 2. LOAD_LOCAL x
# 3. TO_STRING
# 4. CONCAT
```

## Best Practices

### 1. Basit Değişkenler için `$ident`

```sky
string ad = "Sky"
string mesaj = "Merhaba $ad"  # Basit ve okunabilir
```

### 2. Karmaşık İfadeler için `${expr}`

```sky
int x = 10
int y = 20
string sonuç = "Toplam: ${x + y}"  # Açık ve net
```

### 3. F-String'ler için Karmaşık Formatlar

```sky
float pi = 3.14159
string mesaj = f"Pi yaklaşık {pi} değerindedir"
```

### 4. Kaçış Kurallarına Dikkat

```sky
# F-string'de literal süslü parantez
string s1 = f"{{literal}} {x}"

# Normal string'de dolar kaçışı
string s2 = "Fiyat: \\$100"
```

## Örnekler

### Basit Interpolation

```sky
# Değişkenler
string ad = "Sky"
int yaş = 1
bool aktif = true

# Dart tarzı
string mesaj1 = "Merhaba $ad, yaşın $yaş"
string mesaj2 = "Durum: ${aktif ? "Aktif" : "Pasif"}"

# F-string tarzı
string mesaj3 = f"Merhaba {ad}, yaşın {yaş}"
string mesaj4 = f"Durum: {aktif ? "Aktif" : "Pasif"}"
```

### Karmaşık Örnekler

```sky
# Fonksiyon ile
function format_ad(ad: string, yaş: int)
  return f"{ad} ({yaş} yaşında)"

string kişi = format_ad("Sky", 1)

# Liste ile
list öğeler = ["elma", "armut", "kiraz"]
string liste = f"Öğeler: {öğeler}"

# Map ile
map bilgi = {"ad": "Sky", "versiyon": 1}
string detay = f"Bilgi: {bilgi}"
```

### Async Context

```sky
async function işlem()
  int x = 42
  string sonuç = f"İşlem tamamlandı: {x}"
  return sonuç

# async function içinde
async function ana()
  string mesaj = f"Sonuç: {await işlem()}"
  print(mesaj)
```

## String Modülü

Sky ayrıca string işlemleri için modül sağlar:

```sky
# String modülü
var string = import("string")

# Stringify
string sonuç = string.stringify(42)  # "42"

# Concat
string birleşik = string.concat("a", "b", "c")  # "abc"

# Length
int uzunluk = string.length("Sky")  # 3

# Split
list parçalar = string.split("a,b,c", ",")  # ["a", "b", "c"]

# Join
string birleşik2 = string.join(["a", "b", "c"], ",")  # "a,b,c"

# Case conversion
string büyük = string.upper("sky")  # "SKY"
string küçük = string.lower("SKY")  # "sky"

# Trim
string temiz = string.trim("  sky  ")  # "sky"

# Contains
bool var = string.contains("sky", "k")  # true

# Starts/Ends with
bool başlar = string.startsWith("sky", "s")  # true
bool biter = string.endsWith("sky", "y")  # true
```

String interpolation, Sky dilinin güçlü özelliklerinden biridir ve modern programlama dillerinin standart özelliklerini sağlar.
