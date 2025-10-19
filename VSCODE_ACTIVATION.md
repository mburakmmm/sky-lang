# 🎨 VS Code Extension Aktivasyonu - Adım Adım

## ⚠️ Eğer Syntax Highlighting Çalışmıyorsa:

### Yöntem 1: Manuel Language Mode Seçimi

1. **VS Code'da `test_vscode.sky` dosyasını açın**

2. **Sağ alt köşeye bakın** - "Plain Text" yazıyor olabilir

3. **"Plain Text"'e tıklayın** VEYA **Cmd+K M** basın

4. **Arama kutusuna "sky" yazın**

5. **"SKY" veya "sky"'ı seçin**

6. **Artık renklendirme çalışmalı!** 🎉

### Yöntem 2: Developer Reload

1. **Cmd+Shift+P** (Command Palette)
2. "Developer: Reload Window" yazın
3. Enter'a basın
4. VS Code yeniden yüklenecek

### Yöntem 3: Extension'ı Kontrol Edin

1. **Cmd+Shift+X** (Extensions)
2. Arama kutusuna **"@installed sky"** yazın
3. **SKY Language Support** görünüyor mu?
4. Yoksa aşağıdaki komutu çalıştırın:

```bash
# Extension'ı yeniden kur
rm -rf ~/.vscode/extensions/sky-language-0.1.0
./scripts/install_vscode_extension.sh
# VS Code'u restart edin
```

---

## ✅ ÇALIŞIYORSA:

**Göreceğiniz renkler:**
- 🟣 **function**, **class**, **enum** → Mor/Pembe
- 🔵 **if**, **else**, **for**, **while** → Mavi
- 🟢 **int**, **string**, **bool** → Yeşil
- 🟡 **print**, **len**, **range** → Sarı
- 🟠 **"strings"** → Turuncu
- 💬 **# comments** → Gri
- 🔴 **42**, **3.14** → Kırmızı/Turuncu

---

## 🧪 TEST:

Dosyada şu satırı görün:
```sky
function greet(person)
```

**Beklenen:**
- `function` → Mor keyword
- `greet` → Sarı function name
- `person` → Normal text

**Eğer hepsi aynı renkte ise:**
→ Sağ alt köşeden "SKY" language'ı seçin!

---

## 🚀 Snippets Testi:

1. Yeni bir satır açın
2. `fn` yazın
3. **Tab** basın
4. Function template gelecek!

**Diğer snippets:**
- `class<Tab>` → Class
- `if<Tab>` → If
- `enum<Tab>` → Enum
- `match<Tab>` → Match

---

Sorun devam ederse bana ekran görüntüsü veya VS Code'un hangi language mode'unda olduğunu söyleyin!

