#!/usr/bin/env python3
"""
Sky Dili Build ve Deploy Scripti
Geliştirme sürecini otomatikleştirir:
1. Sürüm güncelleme
2. Cargo build
3. Binary'yi ana dizine kopyalama ve PATH'e ekleme
"""

import os
import sys
import subprocess
import shutil
from pathlib import Path

def print_step(step_name):
    """Adım başlığını yazdır"""
    print(f"\n{'='*50}")
    print(f"🔧 {step_name}")
    print(f"{'='*50}")

def print_success(message):
    """Başarı mesajını yazdır"""
    print(f"✅ {message}")

def print_error(message):
    """Hata mesajını yazdır"""
    print(f"❌ {message}")
    sys.exit(1)

def print_info(message):
    """Bilgi mesajını yazdır"""
    print(f"ℹ️  {message}")

def run_command(command, description, check=True):
    """Komut çalıştır ve sonucu kontrol et"""
    print_info(f"{description}...")
    print(f"   Komut: {command}")
    
    try:
        result = subprocess.run(command, shell=True, check=check, capture_output=True, text=True)
        if result.stdout:
            print(f"   Çıktı: {result.stdout.strip()}")
        return result
    except subprocess.CalledProcessError as e:
        print_error(f"{description} başarısız: {e}")
        if e.stderr:
            print(f"   Hata: {e.stderr.strip()}")

def get_current_version():
    """Mevcut sürümü Cargo.toml'dan oku"""
    try:
        with open("Cargo.toml", "r") as f:
            content = f.read()
            for line in content.split('\n'):
                if line.strip().startswith('version ='):
                    version = line.split('=')[1].strip().strip('"')
                    return version
        return "0.1.0"  # Varsayılan sürüm
    except FileNotFoundError:
        print_error("Cargo.toml dosyası bulunamadı!")
    except Exception as e:
        print_error(f"Sürüm okuma hatası: {e}")

def update_version(version_type):
    """Sürümü güncelle"""
    current_version = get_current_version()
    print_info(f"Mevcut sürüm: {current_version}")
    
    # Sürümü parçala
    parts = current_version.split('.')
    if len(parts) != 3:
        print_error(f"Geçersiz sürüm formatı: {current_version}")
    
    major, minor, patch = map(int, parts)
    
    # Sürümü güncelle
    if version_type == "majör":
        major += 1
        minor = 0
        patch = 0
    elif version_type == "minör":
        minor += 1
        patch = 0
    elif version_type == "mini":
        patch += 1
    else:
        print_error("Geçersiz sürüm tipi!")
    
    new_version = f"{major}.{minor}.{patch}"
    print_info(f"Yeni sürüm: {new_version}")
    
    # Cargo.toml'u güncelle
    try:
        with open("Cargo.toml", "r") as f:
            content = f.read()
        
        # Sürüm satırını güncelle
        lines = content.split('\n')
        for i, line in enumerate(lines):
            if line.strip().startswith('version ='):
                lines[i] = f'version = "{new_version}"'
                break
        
        with open("Cargo.toml", "w") as f:
            f.write('\n'.join(lines))
        
        print_success(f"Sürüm {new_version} olarak güncellendi")
        return new_version
        
    except Exception as e:
        print_error(f"Cargo.toml güncelleme hatası: {e}")

def ask_version_type():
    """Kullanıcıdan sürüm tipini sor"""
    print("\n🚀 Sürüm Güncelleme")
    print("Hangi tipte güncelleme yapmak istiyorsunuz?")
    print("1. Majör (1.0.0 → 2.0.0) - Büyük değişiklikler")
    print("2. Minör (1.0.0 → 1.1.0) - Yeni özellikler")
    print("3. Mini  (1.0.0 → 1.0.1) - Hata düzeltmeleri")
    
    while True:
        choice = input("\nSeçiminiz (1/2/3): ").strip()
        if choice == "1":
            return "majör"
        elif choice == "2":
            return "minör"
        elif choice == "3":
            return "mini"
        else:
            print("❌ Geçersiz seçim! Lütfen 1, 2 veya 3 girin.")

def check_path_entry():
    """PATH'te sky binary'sinin olup olmadığını kontrol et"""
    home = Path.home()
    sky_path = home / "sky"
    
    # PATH'i kontrol et
    path_env = os.environ.get("PATH", "")
    path_dirs = path_env.split(":")
    
    sky_in_path = False
    for path_dir in path_dirs:
        if str(home) in path_dir and "sky" in path_dir:
            sky_in_path = True
            break
    
    return sky_in_path, sky_path

def add_to_path():
    """Sky binary'sini PATH'e ekle"""
    home = Path.home()
    shell_configs = [
        home / ".bashrc",
        home / ".zshrc",
        home / ".bash_profile",
        home / ".profile"
    ]
    
    # Hangi shell config dosyası var?
    config_file = None
    for config in shell_configs:
        if config.exists():
            config_file = config
            break
    
    if not config_file:
        print_info("Shell config dosyası bulunamadı, .bashrc oluşturuluyor...")
        config_file = home / ".bashrc"
        config_file.touch()
    
    # PATH ekleme komutu
    path_line = f'export PATH="$HOME:$PATH"'
    
    # Dosyaya ekle
    try:
        with open(config_file, "r") as f:
            content = f.read()
        
        if path_line not in content:
            with open(config_file, "a") as f:
                f.write(f"\n# Sky binary path\n{path_line}\n")
            print_success(f"PATH {config_file.name} dosyasına eklendi")
        else:
            print_info("PATH zaten ekli")
            
    except Exception as e:
        print_error(f"PATH ekleme hatası: {e}")

def main():
    """Ana fonksiyon"""
    print("🌌 Sky Dili Build ve Deploy Scripti")
    print("====================================")
    
    # 1. Sürüm güncelleme
    print_step("SÜRÜM GÜNCELLEME")
    version_type = ask_version_type()
    new_version = update_version(version_type)
    
    # 2. Cargo build
    print_step("CARGO BUILD")
    run_command("cargo build --release", "Release build")
    
    # 3. Binary kopyalama ve PATH kontrolü
    print_step("BINARY KOPYALAMA VE PATH KONTROLÜ")
    
    # Binary'nin varlığını kontrol et
    binary_path = Path("target/release/sky")
    if not binary_path.exists():
        print_error("Release binary bulunamadı!")
    
    # PATH kontrolü
    sky_in_path, sky_home_path = check_path_entry()
    
    if sky_in_path:
        print_info("Sky binary zaten PATH'te")
    else:
        print_info("Sky binary PATH'te değil, ekleniyor...")
    
    # Binary'yi ana dizine kopyala
    home = Path.home()
    destination = home / "sky"
    
    try:
        shutil.copy2(binary_path, destination)
        # Executable permission ver
        os.chmod(destination, 0o755)
        print_success(f"Binary {destination} konumuna kopyalandı")
    except Exception as e:
        print_error(f"Binary kopyalama hatası: {e}")
    
    # PATH'e ekle (gerekirse)
    if not sky_in_path:
        add_to_path()
    
    # 4. Sonuç
    print_step("BAŞARILI TAMAMLANDI")
    print_success(f"Sky dili {new_version} sürümü başarıyla build edildi ve deploy edildi!")
    print_info(f"Binary konumu: {destination}")
    print_info("Yeni terminal penceresi açarak 'sky' komutunu test edebilirsiniz.")
    
    # Test komutu öner
    print("\n🧪 Test komutu:")
    print(f"   {destination} --help")

if __name__ == "__main__":
    main()
