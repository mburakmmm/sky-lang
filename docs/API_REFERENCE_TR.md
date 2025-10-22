# 📚 SKY Programlama Dili - API Referansı (Türkçe)

## 🌟 Giriş

SKY, modern, tip-güvenli, Go ile yazılmış bir programlama dilidir. Python'un sadeliği ile Go'nun performansını birleştirir.

---

## 📖 İçindekiler

1. [Hata Yönetimi](#hata-yönetimi) ⚠️
2. [Temel Sözdizimi](#temel-sözdizimi)
3. [Veri Tipleri](#veri-tipleri)
4. [Fonksiyonlar](#fonksiyonlar)
5. [Kontrol Yapıları](#kontrol-yapıları)
6. [Sınıflar ve OOP](#sınıflar-ve-oop)
7. [Async/Await](#asyncawait)
8. [Pattern Matching](#pattern-matching)
9. [Yerleşik Fonksiyonlar](#yerleşik-fonksiyonlar)
10. [Standart Kütüphane](#standart-kütüphane)
11. [Modül Sistemi](#modül-sistemi)
12. [Paket Yönetimi](#paket-yönetimi)

---

## ⚠️ Hata Yönetimi

SKY'da hata yönetimi **try/catch/finally** modeli ile yapılır. Bu, modern dillerde yaygın olan exception-based error handling yaklaşımıdır.

### Try-Catch-Finally

```sky
# Temel try-catch kullanımı
try
  let result = fs_read_text("dosya.txt")
  print("Dosya okundu: " + result)
catch error
  print("Hata oluştu: " + error)
end

# Finally bloğu ile temizlik
try
  let file = fs_open("data.txt")
  # Dosya işlemleri...
catch error
  print("Dosya hatası: " + error)
finally
  # Her durumda çalışır
  print("Temizlik yapılıyor...")
end
```

### Throw (Hata Fırlatma)

```sky
function divide(a: int, b: int): int
  if b == 0
    throw "Sıfıra bölme hatası!"
  end
  return a / b
end

# Kullanım
try
  let result = divide(10, 0)
  print("Sonuç: " + result)
catch error
  print("Hata yakalandı: " + error)
end
```

### Hata Türleri

SKY'da iki tür hata vardır:

1. **Runtime Hatalar**: Program çalışırken oluşan hatalar
   - `fs_read_text("olmayan_dosya.txt")` → Dosya bulunamadı hatası
   - `list[10]` (5 elemanlı liste) → Index out of range hatası
   - `dict["olmayan_anahtar"]` → Key not found hatası

2. **Throw Hatalar**: Programcının fırlattığı hatalar
   - `throw "Özel hata mesajı"`
   - `throw error_object`

### Hata Yönetimi Best Practices

```sky
# 1. Spesifik hata yakalama
try
  let data = http_get("https://api.example.com/data")
  process_data(data)
catch error
  if error.contains("network")
    print("Ağ hatası, tekrar denenecek...")
  else
    print("Bilinmeyen hata: " + error)
  end
end

# 2. Hata loglama
try
  risky_operation()
catch error
  log_error("Risky operation failed: " + error)
  # Hata yukarı fırlatılabilir
  throw error
end

# 3. Resource cleanup
try
  let connection = db_connect()
  # Veritabanı işlemleri...
catch error
  print("DB hatası: " + error)
finally
  if connection != null
    connection.close()
  end
end
```

---

## 🎯 Temel Sözdizimi

### Değişkenler

```sky
# Değiş

ken tanımlama
let x = 10
let name = "Sky"
let pi = 3.14
let active = true

# Sabit tanımlama
const MAX_SIZE = 100

# Tip belirtme (opsiyonel)
let age: int = 25
let price: float = 19.99
```

### Yorumlar

```sky
# Tek satırlık yorum

# Birden fazla satır için
# her satırda # kullanın
```

---

## 📦 Veri Tipleri

### Primitive Tipler

```sky
# Integer (tam sayı)
let count = 42
let negative = -10

# Float (ondalıklı sayı)
let temperature = 36.6
let epsilon = 0.001

# String (metin)
let greeting = "Merhaba Dünya"
let empty = ""

# Boolean (mantıksal)
let is_valid = true
let is_empty = false

# Nil (boş değer)
let nothing = nil
```

### Koleksiyon Tipleri

#### Liste (List)

```sky
# Liste oluşturma
let numbers = [1, 2, 3, 4, 5]
let names = ["Ali", "Ayşe", "Mehmet"]
let mixed = [1, "hello", true]

# Liste elemanına erişim
print(numbers[0])  # 1

# Liste metodları
let length = len(numbers)
print("Liste uzunluğu:", length)
```

##### Liste Metodları

| Metod | Açıklama | Örnek |
|-------|----------|-------|
| `.append(item)` | Eleman ekle | `list.append(6)` |
| `.pop([index])` | Eleman çıkar | `list.pop()` veya `list.pop(0)` |
| `.remove(item)` | Değere göre sil | `list.remove("Ali")` |
| `.insert(index, item)` | Belirli pozisyona ekle | `list.insert(1, "Veli")` |
| `.sort()` | Sırala | `list.sort()` |
| `.reverse()` | Ters çevir | `list.reverse()` |
| `.clear()` | Temizle | `list.clear()` |
| `.count(item)` | Eleman sayısı | `list.count("Ali")` |
| `.index(item)` | Eleman pozisyonu | `list.index("Ali")` |
| `.copy()` | Kopyala | `let new_list = list.copy()` |

```sky
let fruits = ["elma", "armut", "kiraz"]

# Eleman ekleme
fruits.append("muz")
print(fruits)  # ["elma", "armut", "kiraz", "muz"]

# Eleman çıkarma
let last = fruits.pop()
print(last)    # "muz"
print(fruits)  # ["elma", "armut", "kiraz"]

# Sıralama
fruits.sort()
print(fruits)  # ["armut", "elma", "kiraz"]

# Ters çevirme
fruits.reverse()
print(fruits)  # ["kiraz", "elma", "armut"]
```

#### Sözlük (Dict)

```sky
# Sözlük oluşturma
let person = {
  "name": "Ahmet",
  "age": "30",
  "city": "İstanbul"
}

# Elemana erişim
print(person["name"])  # Ahmet

# Eleman ekleme
person["email"] = "ahmet@example.com"
```

##### Sözlük Metodları

| Metod | Açıklama | Örnek |
|-------|----------|-------|
| `.keys()` | Tüm anahtarları al | `dict.keys()` |
| `.values()` | Tüm değerleri al | `dict.values()` |
| `.items()` | Anahtar-değer çiftleri | `dict.items()` |
| `.get(key, default)` | Güvenli erişim | `dict.get("name", "Bilinmiyor")` |
| `.set(key, value)` | Değer ata | `dict.set("age", 31)` |
| `.has_key(key)` | Anahtar var mı? | `dict.has_key("name")` |
| `.delete(key)` | Anahtarı sil | `dict.delete("age")` |
| `.clear()` | Temizle | `dict.clear()` |
| `.copy()` | Kopyala | `let new_dict = dict.copy()` |
| `.update(other)` | Başka dict ile güncelle | `dict.update(other_dict)` |

```sky
let person = {
  "name": "Ahmet",
  "age": 30,
  "city": "İstanbul"
}

# Anahtarları al
let keys = person.keys()
print(keys)  # ["name", "age", "city"]

# Değerleri al
let values = person.values()
print(values)  # ["Ahmet", 30, "İstanbul"]

# Güvenli erişim
let name = person.get("name", "Bilinmiyor")
let phone = person.get("phone", "Yok")  # "Yok" döner

# Anahtar kontrolü
if person.has_key("age")
  print("Yaş bilgisi mevcut")
end

# Eleman silme
person.delete("age")
print(person)  # {"name": "Ahmet", "city": "İstanbul"}

# Iterasyon
for key, value in person.items()
  print(key + ": " + value)
end
```

---

## 🔧 Fonksiyonlar

### Temel Fonksiyon

```sky
function topla(a, b)
  return a + b
end

let sonuc = topla(5, 3)  # 8
```

### Tip Belirtmeli Fonksiyon

```sky
function carpim(x: int, y: int): int
  return x * y
end
```

### Varsayılan Parametreler

```sky
function selamla(isim = "Misafir")
  print("Merhaba", isim)
end

selamla()          # Merhaba Misafir
selamla("Ali")     # Merhaba Ali
```

### Recursive Fonksiyonlar

```sky
function faktoriyel(n)
  if n <= 1
    return 1
  end
  return n * faktoriyel(n - 1)
end

print(faktoriyel(5))  # 120
```

### Function Type Annotation

SKY dilinde fonksiyon tiplerini belirtmek için `(parametre_tipleri) => dönüş_tipi` syntax'ı kullanılır.

```sky
# Function type annotation örnekleri
function test_callback(callback: (int, string) => bool): void
  let result = callback(42, "test")
  print("Callback sonucu:", result)
end

# Boş parametre listesi için 'any' kullanılır
function test_empty_callback(callback: any): void
  callback()
end

# Lambda expression ile kullanım
test_callback(function(x: int, s: string): bool
  print("Callback çağrıldı: x =", x, "s =", s)
  return true
end)

test_empty_callback(function(): void
  print("Boş callback çağrıldı")
end)
```

**Not:** Boş parametre listesi `() => void` syntax'ı henüz desteklenmiyor. Bunun yerine `any` tipi kullanın.

---

## 🎮 Kontrol Yapıları

### If-Else

```sky
let yas = 18

if yas >= 18
  print("Yetişkin")
else
  print("Çocuk")
end

# Elif ile
let not = 85

if not >= 90
  print("AA")
elif not >= 80
  print("BA")
elif not >= 70
  print("BB")
else
  print("Geçti")
end
```

### While Döngüsü

```sky
let sayac = 0

while sayac < 5
  print(sayac)
  sayac = sayac + 1
end
```

### For Döngüsü

```sky
# Liste üzerinde
let meyveler = ["elma", "armut", "muz"]

for meyve in meyveler
  print(meyve)
end

# Sayı aralığında
for i in range(10)
  print(i)  # 0'dan 9'a kadar
end
```

### Break ve Continue

```sky
for i in range(10)
  if i == 3
    continue  # 3'ü atla
  end
  
  if i == 7
    break  # 7'de dur
  end
  
  print(i)
end
```

---

## 🏛️ Sınıflar ve OOP

### Sınıf Tanımlama

```sky
class Kisi
  function init(isim, yas)
    self.isim = isim
    self.yas = yas
  end
  
  function selamla()
    print("Merhaba, ben", self.isim)
  end
  
  function bilgi()
    print(self.isim, "yaşında", self.yas)
  end
end

# Kullanım
let ahmet = Kisi("Ahmet", 25)
ahmet.selamla()  # Merhaba, ben Ahmet
ahmet.bilgi()    # Ahmet yaşında 25
```

### Erişim Belirleyicileri (Access Modifiers)

SKY'da varsayılan olarak tüm üyeler **public**'tir. Private üyeler için `_` öneki kullanılır:

```sky
class BankaHesabi
  function init(hesap_no: string, bakiye: float)
    self.hesap_no = hesap_no        # Public
    self._bakiye = bakiye           # Private
    self._islem_gecmisi = []        # Private
  end
  
  # Public metodlar
  function bakiye_sorgula(): float
    return self._bakiye
  end
  
  function para_yatir(miktar: float): void
    if miktar > 0
      self._bakiye = self._bakiye + miktar
      self._islem_ekle("Yatırım: " + str(miktar))
    end
  end
  
  # Private metodlar
  function _islem_ekle(aciklama: string): void
    self._islem_gecmisi.append(aciklama)
  end
  
  function _bakiye_kontrol(miktar: float): bool
    return self._bakiye >= miktar
  end
end

# Kullanım
let hesap = BankaHesabi("12345", 1000.0)
print(hesap.bakiye_sorgula())  # ✅ Çalışır
hesap.para_yatir(500.0)        # ✅ Çalışır
# print(hesap._bakiye)         # ❌ Hata: Private member
```

### Static Metodlar ve Özellikler

```sky
class Matematik
  # Static özellik
  static PI = 3.14159
  static E = 2.71828
  
  # Static metod
  static function topla(a: int, b: int): int
    return a + b
  end
  
  static function daire_alan(yaricap: float): float
    return Matematik.PI * yaricap * yaricap
  end
  
  # Instance metod
  function hesapla(x: int): int
    return Matematik.topla(x, 10)
  end
end

# Kullanım
let sonuc = Matematik.topla(5, 3)           # Static metod
let alan = Matematik.daire_alan(5.0)        # Static metod
print("PI değeri:", Matematik.PI)           # Static özellik

let math = Matematik()
let hesaplama = math.hesapla(20)            # Instance metod
```

### Interface/Protocol Tanımlama

```sky
# Interface tanımlama
interface Drawable
  function draw(): void
  function get_area(): float
end

interface Movable
  function move(x: float, y: float): void
  function get_position(): dict
end

# Interface implementasyonu
class Daire
  function init(x: float, y: float, yaricap: float)
    self.x = x
    self.y = y
    self.yaricap = yaricap
  end
  
  # Drawable interface
  function draw(): void
    print("Daire çiziliyor:", self.x, self.y, self.yaricap)
  end
  
  function get_area(): float
    return 3.14159 * self.yaricap * self.yaricap
  end
  
  # Movable interface
  function move(x: float, y: float): void
    self.x = self.x + x
    self.y = self.y + y
  end
  
  function get_position(): dict
    return {"x": self.x, "y": self.y}
  end
end

# Kullanım
let daire = Daire(10, 20, 5)
daire.draw()                    # Interface metodu
let alan = daire.get_area()     # Interface metodu
daire.move(5, 10)               # Interface metodu
```

### Kalıtım (Inheritance)

```sky
class Hayvan
  function init(isim: string)
    self.isim = isim
    self._enerji = 100
  end
  
  function ses_cikar(): void
    print(self.isim, "ses çıkarıyor")
  end
  
  function yemek_ye(): void
    self._enerji = self._enerji + 20
    print(self.isim, "yemek yedi, enerji:", self._enerji)
  end
  
  function _enerji_kontrol(): bool
    return self._enerji > 0
  end
end

class Kedi
  function init(isim: string, cins: string)
    super.init(isim)  # Parent constructor
    self.cins = cins
  end
  
  # Method override
  function ses_cikar(): void
    print(self.isim, "miyavlıyor")
  end
  
  # Yeni metod
  function oyun_oyna(): void
    if self._enerji_kontrol()
      self._enerji = self._enerji - 10
      print(self.isim, "oyun oynuyor")
    else
      print(self.isim, "çok yorgun")
    end
  end
end

# Kullanım
let kedi = Kedi("Pamuk", "Van Kedisi")
kedi.ses_cikar()    # Pamuk miyavlıyor
kedi.yemek_ye()     # Pamuk yemek yedi, enerji: 120
kedi.oyun_oyna()    # Pamuk oyun oynuyor
```

### Çoklu Kalıtım (Multiple Inheritance)

SKY'da çoklu kalıtım `:` operatörü ile desteklenir:

```sky
class Ucan
  function uc(): void
    print("Uçuyor...")
  end
end

class Yuzebilen
  function yuz(): void
    print("Yüzüyor...")
  end
end

class Ordek : Ucan, Yuzebilen
  function init(isim: string)
    self.isim = isim
  end
  
  # Parent sınıfların metodlarını override edebilir
  function uc(): void
    print(self.isim, "uçuyor...")
  end
  
  function yuz(): void
    print(self.isim, "yüzüyor...")
  end
  
  # Kendi metodları
  function yemek_ara(): void
    print(self.isim, "yemek arıyor...")
  end
end

# Kullanım
let ordek = Ordek("Pamuk")
ordek.uc()        # Pamuk uçuyor...
ordek.yuz()       # Pamuk yüzüyor...
ordek.yemek_ara() # Pamuk yemek arıyor...
```

#### Çoklu Kalıtım Kuralları

1. **Metod Çakışması**: Eğer iki parent sınıfta aynı isimde metod varsa, child sınıf bunu override etmelidir
2. **Diamond Problem**: SKY, çoklu kalıtımda diamond problem'i otomatik çözer
3. **Constructor Zinciri**: Parent constructor'lar otomatik çağrılır

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
    super.init()  # Parent constructor'ları çağır
    print("C constructor")
  end
end

let c = C()  # A constructor, B constructor, C constructor
```

### Abstract Class (Soyut Sınıf)

```sky
# Abstract class (interface gibi davranır)
abstract class Sekil
  function init()
    # Abstract class constructor
  end
  
  # Abstract metodlar (implement edilmeli)
  abstract function alan_hesapla(): float
  abstract function cevre_hesapla(): float
  
  # Concrete metod
  function bilgi_yazdir(): void
    print("Alan:", self.alan_hesapla())
    print("Çevre:", self.cevre_hesapla())
  end
end

class Kare
  function init(kenar: float)
    super.init()
    self.kenar = kenar
  end
  
  # Abstract metodları implement et
  function alan_hesapla(): float
    return self.kenar * self.kenar
  end
  
  function cevre_hesapla(): float
    return 4 * self.kenar
  end
end

# Kullanım
let kare = Kare(5.0)
kare.bilgi_yazdir()  # Alan: 25, Çevre: 20
```

---

## ⚡ Async/Await

### Asenkron Fonksiyonlar

```sky
async function veriCek(id)
  print("Veri çekiliyor:", id)
  # Asenkron işlem simülasyonu
  return "Veri-" + str(id)
end

async function main
  let sonuc = await veriCek(42)
  print("Sonuç:", sonuc)
end
```

### Paralel Asenkron İşlemler

#### Sıralı İşlem (Sequential)

```sky
async function kullaniciGetir(id)
  # Simüle edilmiş API çağrısı
  await sleep(100)  # 100ms bekle
  return {"id": str(id), "isim": "User" + str(id)}
end

async function main
  # Sıralı - toplam 200ms sürer
  let user1 = await kullaniciGetir(1)
  let user2 = await kullaniciGetir(2)
  
  print("Kullanıcı 1:", user1)
  print("Kullanıcı 2:", user2)
end
```

#### Paralel İşlem (Concurrent)

```sky
async function main
  # Paralel - toplam 100ms sürer
  let promise1 = kullaniciGetir(1)  # Promise döndürür
  let promise2 = kullaniciGetir(2)  # Promise döndürür
  
  # Her ikisini de bekle
  let user1 = await promise1
  let user2 = await promise2
  
  print("Kullanıcı 1:", user1)
  print("Kullanıcı 2:", user2)
end
```

#### Promise.all() Benzeri İşlem

```sky
async function tumKullanicilariGetir(ids: list): list
  let promises = []
  
  # Tüm promise'leri başlat
  for id in ids
    promises.append(kullaniciGetir(id))
  end
  
  # Tümünü bekle
  let results = []
  for promise in promises
    results.append(await promise)
  end
  
  return results
end
```

#### Built-in Promise.all() Fonksiyonu

SKY'da paralel işlemler için built-in `Promise.all()` fonksiyonu:

```sky
async function main
  let ids = [1, 2, 3, 4, 5]
  
  # Promise.all() ile paralel işlem
  let promises = []
  for id in ids
    promises.append(kullaniciGetir(id))
  end
  
  # Tüm promise'leri paralel olarak bekle
  let results = await Promise.all(promises)
  
  print("Tüm kullanıcılar:", results)
end
```

#### Promise.all() Özellikleri

| Özellik | Açıklama |
|---------|----------|
| **Paralel Çalışma** | Tüm promise'ler aynı anda başlar |
| **Hızlı Başarısızlık** | Bir promise başarısız olursa hemen hata döner |
| **Sıralı Sonuçlar** | Sonuçlar promise'lerin sırasına göre döner |
| **Tip Güvenliği** | Tüm promise'ler aynı tip döndürmelidir |

#### Hata Yönetimi ile Promise.all()

```sky
async function guvenliKullaniciGetir(id)
  try
    return await kullaniciGetir(id)
  catch error
    return {"id": id, "error": "Kullanıcı bulunamadı"}
  end
end

async function main
  let ids = [1, 2, 3, 999]  # 999 geçersiz ID
  
  let promises = []
  for id in ids
    promises.append(guvenliKullaniciGetir(id))
  end
  
  # Hata yönetimi ile Promise.all()
  try
    let results = await Promise.all(promises)
    print("Sonuçlar:", results)
  catch error
    print("Genel hata:", error)
  end
end
```

#### Promise.allSettled() Alternatifi

```sky
# Promise.allSettled() benzeri fonksiyon
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

# Kullanım
let results = await Promise_allSettled(promises)
for result in results
  if result["status"] == "fulfilled"
    print("Başarılı:", result["value"])
  else
    print("Hata:", result["reason"])
  end
end
```

#### Hata Yönetimi ile Paralel İşlemler

```sky
async function guvenliVeriCek(url: string): string
  try
    let response = await http_get(url)
    return response.body
  catch error
    print("Hata:", error)
    return ""
  end
end

async function main
  let urls = [
    "https://api.example.com/data1",
    "https://api.example.com/data2",
    "https://api.example.com/data3"
  ]
  
  # Tüm URL'leri paralel olarak çek
  let promises = []
  for url in urls
    promises.append(guvenliVeriCek(url))
  end
  
  # Sonuçları bekle
  let results = []
  for promise in promises
    let result = await promise
    if result != ""
      results.append(result)
    end
  end
  
  print("Başarılı sonuçlar:", len(results))
end
```

### Event Loop ve Goroutine'ler

SKY, Go tabanlı olduğu için arka planda **Goroutine'ler** kullanır:

```sky
async function arkaPlanIslemi()
  print("Arka plan işlemi başladı")
  await sleep(1000)
  print("Arka plan işlemi bitti")
end

async function main
  # Arka plan işlemini başlat (fire-and-forget)
  arkaPlanIslemi()  # await kullanmıyoruz
  
  # Ana işlem devam eder
  print("Ana işlem devam ediyor")
  await sleep(500)
  print("Ana işlem bitti")
end
```

### Async/Await Best Practices

```sky
# ✅ İyi: Hata yönetimi ile
async function guvenliIslem()
  try
    let result = await riskliIslem()
    return result
  catch error
    print("İşlem başarısız:", error)
    return null
  end
end

# ✅ İyi: Paralel işlemler
async function hizliVeriCekme()
  let promise1 = veriCek(1)
  let promise2 = veriCek(2)
  let promise3 = veriCek(3)
  
  let results = [
    await promise1,
    await promise2,
    await promise3
  ]
  
  return results
end

# ❌ Kötü: Gereksiz await
async function yavasIslem()
  let result1 = await basitIslem()  # await gereksiz
  let result2 = await basitIslem()  # await gereksiz
  return result1 + result2
end
```

---

## 🎯 Pattern Matching

### Enum Tanımlama

```sky
enum Sonuc
  Basarili(int)
  Hata(string)
end
```

### Match İfadesi

```sky
let islem = Basarili(42)

match islem
  Basarili(deger) => print("Başarılı:", deger)
  Hata(mesaj) => print("Hata:", mesaj)
end
```

### Karmaşık Örnek

```sky
enum Durum
  Beklemede
  Isleniyor
  Tamamlandi(string)
  Hatali(int, string)
end

let durum = Tamamlandi("Dosya kaydedildi")

match durum
  Beklemede => print("Bekleniyor...")
  Isleniyor => print("İşleniyor...")
  Tamamlandi(mesaj) => print("Tamamlandı:", mesaj)
  Hatali(kod, aciklama) => print("Hata", kod, ":", aciklama)
end
```

### Exhaustiveness (Kapsayıcılık)

SKY derleyicisi, `match` ifadelerinin tüm durumları kapsadığını kontrol eder:

```sky
enum Renk
  Kirmizi
  Yesil
  Mavi
end

let renk = Kirmizi

match renk
  Kirmizi => print("Kırmızı")
  Yesil => print("Yeşil")
  # Mavi durumu eksik - Derleyici hata verir!
end
```

**Çözüm**: Tüm durumları kapsayın:

```sky
match renk
  Kirmizi => print("Kırmızı")
  Yesil => print("Yeşil")
  Mavi => print("Mavi")
end
```

### Guards (Korumalar)

Koşullu pattern matching için `if` kullanın:

```sky
enum Sayi
  Pozitif(int)
  Negatif(int)
  Sifir
end

let sayi = Pozitif(15)

match sayi
  Pozitif(n) if n > 10 => print("Büyük pozitif sayı:", n)
  Pozitif(n) if n <= 10 => print("Küçük pozitif sayı:", n)
  Negatif(n) => print("Negatif sayı:", n)
  Sifir => print("Sıfır")
end
```

### Wildcard (_) Kullanımı

Bazı durumları yok saymak için `_` kullanın:

```sky
enum Sonuc
  Basarili(int)
  Hata(string)
  Uyari(string)
end

let sonuc = Basarili(42)

match sonuc
  Basarili(deger) => print("Başarılı:", deger)
  _ => print("Hata veya uyarı")  # Hata ve Uyari durumlarını yakalar
end
```

### Nested Pattern Matching

İç içe geçmiş yapıları eşleştirin:

```sky
enum Adres
  Ev(string, int)      # Sokak, numara
  Is(string, string)   # Şirket, departman
end

enum Kisi
  Ogrenci(string, int)           # İsim, yaş
  Calisan(string, int, Adres)    # İsim, yaş, adres
end

let kisi = Calisan("Ahmet", 30, Ev("Cumhuriyet Caddesi", 15))

match kisi
  Ogrenci(isim, yas) => print("Öğrenci:", isim, "yaş:", yas)
  Calisan(isim, yas, Ev(sokak, numara)) => 
    print("Çalışan:", isim, "ev adresi:", sokak, numara)
  Calisan(isim, yas, Is(sirket, departman)) => 
    print("Çalışan:", isim, "iş adresi:", sirket, departman)
end
```

### List Pattern Matching

Liste yapılarını eşleştirin:

```sky
function liste_analiz_et(liste: list): void
  match liste
    [] => print("Boş liste")
    [x] => print("Tek eleman:", x)
    [x, y] => print("İki eleman:", x, y)
    [x, y, z] => print("Üç eleman:", x, y, z)
    [x, ...rest] => print("İlk eleman:", x, "Kalan:", len(rest), "eleman")
  end
end

# Kullanım
liste_analiz_et([])                    # Boş liste
liste_analiz_et([42])                  # Tek eleman: 42
liste_analiz_et([1, 2, 3, 4, 5])      # İlk eleman: 1 Kalan: 4 eleman
```

### Tuple Pattern Matching

Tuple yapılarını eşleştirin:

```sky
function koordinat_analiz_et(koordinat: tuple): void
  match koordinat
    (0, 0) => print("Orijin noktası")
    (x, 0) => print("X ekseni üzerinde:", x)
    (0, y) => print("Y ekseni üzerinde:", y)
    (x, y) if x > 0 and y > 0 => print("1. bölge:", x, y)
    (x, y) if x < 0 and y > 0 => print("2. bölge:", x, y)
    (x, y) if x < 0 and y < 0 => print("3. bölge:", x, y)
    (x, y) if x > 0 and y < 0 => print("4. bölge:", x, y)
    _ => print("Bilinmeyen koordinat")
  end
end

# Kullanım
koordinat_analiz_et((0, 0))      # Orijin noktası
koordinat_analiz_et((5, 3))      # 1. bölge: 5 3
koordinat_analiz_et((-2, 4))     # 2. bölge: -2 4
```

### Pattern Matching Best Practices

```sky
# ✅ İyi: Exhaustive matching
match durum
  Basarili(value) => handle_success(value)
  Hata(message) => handle_error(message)
  Uyari(message) => handle_warning(message)
end

# ✅ İyi: Guard kullanımı
match sayi
  Pozitif(n) if n > 100 => print("Çok büyük sayı")
  Pozitif(n) if n > 10 => print("Büyük sayı")
  Pozitif(n) => print("Küçük pozitif sayı")
  Negatif(n) => print("Negatif sayı")
  Sifir => print("Sıfır")
end

# ✅ İyi: Wildcard kullanımı
match sonuc
  Basarili(value) => process_success(value)
  _ => print("Başarısız işlem")
end

# ❌ Kötü: Eksik durumlar
match renk
  Kirmizi => print("Kırmızı")
  # Diğer renkler eksik!
end

# ❌ Kötü: Gereksiz wildcard
match sayi
  Pozitif(n) => print("Pozitif:", n)
  _ => print("Diğer")  # Negatif ve Sıfır durumları kaybolur
end
```

---

## 🛠️ Yerleşik Fonksiyonlar

SKY'da her zaman kullanılabilir olan temel fonksiyonlar:

### Çıktı Fonksiyonları

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `print(...)` | Değerleri yazdır | `print("Merhaba", 42)` |
| `println(...)` | Değerleri yazdır + yeni satır | `println("Merhaba")` |

### Tip Dönüşümleri

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `int(value)` | String/float'u int'e çevir | `int("42")` → `42` |
| `float(value)` | String/int'i float'a çevir | `float("3.14")` → `3.14` |
| `str(value)` | Herhangi bir değeri string'e çevir | `str(42)` → `"42"` |
| `bool(value)` | Değeri boolean'a çevir | `bool(1)` → `true` |

### Koleksiyon Fonksiyonları

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `len(collection)` | Uzunluk | `len([1,2,3])` → `3` |
| `join(separator, list)` | Listeyi string'e çevir | `join("-", ["a","b"])` → `"a-b"` |
| `range(start, end)` | Sayı aralığı | `range(1, 5)` → `[1,2,3,4]` |

### Tip Kontrolü

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `type(value)` | Değerin tipini döndür | `type(42)` → `"int"` |
| `is_int(value)` | Int mi? | `is_int(42)` → `true` |
| `is_string(value)` | String mi? | `is_string("hello")` → `true` |
| `is_list(value)` | Liste mi? | `is_list([1,2])` → `true` |
| `is_dict(value)` | Sözlük mü? | `is_dict({})` → `true` |

### Zaman Fonksiyonları

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `time_now()` | Şu anki zaman (ms) | `time_now()` → `1698000000000` |
| `sleep(ms)` | Bekle (milisaniye) | `sleep(1000)` |

### Sabitler

| Sabit | Değer | Açıklama |
|-------|-------|----------|
| `null` | `nil` | Boş değer |
| `nil` | `nil` | Boş değer (alias) |
| `true` | `true` | Doğru |
| `false` | `false` | Yanlış |

### Örnek Kullanım

```sky
# Temel çıktı
print("Merhaba", "Dünya")
println("Yeni satır ile")

# Tip dönüşümleri
let num_str = "42"
let num = int(num_str)
let pi_str = str(3.14159)

# Koleksiyon işlemleri
let fruits = ["elma", "armut", "kiraz"]
let joined = join(", ", fruits)
print(joined)  # "elma, armut, kiraz"

# Tip kontrolü
if is_string(value)
  print("Bu bir string: " + value)
end

# Zaman işlemleri
let start = time_now()
# İşlem yap...
let end_time = time_now()
let duration = end_time - start
print("İşlem " + duration + " ms sürdü")
```

---

## 📚 Standart Kütüphane

### Dosya İşlemleri (FS)

#### Temel Dosya İşlemleri

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `fs_read_text(path)` | Dosyayı oku | `fs_read_text("data.txt")` |
| `fs_write_text(path, content)` | Dosyaya yaz | `fs_write_text("out.txt", "data")` |
| `fs_exists(path)` | Dosya var mı? | `fs_exists("file.txt")` |
| `fs_read_bytes(path)` | Binary okuma | `fs_read_bytes("image.png")` |
| `fs_write_bytes(path, data)` | Binary yazma | `fs_write_bytes("out.bin", bytes)` |

#### Dizin İşlemleri

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `fs_mkdir(path)` | Dizin oluştur | `fs_mkdir("new_folder")` |
| `fs_rmdir(path)` | Dizin sil | `fs_rmdir("old_folder")` |
| `fs_list_dir(path)` | Dizin listele | `fs_list_dir(".")` |
| `fs_delete(path)` | Dosya/dizin sil | `fs_delete("file.txt")` |

#### Örnek Kullanım

```sky
# Dosya yazma ve okuma
fs_write_text("test.txt", "Merhaba Dünya")
let icerik = fs_read_text("test.txt")
print(icerik)

# Dosya varlık kontrolü
if fs_exists("test.txt")
  print("Dosya mevcut")
end

# Dizin işlemleri
fs_mkdir("yeni_klasor")
let dosyalar = fs_list_dir(".")
for dosya in dosyalar
  print("Dosya:", dosya)
end

# Hata yönetimi ile dosya okuma
try
  let data = fs_read_text("olmayan_dosya.txt")
  print(data)
catch error
  print("Dosya okunamadı:", error)
end
```

### İşletim Sistemi (OS)

#### Sistem Bilgileri

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `os_platform()` | İşletim sistemi | `os_platform()` → `"darwin"` |
| `os_arch()` | Mimari | `os_arch()` → `"x86_64"` |
| `os_getcwd()` | Çalışma dizini | `os_getcwd()` → `"/home/user"` |
| `os_getenv(name)` | Ortam değişkeni | `os_getenv("HOME")` |
| `os_setenv(name, value)` | Ortam değişkeni ata | `os_setenv("DEBUG", "1")` |

#### Süreç İşlemleri

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `os_run_command(cmd)` | Komut çalıştır | `os_run_command("ls -la")` |
| `os_exit(code)` | Programı sonlandır | `os_exit(0)` |
| `os_args()` | Komut satırı argümanları | `os_args()` → `["prog", "arg1"]` |

#### Örnek Kullanım

```sky
# Platform bilgisi
let platform = os_platform()
let arch = os_arch()
print("Platform:", platform, "Arch:", arch)

# Çalışma dizini
let cwd = os_getcwd()
print("Dizin:", cwd)

# Ortam değişkeni
let home = os_getenv("HOME")
print("Ana dizin:", home)

# Komut çalıştırma
let result = os_run_command("echo 'Merhaba'")
print("Komut çıktısı:", result)

# Komut satırı argümanları
let args = os_args()
print("Argümanlar:", args)
```

### HTTP İşlemleri

#### Temel HTTP Metodları

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `http_get(url)` | GET isteği | `http_get("https://api.example.com")` |
| `http_post(url, data)` | POST isteği | `http_post("https://api.example.com", data)` |
| `http_put(url, data)` | PUT isteği | `http_put("https://api.example.com/1", data)` |
| `http_delete(url)` | DELETE isteği | `http_delete("https://api.example.com/1")` |

#### HTTP Response Özellikleri

| Özellik | Açıklama | Örnek |
|---------|----------|-------|
| `response.status_code` | HTTP durum kodu | `200`, `404`, `500` |
| `response.body` | Response gövdesi | `"{\"name\":\"John\"}"` |
| `response.headers` | Response başlıkları | `{"Content-Type": "application/json"}` |

#### Örnek Kullanım

```sky
# GET isteği
let response = http_get("https://api.github.com/users/octocat")
print("Status:", response.status_code)
print("Body:", response.body)

# POST isteği
let data = {"name": "John", "age": 30}
let response = http_post("https://api.example.com/users", data)
if response.status_code == 201
  print("Kullanıcı oluşturuldu")
end

# Hata yönetimi ile HTTP
try
  let response = http_get("https://api.example.com/data")
  if response.status_code == 200
    print("Başarılı:", response.body)
  else
    print("Hata kodu:", response.status_code)
  end
catch error
  print("Ağ hatası:", error)
end
```

### Kriptografi (Crypto)

#### Hash Fonksiyonları

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `crypto_md5(data)` | MD5 hash | `crypto_md5("şifre123")` |
| `crypto_sha1(data)` | SHA1 hash | `crypto_sha1("şifre123")` |
| `crypto_sha256(data)` | SHA256 hash | `crypto_sha256("şifre123")` |
| `crypto_sha512(data)` | SHA512 hash | `crypto_sha512("şifre123")` |

#### Şifreleme Fonksiyonları

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `crypto_aes_encrypt(data, key)` | AES şifreleme | `crypto_aes_encrypt("data", "key")` |
| `crypto_aes_decrypt(data, key)` | AES şifre çözme | `crypto_aes_decrypt(encrypted, "key")` |
| `crypto_hmac(data, key)` | HMAC imza | `crypto_hmac("data", "secret")` |

#### Örnek Kullanım

```sky
# Hash işlemleri
let hash_md5 = crypto_md5("şifre123")
let hash_sha256 = crypto_sha256("şifre123")
print("MD5:", hash_md5)
print("SHA256:", hash_sha256)

# Şifreleme
let data = "Gizli veri"
let key = "gizli_anahtar"
let encrypted = crypto_aes_encrypt(data, key)
let decrypted = crypto_aes_decrypt(encrypted, key)
print("Şifrelenmiş:", encrypted)
print("Çözülmüş:", decrypted)
```

### JSON İşlemleri

```sky
# JSON encode
let veri = {"isim": "Ahmet", "yas": "25"}
let json_str = json_encode(veri)
print(json_str)  # {"isim":"Ahmet","yas":"25"}

# JSON decode
let parsed = json_decode(json_str)
print(parsed["isim"])  # Ahmet
```

### Zaman ve Tarih (Time)

#### Temel Zaman Fonksiyonları

| Fonksiyon | Açıklama | Örnek |
|-----------|----------|-------|
| `time_now()` | Şu anki zaman (ms) | `time_now()` → `1698000000000` |
| `sleep(ms)` | Bekle (milisaniye) | `sleep(1000)` |
| `time_format(timestamp, format)` | Zamanı formatla | `time_format(now, "%Y-%m-%d %H:%M")` |
| `time_parse(date_string, format)` | String'den zaman | `time_parse("2023-10-22", "%Y-%m-%d")` |
| `time_add(timestamp, duration)` | Zaman ekle | `time_add(now, "1h30m")` |
| `time_diff(timestamp1, timestamp2)` | Zaman farkı | `time_diff(end, start)` |

#### Zaman Formatları

| Format | Açıklama | Örnek |
|--------|----------|-------|
| `%Y` | Yıl (4 haneli) | `2023` |
| `%m` | Ay (01-12) | `10` |
| `%d` | Gün (01-31) | `22` |
| `%H` | Saat (00-23) | `14` |
| `%M` | Dakika (00-59) | `30` |
| `%S` | Saniye (00-59) | `45` |
| `%A` | Haftanın günü | `Pazar` |
| `%B` | Ay adı | `Ekim` |

#### Örnek Kullanım

```sky
# Şu anki zamanı al
let now = time_now()
print("Timestamp:", now)

# Zamanı formatla
let formatted = time_format(now, "%Y-%m-%d %H:%M:%S")
print("Formatted:", formatted)  # 2023-10-22 14:30:45

# String'den zaman parse et
let parsed = time_parse("2023-10-22 14:30:00", "%Y-%m-%d %H:%M:%S")
print("Parsed:", parsed)

# Zaman ekleme
let future = time_add(now, "2h30m")  # 2 saat 30 dakika ekle
let future_formatted = time_format(future, "%Y-%m-%d %H:%M")
print("Future:", future_formatted)

# Zaman farkı
let start = time_now()
sleep(2000)  # 2 saniye bekle
let end_time = time_now()
let duration = time_diff(end_time, start)
print("Duration:", duration, "ms")

# Tarih karşılaştırma
let today = time_now()
let tomorrow = time_add(today, "24h")
if tomorrow > today
  print("Yarın bugünden sonra")
end
```

#### Süre Formatları

| Format | Açıklama | Örnek |
|--------|----------|-------|
| `s` | Saniye | `30s` |
| `m` | Dakika | `5m` |
| `h` | Saat | `2h` |
| `d` | Gün | `7d` |
| `w` | Hafta | `2w` |

```sky
# Süre örnekleri
let short_duration = "30s"      # 30 saniye
let medium_duration = "2h30m"    # 2 saat 30 dakika
let long_duration = "1w3d12h"    # 1 hafta 3 gün 12 saat
```

### Rastgele Sayı (Random)

```sky
# Rastgele tam sayı (0-99)
let sayi = rand_int(100)
print("Rastgele:", sayi)

# UUID oluşturma
let uuid = rand_uuid()
print("UUID:", uuid)
```

### String Metodları

```sky
let metin = "merhaba dünya"

print(metin.upper())       # MERHABA DÜNYA
print(metin.lower())       # merhaba dünya
print(metin.split())       # [merhaba, dünya]
print("  test  ".strip())  # test
```

### Tip Dönüşümleri

```sky
# String'den sayıya
let sayi = int("42")
let ondalik = float("3.14")

# Sayıdan string'e
let metin = str(123)

# Boolean'a
let dogru = bool(1)
let yanlis = bool(0)

# Tip kontrolü
print(type(42))        # int
print(type("hello"))   # string
```

---

## 📁 Modül Sistemi

SKY'da modül sistemi, kodunuzu organize etmenizi ve yeniden kullanılabilir parçalara bölmenizi sağlar.

### Modül İçe Aktarma

#### Temel Import

```sky
# Modülü tam olarak içe aktar
import math

# Kullanım
let result = math.add(5, 3)
```

#### Alias ile Import

```sky
# Modülü farklı isimle içe aktar
import math as matematik

# Kullanım
let result = matematik.add(5, 3)
```

#### Seçici Import

```sky
# Sadece belirli fonksiyonları içe aktar
import math { add, subtract }

# Kullanım
let result = add(5, 3)
let diff = subtract(10, 4)
```

### Modül Oluşturma

#### Basit Modül (math.sky)

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

# Public olmayan fonksiyon (özel)
function _internal_calc(x: int): int
  return x * 2
end
```

#### Kullanım

```sky
# main.sky
import math

function main: void
  let sum = math.add(10, 20)
  let product = math.multiply(5, 6)
  print("Toplam:", sum)
  print("Çarpım:", product)
end
```

### Modül Yapısı

```
proje/
├── main.sky
├── utils/
│   ├── math.sky
│   ├── string.sky
│   └── io.sky
└── models/
    ├── user.sky
    └── product.sky
```

### Dairesel Bağımlılık (Circular Dependencies)

SKY, dairesel bağımlılıkları otomatik olarak tespit eder ve hata verir:

```sky
# A.sky
import B
# ...

# B.sky  
import A  # ❌ Hata: Dairesel bağımlılık!
```

**Çözüm**: Ortak kodu ayrı bir modüle taşıyın:

```sky
# common.sky
function shared_function()
  # Ortak kod
end

# A.sky
import common
# ...

# B.sky
import common
# ...
```

### Modül Arama Yolu

SKY modülleri şu sırayla arar:

1. **Göreceli yol**: `./utils/math.sky`
2. **Proje kökü**: `./math.sky`
3. **Standart kütüphane**: `math` (built-in)
4. **Wing paketleri**: `wing install` ile yüklenen paketler

### Modül Örnekleri

#### HTTP Modülü (http.sky)

```sky
# http.sky
function get(url: string): dict
  # HTTP GET implementasyonu
  return {"status": 200, "body": "data"}
end

function post(url: string, data: dict): dict
  # HTTP POST implementasyonu
  return {"status": 201, "body": "created"}
end
```

#### Kullanım

```sky
import http

let response = http.get("https://api.example.com")
print("Status:", response["status"])
```

---

## 📦 Paket Yönetimi (Wing)

### Yeni Proje Oluşturma

```bash
wing init
```

### Paket Kurulumu

```bash
wing install http
wing install json@1.0.0  # Belirli versiyon
```

### Paket Güncelleme

```bash
wing update           # Tüm paketler
wing update http      # Belirli paket
```

### Proje Derleme

```bash
wing build
```

### Paket Yayınlama

```bash
wing publish
```

---

## 🔒 Unsafe Bloklar

Düşük seviye işlemler için:

```sky
unsafe
  let pointer = 0xDEADBEEF
  # Ham bellek işlemleri
end
```

⚠️ **Dikkat**: Unsafe bloklar dikkatli kullanılmalıdır!

---

## 💡 En İyi Pratikler

### 1. İsimlendirme

```sky
# Değişkenler: snake_case
let kullanici_adi = "ahmet"
let toplam_fiyat = 100

# Fonksiyonlar: snake_case
function hesapla_toplam()
  # ...
end

# Sınıflar: PascalCase
class KullaniciYoneticisi
  # ...
end

# Sabitler: UPPER_CASE
const MAX_DENEME = 3
```

### 2. Hata Yönetimi

```sky
enum Sonuc
  Basarili(string)
  Hata(string)
end

function dosya_oku(yol)
  if fs_exists(yol)
    let icerik = fs_read_text(yol)
    return Basarili(icerik)
  else
    return Hata("Dosya bulunamadı")
  end
end
```

### 3. Dokümantasyon

```sky
# Kullanıcı bilgilerini getirir
#
# Parametreler:
#   id: Kullanıcı ID'si
#
# Döndürür:
#   Kullanıcı bilgileri veya nil
function kullanici_getir(id)
  # ...
end
```

---

## 🚀 Örnek Proje

```sky
# main.sky - Basit bir web scraper

import http
import json

async function sayfa_cek(url)
  let response = await http.get(url)
  return response.body
end

async function veri_isle(html)
  # HTML işleme
  let veri = {"baslik": "Örnek", "icerik": html}
  return veri
end

async function kaydet(veri)
  let json_str = json_encode(veri)
  fs_write_text("sonuc.json", json_str)
  print("Kaydedildi!")
end

async function main
  print("Sayfa çekiliyor...")
  let html = await sayfa_cek("https://example.com")
  
  print("Veri işleniyor...")
  let veri = await veri_isle(html)
  
  print("Kaydediliyor...")
  await kaydet(veri)
  
  print("Tamamlandı!")
end
```

---

## 📖 Daha Fazla Bilgi

- **GitHub**: https://github.com/mburakmmm/sky-lang
- **Örnekler**: `examples/` dizini
- **Testler**: `tests/` dizini

---

**SKY Programlama Dili** - Hızlı, Güvenli, Kolay 🚀

