# 🚀 VS Code Marketplace Yayınlama Rehberi

## 📋 **Adım Adım Süreç:**

### **1. Hazırlık (Tamamlandı ✅)**
- [x] Extension dosyaları hazır
- [x] package.json yapılandırıldı
- [x] Syntax highlighting çalışıyor
- [x] Snippets hazır

### **2. vsce CLI Kurulumu**
```bash
# Terminal'de şu komutu çalıştırın:
npm install -g @vscode/vsce

# Eğer permission hatası alırsanız:
sudo npm install -g @vscode/vsce
```

### **3. Microsoft Azure DevOps Hesabı**
1. [dev.azure.com](https://dev.azure.com) giriş yapın
2. **Personal Access Token** oluşturun:
   - **Scopes:** `Marketplace (manage)` seçin
   - **Expiration:** 1 yıl
3. Token'ı kaydedin (bir daha göremezsiniz!)

### **4. Publisher Oluşturma**
```bash
# Extension dizinine gidin:
cd /Users/melihburakmemis/Documents/sky-go/.vscode/extensions/sky-language

# Publisher oluşturun (ilk kez):
vsce create-publisher melihburakmmm
```

### **5. Package Oluşturma**
```bash
# .vsix dosyası oluşturun:
vsce package

# Çıktı: sky-language-0.1.0.vsix
```

### **6. Marketplace'e Yayınlama**
```bash
# Yayınlayın:
vsce publish

# Veya manuel olarak:
vsce publish -p <YOUR_ACCESS_TOKEN>
```

---

## 🎯 **Marketplace'te Görünecek:**

### **Extension Bilgileri:**
- **İsim:** SKY Language Support
- **Publisher:** melihburakmmm
- **Versiyon:** 0.1.0
- **Açıklama:** Syntax highlighting and snippets for SKY Programming Language

### **Özellikler:**
- ✅ Syntax highlighting (keywords, functions, strings, comments)
- ✅ Code snippets (fn, afn, class, if, for, enum, match)
- ✅ Language configuration
- ✅ Turkish/English support

---

## 🔧 **Yayınlama Sonrası:**

### **Kullanıcılar İçin:**
```bash
# VS Code'da Extension sekmesinde:
# "SKY Language Support" arayın
# "Install" butonuna tıklayın
```

### **Güncellemeler:**
```bash
# package.json'da version'ı artırın (0.1.1, 0.1.2, vs.)
# Sonra tekrar yayınlayın:
vsce publish
```

---

## 📝 **Notlar:**

1. **İlk yayınlama** 24-48 saat sürebilir
2. **Güncellemeler** genelde 1-2 saatte yayınlanır
3. **Marketplace URL:** `https://marketplace.visualstudio.com/items?itemName=melihburakmmm.sky-language-support`

---

## 🚨 **Sorun Giderme:**

### **Permission Hatası:**
```bash
# macOS'ta:
sudo npm install -g @vscode/vsce
```

### **Publisher Hatası:**
```bash
# Mevcut publisher'ı kontrol edin:
vsce show melihburakmmm
```

### **Token Hatası:**
- Azure DevOps'ta yeni token oluşturun
- `Marketplace (manage)` scope'u olduğundan emin olun

---

## ✅ **Başarılı Yayınlama Sonrası:**

Kullanıcılar şu şekilde yükleyebilecek:
1. VS Code açın
2. Extensions sekmesi (Ctrl+Shift+X)
3. "SKY Language Support" arayın
4. Install butonuna tıklayın
5. `.sky` dosyaları otomatik olarak syntax highlighting alacak!

🎉 **Artık SKY dili herkes tarafından kullanılabilir olacak!**
