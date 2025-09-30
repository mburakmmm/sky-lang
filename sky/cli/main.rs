// Sky CLI - Komut Satırı Aracı
// sky run, sky repl, sky fmt komutları

use clap::{Parser, Subcommand};
use std::path::PathBuf;
use std::io::{self, Write, BufRead};
use std::fs;
use std::process;

use sky::compiler::diag::{Diagnostic, Emitter, SourceMap};
use sky::compiler::lexer::lex;
use sky::compiler::binder::bind;
use sky::compiler::bytecode::compiler::Compiler;
use sky::compiler::vm::Vm;
use sky::formatter::{FormatterConfig, format_and_save};

/// Sky programlama dili CLI aracı
#[derive(Parser)]
#[command(name = "sky")]
#[command(version = "0.1.0")]
#[command(about = "Sky programlama dili - Girintili sözdizimi, tip beyanı zorunlu, dinamik & güçlü tip")]
#[command(long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Sky dosyasını çalıştır
    Run {
        /// Çalıştırılacak .sky dosyası
        file: PathBuf,
        
        /// Debug modu - bytecode ve VM detayları
        #[arg(short, long)]
        debug: bool,
        
        /// Hata durumunda stack trace göster
        #[arg(short, long)]
        trace: bool,
    },
    
    /// REPL (Read-Eval-Print Loop) başlat
    Repl {
        /// REPL'de debug modu
        #[arg(short, long)]
        debug: bool,
    },
    
    /// Sky dosyasını biçimlendir
    Fmt {
        /// Biçimlendirilecek dosya veya klasör
        path: PathBuf,
        
        /// Girinti boyutu (2 veya 4)
        #[arg(short, long, default_value = "2")]
        indent: usize,
        
        /// Maksimum satır uzunluğu
        #[arg(short, long, default_value = "120")]
        max_line_length: usize,
        
        /// Değişiklikleri dosyaya yaz (varsayılan: sadece göster)
        #[arg(short, long)]
        write: bool,
        
        /// Tüm dosyaları recursive olarak biçimlendir
        #[arg(short, long)]
        recursive: bool,
    },
    
    /// Sky dosyasını parse et ve AST'i göster
    Parse {
        /// Parse edilecek .sky dosyası
        file: PathBuf,
        
        /// AST'i JSON olarak göster
        #[arg(short, long)]
        json: bool,
    },
    
    /// Sky dosyasını lex et ve token'ları göster
    Lex {
        /// Lex edilecek .sky dosyası
        file: PathBuf,
        
        /// Token'ları JSON olarak göster
        #[arg(short, long)]
        json: bool,
    },
    
    /// Sky dosyasını compile et ve bytecode'u göster
    Compile {
        /// Compile edilecek .sky dosyası
        file: PathBuf,
        
        /// Bytecode'u disassemble et
        #[arg(short, long)]
        disassemble: bool,
        
        /// Bytecode'u hex olarak göster
        #[arg(long)]
        hex: bool,
    },
    
    /// Sky dosyasının syntax'ını kontrol et
    Check {
        /// Kontrol edilecek .sky dosyası
        file: PathBuf,
        
        /// Sadece syntax hatalarını göster
        #[arg(short, long)]
        syntax_only: bool,
    },
    
    /// Sky sürüm bilgisi
    Version,
    
    /// Sky hakkında bilgi
    About,
}

fn main() {
    let cli = Cli::parse();
    
    match &cli.command {
        Commands::Run { file, debug, trace } => {
            run_file(file, *debug, *trace);
        }
        Commands::Repl { debug } => {
            start_repl(*debug);
        }
        Commands::Fmt { path, indent, max_line_length, write, recursive } => {
            format_files(path, *indent, *max_line_length, *write, *recursive);
        }
        Commands::Parse { file, json } => {
            parse_file(file, *json);
        }
        Commands::Lex { file, json } => {
            lex_file(file, *json);
        }
        Commands::Compile { file, disassemble, hex } => {
            compile_file(file, *disassemble, *hex);
        }
        Commands::Check { file, syntax_only } => {
            check_file(file, *syntax_only);
        }
        Commands::Version => {
            print_version();
        }
        Commands::About => {
            print_about();
        }
    }
}

/// Sky dosyasını çalıştır
fn run_file(file: &PathBuf, debug: bool, trace: bool) {
    if !file.exists() {
        eprintln!("Hata: Dosya bulunamadı: {:?}", file);
        process::exit(1);
    }
    
    if !file.extension().map_or(false, |ext| ext == "sky") {
        eprintln!("Hata: Dosya .sky uzantısına sahip olmalı: {:?}", file);
        process::exit(1);
    }
    
    let content = match fs::read_to_string(file) {
        Ok(content) => content,
        Err(e) => {
            eprintln!("Hata: Dosya okunamadı: {}", e);
            process::exit(1);
        }
    };
    
    if debug {
        println!("=== Sky Debug Modu ===");
        println!("Dosya: {:?}", file);
        println!("Boyut: {} byte", content.len());
        println!("Satır sayısı: {}", content.lines().count());
        println!();
    }
    
    // Lexer
    let tokens = match lex(&content) {
        Ok(tokens) => {
            if debug {
                println!("=== Token'lar ===");
                for (i, token) in tokens.iter().enumerate() {
                    println!("  {}: {:?}", i, token);
                }
                println!();
            }
            tokens
        }
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    // Parser
    let ast = match sky::compiler::parser::parse_with_source(tokens, content) {
        Ok(ast) => {
            if debug {
                println!("=== AST ===");
                println!("{}", format_ast(&ast));
                println!();
            }
            ast
        }
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    // Binder
    let bound_ast = match bind(ast) {
        Ok(bound_ast) => {
            if debug {
                println!("=== Bound AST ===");
                println!("{}", format_bound_ast(&bound_ast));
                println!();
            }
            bound_ast
        }
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    // Compiler
    let mut compiler = Compiler::new();
    let chunk = match compiler.compile_ast(bound_ast) {
        Ok(chunk) => {
            if debug {
                println!("=== Bytecode ===");
                println!("{}", chunk.disassemble());
                println!();
            }
            chunk
        }
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    // Debug: Bytecode'u disassemble et - KALDIRILDI
    // print!("{}", chunk.disassemble());
    
    // VM
    let functions = compiler.get_functions().clone();
    let mut vm = Vm::new_with_functions(functions);
    match vm.run(chunk) {
        Ok(result) => {
            if debug {
                println!("=== Sonuç ===");
                println!("{:?}", result);
            }
        }
        Err(e) => {
            if trace {
                eprintln!("Hata: {:?}", e);
            } else {
                eprintln!("Hata: {}", e);
            }
            process::exit(1);
        }
    }
}

/// REPL başlat
fn start_repl(debug: bool) {
    println!("Sky REPL v0.1.0");
    println!("Yardım için 'help' yazın, çıkmak için 'exit' yazın.");
    println!();
    
    let stdin = io::stdin();
    let mut input = String::new();
    let mut vm = Vm::new();
    
    loop {
        print!("sky> ");
        io::stdout().flush().unwrap();
        
        input.clear();
        stdin.lock().read_line(&mut input).unwrap();
        
        let line = input.trim();
        
        match line {
            "exit" | "quit" => {
                println!("Güle güle!");
                break;
            }
            "help" => {
                print_help();
                continue;
            }
            "clear" => {
                print!("\x1B[2J\x1B[1;1H");
                io::stdout().flush().unwrap();
                continue;
            }
            "" => continue,
            _ => {}
        }
        
        // REPL'de tek satırlık ifadeleri çalıştır
        if let Err(e) = run_repl_line(line, &mut vm, debug) {
            eprintln!("Hata: {}", e);
        }
    }
}

/// REPL satırını çalıştır
fn run_repl_line(line: &str, vm: &mut Vm, debug: bool) -> Result<(), String> {
    // Lexer
    let tokens = lex(line).map_err(|e| format!("Lexer hatası: {:?}", e))?;
    
    // Parser
    let ast = sky::compiler::parser::parse(tokens).map_err(|e| format!("Parser hatası: {:?}", e))?;
    
    // Binder
    let bound_ast = bind(ast).map_err(|e| format!("Binder hatası: {:?}", e))?;
    
    // Compiler
    let mut compiler = Compiler::new();
    let chunk = compiler.compile_ast(bound_ast).map_err(|e| format!("Compiler hatası: {:?}", e))?;
    
    if debug {
        println!("Bytecode: {}", chunk.disassemble());
    }
    
    // VM
    match vm.run(chunk) {
        Ok(result) => {
            if result != sky::compiler::vm::Value::Null {
                println!("{:?}", result);
            }
            Ok(())
        }
        Err(e) => Err(format!("VM hatası: {:?}", e)),
    }
}

/// REPL yardım mesajı
fn print_help() {
    println!("Sky REPL Komutları:");
    println!("  help     - Bu yardım mesajını göster");
    println!("  exit     - REPL'den çık");
    println!("  quit     - REPL'den çık");
    println!("  clear    - Ekranı temizle");
    println!();
    println!("Örnekler:");
    println!("  int x = 42");
    println!("  print(x)");
    println!("  function selam(isim: string)");
    println!("    return \"Merhaba \" + isim");
    println!("  end");
    println!("  selam(\"Sky\")");
    println!();
}

/// Dosyaları biçimlendir
fn format_files(path: &PathBuf, indent: usize, max_line_length: usize, write: bool, recursive: bool) {
    if !path.exists() {
        eprintln!("Hata: Dosya/klasör bulunamadı: {:?}", path);
        process::exit(1);
    }
    
    let config = FormatterConfig {
        indent_size: indent,
        max_line_length,
        use_tabs: false,
        trailing_commas: true,
        newline_at_end: true,
        remove_trailing_spaces: true,
    };
    
    if path.is_file() {
        format_single_file(path, &config, write);
    } else if path.is_dir() {
        format_directory(path, &config, write, recursive);
    }
}

/// Tek dosyayı biçimlendir
fn format_single_file(file: &PathBuf, config: &FormatterConfig, write: bool) {
    if !file.extension().map_or(false, |ext| ext == "sky") {
        eprintln!("Uyarı: Dosya .sky uzantısına sahip değil: {:?}", file);
        return;
    }
    
    let result = match format_and_save(file, Some(config.clone())) {
        Ok(result) => result,
        Err(e) => {
            eprintln!("Hata: Dosya biçimlendirilemedi {:?}: {}", file, e);
            return;
        }
    };
    
    if result.changes_made {
        if write {
            println!("Biçimlendirildi: {:?} ({} satır değişti)", file, result.lines_changed);
        } else {
            println!("Değişiklikler ({} satır):", result.lines_changed);
            print!("{}", result.formatted_code);
        }
    } else {
        println!("Değişiklik yok: {:?}", file);
    }
    
    // Diagnostic'leri göster
    if !result.diagnostics.is_empty() {
        print_diagnostics(&result.diagnostics);
    }
}

/// Klasörü biçimlendir
fn format_directory(dir: &PathBuf, config: &FormatterConfig, write: bool, recursive: bool) {
    let entries = match fs::read_dir(dir) {
        Ok(entries) => entries,
        Err(e) => {
            eprintln!("Hata: Klasör okunamadı {:?}: {}", dir, e);
            return;
        }
    };
    
    for entry in entries {
        let entry = match entry {
            Ok(entry) => entry,
            Err(_) => continue,
        };
        
        let path = entry.path();
        
        if path.is_file() && path.extension().map_or(false, |ext| ext == "sky") {
            format_single_file(&path, config, write);
        } else if path.is_dir() && recursive {
            format_directory(&path, config, write, recursive);
        }
    }
}

/// Dosyayı parse et
fn parse_file(file: &PathBuf, json: bool) {
    let content = match fs::read_to_string(file) {
        Ok(content) => content,
        Err(e) => {
            eprintln!("Hata: Dosya okunamadı: {}", e);
            process::exit(1);
        }
    };
    
    let tokens = match lex(&content) {
        Ok(tokens) => {
            tokens
        },
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    let ast = match sky::compiler::parser::parse_with_source(tokens, content) {
        Ok(ast) => {
            ast
        },
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    if json {
        println!("JSON serialization henüz implement edilmedi");
    } else {
        println!("AST:");
        println!("{}", format_ast(&ast));
    }
}

/// Dosyayı lex et
fn lex_file(file: &PathBuf, json: bool) {
    let content = match fs::read_to_string(file) {
        Ok(content) => content,
        Err(e) => {
            eprintln!("Hata: Dosya okunamadı: {}", e);
            process::exit(1);
        }
    };
    
    let tokens = match lex(&content) {
        Ok(tokens) => tokens,
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    if json {
        println!("JSON serialization henüz implement edilmedi");
    } else {
        for (i, token) in tokens.iter().enumerate() {
            println!("  {}: {:?}", i, token);
        }
    }
}

/// Dosyayı compile et
fn compile_file(file: &PathBuf, disassemble: bool, hex: bool) {
    let content = match fs::read_to_string(file) {
        Ok(content) => content,
        Err(e) => {
            eprintln!("Hata: Dosya okunamadı: {}", e);
            process::exit(1);
        }
    };
    
    let tokens = match lex(&content) {
        Ok(tokens) => tokens,
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    let ast = match sky::compiler::parser::parse_with_source(tokens, content) {
        Ok(ast) => ast,
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    let bound_ast = match bind(ast) {
        Ok(bound_ast) => bound_ast,
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    let mut compiler = Compiler::new();
    let chunk = match compiler.compile_ast(bound_ast) {
        Ok(chunk) => chunk,
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    if disassemble {
        println!("{}", chunk.disassemble());
    } else if hex {
        println!("Bytecode (hex):");
        for (i, byte) in chunk.code.iter().enumerate() {
            if i % 16 == 0 {
                print!("{:04x}: ", i);
            }
            print!("{:02x} ", byte);
            if i % 16 == 15 {
                println!();
            }
        }
        if chunk.code.len() % 16 != 0 {
            println!();
        }
    } else {
        println!("Bytecode boyutu: {} byte", chunk.code.len());
        println!("Sabit sayısı: {}", chunk.consts.len());
    }
}

/// Dosyayı kontrol et
fn check_file(file: &PathBuf, syntax_only: bool) {
    let content = match fs::read_to_string(file) {
        Ok(content) => content,
        Err(e) => {
            eprintln!("Hata: Dosya okunamadı: {}", e);
            process::exit(1);
        }
    };
    
    let tokens = match lex(&content) {
        Ok(tokens) => tokens,
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    let ast = match sky::compiler::parser::parse_with_source(tokens, content) {
        Ok(ast) => ast,
        Err(diag) => {
            print_diagnostics(&[diag]);
            process::exit(1);
        }
    };
    
    if !syntax_only {
        let _bound_ast = match bind(ast) {
            Ok(_) => {
                println!("✓ Dosya başarıyla kontrol edildi");
                return;
            }
            Err(diag) => {
                print_diagnostics(&[diag]);
                process::exit(1);
            }
        };
    }
    
    println!("✓ Syntax kontrolü başarılı");
}

/// Sürüm bilgisi
fn print_version() {
    println!("Sky v0.1.0");
    println!("Rust sürümü: {}", std::env::var("RUSTC_VERSION").unwrap_or_else(|_| "Unknown".to_string()));
    println!("Derleme tarihi: {}", std::env::var("BUILD_DATE").unwrap_or_else(|_| "Unknown".to_string()));
}

/// Hakkında bilgi
fn print_about() {
    println!("Sky Programlama Dili");
    println!("====================");
    println!();
    println!("Sky, Python-benzeri girintili sözdizimi ve basit tasarımı olan");
    println!("bir programlama dilidir. Dinamik & güçlü tipli çalışma zamanı,");
    println!("async/await, coroutines, VM bytecode ve Python/JS köprüleri");
    println!("destekler.");
    println!();
    println!("Özellikler:");
    println!("  • Girintili sözdizimi (Python-benzeri)");
    println!("  • Tip beyanı zorunlu (var, int, float, bool, string, list, map)");
    println!("  • Dinamik & güçlü tip (runtime kontrolleri)");
    println!("  • Async/await (event loop)");
    println!("  • Coroutines (coop/yield)");
    println!("  • VM bytecode interpreter");
    println!("  • Garbage collection");
    println!("  • Python ve JS köprüleri");
    println!("  • Unicode identifier desteği (Türkçe karakterler)");
    println!();
    println!("Lisans: MIT");
    println!("Implementasyon: Rust 1.79+");
}

/// Diagnostic'leri yazdır
fn print_diagnostics(diagnostics: &[Diagnostic]) {
    let source_map = SourceMap::new();
    let emitter = Emitter::new(source_map);
    for diag in diagnostics {
        print!("{}", emitter.emit(diag));
    }
}

/// AST'i formatla
fn format_ast(ast: &sky::compiler::parser::Ast) -> String {
    let mut result = String::new();
    result.push_str("AST:\n");
    
    for (i, stmt) in ast.statements.iter().enumerate() {
        result.push_str(&format!("  {}: {:?}\n", i, stmt));
    }
    
    result
}

/// Bound AST'i formatla
fn format_bound_ast(bound_ast: &sky::compiler::binder::BoundAst) -> String {
    let mut result = String::new();
    result.push_str("Bound AST:\n");
    
    for (i, stmt) in bound_ast.statements.iter().enumerate() {
        result.push_str(&format!("  {}: {:?}\n", i, stmt));
    }
    
    result
}
