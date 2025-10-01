// Binder - Sembol & Kapsam Çözümlemesi
// İsim çözümleme (scope stack), gölgelemeyi ve global/yerel ayırımını belirler

pub mod scope;
pub mod symbols;

use scope::{Scope, ScopeStack};
use symbols::{SymbolInfo, SymbolKind, Slot};
use crate::compiler::parser::ast::{Ast, Stmt, Expr, TypeDecl as AstTypeDecl};
use crate::compiler::types::{TypeDecl, Param};
use crate::compiler::diag::{Diagnostic, Span};
use crate::compiler::parser::ast::PrimitiveType;

impl From<AstTypeDecl> for TypeDecl {
    fn from(ast_ty: AstTypeDecl) -> Self {
        match ast_ty {
            AstTypeDecl::Var => TypeDecl::Var,
            AstTypeDecl::Int => TypeDecl::Int,
            AstTypeDecl::Float => TypeDecl::Float,
            AstTypeDecl::Bool => TypeDecl::Bool,
            AstTypeDecl::String => TypeDecl::String,
            AstTypeDecl::List => TypeDecl::List,
            AstTypeDecl::Map => TypeDecl::Map,
            AstTypeDecl::ListParam(param) => {
                let decl_param = match param {
                    PrimitiveType::Int => crate::compiler::types::decl::PrimitiveType::Int,
                    PrimitiveType::Float => crate::compiler::types::decl::PrimitiveType::Float,
                    PrimitiveType::Bool => crate::compiler::types::decl::PrimitiveType::Bool,
                    PrimitiveType::String => crate::compiler::types::decl::PrimitiveType::String,
                };
                TypeDecl::ListParam(decl_param)
            }
        }
    }
}

/// Bağlı AST - sembol bilgileri ile zenginleştirilmiş
#[derive(Debug, Clone)]
pub struct BoundAst {
    pub statements: Vec<BoundStmt>,
    pub scopes: Vec<Scope>,
    pub symbols: Vec<SymbolInfo>,
}

/// Bağlı statement
#[derive(Debug, Clone)]
pub enum BoundStmt {
    VarDecl {
        ty: TypeDecl,
        name: String,
        value: BoundExpr,
        symbol: SymbolInfo,
        span: Span,
    },
    Func {
        kind: crate::compiler::parser::ast::FuncKind,
        name: String,
        params: Vec<Param>,
        body: Vec<BoundStmt>,
        symbol: SymbolInfo,
        span: Span,
    },
    If {
        condition: BoundExpr,
        then_branch: Vec<BoundStmt>,
        elif_branches: Vec<(BoundExpr, Vec<BoundStmt>)>,
        else_branch: Option<Vec<BoundStmt>>,
        span: Span,
    },
    For {
        variable: String,
        iterable: BoundExpr,
        body: Vec<BoundStmt>,
        symbol: SymbolInfo,
        span: Span,
    },
    While {
        condition: BoundExpr,
        body: Vec<BoundStmt>,
        span: Span,
    },
    Return {
        value: Option<BoundExpr>,
        span: Span,
    },
    Break {
        span: Span,
    },
    Continue {
        span: Span,
    },
    Import {
        module_name: String,
        span: Span,
    },
    ExprStmt {
        expr: BoundExpr,
        span: Span,
    },
}

impl BoundStmt {
    pub fn span(&self) -> Span {
        match self {
            BoundStmt::VarDecl { span, .. } => *span,
            BoundStmt::Func { span, .. } => *span,
            BoundStmt::If { span, .. } => *span,
            BoundStmt::For { span, .. } => *span,
            BoundStmt::While { span, .. } => *span,
            BoundStmt::Return { span, .. } => *span,
            BoundStmt::Break { span, .. } => *span,
            BoundStmt::Continue { span, .. } => *span,
            BoundStmt::Import { span, .. } => *span,
            BoundStmt::ExprStmt { span, .. } => *span,
        }
    }
}

/// Bağlı expression
#[derive(Debug, Clone)]
pub enum BoundExpr {
    Lit(crate::compiler::parser::ast::Literal),
    Ident {
        name: String,
        symbol: SymbolInfo,
        span: Span,
    },
    Call {
        callee: Box<BoundExpr>,
        args: Vec<BoundExpr>,
        span: Span,
    },
    Attr {
        object: Box<BoundExpr>,
        attr: String,
        span: Span,
    },
    Index {
        object: Box<BoundExpr>,
        index: Box<BoundExpr>,
        span: Span,
    },
    Unary {
        op: crate::compiler::parser::ast::UnaryOp,
        expr: Box<BoundExpr>,
        span: Span,
    },
    Binary {
        left: Box<BoundExpr>,
        op: crate::compiler::parser::ast::BinaryOp,
        right: Box<BoundExpr>,
        span: Span,
    },
    Await {
        expr: Box<BoundExpr>,
        span: Span,
    },
    Yield {
        expr: Option<Box<BoundExpr>>,
        span: Span,
    },
    Range {
        start: Box<BoundExpr>,
        end: Box<BoundExpr>,
        span: Span,
    },
}

/// Ana bağlama fonksiyonu
pub fn bind(ast: Ast) -> Result<BoundAst, Diagnostic> {
    let mut binder = Binder::new();
    binder.bind_program(ast)
}

/// Bağlayıcı
pub struct Binder {
    scopes: ScopeStack,
    symbols: Vec<SymbolInfo>,
    next_slot: u32,
    in_function: bool, // Fonksiyon bağlamında mı?
    in_loop: bool, // Döngü bağlamında mı?
}

impl Binder {
    pub fn new() -> Self {
        let mut binder = Self {
            scopes: ScopeStack::new(),
            symbols: Vec::new(),
            next_slot: 0,
            in_function: false,
            in_loop: false,
        };
        
        // Global scope oluştur
        binder.scopes.push(Scope::new());
        
        // Built-in fonksiyonları ekle
        binder.declare_builtin_functions();
        
        binder
    }
    
    fn declare_builtin_functions(&mut self) {
        // print fonksiyonu
        let print_symbol = SymbolInfo::new(
            "print".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var), // print herhangi bir tip alabilir
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(print_symbol.clone());
        self.scopes.current_mut().insert("print".to_string(), print_symbol);
        
        // stringify fonksiyonu
        let stringify_symbol = SymbolInfo::new(
            "stringify".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(stringify_symbol.clone());
        self.scopes.current_mut().insert("stringify".to_string(), stringify_symbol);
        
        // now fonksiyonu
        let now_symbol = SymbolInfo::new(
            "now".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(now_symbol.clone());
        self.scopes.current_mut().insert("now".to_string(), now_symbol);
        
        // python_import fonksiyonu
        let python_import_symbol = SymbolInfo::new(
            "python_import".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(python_import_symbol.clone());
        self.scopes.current_mut().insert("python_import".to_string(), python_import_symbol);
        
        // js_eval fonksiyonu
        let js_eval_symbol = SymbolInfo::new(
            "js_eval".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(js_eval_symbol.clone());
        self.scopes.current_mut().insert("js_eval".to_string(), js_eval_symbol);
        
        // math modülü (slot 6'da)
        let math_symbol = SymbolInfo::new(
            "math".to_string(),
            SymbolKind::Variable,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(math_symbol.clone());
        self.scopes.current_mut().insert("math".to_string(), math_symbol);
        
        // IO fonksiyonları
        let input_symbol = SymbolInfo::new(
            "input".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(input_symbol.clone());
        self.scopes.current_mut().insert("input".to_string(), input_symbol);
        
        let write_symbol = SymbolInfo::new(
            "write".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(write_symbol.clone());
        self.scopes.current_mut().insert("write".to_string(), write_symbol);
        
        let error_symbol = SymbolInfo::new(
            "error".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(error_symbol.clone());
        self.scopes.current_mut().insert("error".to_string(), error_symbol);
        
        let format_symbol = SymbolInfo::new(
            "format".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(format_symbol.clone());
        self.scopes.current_mut().insert("format".to_string(), format_symbol);
        
        let len_symbol = SymbolInfo::new(
            "len".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(len_symbol.clone());
        self.scopes.current_mut().insert("len".to_string(), len_symbol);
        
        let type_symbol = SymbolInfo::new(
            "type".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(type_symbol.clone());
        self.scopes.current_mut().insert("type".to_string(), type_symbol);
        
        // Time fonksiyonları
        let sleep_symbol = SymbolInfo::new(
            "sleep".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(sleep_symbol.clone());
        self.scopes.current_mut().insert("sleep".to_string(), sleep_symbol);
        
        let date_string_symbol = SymbolInfo::new(
            "date_string".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(date_string_symbol.clone());
        self.scopes.current_mut().insert("date_string".to_string(), date_string_symbol);
        
        let timer_symbol = SymbolInfo::new(
            "timer".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(timer_symbol.clone());
        self.scopes.current_mut().insert("timer".to_string(), timer_symbol);
        
        let benchmark_symbol = SymbolInfo::new(
            "benchmark".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(benchmark_symbol.clone());
        self.scopes.current_mut().insert("benchmark".to_string(), benchmark_symbol);
        
        let timezone_symbol = SymbolInfo::new(
            "timezone".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(timezone_symbol.clone());
        self.scopes.current_mut().insert("timezone".to_string(), timezone_symbol);
        
        let micros_symbol = SymbolInfo::new(
            "micros".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(micros_symbol.clone());
        self.scopes.current_mut().insert("micros".to_string(), micros_symbol);
        
        let nanos_symbol = SymbolInfo::new(
            "nanos".to_string(),
            SymbolKind::Function,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(nanos_symbol.clone());
        self.scopes.current_mut().insert("nanos".to_string(), nanos_symbol);
        
        // HTTP modülü (slot 19)
        let http_symbol = SymbolInfo::new(
            "http".to_string(),
            SymbolKind::Variable,
            Some(TypeDecl::Var),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(http_symbol.clone());
        self.scopes.current_mut().insert("http".to_string(), http_symbol);
        
        // Özel değişkenler
        let args_symbol = SymbolInfo::new(
            "__args__".to_string(),
            SymbolKind::Variable,
            Some(TypeDecl::List),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(args_symbol.clone());
        self.scopes.current_mut().insert("__args__".to_string(), args_symbol);
        
        let name_symbol = SymbolInfo::new(
            "__name__".to_string(),
            SymbolKind::Variable,
            Some(TypeDecl::String),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(name_symbol.clone());
        self.scopes.current_mut().insert("__name__".to_string(), name_symbol);
        
        let file_symbol = SymbolInfo::new(
            "__file__".to_string(),
            SymbolKind::Variable,
            Some(TypeDecl::String),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(file_symbol.clone());
        self.scopes.current_mut().insert("__file__".to_string(), file_symbol);
        
        let dir_symbol = SymbolInfo::new(
            "__dir__".to_string(),
            SymbolKind::Variable,
            Some(TypeDecl::String),
            Slot::Global(self.next_slot),
            Span::new(0, 0, 0),
        );
        self.next_slot += 1;
        self.symbols.push(dir_symbol.clone());
        self.scopes.current_mut().insert("__dir__".to_string(), dir_symbol);
    }

    fn bind_program(&mut self, ast: Ast) -> Result<BoundAst, Diagnostic> {
        let mut bound_statements = Vec::new();
        
        for stmt in ast.statements {
            let bound_stmt = self.bind_statement(stmt)?;
            bound_statements.push(bound_stmt);
        }

        let scopes = self.scopes.all_scopes();
        
        Ok(BoundAst {
            statements: bound_statements,
            scopes,
            symbols: self.symbols.clone(),
        })
    }

    fn bind_statement(&mut self, stmt: Stmt) -> Result<BoundStmt, Diagnostic> {
        match stmt {
            Stmt::VarDecl { ty, name, value, span } => {
                let symbol = self.declare_variable(&name, &TypeDecl::from(ty.clone()), span)?;
                let bound_value = self.bind_expression(value)?;
                
                Ok(BoundStmt::VarDecl {
                    ty: TypeDecl::from(ty),
                    name,
                    value: bound_value,
                    symbol,
                    span,
                })
            }
            
            Stmt::Func { kind, name, params, body, span } => {
                let symbol = self.declare_function(&name, &kind, span)?;
                
                // Fonksiyon parametrelerini scope'a ekle
                self.scopes.push(Scope::new());
                let mut bound_params = Vec::new();
                
                for (i, param) in params.iter().enumerate() {
                    let _param_symbol = self.declare_parameter(&param.name, &TypeDecl::from(param.ty.clone()), param.span, i as u16)?;
                    bound_params.push(Param::new(param.name.clone(), TypeDecl::from(param.ty.clone()), param.span));
                }
                
                // Fonksiyon bağlamını işaretle
                let old_in_function = self.in_function;
                self.in_function = true;
                
                // Fonksiyon gövdesini bağla
                let mut bound_body = Vec::new();
                for body_stmt in body {
                    let bound_stmt = self.bind_statement(body_stmt)?;
                    bound_body.push(bound_stmt);
                }
                
                // Fonksiyon bağlamını geri yükle
                self.in_function = old_in_function;
                
                self.scopes.pop();
                
                Ok(BoundStmt::Func {
                    kind,
                    name,
                    params: bound_params,
                    body: bound_body,
                    symbol,
                    span,
                })
            }
            
            Stmt::If { condition, then_branch, elif_branches, else_branch, span } => {
                let bound_condition = self.bind_expression(condition)?;
                
                // Then branch
                self.scopes.push(Scope::new());
                let mut bound_then = Vec::new();
                for stmt in then_branch {
                    bound_then.push(self.bind_statement(stmt)?);
                }
                self.scopes.pop();
                
                // Elif branches
                let mut bound_elifs = Vec::new();
                for (cond, branch) in elif_branches {
                    let bound_cond = self.bind_expression(cond)?;
                    self.scopes.push(Scope::new());
                    let mut bound_branch = Vec::new();
                    for stmt in branch {
                        bound_branch.push(self.bind_statement(stmt)?);
                    }
                    self.scopes.pop();
                    bound_elifs.push((bound_cond, bound_branch));
                }
                
                // Else branch
                let bound_else = if let Some(else_branch) = else_branch {
                    self.scopes.push(Scope::new());
                    let mut bound_else_branch = Vec::new();
                    for stmt in else_branch {
                        bound_else_branch.push(self.bind_statement(stmt)?);
                    }
                    self.scopes.pop();
                    Some(bound_else_branch)
                } else {
                    None
                };
                
                Ok(BoundStmt::If {
                    condition: bound_condition,
                    then_branch: bound_then,
                    elif_branches: bound_elifs,
                    else_branch: bound_else,
                    span,
                })
            }
            
            Stmt::For { variable, iterable, body, span } => {
                let bound_iterable = self.bind_expression(iterable)?;
                
                // For loop için yeni scope oluştur
                self.scopes.push(Scope::new());
                let symbol = self.declare_variable(&variable, &TypeDecl::Var, span)?;
                
                // Döngü bağlamını işaretle
                let old_in_loop = self.in_loop;
                self.in_loop = true;
                
                let mut bound_body = Vec::new();
                for stmt in body {
                    let bound_stmt = self.bind_statement(stmt)?;
                    bound_body.push(bound_stmt);
                }
                
                // Döngü bağlamını geri yükle
                self.in_loop = old_in_loop;
                
                self.scopes.pop();
                
                Ok(BoundStmt::For {
                    variable,
                    iterable: bound_iterable,
                    body: bound_body,
                    symbol,
                    span,
                })
            }
            
            Stmt::While { condition, body, span } => {
                let bound_condition = self.bind_expression(condition)?;
                
                self.scopes.push(Scope::new());
                
                // Döngü bağlamını işaretle
                let old_in_loop = self.in_loop;
                self.in_loop = true;
                
                let mut bound_body = Vec::new();
                for stmt in body {
                    let bound_stmt = self.bind_statement(stmt)?;
                    bound_body.push(bound_stmt);
                }
                
                // Döngü bağlamını geri yükle
                self.in_loop = old_in_loop;
                
                self.scopes.pop();
                
                Ok(BoundStmt::While {
                    condition: bound_condition,
                    body: bound_body,
                    span,
                })
            }
            
            Stmt::Return { value, span } => {
                // Return sadece fonksiyon içinde kullanılabilir
                if !self.in_function {
                    return Err(Diagnostic::error(
                        "E0203",
                        "return statement outside function",
                        span,
                    ));
                }
                
                let bound_value = if let Some(value) = value {
                    Some(self.bind_expression(value)?)
                } else {
                    None
                };
                
                Ok(BoundStmt::Return {
                    value: bound_value,
                    span,
                })
            }
            
            Stmt::Break { span } => {
                // Break sadece döngü içinde kullanılabilir
                if !self.in_loop {
                    return Err(Diagnostic::error(
                        "E0204",
                        "break statement outside loop",
                        span,
                    ));
                }
                
                Ok(BoundStmt::Break { span })
            },
            Stmt::Continue { span } => {
                // Continue sadece döngü içinde kullanılabilir
                if !self.in_loop {
                    return Err(Diagnostic::error(
                        "E0205",
                        "continue statement outside loop",
                        span,
                    ));
                }
                
                Ok(BoundStmt::Continue { span })
            },
            
            Stmt::Import { module_name, span } => {
                // Import statement'ı şimdilik passthrough olarak bırak
                // Gerçek modül yükleme VM'de yapılacak
                Ok(BoundStmt::Import {
                    module_name,
                    span,
                })
            }
            
            Stmt::ExprStmt { expr, span } => {
                let bound_expr = self.bind_expression(expr)?;
                Ok(BoundStmt::ExprStmt {
                    expr: bound_expr,
                    span,
                })
            }
        }
    }

    fn bind_expression(&mut self, expr: Expr) -> Result<BoundExpr, Diagnostic> {
        match expr {
            Expr::Lit(lit) => Ok(BoundExpr::Lit(lit)),
            
            Expr::Ident(name, span) => {
                let symbol = self.resolve_variable(&name, span)?;
                Ok(BoundExpr::Ident {
                    name,
                    symbol,
                    span,
                })
            }
            
            Expr::Call { callee, args, span } => {
                let bound_callee = self.bind_expression(*callee)?;
                let mut bound_args = Vec::new();
                for arg in args {
                    bound_args.push(self.bind_expression(arg)?);
                }
                
                Ok(BoundExpr::Call {
                    callee: Box::new(bound_callee),
                    args: bound_args,
                    span,
                })
            }
            
            Expr::Attr { object, attr, span } => {
                let bound_object = self.bind_expression(*object)?;
                Ok(BoundExpr::Attr {
                    object: Box::new(bound_object),
                    attr,
                    span,
                })
            }
            
            Expr::Index { object, index, span } => {
                let bound_object = self.bind_expression(*object)?;
                let bound_index = self.bind_expression(*index)?;
                Ok(BoundExpr::Index {
                    object: Box::new(bound_object),
                    index: Box::new(bound_index),
                    span,
                })
            }
            
            Expr::Unary { op, expr, span } => {
                let bound_expr = self.bind_expression(*expr)?;
                Ok(BoundExpr::Unary {
                    op,
                    expr: Box::new(bound_expr),
                    span,
                })
            }
            
            Expr::Binary { left, op, right, span } => {
                let bound_left = self.bind_expression(*left)?;
                let bound_right = self.bind_expression(*right)?;
                Ok(BoundExpr::Binary {
                    left: Box::new(bound_left),
                    op,
                    right: Box::new(bound_right),
                    span,
                })
            }
            
            Expr::Await { expr, span } => {
                let bound_expr = self.bind_expression(*expr)?;
                Ok(BoundExpr::Await {
                    expr: Box::new(bound_expr),
                    span,
                })
            }
            
            Expr::Yield { expr, span } => {
                let bound_expr = if let Some(expr) = expr {
                    Some(Box::new(self.bind_expression(*expr)?))
                } else {
                    None
                };
                Ok(BoundExpr::Yield {
                    expr: bound_expr,
                    span,
                })
            }
            
            Expr::Interpolated { parts, span } => {
                // String interpolation binding - şimdilik basit implementation
                Ok(BoundExpr::Lit(crate::compiler::parser::ast::Literal::String("interpolated".to_string())))
            }
            
            Expr::Range { start, end, span } => {
                // Range expression binding
                let bound_start = self.bind_expression(*start)?;
                let bound_end = self.bind_expression(*end)?;
                Ok(BoundExpr::Range {
                    start: Box::new(bound_start),
                    end: Box::new(bound_end),
                    span,
                })
            }
            
            Expr::Ternary { condition, true_expr, false_expr, span } => {
                // Ternary expression binding - şimdilik basit implementation
                let bound_condition = self.bind_expression(*condition)?;
                let bound_true = self.bind_expression(*true_expr)?;
                let bound_false = self.bind_expression(*false_expr)?;
                Ok(BoundExpr::Lit(crate::compiler::parser::ast::Literal::String("ternary".to_string())))
            }
        }
    }

    fn declare_variable(&mut self, name: &str, ty: &TypeDecl, span: Span) -> Result<SymbolInfo, Diagnostic> {
        if self.scopes.current().contains(name) {
            return Err(Diagnostic::error(
                "E0002",
                &format!("Variable '{}' already declared in this scope", name),
                span,
            ));
        }

        // Global scope'ta ise Global slot kullan, değilse Local slot kullan
        let slot = if self.scopes.is_global_scope() {
            Slot::Global(self.next_slot)
        } else {
            Slot::Local(self.next_slot)
        };
        self.next_slot += 1;

        let symbol = SymbolInfo::new(
            name.to_string(),
            SymbolKind::Variable,
            Some(ty.clone()),
            slot,
            span,
        );

        self.scopes.current_mut().insert(name.to_string(), symbol.clone());
        self.symbols.push(symbol.clone());
        
        Ok(symbol)
    }

    fn declare_parameter(&mut self, name: &str, ty: &TypeDecl, span: Span, slot_index: u16) -> Result<SymbolInfo, Diagnostic> {
        if self.scopes.current().contains(name) {
            return Err(Diagnostic::error(
                "E0002",
                &format!("Parameter '{}' already declared in this scope", name),
                span,
            ));
        }

        let slot = Slot::Local(slot_index as u32);

        let symbol = SymbolInfo::new(
            name.to_string(),
            SymbolKind::Variable,
            Some(ty.clone()),
            slot,
            span,
        );

        self.scopes.current_mut().insert(name.to_string(), symbol.clone());
        self.symbols.push(symbol.clone());
        
        Ok(symbol)
    }

    fn declare_function(&mut self, name: &str, kind: &crate::compiler::parser::ast::FuncKind, span: Span) -> Result<SymbolInfo, Diagnostic> {
        if self.scopes.current().contains(name) {
            return Err(Diagnostic::error(
                "E0003",
                &format!("Function '{}' already declared in this scope", name),
                span,
            ));
        }

        let slot = Slot::Global(self.next_slot);
        self.next_slot += 1;

        let symbol = SymbolInfo::new(
            name.to_string(),
            SymbolKind::Function,
            None, // Fonksiyonların dönüş tipi MVP'de belirtilmiyor
            slot,
            span,
        );

        self.scopes.current_mut().insert(name.to_string(), symbol.clone());
        self.symbols.push(symbol.clone());
        
        Ok(symbol)
    }

    fn resolve_variable(&mut self, name: &str, span: Span) -> Result<SymbolInfo, Diagnostic> {
        if let Some(symbol) = self.scopes.resolve(name) {
            // Visibility kontrolü - private symbol'lere erişim kontrolü
            // Şimdilik sadece temel kontrol, import sistemi olunca genişletilecek
            if symbol.is_private() && !self.is_in_same_module(&symbol) {
                return Err(Diagnostic::error(
                    crate::compiler::diag::codes::PRIVATE_SYMBOL_ACCESS,
                    &format!("Cannot access private symbol '{}'", name),
                    span,
                ));
            }
            Ok(symbol.clone())
        } else {
            Err(Diagnostic::error(
                "E0004",
                &format!("Undefined variable '{}'", name),
                span,
            ))
        }
    }
    
    /// Aynı modül içinde olup olmadığını kontrol et (şimdilik her zaman true)
    fn is_in_same_module(&self, _symbol: &SymbolInfo) -> bool {
        // Import sistemi olmadığı için şimdilik her zaman true
        true
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::compiler::parser::{parse, ast::Ast};
    use crate::compiler::lexer::lex;

    #[test]
    fn test_variable_binding() {
        let tokens = lex("int x = 42").unwrap();
        let ast = parse(tokens).unwrap();
        let bound_ast = bind(ast).unwrap();
        
        assert_eq!(bound_ast.statements.len(), 1);
        if let BoundStmt::VarDecl { name, symbol, .. } = &bound_ast.statements[0] {
            assert_eq!(name, "x");
            assert_eq!(symbol.kind, SymbolKind::Variable);
        } else {
            panic!("Expected VarDecl");
        }
    }

    #[test]
    fn test_function_binding() {
        let tokens = lex("function test(x: int)\n  return x").unwrap();
        let ast = parse(tokens).unwrap();
        let bound_ast = bind(ast).unwrap();
        
        assert_eq!(bound_ast.statements.len(), 1);
        if let BoundStmt::Func { name, symbol, .. } = &bound_ast.statements[0] {
            assert_eq!(name, "test");
            assert_eq!(symbol.kind, SymbolKind::Function);
        } else {
            panic!("Expected Func");
        }
    }

    #[test]
    fn test_undefined_variable() {
        let tokens = lex("print(x)").unwrap();
        let ast = parse(tokens).unwrap();
        let result = bind(ast);
        assert!(result.is_err());
        if let Err(diag) = result {
            assert_eq!(diag.code, "E0004");
        }
    }

    #[test]
    fn test_variable_shadowing() {
        let tokens = lex("int x = 1\nif true\n  int x = 2").unwrap();
        let ast = parse(tokens).unwrap();
        let result = bind(ast);
        // Bu test başarılı olmalı - farklı scope'larda aynı isim kullanılabilir
        assert!(result.is_ok());
    }
}
