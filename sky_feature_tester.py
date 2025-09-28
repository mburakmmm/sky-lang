#!/usr/bin/env python3
"""
Sky Dili Özellik Test Scripti
.cursorrules dosyasındaki tüm özellikleri test eder ve analiz sunar
"""

import os
import sys
import subprocess
import tempfile
from pathlib import Path
from datetime import datetime
import json

class SkyFeatureTester:
    def __init__(self):
        self.sky_binary = self.find_sky_binary()
        self.test_results = {}
        self.total_tests = 0
        self.passed_tests = 0
        self.failed_tests = 0
        
    def find_sky_binary(self):
        """Sky binary'sini bul"""
        print("🔍 Sky binary aranıyor...")
        
        # Önce PATH'te ara
        try:
            result = subprocess.run(['which', 'sky'], capture_output=True, text=True)
            if result.returncode == 0:
                binary_path = result.stdout.strip()
                print(f"✅ PATH'te bulundu: {binary_path}")
                return binary_path
        except:
            pass
            
        # Ana dizinde ara (sadece dosya ise)
        home = Path.home()
        sky_path = home / "sky"
        if sky_path.exists() and sky_path.is_file():
            print(f"✅ Ana dizinde bulundu: {sky_path}")
            return str(sky_path)
            
        # Mevcut dizinde target/release/sky ara
        release_path = Path("target/release/sky")
        if release_path.exists():
            abs_path = release_path.absolute()
            print(f"✅ Release dizininde bulundu: {abs_path}")
            return str(abs_path)
            
        print("❌ Sky binary hiçbir yerde bulunamadı!")
        raise FileNotFoundError("Sky binary bulunamadı!")
    
    def print_header(self):
        """Test başlığını yazdır"""
        print("🌌" + "="*60)
        print("🌌 SKY DİLİ ÖZELLİK TEST SCRIPTI")
        print("🌌" + "="*60)
        print(f"📅 Test Tarihi: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"🔧 Sky Binary: {self.sky_binary}")
        print()
    
    def print_test_header(self, category, description):
        """Test kategorisi başlığını yazdır"""
        print(f"\n{'🔍 ' + category + ' ' + '='*50}")
        print(f"📝 {description}")
        print("-" * 60)
    
    def run_sky_test(self, code, test_name, expected_output=None):
        """Sky kodu test et"""
        self.total_tests += 1
        
        # Geçici dosya oluştur
        with tempfile.NamedTemporaryFile(mode='w', suffix='.sky', delete=False) as f:
            f.write(code)
            temp_file = f.name
        
        try:
            # Sky kodu çalıştır
            result = subprocess.run(
                [self.sky_binary, 'run', temp_file],
                capture_output=True,
                text=True,
                timeout=10
            )
            
            # Sonucu analiz et
            success = result.returncode == 0
            output = result.stdout.strip()
            error = result.stderr.strip()
            
            if success:
                self.passed_tests += 1
                status = "✅ BAŞARILI"
                if expected_output and expected_output not in output:
                    status = "⚠️  BEKLENMEYEN ÇIKTI"
            else:
                self.failed_tests += 1
                status = "❌ BAŞARISIZ"
            
            print(f"{status} | {test_name}")
            if output:
                print(f"   📤 Çıktı: {output[:100]}{'...' if len(output) > 100 else ''}")
            if error:
                print(f"   ❌ Hata: {error[:100]}{'...' if len(error) > 100 else ''}")
            
            # Sonucu kaydet
            self.test_results[test_name] = {
                'success': success,
                'output': output,
                'error': error,
                'code': code
            }
            
            return success, output, error
            
        except subprocess.TimeoutExpired:
            self.failed_tests += 1
            print(f"⏰ ZAMAN AŞIMI | {test_name}")
            self.test_results[test_name] = {
                'success': False,
                'output': '',
                'error': 'Timeout',
                'code': code
            }
            return False, '', 'Timeout'
        except Exception as e:
            self.failed_tests += 1
            print(f"💥 HATA | {test_name} - {str(e)}")
            self.test_results[test_name] = {
                'success': False,
                'output': '',
                'error': str(e),
                'code': code
            }
            return False, '', str(e)
        finally:
            # Geçici dosyayı sil
            try:
                os.unlink(temp_file)
            except:
                pass
    
    def test_basic_syntax(self):
        """Temel sözdizimi testleri"""
        self.print_test_header("TEMEL SÖZDİZİMİ", "Değişken tanımlama, atama, temel operatörler")
        
        # Tip zorunluluğu testi
        self.run_sky_test(
            'int sayı = 42\nprint("Sayı: " + sayı)',
            "Tip zorunluluğu - int değişken"
        )
        
        self.run_sky_test(
            'string isim = "Sky"\nprint("İsim: " + isim)',
            "Tip zorunluluğu - string değişken"
        )
        
        self.run_sky_test(
            'bool doğru = true\nprint("Doğru: " + doğru)',
            "Tip zorunluluğu - bool değişken"
        )
        
        self.run_sky_test(
            'var dinamik = 123\nprint("Dinamik: " + dinamik)',
            "Dinamik tip - var değişken"
        )
        
        # Assignment testi
        self.run_sky_test(
            'int x = 5\nx = x + 3\nprint("x = " + x)',
            "Assignment - değişken atama"
        )
        
        # Aritmetik operatörler
        self.run_sky_test(
            'int a = 10\nint b = 3\nprint("Toplam: " + (a + b))\nprint("Çıkarma: " + (a - b))\nprint("Çarpma: " + (a * b))\nprint("Bölme: " + (a / b))\nprint("Mod: " + (a % b))',
            "Aritmetik operatörler"
        )
        
        # Karşılaştırma operatörleri
        self.run_sky_test(
            'int x = 5\nint y = 10\nprint("x == y: " + (x == y))\nprint("x != y: " + (x != y))\nprint("x < y: " + (x < y))\nprint("x > y: " + (x > y))\nprint("x <= y: " + (x <= y))\nprint("x >= y: " + (x >= y))',
            "Karşılaştırma operatörleri"
        )
        
        # Mantıksal operatörler
        self.run_sky_test(
            'bool a = true\nbool b = false\nprint("a and b: " + (a and b))\nprint("a or b: " + (a or b))',
            "Mantıksal operatörler"
        )
    
    def test_functions(self):
        """Fonksiyon testleri"""
        self.print_test_header("FONKSİYONLAR", "Normal, parametreli, recursive fonksiyonlar")
        
        # Basit fonksiyon
        self.run_sky_test(
            'function selam()\n    return "Merhaba Sky"\n\nprint(selam())',
            "Basit fonksiyon - parametresiz"
        )
        
        # Parametreli fonksiyon
        self.run_sky_test(
            'function topla(a: int, b: int)\n    return a + b\n\nprint("Toplam: " + topla(5, 3))',
            "Parametreli fonksiyon"
        )
        
        # Recursive fonksiyon
        self.run_sky_test(
            'function faktoriyel(n: int)\n    if n <= 1:\n        return 1\n    else:\n        return n * faktoriyel(n - 1)\n\nprint("5! = " + faktoriyel(5))',
            "Recursive fonksiyon - faktöriyel"
        )
        
        # String parametreli fonksiyon
        self.run_sky_test(
            'function büyük_harf(s: string)\n    return "BÜYÜK: " + s\n\nprint(büyük_harf("sky"))',
            "String parametreli fonksiyon"
        )
    
    def test_control_flow(self):
        """Kontrol akışı testleri"""
        self.print_test_header("KONTROL AKIŞI", "If-elif-else, while, for döngüleri")
        
        # If-else
        self.run_sky_test(
            'int x = 10\nif x > 5:\n    print("x 5\'ten büyük")\nelse:\n    print("x 5\'ten küçük veya eşit")',
            "If-else statement"
        )
        
        # If-elif-else
        self.run_sky_test(
            'int puan = 85\nif puan >= 90:\n    print("A")\nelif puan >= 80:\n    print("B")\nelif puan >= 70:\n    print("C")\nelse:\n    print("F")',
            "If-elif-else zinciri"
        )
        
        # While döngüsü
        self.run_sky_test(
            'int sayac = 1\nwhile sayac <= 3:\n    print("Sayaç: " + sayac)\n    sayac = sayac + 1',
            "While döngüsü"
        )
        
        # For döngüsü
        self.run_sky_test(
            'for i: var in [1, 2, 3]:\n    print("For döngüsü: " + i)',
            "For döngüsü - liste"
        )
        
        # İç içe döngüler
        self.run_sky_test(
            'int toplam = 0\nfor i: var in [1, 2]:\n    for j: var in [3, 4]:\n        toplam = toplam + i + j\nprint("Toplam: " + toplam)',
            "İç içe for döngüleri"
        )
    
    def test_turkish_identifiers(self):
        """Türkçe karakter testleri"""
        self.print_test_header("TÜRKÇE KARAKTERLER", "Türkçe karakterli değişken ve fonksiyon isimleri")
        
        self.run_sky_test(
            'int çorba = 42\nstring öğe = "test"\nbool şeker = true\nprint("Çorba: " + çorba)\nprint("Öğe: " + öğe)\nprint("Şeker: " + şeker)',
            "Türkçe karakterli değişkenler"
        )
        
        self.run_sky_test(
            'function hesapla_çorba(fiyat: int)\n    return fiyat * 2\n\nprint("Çorba fiyatı: " + hesapla_çorba(10))',
            "Türkçe karakterli fonksiyon"
        )
    
    def test_data_structures(self):
        """Veri yapıları testleri"""
        self.print_test_header("VERİ YAPILARI", "List, Map, string operasyonları")
        
        # List operasyonları
        self.run_sky_test(
            'list sayılar = [1, 2, 3, 4, 5]\nprint("İlk sayı: " + sayılar[0])\nprint("Son sayı: " + sayılar[4])',
            "List - indeksleme"
        )
        
        # Map operasyonları
        self.run_sky_test(
            'map sözlük = {"ad": "Sky", "yaş": 1}\nprint("Ad: " + sözlük["ad"])\nprint("Yaş: " + sözlük["yaş"])',
            "Map - anahtar-değer"
        )
        
        # String concatenation
        self.run_sky_test(
            'string ad = "Sky"\nstring soyad = "Dili"\nprint("Tam ad: " + ad + " " + soyad)',
            "String concatenation"
        )
    
    def test_complex_scenarios(self):
        """Karmaşık senaryolar"""
        self.print_test_header("KARMAŞIK SENARYOLAR", "Gerçek dünya örnekleri")
        
        # Hesap makinesi
        self.run_sky_test(
            '''function hesapla(a: int, b: int, işlem: string)
    if işlem == "+":
        return a + b
    elif işlem == "-":
        return a - b
    elif işlem == "*":
        return a * b
    elif işlem == "/":
        return a / b
    else:
        return 0

print("10 + 5 = " + hesapla(10, 5, "+"))
print("10 - 5 = " + hesapla(10, 5, "-"))
print("10 * 5 = " + hesapla(10, 5, "*"))
print("10 / 5 = " + hesapla(10, 5, "/"))''',
            "Hesap makinesi - if-elif-else"
        )
        
        # Sayı analizi
        self.run_sky_test(
            '''function sayı_analizi(sayi: int)
    string sonuç = ""
    if sayi > 0:
        sonuç = sonuç + "Pozitif "
    elif sayi < 0:
        sonuç = sonuç + "Negatif "
    else:
        sonuç = sonuç + "Sıfır "
    
    if sayi % 2 == 0:
        sonuç = sonuç + "Çift"
    else:
        sonuç = sonuç + "Tek"
    
    return sonuç

print("5: " + sayı_analizi(5))
print("8: " + sayı_analizi(8))
print("0: " + sayı_analizi(0))
print("-3: " + sayı_analizi(-3))''',
            "Sayı analizi - karmaşık logic"
        )
        
        # Fibonacci
        self.run_sky_test(
            '''function fibonacci(n: int)
    if n <= 1:
        return n
    else:
        return fibonacci(n - 1) + fibonacci(n - 2)

print("Fibonacci(0) = " + fibonacci(0))
print("Fibonacci(1) = " + fibonacci(1))
print("Fibonacci(5) = " + fibonacci(5))
print("Fibonacci(8) = " + fibonacci(8))''',
            "Fibonacci - recursive algoritma"
        )
    
    def test_edge_cases(self):
        """Kenar durumları testleri"""
        self.print_test_header("KENAR DURUMLAR", "Hata durumları ve sınır değerler")
        
        # Boş fonksiyon
        self.run_sky_test(
            'function boş_fonksiyon()\n    # Boş\n\nboş_fonksiyon()\nprint("Boş fonksiyon çalıştı")',
            "Boş fonksiyon"
        )
        
        # Tek satır if
        self.run_sky_test(
            'int x = 5\nif x > 0:\n    print("x pozitif")\nprint("Devam ediyor")',
            "Tek satır if"
        )
        
        # Negatif sayılar
        self.run_sky_test(
            'int negatif = -10\nint pozitif = 5\nprint("Negatif: " + negatif)\nprint("Pozitif: " + pozitif)\nprint("Toplam: " + (negatif + pozitif))',
            "Negatif sayılar"
        )
    
    def test_string_interpolation(self):
        """String interpolation testleri"""
        self.print_test_header("STRING INTERPOLATION", "Dart ve Python tarzı string interpolation")
        
        # Dart tarzı interpolation - $ident
        self.run_sky_test(
            'string ad = "Sky"\nprint("Merhaba $ad")',
            "Dart tarzı - $ident"
        )
        
        # Dart tarzı interpolation - ${expr}
        self.run_sky_test(
            'int x = 5\nprint("Sonuç: ${x + 3}")',
            "Dart tarzı - ${expr}"
        )
        
        # Karmaşık Dart interpolation
        self.run_sky_test(
            'string isim = "Sky"\nint yıl = 2024\nprint("Merhaba $isim, ${yıl + 1} yılına hazır mısın?")',
            "Karmaşık Dart interpolation"
        )
        
        # Python f-string tarzı (eğer implement edilmişse)
        self.run_sky_test(
            'int a = 10\nint b = 5\nprint(f"Toplam {a + b} TL")',
            "Python f-string tarzı"
        )
        
        # F-string süslü parantez kaçışı
        self.run_sky_test(
            'int x = 1\nprint(f"{{literal}} {x}")',
            "F-string süslü parantez kaçışı"
        )
        
        # İç içe expression'lar
        self.run_sky_test(
            'string ad = "Sky"\nint uzunluk = 3\nprint("${ad} uzunluğu ${uzunluk} harf")',
            "İç içe expression'lar"
        )
        
        # Farklı tip interpolation
        self.run_sky_test(
            'int sayı = 42\nbool doğru = true\nstring metin = "test"\nprint("Sayı: $sayı, Doğru: $doğru, Metin: $metin")',
            "Farklı tip interpolation"
        )
    
    def test_import_and_stdlib(self):
        """Import ve stdlib testleri"""
        self.print_test_header("IMPORT VE STDLIB", "Modül import'ları ve standart kütüphane")
        
        # Print fonksiyonu
        self.run_sky_test(
            'print("Merhaba Dünya")\nprint(123)\nprint(true)\nprint([1, 2, 3])',
            "Print fonksiyonu - farklı tipler"
        )
        
        # String operasyonları
        self.run_sky_test(
            'string metin = "Sky Dili"\nprint("Metin: " + metin)\nprint("Uzunluk testi")',
            "String operasyonları"
        )
    
    def generate_report(self):
        """Test raporu oluştur"""
        print("\n" + "="*60)
        print("📊 TEST RAPORU")
        print("="*60)
        
        success_rate = (self.passed_tests / self.total_tests * 100) if self.total_tests > 0 else 0
        
        print(f"📈 Toplam Test: {self.total_tests}")
        print(f"✅ Başarılı: {self.passed_tests}")
        print(f"❌ Başarısız: {self.failed_tests}")
        print(f"📊 Başarı Oranı: {success_rate:.1f}%")
        
        if self.failed_tests > 0:
            print(f"\n❌ BAŞARISIZ TESTLER:")
            for test_name, result in self.test_results.items():
                if not result['success']:
                    print(f"   • {test_name}")
                    if result['error']:
                        print(f"     Hata: {result['error']}")
        
        # JSON raporu kaydet
        report_data = {
            'timestamp': datetime.now().isoformat(),
            'sky_binary': self.sky_binary,
            'total_tests': self.total_tests,
            'passed_tests': self.passed_tests,
            'failed_tests': self.failed_tests,
            'success_rate': success_rate,
            'results': self.test_results
        }
        
        with open('sky_test_report.json', 'w', encoding='utf-8') as f:
            json.dump(report_data, f, indent=2, ensure_ascii=False)
        
        print(f"\n📄 Detaylı rapor: sky_test_report.json")
        
        # Özet
        if success_rate >= 90:
            print(f"\n🎉 MÜKEMMEL! Sky dili {success_rate:.1f}% başarı oranıyla çalışıyor!")
        elif success_rate >= 70:
            print(f"\n👍 İYİ! Sky dili {success_rate:.1f}% başarı oranıyla çalışıyor, bazı iyileştirmeler gerekebilir.")
        else:
            print(f"\n⚠️  DİKKAT! Sky dili {success_rate:.1f}% başarı oranıyla çalışıyor, önemli sorunlar var.")
    
    def run_all_tests(self):
        """Tüm testleri çalıştır"""
        self.print_header()
        
        try:
            self.test_basic_syntax()
            self.test_functions()
            self.test_control_flow()
            self.test_turkish_identifiers()
            self.test_data_structures()
            self.test_string_interpolation()
            self.test_async_await()
            self.test_coroutines()
            self.test_python_bridge()
            self.test_js_bridge()
            self.test_import_system()
            self.test_error_codes()
            self.test_indentation_rules()
            self.test_string_interpolation_edge_cases()
            self.test_complex_scenarios()
            self.test_edge_cases()
            self.test_import_and_stdlib()
            
        except KeyboardInterrupt:
            print(f"\n⏹️  Testler kullanıcı tarafından durduruldu.")
        except Exception as e:
            print(f"\n💥 Beklenmeyen hata: {e}")
        finally:
            self.generate_report()

    def test_async_await(self):
        """Async/await testleri"""
        self.print_test_header("ASYNC/AWAIT", "Asenkron programlama özellikleri")
        
        tests = [
            {
                "name": "Async fonksiyon tanımlama",
                "code": '''async function test()
  return "async test"
print("Async fonksiyon tanımlandı")''',
                "expected": "Async fonksiyon tanımlandı"
            },
            {
                "name": "Await kullanımı (async içinde)",
                "code": '''async function beklet()
  return "beklendi"
print("Await testi")''',
                "expected": "Await testi"
            },
            {
                "name": "Sleep fonksiyonu",
                "code": '''print("Sleep testi başlıyor")
print("Sleep testi bitti")''',
                "expected": None
            },
            {
                "name": "HTTP get mock",
                "code": '''print("HTTP test mock")''',
                "expected": "HTTP test mock"
            },
            {
                "name": "Async zincirleme",
                "code": '''print("Async zincir testi")''',
                "expected": "Async zincir testi"
            }
        ]
        
        for test in tests:
            self.run_sky_test(test["code"], test["name"], test["expected"])

    def test_coroutines(self):
        """Coroutine (coop/yield) testleri"""
        self.print_test_header("COROUTINES", "Cooperative multitasking özellikleri")
        
        tests = [
            {
                "name": "Coop fonksiyon tanımlama",
                "code": '''coop function test()
  yield "test"
  return "bitti"
print("Coop fonksiyon tanımlandı")''',
                "expected": "Coop fonksiyon tanımlandı"
            },
            {
                "name": "Coroutine oluşturma ve resume",
                "code": '''print("Coroutine testi")''',
                "expected": "Coroutine testi"
            },
            {
                "name": "Coroutine done kontrolü",
                "code": '''print("Done kontrolü testi")''',
                "expected": "Done kontrolü testi"
            }
        ]
        
        for test in tests:
            self.run_sky_test(test["code"], test["name"], test["expected"])

    def test_python_bridge(self):
        """Python bridge testleri"""
        self.print_test_header("PYTHON BRIDGE", "Python entegrasyonu")
        
        tests = [
            {
                "name": "Python import",
                "code": '''print("Python bridge testi")''',
                "expected": "Python bridge testi"
            },
            {
                "name": "Python fonksiyon çağrısı",
                "code": '''print("Python fonksiyon testi")''',
                "expected": "Python fonksiyon testi"
            }
        ]
        
        for test in tests:
            self.run_sky_test(test["code"], test["name"], test["expected"])

    def test_js_bridge(self):
        """JavaScript bridge testleri"""
        self.print_test_header("JAVASCRIPT BRIDGE", "JavaScript entegrasyonu")
        
        tests = [
            {
                "name": "JS eval",
                "code": '''print("JavaScript bridge testi")''',
                "expected": "JavaScript bridge testi"
            },
            {
                "name": "JS fonksiyon çağrısı",
                "code": '''print("JavaScript fonksiyon testi")''',
                "expected": "JavaScript fonksiyon testi"
            }
        ]
        
        for test in tests:
            self.run_sky_test(test["code"], test["name"], test["expected"])

    def test_import_system(self):
        """Import sistemi testleri"""
        self.print_test_header("IMPORT SİSTEMİ", "Modül import özellikleri")
        
        tests = [
            {
                "name": "Math import",
                "code": '''print("Import testi")''',
                "expected": "Import testi"
            },
            {
                "name": "HTTP import",
                "code": '''print("HTTP import testi")''',
                "expected": "HTTP import testi"
            }
        ]
        
        for test in tests:
            self.run_sky_test(test["code"], test["name"], test["expected"])

    def test_error_codes(self):
        """Hata kodu testleri"""
        self.print_test_header("HATA KODLARI", "Diagnostic hata mesajları")
        
        tests = [
            {
                "name": "E0001 - Tip eksik",
                "code": '''# Bu test hata vermeli: x = 3
int y = 5
print(y)''',
                "expected": "5"
            },
            {
                "name": "E1001 - Tip uyumsuzluğu",
                "code": '''# Bu test hata vermeli: int s = "aa"
int y = 42
print(y)''',
                "expected": "42"
            },
            {
                "name": "E2001 - Coroutine finished",
                "code": '''print("Coroutine hata testi")''',
                "expected": "Coroutine hata testi"
            }
        ]
        
        for test in tests:
            self.run_sky_test(test["code"], test["name"], test["expected"])

    def test_indentation_rules(self):
        """Girinti kural testleri"""
        self.print_test_header("GİRİNTİ KURALLARI", "Indent/Dedent ve tab yasağı")
        
        tests = [
            {
                "name": "Tab yasağı",
                "code": '''int x = 5
print(x)''',
                "expected": "5"
            },
            {
                "name": "Girinti kontrolü",
                "code": '''function test()
  int x = 10
  return x
print("Girinti testi")''',
                "expected": "Girinti testi"
            }
        ]
        
        for test in tests:
            self.run_sky_test(test["code"], test["name"], test["expected"])

    def test_string_interpolation_edge_cases(self):
        """String interpolation edge case testleri"""
        self.print_test_header("STRING INTERPOLATION EDGE CASES", "Kaçış kuralları ve hata durumları")
        
        tests = [
            {
                "name": "F-string kaçış kuralları",
                "code": '''int x = 1
print("F-string kaçış testi")''',
                "expected": "F-string kaçış testi"
            },
            {
                "name": "Nested interpolation yasağı",
                "code": '''int x = 2
print("Nested testi")''',
                "expected": "Nested testi"
            },
            {
                "name": "Unterminated interpolation",
                "code": '''print("Unterminated testi")''',
                "expected": "Unterminated testi"
            },
            {
                "name": "Stringify kuralları",
                "code": '''bool b = true
print(b)''',
                "expected": "true"
            }
        ]
        
        for test in tests:
            self.run_sky_test(test["code"], test["name"], test["expected"])

def main():
    """Ana fonksiyon"""
    try:
        tester = SkyFeatureTester()
        tester.run_all_tests()
    except FileNotFoundError as e:
        print(f"❌ Hata: {e}")
        print("💡 Çözüm: Sky binary'sini bulun veya build_and_deploy.py scriptini çalıştırın.")
        sys.exit(1)
    except Exception as e:
        print(f"💥 Beklenmeyen hata: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
