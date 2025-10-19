# 📚 SKY Programlama Dili - API Referansı (Türkçe)

## 🌟 Giriş

SKY, modern, tip-güvenli, Go ile yazılmış bir programlama dilidir. Python'un sadeliği ile Go'nun performansını birleştirir.

---

## 📖 İçindekiler

1. [Temel Sözdizimi](#temel-sözdizimi)
2. [Veri Tipleri](#veri-tipleri)
3. [Fonksiyonlar](#fonksiyonlar)
4. [Kontrol Yapıları](#kontrol-yapıları)
5. [Sınıflar ve OOP](#sınıflar-ve-oop)
6. [Async/Await](#asyncawait)
7. [Pattern Matching](#pattern-matching)
8. [Standart Kütüphane](#standart-kütüphane)
9. [Paket Yönetimi](#paket-yönetimi)

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
print(names.upper())  # Listeyi büyük harfe çevir
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

### Kalıtım (Inheritance)

```sky
class Hayvan
  function init(isim)
    self.isim = isim
  end
  
  function ses_cikar()
    print(self.isim, "ses çıkarıyor")
  end
end

class Kedi
  function init(isim, cins)
    super.init(isim)
    self.cins = cins
  end
  
  function ses_cikar()
    print(self.isim, "miyavlıyor")
  end
end

let kedi = Kedi("Pamuk", "Van")
kedi.ses_cikar()  # Pamuk miyavlıyor
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

### Birden Fazla Asenkron İşlem

```sky
async function kullaniciGetir(id)
  return {"id": str(id), "isim": "User" + str(id)}
end

async function main
  let user1 = await kullaniciGetir(1)
  let user2 = await kullaniciGetir(2)
  
  print("Kullanıcı 1:", user1)
  print("Kullanıcı 2:", user2)
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

---

## 📚 Standart Kütüphane

### Dosya İşlemleri (FS)

```sky
# Dosya yazma
fs_write_text("test.txt", "Merhaba Dünya")

# Dosya okuma
let icerik = fs_read_text("test.txt")
print(icerik)

# Dosya varlık kontrolü
if fs_exists("test.txt")
  print("Dosya mevcut")
end
```

### İşletim Sistemi (OS)

```sky
# Platform bilgisi
let platform = os_platform()
print("Platform:", platform)

# Çalışma dizini
let cwd = os_getcwd()
print("Dizin:", cwd)

# Ortam değişkeni
let home = os_getenv("HOME")
print("Ana dizin:", home)
```

### Kriptografi (Crypto)

```sky
# MD5 hash
let hash_md5 = crypto_md5("şifre123")
print("MD5:", hash_md5)

# SHA256 hash
let hash_sha256 = crypto_sha256("şifre123")
print("SHA256:", hash_sha256)
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

```sky
# Şu anki zaman damgası
let simdi = time_now()
print("Timestamp:", simdi)

# Bekleme (milisaniye)
time_sleep(1000)  # 1 saniye bekle
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

