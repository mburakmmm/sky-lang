#!/usr/bin/env python3
"""
Sky Dili Sıkı Özellik Test Scripti
Daha katı ve detaylı testler ile Sky dilinin tüm özelliklerini test eder
"""

import os
import sys
import subprocess
import tempfile
from pathlib import Path
from datetime import datetime
import json
import re
import time

class SkyStrictTester:
    def __init__(self):
        self.sky_binary = self.find_sky_binary()
        self.test_results = {}
        self.total_tests = 0
        self.passed_tests = 0
        self.failed_tests = 0
        self.warning_tests = 0
        self.performance_tests = {}
        
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
            
        # Mevcut dizinde target/release/sky ara
        release_path = Path("target/release/sky")
        if release_path.exists():
            abs_path = release_path.absolute()
            print(f"✅ Release dizininde bulundu: {abs_path}")
            return str(abs_path)
            
        print("❌ Sky binary bulunamadı!")
        sys.exit(1)
    
    def run_sky_code(self, code, timeout=10):
        """Sky kodunu çalıştır ve sonucu döndür"""
        try:
            with tempfile.NamedTemporaryFile(mode='w', suffix='.sky', delete=False) as f:
                f.write(code)
                temp_file = f.name
            
            start_time = time.time()
            result = subprocess.run(
                [self.sky_binary, 'run', temp_file],
                capture_output=True,
                text=True,
                timeout=timeout
            )
            end_time = time.time()
            
            os.unlink(temp_file)
            
            return {
                'returncode': result.returncode,
                'stdout': result.stdout,
                'stderr': result.stderr,
                'execution_time': end_time - start_time
            }
        except subprocess.TimeoutExpired:
            return {
                'returncode': -1,
                'stdout': '',
                'stderr': f'Timeout after {timeout} seconds',
                'execution_time': timeout
            }
        except Exception as e:
            return {
                'returncode': -1,
                'stdout': '',
                'stderr': str(e),
                'execution_time': 0
            }
    
    def test_syntax_strictness(self):
        """Sözdizimi sıkılığı testleri"""
        print("\n🔍 SÖZDİZİMİ SIKILIĞI ==================================================")
        print("📝 Tip zorunluluğu, girinti kuralları, sözdizimi hataları")
        print("-" * 60)
        
        tests = [
            {
                'name': 'Tip zorunluluğu - eksik tip',
                'code': 'x = 42',
                'should_fail': True,
                'expected_error': 'missing type annotation'
            },
            {
                'name': 'Tip zorunluluğu - geçersiz tip',
                'code': 'int x = "string"',
                'should_fail': True,
                'expected_error': 'type mismatch'
            },
            {
                'name': 'Tip zorunluluğu - doğru kullanım',
                'code': 'int x = 42\nprint(x)',
                'should_fail': False
            },
            {
                'name': 'Tab yasağı - tab karakteri',
                'code': '\tint x = 42',
                'should_fail': True,
                'expected_error': 'tab'
            },
            {
                'name': 'Girinti tutarsızlığı',
                'code': 'if true:\n  print("test")\n    print("wrong indent")',
                'should_fail': True,
                'expected_error': 'indentation'
            },
            {
                'name': 'Geçersiz karakter - özel karakter',
                'code': 'int x@ = 42',
                'should_fail': True,
                'expected_error': 'invalid character'
            },
            {
                'name': 'Eksik iki nokta - if statement',
                'code': 'if true\n    print("test")',
                'should_fail': False,
                'expected_output': 'test'
            },
            {
                'name': 'Eksik iki nokta - function parameter',
                'code': 'function test(x: int)\n    return x\nprint(test(42))',
                'should_fail': False,
                'expected_output': '42'
            },
            {
                'name': 'Geçersiz keyword kullanımı',
                'code': 'return 42',
                'should_fail': True,
                'expected_error': 'return outside function'
            },
            {
                'name': 'Eksik parantez - function call',
                'code': 'print "test"',
                'should_fail': True,
                'expected_error': 'expected \'(\''
            }
        ]
        
        for test in tests:
            self.run_test(test)
    
    def test_type_system_strictness(self):
        """Tip sistemi sıkılığı testleri"""
        print("\n🔍 TİP SİSTEMİ SIKILIĞI ==================================================")
        print("📝 Runtime tip kontrolü, tip dönüşümleri, tip güvenliği")
        print("-" * 60)
        
        tests = [
            {
                'name': 'Int tip kontrolü - string atama',
                'code': 'int x = 42\nx = "string"',
                'should_fail': True,
                'expected_error': 'type mismatch'
            },
            {
                'name': 'Float tip kontrolü - int atama',
                'code': 'float x = 3.14\nx = 42',
                'should_fail': False  # int -> float dönüşümü olmalı
            },
            {
                'name': 'Bool tip kontrolü - string atama',
                'code': 'bool x = true\nx = "false"',
                'should_fail': True,
                'expected_error': 'type mismatch'
            },
            {
                'name': 'String tip kontrolü - int atama',
                'code': 'string x = "hello"\nx = 42',
                'should_fail': True,
                'expected_error': 'type mismatch'
            },
            {
                'name': 'Var tip - herhangi bir değer',
                'code': 'var x = 42\nx = "string"\nx = true\nprint(x)',
                'should_fail': False
            },
            {
                'name': 'List tip kontrolü - farklı tip eleman',
                'code': 'list x = [1, 2, 3]\nx[0] = "string"',
                'should_fail': True,
                'expected_error': 'type mismatch'
            },
            {
                'name': 'Map tip kontrolü - farklı tip değer',
                'code': 'map x = {"key": "value"}\nx["key"] = 42',
                'should_fail': True,
                'expected_error': 'type mismatch'
            },
            {
                'name': 'Function return tip kontrolü',
                'code': 'function test()\n    return "string"\nprint(test())',
                'should_fail': False,
                'expected_output': 'string'
            },
            {
                'name': 'Aritmetik operasyon tip kontrolü',
                'code': 'int x = 42\nstring y = "hello"\nint z = x + y',
                'should_fail': True,
                'expected_error': 'invalid operation'
            },
            {
                'name': 'Karşılaştırma operasyon tip kontrolü',
                'code': 'int x = 42\nstring y = "hello"\nbool z = x == y',
                'should_fail': True,
                'expected_error': 'invalid operation'
            }
        ]
        
        for test in tests:
            self.run_test(test)
    
    def test_control_flow_strictness(self):
        """Kontrol akışı sıkılığı testleri"""
        print("\n🔍 KONTROL AKIŞI SIKILIĞI ==================================================")
        print("📝 If-elif-else, döngüler, break/continue kuralları")
        print("-" * 60)
        
        tests = [
            {
                'name': 'If statement - doğru sözdizimi',
                'code': 'if true\n    print("test")',
                'should_fail': False,
                'expected_output': 'test'
            },
            {
                'name': 'Elif statement - doğru sözdizimi',
                'code': 'if false\n    print("test")\nelif true\n    print("elif")',
                'should_fail': False,
                'expected_output': 'elif'
            },
            {
                'name': 'Else statement - doğru sözdizimi',
                'code': 'if false\n    print("test")\nelse\n    print("else")',
                'should_fail': False,
                'expected_output': 'else'
            },
            {
                'name': 'While döngüsü - doğru sözdizimi',
                'code': 'int i = 0\nwhile i < 1\n    print("test")\n    i = i + 1',
                'should_fail': False,
                'expected_output': 'test'
            },
            {
                'name': 'For döngüsü - doğru sözdizimi',
                'code': 'list liste = [1,2,3]\nfor x: var in liste\n    print(x)',
                'should_fail': False,
                'expected_output': '1\n2\n3'
            },
            {
                'name': 'Break - döngü dışında',
                'code': 'break',
                'should_fail': True,
                'expected_error': 'break outside loop'
            },
            {
                'name': 'Continue - döngü dışında',
                'code': 'continue',
                'should_fail': True,
                'expected_error': 'continue outside loop'
            },
            {
                'name': 'Return - fonksiyon dışında',
                'code': 'return 42',
                'should_fail': True,
                'expected_error': 'return outside function'
            },
            {
                'name': 'Await - async dışında',
                'code': 'await sleep(1000)',
                'should_fail': True,
                'expected_error': 'await outside async function'
            },
            {
                'name': 'Yield - coop dışında',
                'code': 'yield 42',
                'should_fail': True,
                'expected_error': 'yield outside coop function'
            }
        ]
        
        for test in tests:
            self.run_test(test)
    
    def test_function_strictness(self):
        """Fonksiyon sıkılığı testleri"""
        print("\n🔍 FONKSİYON SIKILIĞI ==================================================")
        print("📝 Fonksiyon tanımlama, parametreler, return değerleri")
        print("-" * 60)
        
        tests = [
            {
                'name': 'Fonksiyon - eksik parametre tipi',
                'code': 'function test(x)\n    return x',
                'should_fail': True,
                'expected_error': 'missing type annotation'
            },
            {
                'name': 'Fonksiyon - doğru sözdizimi',
                'code': 'function test(x: int)\n    return x\nprint(test(42))',
                'should_fail': False,
                'expected_output': '42'
            },
            {
                'name': 'Async fonksiyon - doğru sözdizimi',
                'code': 'async function test(x: int)\n    return x\nvar future = test(42)\nprint(await future)',
                'should_fail': False,
                'expected_output': '42'
            },
            {
                'name': 'Coop fonksiyon - doğru sözdizimi',
                'code': 'coop function test(x: int)\n    yield x\n    return x + 1\nvar c = test(42)\nprint(c.resume())',
                'should_fail': False,
                'expected_output': '42'
            },
            {
                'name': 'Fonksiyon - geçersiz parametre tipi',
                'code': 'function test(x: invalid_type)\n    return x',
                'should_fail': True,
                'expected_error': 'unknown type'
            },
            {
                'name': 'Fonksiyon - eksik return değeri',
                'code': 'function test()\n    print("test")\nprint(test())',
                'should_fail': False,
                'expected_output': 'test\nnull'
            },
            {
                'name': 'Fonksiyon - yanlış return tipi',
                'code': 'function test()\n    return "string"\nprint(test())',
                'should_fail': False,
                'expected_error': 'type mismatch'
            },
            {
                'name': 'Fonksiyon - çoklu return tipi',
                'code': 'function test()\n    if true\n        return 42\n    else\n        return "string"',
                'should_fail': False,
                'expected_output': '42'
            },
            {
                'name': 'Fonksiyon - eksik parametre',
                'code': 'function test(x: int, y: int)\n    return x + y\n\ntest(42)',
                'should_fail': True,
                'expected_error': 'missing argument'
            },
            {
                'name': 'Fonksiyon - fazla parametre',
                'code': 'function test(x: int)\n    return x\n\ntest(42, 43)',
                'should_fail': True,
                'expected_error': 'too many arguments'
            }
        ]
        
        for test in tests:
            self.run_test(test)
    
    def test_performance_strictness(self):
        """Performans sıkılığı testleri"""
        print("\n🔍 PERFORMANS SIKILIĞI ==================================================")
        print("📝 Execution time, memory usage, optimization")
        print("-" * 60)
        
        tests = [
            {
                'name': 'Basit aritmetik - 1000 iterasyon',
                'code': 'int sum = 0\nint i = 1\nwhile i <= 1000\n    sum = sum + i\n    i = i + 1\nprint(sum)',
                'max_time': 1.0,
                'should_fail': False
            },
            {
                'name': 'String concatenation - 100 iterasyon',
                'code': 'string result = ""\nint i = 1\nwhile i <= 100\n    result = result + "test"\n    i = i + 1\nprint(len(result))',
                'max_time': 2.0,
                'should_fail': False
            },
            {
                'name': 'Recursive factorial - 20',
                'code': 'function factorial(n: int)\n    if n <= 1\n        return 1\n    else\n        return n * factorial(n - 1)\n\nprint(factorial(20))',
                'max_time': 3.0,
                'should_fail': False
            },
            {
                'name': 'List operations - 1000 eleman',
                'code': 'list numbers = []\nint i = 1\nwhile i <= 1000\n    numbers = numbers + [i]\n    i = i + 1\nprint(len(numbers))',
                'max_time': 2.0,
                'should_fail': False
            },
            {
                'name': 'Map operations - 1000 eleman',
                'code': 'map data = {}\nint i = 1\nwhile i <= 1000\n    data["key" + i] = i\n    i = i + 1\nprint(len(data))',
                'max_time': 2.0,
                'should_fail': False
            }
        ]
        
        for test in tests:
            self.run_performance_test(test)
    
    def test_memory_strictness(self):
        """Bellek sıkılığı testleri"""
        print("\n🔍 BELLEK SIKILIĞI ==================================================")
        print("📝 Memory leaks, garbage collection, memory usage")
        print("-" * 60)
        
        tests = [
            {
                'name': 'Büyük liste oluşturma',
                'code': 'list big_list = []\nint i = 1\nwhile i <= 10000\n    big_list = big_list + [i]\n    i = i + 1\nprint("List created")\nbig_list = null',
                'should_fail': False
            },
            {
                'name': 'Büyük map oluşturma',
                'code': 'map big_map = {}\nint i = 1\nwhile i <= 10000\n    big_map["key" + i] = "value" + i\n    i = i + 1\nprint("Map created")\nbig_map = null',
                'should_fail': False
            },
            {
                'name': 'String concatenation - büyük string',
                'code': 'string big_string = ""\nint i = 1\nwhile i <= 1000\n    big_string = big_string + "very long string that takes memory"\n    i = i + 1\nprint("String created")\nbig_string = null',
                'should_fail': False
            },
            {
                'name': 'Nested structures',
                'code': 'list outer = []\nint i = 1\nwhile i <= 100\n    list inner = []\n    int j = 1\n    while j <= 100\n        inner = inner + [i * j]\n        j = j + 1\n    outer = outer + [inner]\n    i = i + 1\nprint("Nested structures created")\nouter = null',
                'should_fail': False
            }
        ]
        
        for test in tests:
            self.run_test(test)
    
    def test_error_handling_strictness(self):
        """Hata yönetimi sıkılığı testleri"""
        print("\n🔍 HATA YÖNETİMİ SIKILIĞI ==================================================")
        print("📝 Error messages, exception handling, edge cases")
        print("-" * 60)
        
        tests = [
            {
                'name': 'Division by zero',
                'code': 'int x = 42\nint y = 0\nint z = x / y',
                'should_fail': True,
                'expected_error': 'division by zero'
            },
            {
                'name': 'Modulo by zero',
                'code': 'int x = 42\nint y = 0\nint z = x % y',
                'should_fail': True,
                'expected_error': 'modulo by zero'
            },
            {
                'name': 'Array index out of bounds',
                'code': 'list arr = [1, 2, 3]\nint x = arr[10]',
                'should_fail': True,
                'expected_error': 'index out of bounds'
            },
            {
                'name': 'Map key not found',
                'code': 'map data = {"key": "value"}\nstring x = data["nonexistent"]',
                'should_fail': True,
                'expected_error': 'key not found'
            },
            {
                'name': 'Stack overflow - deep recursion',
                'code': 'function recurse(n: int)\n    if n <= 0\n        return 0\n    else\n        return recurse(n - 1)\n\nprint(recurse(10000))',
                'should_fail': True,
                'expected_error': 'stack overflow'
            },
            {
                'name': 'Infinite loop detection',
                'code': 'while true\n    print("infinite")',
                'should_fail': True,
                'expected_error': 'infinite loop'
            }
        ]
        
        for test in tests:
            self.run_test(test)
    
    def test_unicode_strictness(self):
        """Unicode sıkılığı testleri"""
        print("\n🔍 UNICODE SIKILIĞI ==================================================")
        print("📝 Türkçe karakterler, Unicode identifiers, string handling")
        print("-" * 60)
        
        tests = [
            {
                'name': 'Türkçe karakterli değişken',
                'code': 'int çorba = 42\nprint(çorba)',
                'should_fail': False
            },
            {
                'name': 'Türkçe karakterli fonksiyon',
                'code': 'function sayı_analizi(sayi: int)\n    return "test"\n\nprint(sayı_analizi(5))',
                'should_fail': False
            },
            {
                'name': 'Unicode string literal',
                'code': 'string emoji = "🚀 Sky Dili 🎉"\nprint(emoji)',
                'should_fail': False
            },
            {
                'name': 'Türkçe karakterli string interpolation',
                'code': 'string isim = "Melih"\nstring mesaj = "Merhaba $isim"\nprint(mesaj)',
                'should_fail': False
            },
            {
                'name': 'Unicode identifier - geçersiz karakter',
                'code': 'int x@ = 42',
                'should_fail': True,
                'expected_error': 'invalid character'
            },
            {
                'name': 'Unicode identifier - sayı ile başlama',
                'code': 'int 2x = 42',
                'should_fail': True,
                'expected_error': 'identifier cannot start with digit'
            }
        ]
        
        for test in tests:
            self.run_test(test)
    
    def run_test(self, test):
        """Tek bir testi çalıştır"""
        self.total_tests += 1
        
        result = self.run_sky_code(test['code'])
        
        if test.get('should_fail', False):
            # Test başarısız olmalı
            if result['returncode'] != 0:
                if 'expected_error' in test:
                    if test['expected_error'].lower() in result['stderr'].lower():
                        self.passed_tests += 1
                        print(f"✅ BAŞARILI | {test['name']}")
                        print(f"   📤 Hata: {result['stderr'].strip()}")
                    else:
                        self.failed_tests += 1
                        print(f"❌ BAŞARISIZ | {test['name']}")
                        print(f"   📤 Beklenen: {test['expected_error']}")
                        print(f"   📤 Gerçek: {result['stderr'].strip()}")
                else:
                    self.passed_tests += 1
                    print(f"✅ BAŞARILI | {test['name']}")
                    print(f"   📤 Hata: {result['stderr'].strip()}")
            else:
                self.failed_tests += 1
                print(f"❌ BAŞARISIZ | {test['name']}")
                print(f"   📤 Beklenen: Hata")
                print(f"   📤 Gerçek: Başarılı")
        else:
            # Test başarılı olmalı
            if result['returncode'] == 0:
                self.passed_tests += 1
                print(f"✅ BAŞARILI | {test['name']}")
                if result['stdout'].strip():
                    print(f"   📤 Çıktı: {result['stdout'].strip()}")
            else:
                self.failed_tests += 1
                print(f"❌ BAŞARISIZ | {test['name']}")
                print(f"   📤 Hata: {result['stderr'].strip()}")
        
        # Test sonucunu kaydet
        self.test_results[test['name']] = {
            'passed': result['returncode'] == 0 if not test.get('should_fail', False) else result['returncode'] != 0,
            'returncode': result['returncode'],
            'stdout': result['stdout'],
            'stderr': result['stderr'],
            'execution_time': result['execution_time'],
            'category': self.get_test_category(test['name'])
        }
    
    def run_performance_test(self, test):
        """Performans testini çalıştır"""
        self.total_tests += 1
        
        result = self.run_sky_code(test['code'])
        
        if result['returncode'] == 0:
            execution_time = result['execution_time']
            max_time = test.get('max_time', 5.0)
            
            if execution_time <= max_time:
                self.passed_tests += 1
                print(f"✅ BAŞARILI | {test['name']}")
                print(f"   📤 Süre: {execution_time:.3f}s (max: {max_time}s)")
            else:
                self.warning_tests += 1
                print(f"⚠️  UYARI | {test['name']}")
                print(f"   📤 Süre: {execution_time:.3f}s (max: {max_time}s)")
            
            self.performance_tests[test['name']] = execution_time
        else:
            self.failed_tests += 1
            print(f"❌ BAŞARISIZ | {test['name']}")
            print(f"   📤 Hata: {result['stderr'].strip()}")
        
        # Test sonucunu kaydet
        self.test_results[test['name']] = {
            'passed': result['returncode'] == 0,
            'returncode': result['returncode'],
            'stdout': result['stdout'],
            'stderr': result['stderr'],
            'execution_time': result['execution_time'],
            'category': self.get_test_category(test['name'])
        }
    
    def run_all_tests(self):
        """Tüm testleri çalıştır"""
        print("🌌============================================================")
        print("🌌 SKY DİLİ SIKI ÖZELLİK TEST SCRIPTI")
        print("🌌============================================================")
        print(f"📅 Test Tarihi: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"🔧 Sky Binary: {self.sky_binary}")
        print()
        
        # Tüm test kategorilerini çalıştır
        self.test_syntax_strictness()
        self.test_type_system_strictness()
        self.test_control_flow_strictness()
        self.test_function_strictness()
        self.test_performance_strictness()
        self.test_memory_strictness()
        self.test_error_handling_strictness()
        self.test_unicode_strictness()
        
        # Sonuçları göster
        self.show_results()
    
    def show_results(self):
        """Test sonuçlarını göster"""
        print("\n" + "=" * 60)
        print("📊 SIKI TEST RAPORU")
        print("=" * 60)
        print(f"📈 Toplam Test: {self.total_tests}")
        print(f"✅ Başarılı: {self.passed_tests}")
        print(f"⚠️  Uyarı: {self.warning_tests}")
        print(f"❌ Başarısız: {self.failed_tests}")
        
        success_rate = (self.passed_tests / self.total_tests) * 100 if self.total_tests > 0 else 0
        print(f"📊 Başarı Oranı: {success_rate:.1f}%")
        
        if self.failed_tests > 0:
            print(f"\n❌ BAŞARISIZ TESTLER:")
            for name, result in self.test_results.items():
                if not result['passed']:
                    print(f"   • {name}")
                    print(f"     Hata: {result['stderr'].strip()}")
        
        if self.warning_tests > 0:
            print(f"\n⚠️  UYARI TESTLER:")
            for name, result in self.test_results.items():
                if result['passed'] and result['execution_time'] > 1.0:
                    print(f"   • {name}")
                    print(f"     Süre: {result['execution_time']:.3f}s")
        
        # Performans istatistikleri
        if self.performance_tests:
            print(f"\n📈 PERFORMANS İSTATİSTİKLERİ:")
            avg_time = sum(self.performance_tests.values()) / len(self.performance_tests)
            max_time = max(self.performance_tests.values())
            min_time = min(self.performance_tests.values())
            print(f"   • Ortalama süre: {avg_time:.3f}s")
            print(f"   • En yavaş test: {max_time:.3f}s")
            print(f"   • En hızlı test: {min_time:.3f}s")
        
        # JSON raporu oluştur
        report = {
            'timestamp': datetime.now().isoformat(),
            'sky_binary': self.sky_binary,
            'binary_exists': os.path.exists(self.sky_binary),
            'binary_executable': os.access(self.sky_binary, os.X_OK),
            'total_tests': self.total_tests,
            'passed_tests': self.passed_tests,
            'warning_tests': self.warning_tests,
            'failed_tests': self.failed_tests,
            'success_rate': success_rate,
            'performance_tests': self.performance_tests,
            'test_categories': {
                'syntax_strictness': self.get_category_results('Sözdizimi'),
                'type_system_strictness': self.get_category_results('Tip Sistemi'),
                'control_flow_strictness': self.get_category_results('Kontrol Akışı'),
                'function_strictness': self.get_category_results('Fonksiyon'),
                'performance_strictness': self.get_category_results('Performans'),
                'memory_strictness': self.get_category_results('Bellek'),
                'error_handling_strictness': self.get_category_results('Hata Yönetimi'),
                'unicode_strictness': self.get_category_results('Unicode')
            },
            'test_results': self.test_results,
            'summary': {
                'excellent': success_rate >= 95,
                'good': success_rate >= 90,
                'average': success_rate >= 80,
                'poor': success_rate < 80
            }
        }
        
        report_path = 'sky_strict_test_report.json'
        with open(report_path, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)
        
        print(f"\n📄 Detaylı rapor: {os.path.abspath(report_path)}")
        print(f"📁 Rapor boyutu: {os.path.getsize(report_path)} bytes")
        
        if success_rate >= 95:
            print("\n🎉 MÜKEMMEL! Sky dili sıkı testlerde %95+ başarı gösteriyor!")
        elif success_rate >= 90:
            print("\n👍 İYİ! Sky dili sıkı testlerde %90+ başarı gösteriyor!")
        elif success_rate >= 80:
            print("\n⚠️  ORTA! Sky dili sıkı testlerde %80+ başarı gösteriyor!")
        else:
            print("\n❌ DÜŞÜK! Sky dili sıkı testlerde %80 altında başarı gösteriyor!")
    
    def get_test_category(self, test_name):
        """Test adından kategoriyi belirle"""
        if 'Tip zorunluluğu' in test_name or 'Tab yasağı' in test_name or 'Girinti' in test_name or 'Geçersiz karakter' in test_name or 'Eksik iki nokta' in test_name or 'Geçersiz keyword' in test_name or 'Eksik parantez' in test_name:
            return 'Sözdizimi'
        elif 'tip kontrolü' in test_name or 'tip dönüşümü' in test_name or 'Var tip' in test_name or 'List tip' in test_name or 'Map tip' in test_name or 'Function return tip' in test_name or 'Aritmetik operasyon tip' in test_name or 'Karşılaştırma operasyon tip' in test_name:
            return 'Tip Sistemi'
        elif 'If statement' in test_name or 'Elif statement' in test_name or 'Else statement' in test_name or 'While döngüsü' in test_name or 'For döngüsü' in test_name or 'Break' in test_name or 'Continue' in test_name or 'Return' in test_name or 'Await' in test_name or 'Yield' in test_name:
            return 'Kontrol Akışı'
        elif 'Fonksiyon' in test_name or 'Async fonksiyon' in test_name or 'Coop fonksiyon' in test_name:
            return 'Fonksiyon'
        elif 'aritmetik' in test_name or 'String concatenation' in test_name or 'Recursive factorial' in test_name or 'List operations' in test_name or 'Map operations' in test_name:
            return 'Performans'
        elif 'Büyük liste' in test_name or 'Büyük map' in test_name or 'Nested structures' in test_name:
            return 'Bellek'
        elif 'Division by zero' in test_name or 'Modulo by zero' in test_name or 'Array index out of bounds' in test_name or 'Map key not found' in test_name or 'Stack overflow' in test_name or 'Infinite loop' in test_name:
            return 'Hata Yönetimi'
        elif 'Türkçe karakter' in test_name or 'Unicode' in test_name or 'emoji' in test_name:
            return 'Unicode'
        else:
            return 'Diğer'
    
    def get_category_results(self, category_name):
        """Kategoriye göre test sonuçlarını döndür"""
        category_tests = {name: result for name, result in self.test_results.items() 
                         if result.get('category') == category_name}
        
        if not category_tests:
            return {'total': 0, 'passed': 0, 'failed': 0, 'success_rate': 0}
        
        total = len(category_tests)
        passed = sum(1 for result in category_tests.values() if result['passed'])
        failed = total - passed
        success_rate = (passed / total) * 100 if total > 0 else 0
        
        return {
            'total': total,
            'passed': passed,
            'failed': failed,
            'success_rate': success_rate,
            'tests': category_tests
        }

if __name__ == "__main__":
    tester = SkyStrictTester()
    tester.run_all_tests()
