# SKY Programlama Dili

SKY, modern ve güvenli bir programlama dilidir. Go ile yazılmış, LLVM JIT tabanlı çalışan, girintileme tabanlı blok yapısına sahip bir dildir.

## Özellikler

- 🚀 **JIT Derleyici**: LLVM tabanlı anlık derleme ve çalıştırma
- 🎯 **Tip Güvenliği**: İsteğe bağlı statik tip sistemi ve tip çıkarımı
- ⚡ **Async/Await**: Modern eşzamansız programlama desteği
- 🔒 **Güvenlik**: `unsafe` blokları ile kontrollü düşük seviye erişim
- 🔄 **Coroutines**: `coop`/`yield` ile hafif eşzamanlılık
- 🌐 **FFI**: C kütüphaneleri ile kolay entegrasyon
- 📦 **Wing**: Güçlü paket yöneticisi
- 🛠️ **LSP**: Modern editör desteği

## Kurulum

```bash
make build
make install
```

## Hızlı Başlangıç

```sky
# hello.sky
function main
  let name = "SKY"
  print("Hello, " + name + "!")
end
```

Çalıştırma:

```bash
sky run hello.sky
```

## CLI Araçları

### sky - Ana Derleyici ve Çalıştırıcı

```bash
sky run <file.sky>          # JIT ile derle ve çalıştır
sky build <file.sky>        # AOT derleyip ikili üret
sky test                    # Testleri çalıştır
sky repl                    # Etkileşimli REPL
sky dump --tokens <file>    # Lexer çıktısını göster
sky dump --ast <file>       # AST çıktısını göster
sky check <file>            # Semantik analiz yap
```

### wing - Paket Yöneticisi

```bash
wing init                   # Yeni proje başlat
wing install <paket>        # Paket kur
wing update                 # Paketleri güncelle
wing build                  # Projeyi derle
wing publish                # Paket yayınla
```

### skyls - Language Server

LSP desteği ile VS Code, Vim, Emacs gibi editörlerde otomatik tamamlama, hata gösterimi ve daha fazlası.

### skydbg - Debugger

LLDB/GDB köprüsü ile gelişmiş debugging desteği.

## Dil Özellikleri

### Değişkenler ve Sabitler

```sky
let x = 10              # Değişken (tip çıkarımı: int)
let y: float = 3.14     # Açık tip tanımı
const PI = 3.14159      # Sabit (yeniden atanamaz)
```

### Fonksiyonlar

```sky
function add(a: int, b: int): int
  return a + b
end

function greet(name)
  print("Hello, " + name)
end
```

### Kontrol Yapıları

```sky
# If-Else
if x < 10
  print("küçük")
else
  print("büyük")
end

# While
while x > 0
  x = x - 1
end

# For
for item in collection
  print(item)
end
```

### Async/Await

```sky
async function fetchData(url: string): string
  let response = await http_get(url)
  return response.body
end

function main
  let data = await fetchData("https://api.example.com")
  print(data)
end
```

### Coroutines

```sky
coop function generator(n: int)
  for i in range(n)
    yield i
  end
end

function main
  for val in generator(5)
    print(val)  # 0, 1, 2, 3, 4
  end
end
```

### Unsafe Blokları

```sky
function directMemoryAccess
  unsafe
    let ptr = malloc(1024)
    # Ham pointer işlemleri
    free(ptr)
  end
end
```

### FFI - C Entegrasyonu

```sky
import ffi

function useStrlen
  let clib = ffi.load("libc")
  let strlen = clib.symbol("strlen")
  let result = strlen("hello")
  print(result)  # 5
end
```

## Tip Sistemi

SKY, isteğe bağlı statik tip sistemine sahiptir:

- `int` - Tam sayılar
- `float` - Ondalık sayılar
- `string` - Metinler
- `bool` - Boolean değerler
- `any` - Dinamik tip

## Geliştirme

### Gereksinimler

- Go 1.22 veya üzeri
- LLVM 15+ (C API)
- Make
- golangci-lint

### Build

```bash
make build          # Tüm araçları derle
make test           # Testleri çalıştır
make lint           # Lint kontrolü
make e2e            # E2E testleri
```

### Test

```bash
go test ./...              # Unit testler
go test ./... -race        # Race detection
./scripts/e2e.sh           # E2E testler
```

## Proje Yapısı

```
sky-go/
├── cmd/                   # CLI araçları
│   ├── sky/              # Ana derleyici
│   ├── wing/             # Paket yöneticisi
│   ├── skyls/            # LSP server
│   └── skydbg/           # Debugger
├── internal/             # İç implementasyon
│   ├── lexer/           # Tokenizasyon
│   ├── parser/          # AST üretimi
│   ├── ast/             # AST düğümleri
│   ├── sema/            # Semantik analiz
│   ├── ir/              # LLVM IR üretimi
│   ├── jit/             # JIT engine
│   ├── runtime/         # Runtime & GC
│   ├── ffi/             # FFI köprüsü
│   └── ...
├── examples/             # Örnek kodlar
├── docs/                # Dokümantasyon
└── tests/               # Test dosyaları
```

## Sprint Planı

Proje 6 sprint ile geliştirilmektedir:

1. **S1**: Temel Tasarım & Gramer İskeleti (Lexer, EBNF)
2. **S2**: Parser & AST
3. **S3**: Semantik Analiz & Tip Denetimi
4. **S4**: LLVM IR & JIT (MVP)
5. **S5**: Runtime, GC, FFI, Unsafe
6. **S6**: Async/Await, LSP, Debugger, Wing

## Lisans

MIT License - Detaylar için LICENSE dosyasına bakın.

## Katkıda Bulunma

Katkılarınızı bekliyoruz! Lütfen conventional commits kullanın.

## Topluluk

- GitHub Issues: Hata raporları ve özellik istekleri
- Discussions: Sorular ve tartışmalar

---

**Not**: Bu proje aktif geliştirme aşamasındadır.

