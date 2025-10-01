// Compiler - AST'ten Bytecode Üretici
// Bound AST'yi bytecode'a dönüştürür

use super::chunk::{Chunk, Value};
use super::op::OpCode;
use crate::compiler::binder::{BoundAst, BoundStmt, BoundExpr, symbols::Slot};
use crate::compiler::parser::ast::{Literal, BinaryOp, UnaryOp, Expr};
use crate::compiler::diag::{Diagnostic, Span};

/// Bytecode derleyicisi
pub struct Compiler {
    chunk: Chunk,
    functions: Vec<Chunk>,
    function_info: Vec<FunctionInfo>, // Fonksiyon bilgileri
    loop_stack: Vec<LoopContext>, // Loop context stack
    global_types: Vec<Option<crate::compiler::types::decl::TypeDecl>>, // Global değişken tipleri
}

/// Fonksiyon bilgisi
#[derive(Debug, Clone)]
pub struct FunctionInfo {
    pub name: String,
    pub param_count: usize,
    pub chunk_index: usize,
}

/// Loop context bilgisi
#[derive(Debug, Clone)]
struct LoopContext {
    start_pos: usize,    // Loop başlangıç pozisyonu
    break_positions: Vec<usize>, // Break statement'larının pozisyonları
    continue_positions: Vec<usize>, // Continue statement'larının pozisyonları
}

impl Compiler {
    pub fn new() -> Self {
        Self {
            chunk: Chunk::new(),
            functions: Vec::new(),
            function_info: Vec::new(),
            loop_stack: Vec::new(),
            global_types: Vec::new(),
        }
    }

    pub fn compile_ast(&mut self, ast: BoundAst) -> Result<Chunk, Diagnostic> {
        for stmt in ast.statements {
            self.compile_statement(stmt)?;
        }
        
        self.chunk.write_op(OpCode::Return, 0);
        Ok(self.chunk.clone())
    }

    pub fn get_functions(&self) -> &Vec<Chunk> {
        &self.functions
    }
    
    pub fn get_function_info(&self) -> &Vec<FunctionInfo> {
        &self.function_info
    }
    
    pub fn get_global_types(&self) -> &Vec<Option<crate::compiler::types::decl::TypeDecl>> {
        &self.global_types
    }

    fn compile_statement(&mut self, stmt: BoundStmt) -> Result<(), Diagnostic> {
        match stmt {
            BoundStmt::VarDecl { value, symbol, .. } => {
                self.compile_expression(value)?;
                
                // Tip kontrolü ile store işlemi
                if let Some(ty) = &symbol.ty {
                    let type_code = ty.to_bytecode_type();
                    match symbol.slot {
                        crate::compiler::binder::symbols::Slot::Local(idx) => {
                            self.chunk.write_op(OpCode::StoreLocalTyped(idx.try_into().unwrap(), type_code), 0);
                        }
                        crate::compiler::binder::symbols::Slot::Global(idx) => {
                            // Global değişken tipini kaydet
                            let global_idx = idx.try_into().unwrap();
                            while self.global_types.len() <= global_idx as usize {
                                self.global_types.push(None);
                            }
                            self.global_types[global_idx as usize] = Some(ty.clone());
                            
                            self.chunk.write_op(OpCode::StoreGlobalTyped(global_idx, type_code), 0);
                        }
                    }
                } else {
                    // Tip bilgisi yoksa normal store
                    match symbol.slot {
                        crate::compiler::binder::symbols::Slot::Local(idx) => {
                            self.chunk.write_op(OpCode::StoreLocal(idx.try_into().unwrap()), 0);
                        }
                        crate::compiler::binder::symbols::Slot::Global(idx) => {
                            self.chunk.write_op(OpCode::StoreGlobal(idx.try_into().unwrap()), 0);
                        }
                    }
                }
            }
            
            BoundStmt::Func { kind, name, params, body, symbol, .. } => {
                // Fonksiyon için yeni chunk oluştur
                let mut func_chunk = Chunk::new();
                let param_count = params.len();
                
                // Fonksiyon gövdesini derle (parametreler VM tarafından local scope'a yerleştirilir)
                for body_stmt in body {
                    self.compile_statement_to_chunk(&mut func_chunk, body_stmt)?;
                }
                
                // Function sonuna RETURN opcode'u ekle (eğer yoksa)
                if func_chunk.code.is_empty() || func_chunk.code[func_chunk.code.len() - 1] != 0x62 {
                    func_chunk.write_op(OpCode::Return, 0);
                }
                
                // Fonksiyon chunk'ını kaydet
                let func_index = self.functions.len();
                self.functions.push(func_chunk);
                
                // Fonksiyon bilgilerini kaydet
                self.function_info.push(FunctionInfo {
                    name: name.clone(),
                    param_count,
                    chunk_index: func_index,
                });
                
                // Fonksiyon oluşturma opcode'u - parametre sayısını da kaydet
                match kind {
                    crate::compiler::parser::ast::FuncKind::Normal => {
                        self.chunk.write_op(OpCode::MakeFunction(func_index as u16, param_count as u8), 0);
                    }
                    crate::compiler::parser::ast::FuncKind::Async => {
                        // Async function'lar da normal function gibi compile edilir
                        // Event loop ve await handling VM'de yapılır
                        self.chunk.write_op(OpCode::MakeFunction(func_index as u16, param_count as u8), 0);
                    }
                    crate::compiler::parser::ast::FuncKind::Coop => {
                        self.chunk.write_op(OpCode::MakeCoopFunction(func_index as u16, param_count as u8), 0);
                    }
                    _ => {
                        self.chunk.write_op(OpCode::MakeFunction(func_index as u16, param_count as u8), 0);
                    }
                }
                
                // Global'e kaydet
                match symbol.slot {
                    crate::compiler::binder::symbols::Slot::Global(idx) => {
                        self.chunk.write_op(OpCode::StoreGlobal(idx.try_into().unwrap()), 0);
                    }
                    _ => return Err(Diagnostic::error("E0000", "Function must be global", Span::new(0, 0, 0))),
                }
            }
            
            BoundStmt::If { condition, then_branch, elif_branches, else_branch, .. } => {
                // Koşulu derle
                self.compile_expression(condition)?;
                
                // Jump if false to elif/else
                let jump_offset = self.chunk.len();
                self.chunk.write_op(OpCode::JumpIfFalse(0), 0); // Placeholder - sonra düzeltilecek
                
                // Then branch
                for stmt in then_branch {
                    self.compile_statement(stmt)?;
                }
                
                // Jump to end of if-elif-else chain
                let mut end_jump_offset = self.chunk.len();
                self.chunk.write_op(OpCode::Jump(0), 0); // Placeholder - sonra düzeltilecek
                
                // Then branch jump offset'ini düzelt (if statement'ın sonuna atla)
                let if_end_pos = self.chunk.len();
                let offset = if_end_pos as i32 - (jump_offset + 3) as i32;
                if offset >= 0 && offset <= u16::MAX as i32 {
                    self.chunk.code[jump_offset + 1..jump_offset + 3].copy_from_slice(&(offset as u16).to_le_bytes());
                }
                
                // Elif branches
                for (elif_condition, elif_body) in elif_branches {
                    // Elif koşulu derle
                    self.compile_expression(elif_condition)?;
                    
                    // Jump if false to next elif/else
                    let elif_jump_offset = self.chunk.len();
                    self.chunk.write_op(OpCode::JumpIfFalse(0), 0); // Placeholder
                    
                    // Elif body
                    for stmt in elif_body {
                        self.compile_statement(stmt)?;
                    }
                    
                    // Jump to end
                    let elif_end_jump_offset = self.chunk.len();
                    self.chunk.write_op(OpCode::Jump(0), 0); // Placeholder
                    
                    // Elif jump offset'ini düzelt
                    let next_pos = self.chunk.len();
                    let elif_offset_bytes = next_pos.to_le_bytes();
                    self.chunk.code[elif_jump_offset + 1..elif_jump_offset + 3].copy_from_slice(&elif_offset_bytes[..2]);
                    
                    // Patch previous end jump to current elif
                    let prev_end_jump = end_jump_offset;
                    let prev_offset_bytes = next_pos.to_le_bytes();
                    self.chunk.code[prev_end_jump + 1..prev_end_jump + 3].copy_from_slice(&prev_offset_bytes[..2]);
                    
                    // Update end_jump_offset for next iteration
                    end_jump_offset = elif_end_jump_offset;
                }
                
                // Else branch
                if let Some(else_body) = else_branch {
                    for stmt in else_body {
                        self.compile_statement(stmt)?;
                    }
                }
                
                // Final end jump offset'ini düzelt
                let final_end_pos = self.chunk.len();
                let final_offset_bytes = final_end_pos.to_le_bytes();
                self.chunk.code[end_jump_offset + 1..end_jump_offset + 3].copy_from_slice(&final_offset_bytes[..2]);
            }
            
            BoundStmt::For { variable, iterable, body, symbol, .. } => {
                // For loop implementasyonu - iterator pattern kullanarak
                // 1. Iterable'ı stack'e yükle
                self.compile_expression(iterable)?;
                
                // 2. Iterator oluştur
                self.chunk.write_op(OpCode::IterNew, 0);
                
                // 3. Loop başlangıç pozisyonu
                let loop_start = self.chunk.len();
                
                // 4. Loop context'i stack'e ekle
                let loop_context = LoopContext {
                    start_pos: loop_start,
                    break_positions: Vec::new(),
                    continue_positions: Vec::new(),
                };
                self.loop_stack.push(loop_context);
                
                // 5. Iterator'dan sonraki elemanı al
                self.chunk.write_op(OpCode::IterNext, 0);
                
                // 6. Iterator bitti mi kontrol et
                self.chunk.write_op(OpCode::IterDone, 0);
                
                // 7. Eğer bitmişse döngüden çık
                self.chunk.write_op(OpCode::JumpIfFalse(0), 0); // Placeholder - sonra düzeltilecek
                let loop_end_offset = self.chunk.len() - 3;
                
                // 8. Variable'a ata (symbol'dan slot al)
                match &symbol.slot {
                    crate::compiler::binder::symbols::Slot::Local(slot) => {
                        self.chunk.write_op(OpCode::StoreLocal((*slot) as u16), 0);
                    }
                    crate::compiler::binder::symbols::Slot::Global(slot) => {
                        self.chunk.write_op(OpCode::StoreGlobal((*slot) as u16), 0);
                    }
                }
                
                // 9. Loop body'sini compile et
                for body_stmt in body {
                    self.compile_statement(body_stmt)?;
                }
                
                // 10. Continue statement'larını patch et
                if let Some(loop_context) = self.loop_stack.last_mut() {
                    let continue_target = loop_context.start_pos;
                    for &continue_pos in &loop_context.continue_positions {
                        self.chunk.patch_jump(continue_pos, continue_target);
                    }
                }
                
                // 11. Loop'a geri dön
                let back_jump_pos = self.chunk.len();
                self.chunk.write_op(OpCode::Jump(0), 0); // Placeholder
                self.chunk.patch_jump(back_jump_pos, loop_start);
                
                // 12. Loop end jump offset'ini düzelt - for loop'un sonuna jump et
                let end_pos = self.chunk.len();
                let jump_target = end_pos; // for loop'un sonu
                let offset = jump_target as i32 - (loop_end_offset + 3) as i32;
                if offset >= 0 && offset <= u16::MAX as i32 {
                    self.chunk.code[loop_end_offset + 1..loop_end_offset + 3].copy_from_slice(&(offset as u16).to_le_bytes());
                }
                
                // 13. Break statement'larını patch et
                if let Some(loop_context) = self.loop_stack.pop() {
                    for &break_pos in &loop_context.break_positions {
                        self.chunk.patch_jump(break_pos, end_pos); // Loop'un sonuna jump et
                    }
                }
            }
            
            BoundStmt::While { condition, body, .. } => {
                let loop_start = self.chunk.len();
                
                // Loop context'i stack'e ekle
                let loop_context = LoopContext {
                    start_pos: loop_start,
                    break_positions: Vec::new(),
                    continue_positions: Vec::new(),
                };
                self.loop_stack.push(loop_context);
                
                // Koşulu derle
                self.compile_expression(condition)?;
                
                // Jump if false
                let jump_offset = self.chunk.len();
                self.chunk.write_op(OpCode::JumpIfFalse(0), 0); // Placeholder - sonra düzeltilecek
                
                // Loop body
                for stmt in body {
                    self.compile_statement(stmt)?;
                }
                
                // Continue statement'larını patch et
                if let Some(loop_context) = self.loop_stack.last_mut() {
                    let continue_target = loop_context.start_pos;
                    for &continue_pos in &loop_context.continue_positions {
                        self.chunk.patch_jump(continue_pos, continue_target);
                    }
                }
                
                // Loop'a geri dön
                let back_jump_pos = self.chunk.len();
                self.chunk.write_op(OpCode::Jump(0), 0); // Placeholder
                self.chunk.patch_jump(back_jump_pos, loop_start);
                
                // Exit jump offset'ini düzelt - while loop'un sonuna jump et
                let end_pos = self.chunk.len();
                let jump_target = end_pos; // while loop'un sonu
                let offset = jump_target as i32 - (jump_offset + 3) as i32;
                if offset >= 0 && offset <= u16::MAX as i32 {
                    self.chunk.code[jump_offset + 1..jump_offset + 3].copy_from_slice(&(offset as u16).to_le_bytes());
                }
                
                // Break statement'larını patch et - while loop'un sonuna jump et
                if let Some(loop_context) = self.loop_stack.pop() {
                    for &break_pos in &loop_context.break_positions {
                        self.chunk.patch_jump(break_pos, end_pos);
                    }
                }
            }
            
            BoundStmt::Return { value, .. } => {
                if let Some(value) = value {
                    self.compile_expression(value)?;
                } else {
                    let const_idx = self.chunk.add_constant(Value::Null);
                    self.chunk.write_op(OpCode::Const(const_idx), 0);
                }
                self.chunk.write_op(OpCode::Return, 0);
            }
            
            BoundStmt::Break { .. } => {
                // Break implementation - loop'tan çık
                let jump_offset = self.chunk.len();
                self.chunk.write_op(OpCode::Jump(0), 0); // Placeholder - sonra düzeltilecek
                
                // Loop context'e break pozisyonunu ekle
                if let Some(loop_context) = self.loop_stack.last_mut() {
                    loop_context.break_positions.push(jump_offset);
                }
                // Break statement'ı hemen patch etme - while loop'un sonunda patch edilecek
            }
            
            BoundStmt::Continue { .. } => {
                // Continue implementation - loop başına dön
                let jump_offset = self.chunk.len();
                self.chunk.write_op(OpCode::Jump(0), 0); // Placeholder - sonra düzeltilecek
                
                // Loop context'e continue pozisyonunu ekle
                if let Some(loop_context) = self.loop_stack.last_mut() {
                    loop_context.continue_positions.push(jump_offset);
                }
            }
            
            BoundStmt::Import { ref module_name, .. } => {
                // Import implementation - modülü global namespace'e ekle
                // Modül adını constant olarak ekle
                let const_idx = self.chunk.add_constant(Value::String(module_name.clone()));
                self.chunk.write_op(OpCode::Const(const_idx), stmt.span().start_line());
                
                // Modülü global scope'a kaydet
                // Stdlib modülleri için özel handling
                match module_name.as_str() {
                    "math" => {
                        // Math modülü için mock implementation
                        let math_module = Value::String("math_module".to_string());
                        let math_idx = self.chunk.add_constant(math_module);
                        self.chunk.write_op(OpCode::Const(math_idx), stmt.span().start_line());
                        self.chunk.write_op(OpCode::StoreGlobal(6), stmt.span().start_line()); // math global slot
                    }
                    _ => {
                        // Diğer modüller için genel handling
                        let module_value = Value::String(format!("{}_module", module_name));
                        let module_idx = self.chunk.add_constant(module_value);
                        self.chunk.write_op(OpCode::Const(module_idx), stmt.span().start_line());
                        // Global slot assignment burada yapılacak
                    }
                }
            }
            
            BoundStmt::Import { module_name, .. } => {
                // Import statement'ı şimdilik ignore et
                // Gerçek modül yükleme VM'de yapılacak
                println!("Import statement: {}", module_name);
            }
            
            BoundStmt::ExprStmt { expr, .. } => {
                self.compile_expression(expr)?;
                self.chunk.write_op(OpCode::Pop, 0);
            }
        }
        
        Ok(())
    }

    fn compile_statement_to_chunk(&mut self, chunk: &mut Chunk, stmt: BoundStmt) -> Result<(), Diagnostic> {
        // Geçici olarak chunk'ı değiştir
        let old_chunk = std::mem::replace(&mut self.chunk, chunk.clone());
        let result = self.compile_statement(stmt);
        *chunk = std::mem::replace(&mut self.chunk, old_chunk);
        result
    }

    /// Expression'ı Value olarak compile et (constant için)
    fn compile_expr_to_value(&self, expr: &Expr) -> Result<Value, Diagnostic> {
        match expr {
            Expr::Lit(value) => {
                match value {
                    Literal::Int(i) => Ok(Value::Int(*i)),
                    Literal::Float(f) => Ok(Value::Float(*f)),
                    Literal::String(s) => Ok(Value::String(s.clone())),
                    Literal::Bool(b) => Ok(Value::Bool(*b)),
                    Literal::Null => Ok(Value::Null),
                    Literal::List(items) => {
                        let mut list_values = Vec::new();
                        for item in items {
                            list_values.push(self.compile_expr_to_value(&item)?);
                        }
                        Ok(Value::List(list_values))
                    },
                    Literal::Map(entries) => {
                        let mut map_values = Vec::new();
                        for (key, value) in entries {
                            let compiled_value = self.compile_expr_to_value(&value)?;
                            map_values.push((key.clone(), compiled_value));
                        }
                        Ok(Value::Map(map_values))
                    }
                }
            },
            _ => Err(Diagnostic {
                code: "E9999".to_string(),
                message: "Non-constant expression in constant context".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
                severity: crate::compiler::diag::Severity::Error,
                notes: vec![],
            }),
        }
    }

    fn compile_expression(&mut self, expr: BoundExpr) -> Result<(), Diagnostic> {
        match expr {
            BoundExpr::Lit(lit) => {
                let value = match lit {
                    Literal::Int(i) => Value::Int(i),
                    Literal::Float(f) => Value::Float(f),
                    Literal::String(s) => Value::String(s),
                    Literal::Bool(b) => Value::Bool(b),
                    Literal::Null => Value::Null,
                    Literal::List(items) => {
                        let mut list_values = Vec::new();
                        for item in items {
                            list_values.push(self.compile_expr_to_value(&item)?);
                        }
                        Value::List(list_values)
                    },
                    Literal::Map(entries) => {
                        let mut map_values = Vec::new();
                        for (key, value) in entries {
                            let compiled_value = self.compile_expr_to_value(&value)?;
                            map_values.push((key.clone(), compiled_value));
                        }
                        Value::Map(map_values)
                    }
                };
                let const_idx = self.chunk.add_constant(value.clone());
                self.chunk.write_op(OpCode::Const(const_idx), 0);
            }
            
            BoundExpr::Ident { symbol, .. } => {
                match symbol.slot {
                    crate::compiler::binder::symbols::Slot::Local(idx) => {
                        self.chunk.write_op(OpCode::LoadLocal(idx.try_into().unwrap()), 0);
                    }
                    crate::compiler::binder::symbols::Slot::Global(idx) => {
                        let idx_u16: u16 = idx.try_into().unwrap();
                        self.chunk.write_op(OpCode::LoadGlobal(idx_u16), 0);
                    }
                }
            }
            
            BoundExpr::Call { callee, args, .. } => {
                
                // Callee'yi derle
                self.compile_expression(*callee)?;
                
                // Argümanları derle
                let args_len = args.len();
                for arg in args.iter() {
                    self.compile_expression(arg.clone())?;
                }
                
                // Fonksiyonu çağır
                self.chunk.write_op(OpCode::Call(args_len as u8), 0);
            }
            
            BoundExpr::Unary { op, expr, .. } => {
                self.compile_expression(*expr)?;
                
                match op {
                    UnaryOp::Neg => {
                        // Negatif sayı için: 0 - expr
                        let const_idx = self.chunk.add_constant(Value::Int(0));
                        self.chunk.write_op(OpCode::Const(const_idx), 0);
                        self.chunk.write_op(OpCode::Sub, 0);
                    },
                    UnaryOp::Not => self.chunk.write_op(OpCode::Not, 0),
                }
            }
            
            BoundExpr::Binary { left, op, right, .. } => {
                match op {
                BinaryOp::Assign => {
                    // Assignment: önce sağ tarafı (value) derle, sonra sol tarafı (identifier) derle
                    self.compile_expression(*right)?;

                    if let BoundExpr::Ident { name, symbol, .. } = *left {
                        // Değeri duplicate et çünkü assignment sonuç döndürmeli
                        self.chunk.write_op(OpCode::Dup, 0);
                        
                        // Tip kontrolü ile store işlemi
                        if let Some(ty) = &symbol.ty {
                            let type_code = ty.to_bytecode_type();
                            match symbol.slot {
                                Slot::Local(slot) => {
                                    self.chunk.write_op(OpCode::StoreLocalTyped(slot as u16, type_code), 0);
                                }
                                Slot::Global(slot) => {
                                    self.chunk.write_op(OpCode::StoreGlobalTyped(slot as u16, type_code), 0);
                                }
                            }
                        } else {
                            // Tip bilgisi yoksa normal store
                            match symbol.slot {
                                Slot::Local(slot) => {
                                    self.chunk.write_op(OpCode::StoreLocal(slot as u16), 0);
                                }
                                Slot::Global(slot) => {
                                    self.chunk.write_op(OpCode::StoreGlobal(slot as u16), 0);
                                }
                            }
                        }
                    } else if let BoundExpr::Index { object, index, .. } = *left {
                        // Index assignment: object[index] = value
                        // SET_INDEX bekliyor: value, index, container
                        // Önce value'yu compile et (zaten stack'te)
                        self.compile_expression(*object)?;  // container
                        self.compile_expression(*index)?;   // index
                        self.chunk.write_op(OpCode::SetIndex, 0);
                    } else {
                        return Err(Diagnostic::error("E0001", "Assignment target must be an identifier or index", Span::new(0, 0, 0)));
                    }
                }
                    _ => {
                        // Diğer binary operatörler için normal sıralama
                        self.compile_expression(*left)?;
                        self.compile_expression(*right)?;
                    }
                }
                
                match op {
                    BinaryOp::Add => self.chunk.write_op(OpCode::Add, 0),
                    BinaryOp::Sub => self.chunk.write_op(OpCode::Sub, 0),
                    BinaryOp::Mul => self.chunk.write_op(OpCode::Mul, 0),
                    BinaryOp::Div => self.chunk.write_op(OpCode::Div, 0),
                    BinaryOp::Mod => self.chunk.write_op(OpCode::Mod, 0),
                    BinaryOp::Eq => self.chunk.write_op(OpCode::Equal, 0),
                    BinaryOp::Ne => self.chunk.write_op(OpCode::NotEqual, 0),
                    BinaryOp::Lt => self.chunk.write_op(OpCode::Less, 0),
                    BinaryOp::Le => self.chunk.write_op(OpCode::LessEqual, 0),
                    BinaryOp::Gt => self.chunk.write_op(OpCode::Greater, 0),
                    BinaryOp::Ge => self.chunk.write_op(OpCode::GreaterEqual, 0),
                    BinaryOp::And => self.chunk.write_op(OpCode::And, 0),
                    BinaryOp::Or => self.chunk.write_op(OpCode::Or, 0),
                    BinaryOp::Assign => {
                        // Assignment işlemi yukarıda özel olarak işlendi
                        // Burada hiçbir şey yapmaya gerek yok
                    }
                }
            }
            
            BoundExpr::Await { expr, .. } => {
                self.compile_expression(*expr)?;
                self.chunk.write_op(OpCode::Await, 0);
            }
            
            BoundExpr::Yield { expr, .. } => {
                if let Some(expr) = expr {
                    self.compile_expression(*expr)?;
                } else {
                    let const_idx = self.chunk.add_constant(Value::Null);
                    self.chunk.write_op(OpCode::Const(const_idx), 0);
                }
                self.chunk.write_op(OpCode::Yield, 0);
            }
            
            BoundExpr::Range { start, end, .. } => {
                // Range expression: start ve end değerlerini compile et
                self.compile_expression(*start)?;
                self.compile_expression(*end)?;
                // Range opcode - start ve end değerlerini alıp Range value oluştur
                self.chunk.write_op(OpCode::MakeRange, 0);
            }
            
            BoundExpr::Attr { object, attr, .. } => {
                self.compile_expression(*object)?;
                // Attribute access - object.field
                // Attribute adını stack'e push et
                let const_idx = self.chunk.add_constant(Value::String(attr.clone()));
                self.chunk.write_op(OpCode::Const(const_idx), 0);
                // Attribute access opcode
                self.chunk.write_op(OpCode::GetAttr, 0);
            }
            
            BoundExpr::Index { object, index, .. } => {
                self.compile_expression(*object)?;
                self.compile_expression(*index)?;
                // Index access opcode
                self.chunk.write_op(OpCode::GetIndex, 0);
            }
        }
        
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::compiler::lexer::lex;
    use crate::compiler::parser::parse;
    use crate::compiler::binder::bind;

    #[test]
    fn test_simple_arithmetic_compilation() {
        let tokens = lex("int x = 1 + 2").unwrap();
        let ast = parse(tokens).unwrap();
        let bound_ast = bind(ast).unwrap();
        let mut compiler = Compiler::new();
        let chunk = compiler.compile_ast(bound_ast).unwrap();
        
        assert!(!chunk.code.is_empty());
    }

    #[test]
    fn test_function_compilation() {
        let tokens = lex("function test()\n  return 42").unwrap();
        let ast = parse(tokens).unwrap();
        let bound_ast = bind(ast).unwrap();
        let mut compiler = Compiler::new();
        let chunk = compiler.compile_ast(bound_ast).unwrap();
        
        assert!(!chunk.code.is_empty());
    }
}
