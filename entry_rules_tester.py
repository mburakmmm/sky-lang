#!/usr/bin/env python3
"""
Sky Entry Rules Tester
entry.cursorrules kural dosyasındaki tüm özellikleri test eder
"""

import subprocess
import sys
import os
import tempfile
from pathlib import Path

class SkyTester:
    def __init__(self):
        self.sky_binary = "./target/release/sky"
        self.test_dir = Path("entry_tests")
        self.test_dir.mkdir(exist_ok=True)
        self.passed = 0
        self.failed = 0
        self.total = 0
        
    def run_sky(self, code, args=None, should_fail=False, expected_error=None, expected_output=None):
        """Sky kodunu çalıştır ve sonucu kontrol et"""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.sky', delete=False) as f:
            f.write(code)
            temp_file = f.name
        
        try:
            cmd = [self.sky_binary, "run", temp_file]
            if args:
                cmd.extend(args)
            
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=10)
            
            if should_fail:
                if result.returncode == 0:
                    return False, f"Expected failure but got success. Output: {result.stdout}"
                if expected_error and expected_error not in result.stderr:
                    return False, f"Expected error '{expected_error}' not found in stderr: {result.stderr}"
                return True, "Failed as expected"
            else:
                if result.returncode != 0:
                    return False, f"Expected success but got failure. Error: {result.stderr}"
                if expected_output and expected_output not in result.stdout:
                    return False, f"Expected output '{expected_output}' not found in stdout: {result.stdout}"
                return True, "Success as expected"
                
        except subprocess.TimeoutExpired:
            return False, "Test timed out"
        except Exception as e:
            return False, f"Exception: {e}"
        finally:
            os.unlink(temp_file)
    
    def test(self, name, code, args=None, should_fail=False, expected_error=None, expected_output=None):
        """Tek bir testi çalıştır"""
        self.total += 1
        print(f"Testing: {name}")
        
        success, message = self.run_sky(code, args, should_fail, expected_error, expected_output)
        
        if success:
            print(f"✅ PASS: {name}")
            self.passed += 1
        else:
            print(f"❌ FAIL: {name} - {message}")
            self.failed += 1
        
        return success
    
    def run_all_tests(self):
        """Tüm testleri çalıştır"""
        print("🚀 Sky Entry Rules Tester")
        print("=" * 50)
        
        # 1. Main Entry Point Tests (E6001, E6002, E6003)
        print("\n📋 1. Main Entry Point Tests")
        
        # Main fonksiyonu süslü parantez ile zorunlu
        self.test(
            "Main - süslü parantez zorunlu (başarılı)",
            """function main()
{
  print("Main çalıştı")
}"""
        )
        
        # Main fonksiyonu girintili blok - hata
        self.test(
            "Main - girintili blok (hata)",
            """function main()
  print("Main çalıştı")""",
            should_fail=True,
            expected_error="main function body must be brace-delimited"
        )
        
        # Diğer fonksiyonlar girintili blok kullanmalı
        self.test(
            "Non-main - girintili blok (başarılı)",
            """function test()
  print("Test çalıştı")

function main()
{
  test()
}"""
        )
        
        # Diğer fonksiyonlar süslü parantez - hata
        self.test(
            "Non-main - süslü parantez (hata)",
            """function test()
{
  print("Test çalıştı")
}""",
            should_fail=True,
            expected_error="only main function may use brace-delimited body"
        )
        
        # Main parametresi: args: list[string] (başarılı)
        self.test(
            "Main - args parametresi (başarılı)",
            """function main(args: list[string])
{
  print("Args sayısı: " + len(args))
}""",
            args=["arg1", "arg2"]
        )
        
        # Main parametresi: geçersiz tip (hata)
        self.test(
            "Main - geçersiz parametre tipi (hata)",
            """function main(x: int)
{
  print("Test")
}""",
            should_fail=True,
            expected_error="main function parameter must be 'args: list[string]'"
        )
        
        # Main parametresi: çoklu parametre (hata)
        self.test(
            "Main - çoklu parametre (hata)",
            """function main(args: list[string], extra: int)
{
  print("Test")
}""",
            should_fail=True,
            expected_error="main function must have no parameters or exactly one"
        )
        
        # 2. List Tip Parametreleri Tests
        print("\n📋 2. List Tip Parametreleri Tests")
        
        # list[int] (başarılı)
        self.test(
            "List[int] - başarılı",
            """list[int] sayılar = [1, 2, 3, 4, 5]

function main()
{
  print("Liste uzunluğu: " + len(sayılar))
}"""
        )
        
        # list[string] (başarılı)
        self.test(
            "List[string] - başarılı",
            """list[string] isimler = ["Ali", "Veli", "Ayşe"]

function main()
{
  print("İsim sayısı: " + len(isimler))
}"""
        )
        
        # list[float] (başarılı)
        self.test(
            "List[float] - başarılı",
            """list[float] sayılar = [1.5, 2.7, 3.14]

function main()
{
  print("Float sayı sayısı: " + len(sayılar))
}"""
        )
        
        # list[bool] (başarılı)
        self.test(
            "List[bool] - başarılı",
            """list[bool] değerler = [true, false, true]

function main()
{
  print("Boolean sayısı: " + len(değerler))
}"""
        )
        
        # list[T] tip uyumsuzluğu (hata)
        self.test(
            "List[int] - tip uyumsuzluğu (hata)",
            """list[int] sayılar = ["string", 2, 3]

function main()
{
  print("Test")
}""",
            should_fail=True,
            expected_error="Type mismatch"
        )
        
        # 3. Visibility System Tests (E0404)
        print("\n📋 3. Visibility System Tests")
        
        # Private değişken (başarılı - aynı modülde)
        self.test(
            "Private değişken - aynı modül (başarılı)",
            """int _private_var = 42

function main()
{
  print(_private_var)
}"""
        )
        
        # Public değişken (başarılı)
        self.test(
            "Public değişken (başarılı)",
            """int public_var = 42

function main()
{
  print(public_var)
}"""
        )
        
        # 4. CLI Arguments Tests
        print("\n📋 4. CLI Arguments Tests")
        
        # __args__ değişkeni
        self.test(
            "CLI Arguments - __args__",
            """function main()
{
  print("Argüman sayısı: " + len(__args__))
}""",
            args=["arg1", "arg2", "arg3"]
        )
        
        # main(args: list[string]) ile argümanlar
        self.test(
            "Main args parametresi",
            """function main(args: list[string])
{
  print("İlk argüman: " + args[0])
}""",
            args=["test_arg"]
        )
        
        # 5. Special Variables Tests
        print("\n📋 5. Special Variables Tests")
        
        # __name__ değişkeni
        self.test(
            "Special Variables - __name__",
            """function main()
{
  print(__name__)
}"""
        )
        
        # __file__ değişkeni
        self.test(
            "Special Variables - __file__",
            """function main()
{
  print("Dosya yolu içeriyor: " + (__file__ != ""))
}"""
        )
        
        # __dir__ değişkeni
        self.test(
            "Special Variables - __dir__",
            """function main()
{
  print("Dizin yolu içeriyor: " + (__dir__ != ""))
}"""
        )
        
        # Tüm özel değişkenler
        self.test(
            "Special Variables - hepsi",
            """function main()
{
  print("__name__: " + __name__)
  print("__file__: " + __file__)
  print("__dir__: " + __dir__)
  print("__args__ uzunluğu: " + len(__args__))
}""",
            args=["test"]
        )
        
        # 6. Import System Tests
        print("\n📋 6. Import System Tests")
        
        # Import statement (başarılı)
        self.test(
            "Import - math modülü",
            """import math

function main()
{
  print("Math modülü import edildi")
}"""
        )
        
        # Import statement (başarılı)
        self.test(
            "Import - os modülü",
            """import os

function main()
{
  print("OS modülü import edildi")
}"""
        )
        
        # Import statement (başarılı)
        self.test(
            "Import - sys modülü",
            """import sys

function main()
{
  print("SYS modülü import edildi")
}"""
        )
        
        # 7. Complex Integration Tests
        print("\n📋 7. Complex Integration Tests")
        
        # Kapsamlı örnek
        self.test(
            "Complex - kapsamlı örnek",
            """import math

int _private_counter = 0
list[string] _private_list = ["a", "b", "c"]

function _private_function(x: int)
  _private_counter = _private_counter + x
  return _private_counter

function public_function()
  return _private_function(5)

function main(args: list[string])
{
  print("Modül adı: " + __name__)
  print("Dosya: " + __file__)
  print("Argüman sayısı: " + len(args))
  
  int result = public_function()
  print("Sonuç: " + result)
  
  print("Private liste uzunluğu: " + len(_private_list))
}""",
            args=["test_arg1", "test_arg2"]
        )
        
        # 8. Error Code Tests
        print("\n📋 8. Error Code Tests")
        
        # E6001 - Invalid main signature
        self.test(
            "Error E6001 - Invalid main signature",
            """function main(x: int)
{
  print("Test")
}""",
            should_fail=True,
            expected_error="E6001"
        )
        
        # E6002 - Brace body non-main
        self.test(
            "Error E6002 - Brace body non-main",
            """function test()
{
  print("Test")
}""",
            should_fail=True,
            expected_error="E6002"
        )
        
        # E6003 - Main must use braces
        self.test(
            "Error E6003 - Main must use braces",
            """function main()
  print("Test")""",
            should_fail=True,
            expected_error="E6003"
        )
        
        # 9. Grammar Compliance Tests
        print("\n📋 9. Grammar Compliance Tests")
        
        # Sky gramerine uygun kod
        self.test(
            "Grammar - Sky syntax compliance",
            """import math

list[int] sayılar = [1, 2, 3, 4, 5]
map sözlük = {"anahtar": "değer", "sayı": 42}

function hesapla(x: int, y: int)
  int sonuç = x * y + 10
  return sonuç

function main(args: list[string])
{
  int x = hesapla(5, 3)
  print("Hesaplama sonucu: " + x)
  
  print("Liste uzunluğu: " + len(sayılar))
  print("Sözlük anahtarı: " + sözlük["anahtar"])
  
  if x > 20
    print("Sonuç büyük")
  else
    print("Sonuç küçük")
}""",
            args=["grammar_test"]
        )
        
        # 10. Edge Cases Tests
        print("\n📋 10. Edge Cases Tests")
        
        # Boş main fonksiyonu
        self.test(
            "Edge Case - boş main",
            """function main()
{
}"""
        )
        
        # Parametresiz main
        self.test(
            "Edge Case - parametresiz main",
            """function main()
{
  print("Parametresiz main")
}"""
        )
        
        # Boş liste parametreleri
        self.test(
            "Edge Case - boş list parametreleri",
            """list[int] boş_liste = []
list[string] boş_string_liste = []

function main()
{
  print("Boş liste uzunlukları: " + len(boş_liste) + ", " + len(boş_string_liste))
}"""
        )
        
        print("\n" + "=" * 50)
        print(f"📊 Test Sonuçları:")
        print(f"✅ Başarılı: {self.passed}")
        print(f"❌ Başarısız: {self.failed}")
        print(f"📈 Toplam: {self.total}")
        print(f"🎯 Başarı Oranı: {(self.passed/self.total*100):.1f}%")
        
        if self.failed == 0:
            print("\n🎉 Tüm testler başarılı!")
            return True
        else:
            print(f"\n⚠️  {self.failed} test başarısız!")
            return False

def main():
    tester = SkyTester()
    success = tester.run_all_tests()
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()