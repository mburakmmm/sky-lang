// Value System - Değer Sistemi
// VM'de çalışan tüm değer türlerini tanımlar

use std::collections::HashMap;
use std::rc::Rc;
use crate::compiler::types::runtime::ValueKind;
use super::RuntimeError;
use super::call::CallFrame;

/// VM değer türleri
#[derive(Debug, Clone, PartialEq)]
pub enum Value {
    Int(i64),
    Float(f64),
    Bool(bool),
    String(String),
    List(Vec<Value>),
    Map(HashMap<String, Value>),
    Function(FunctionValue),
    NativeFn(NativeFunction),
    Future(FutureValue),
    Coroutine(CoroutineValue),
    Range {
        start: i64,
        end: i64,
    },
    Null,
}

unsafe impl Send for Value {}
unsafe impl Sync for Value {}

/// Fonksiyon değeri
#[derive(Debug, Clone, PartialEq)]
pub struct FunctionValue {
    pub chunk_index: usize,
    pub name: String,
    pub arity: u8,
}

/// Native fonksiyon
#[derive(Clone)]
pub struct NativeFunction {
    pub name: String,
    pub arity: u8,
    pub func: Rc<dyn Fn(&[Value]) -> Result<Value, RuntimeError>>,
}

impl std::fmt::Debug for NativeFunction {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("NativeFunction")
            .field("name", &self.name)
            .field("arity", &self.arity)
            .field("func", &"<function>")
            .finish()
    }
}

// NativeFunction şimdi Rc kullanarak clone edilebilir

impl PartialEq for NativeFunction {
    fn eq(&self, other: &Self) -> bool {
        self.name == other.name && self.arity == other.arity
    }
}

/// Future değeri (async)
#[derive(Debug, Clone, PartialEq)]
pub struct FutureValue {
    pub id: u64,
    pub resolved: bool,
    pub value: Option<Box<Value>>,
    pub operation_type: String,
}

impl FutureValue {
    pub fn new(id: u64, operation_type: String) -> Self {
        Self {
            id,
            resolved: false,
            value: None,
            operation_type,
        }
    }

    pub fn is_completed(&self) -> bool {
        self.resolved
    }

    pub fn get_result(&self) -> Value {
        self.value.as_ref().map(|v| (**v).clone()).unwrap_or(Value::Null)
    }

    pub fn mark_completed(&mut self, value: Value) {
        self.resolved = true;
        self.value = Some(Box::new(value));
    }

    pub fn resolve(&mut self, value: Value) {
        self.resolved = true;
        self.value = Some(Box::new(value));
    }
}

/// Coroutine değeri
#[derive(Debug, Clone, PartialEq)]
pub struct CoroutineValue {
    pub id: u64,
    pub state: CoroutineState,
    pub frame: Option<CallFrame>,
}

impl CoroutineValue {
    pub fn new(id: u64) -> Self {
        Self {
            id,
            state: CoroutineState::Suspended,
            frame: None,
        }
    }

    pub fn resume(&mut self, _args: &[Value]) -> Result<Value, RuntimeError> {
        // Basit implementation - gerçekte frame state'i restore etmeli
        match self.state {
            CoroutineState::Suspended => {
                self.state = CoroutineState::Running;
                Ok(Value::Int(42)) // Placeholder return value
            },
            CoroutineState::Running => {
                self.state = CoroutineState::Done;
                Ok(Value::Int(43)) // Placeholder return value
            },
            CoroutineState::Done => {
                Err(RuntimeError::InvalidOperation {
                    op: "Cannot resume completed coroutine".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                })
            }
        }
    }

    pub fn is_done(&self) -> bool {
        self.state == CoroutineState::Done
    }
}

/// Coroutine durumları
#[derive(Debug, Clone, PartialEq)]
pub enum CoroutineState {
    Suspended,
    Running,
    Done,
}

impl Value {
    /// Değerin türünü döndür
    pub fn kind(&self) -> ValueKind {
        match self {
            Self::Int(_) => ValueKind::Int,
            Self::Float(_) => ValueKind::Float,
            Self::Bool(_) => ValueKind::Bool,
            Self::String(_) => ValueKind::String,
            Self::List(_) => ValueKind::List,
            Self::Map(_) => ValueKind::Map,
            Self::Function(_) => ValueKind::Function,
            Self::NativeFn(_) => ValueKind::NativeFn,
            Self::Future(_) => ValueKind::Future,
            Self::Coroutine(_) => ValueKind::Coroutine,
            Self::Range { .. } => ValueKind::Range,
            Self::Null => ValueKind::Null,
        }
    }

    /// Değerin türünü döndür (kind() ile aynı)
    pub fn get_type(&self) -> ValueKind {
        self.kind()
    }

    /// Değerin truthy olup olmadığını kontrol et
    pub fn is_truthy(&self) -> bool {
        match self {
            Self::Bool(false) | Self::Null => false,
            Self::Int(0) | Self::Float(0.0) => false,
            Self::String(s) => !s.is_empty(),
            Self::List(l) => !l.is_empty(),
            Self::Map(m) => !m.is_empty(),
            _ => true,
        }
    }

    /// Değeri string'e dönüştür
    pub fn to_string(&self) -> String {
        match self {
            Self::Int(i) => i.to_string(),
            Self::Float(f) => {
                if f.fract() == 0.0 {
                    format!("{:.0}", f)
                } else {
                    f.to_string()
                }
            }
            Self::Bool(b) => b.to_string(),
            Self::String(s) => s.clone(),
            Self::List(l) => {
                let elements: Vec<String> = l.iter().map(|v| v.to_string()).collect();
                format!("[{}]", elements.join(", "))
            }
            Self::Map(m) => {
                let pairs: Vec<String> = m.iter()
                    .map(|(k, v)| format!("{}: {}", k, v.to_string()))
                    .collect();
                format!("{{{}}}", pairs.join(", "))
            }
            Self::Function(f) => format!("<function:{}>", f.name),
            Self::NativeFn(f) => format!("<native:{}>", f.name),
            Self::Future(f) => format!("<future:{}>", f.id),
            Self::Coroutine(c) => format!("<coroutine:{}>", c.id),
            Self::Range { start, end } => format!("{}..{}", start, end),
            Self::Null => "null".to_string(),
        }
    }
    
    /// Değeri int olarak döndür (eğer mümkünse)
    pub fn as_int(&self) -> Result<i64, RuntimeError> {
        match self {
            Self::Int(i) => Ok(*i),
            Self::Float(f) => Ok(*f as i64),
            _ => Err(RuntimeError::TypeError {
                expected: "int".to_string(),
                found: self.kind().to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }
    
    /// Değeri list olarak döndür (eğer mümkünse)
    pub fn as_list(&self) -> Result<&Vec<Self>, RuntimeError> {
        match self {
            Self::List(list) => Ok(list),
            _ => Err(RuntimeError::TypeError {
                expected: "list".to_string(),
                found: self.kind().to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }
    
    /// Değeri bool olarak döndür (eğer mümkünse)
    pub fn as_bool(&self) -> Result<bool, RuntimeError> {
        match self {
            Self::Bool(b) => Ok(*b),
            _ => Err(RuntimeError::TypeError {
                expected: "bool".to_string(),
                found: self.kind().to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    /// Değerin equal olup olmadığını kontrol et
    pub fn is_equal(&self, other: &Value) -> bool {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => a == b,
            (Self::Float(a), Self::Float(b)) => (a - b).abs() < f64::EPSILON,
            (Self::Bool(a), Self::Bool(b)) => a == b,
            (Self::String(a), Self::String(b)) => a == b,
            (Self::Null, Self::Null) => true,
            _ => false,
        }
    }

    /// Equal işlemi (binary_op için)
    pub fn equal(&self, other: &Value) -> Result<Value, RuntimeError> {
        Ok(Value::Bool(self.is_equal(other)))
    }

    /// Not equal işlemi (binary_op için)
    pub fn not_equal(&self, other: &Value) -> Result<Value, RuntimeError> {
        Ok(Value::Bool(!self.is_equal(other)))
    }

    /// Less equal işlemi (binary_op için)
    pub fn less_equal(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => Ok(Self::Bool(a <= b)),
            (Self::Float(a), Self::Float(b)) => Ok(Self::Bool(a <= b)),
            _ => Err(RuntimeError::InvalidOperation {
                op: "less_equal".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    /// Greater equal işlemi (binary_op için)
    pub fn greater_equal(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => Ok(Self::Bool(a >= b)),
            (Self::Float(a), Self::Float(b)) => Ok(Self::Bool(a >= b)),
            _ => Err(RuntimeError::InvalidOperation {
                op: "greater_equal".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    /// And işlemi (binary_op için)
    pub fn and(&self, other: &Value) -> Result<Value, RuntimeError> {
        self.and_op(other)
    }

    /// Or işlemi (binary_op için)
    pub fn or(&self, other: &Value) -> Result<Value, RuntimeError> {
        self.or_op(other)
    }

    /// Not işlemi (unary_op için)
    pub fn not(&self) -> Result<Value, RuntimeError> {
        self.not_op()
    }

    /// Concat işlemi (binary_op için)
    pub fn concat(&self, other: &Value) -> Result<Value, RuntimeError> {
        self.add(other) // String concatenation için add kullan
    }

    /// To string işlemi (unary_op için)
    pub fn to_string_op(&self) -> Result<Value, RuntimeError> {
        Ok(Value::String(self.to_string()))
    }

    /// Aritmetik işlemler
    pub fn add(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => Ok(Self::Int(a + b)),
            (Self::Float(a), Self::Float(b)) => Ok(Self::Float(a + b)),
            (Self::Int(a), Self::Float(b)) => Ok(Self::Float(*a as f64 + b)),
            (Self::Float(a), Self::Int(b)) => Ok(Self::Float(a + *b as f64)),
            (Self::String(a), Self::String(b)) => Ok(Self::String(a.clone() + b)),
            (Self::String(a), Self::Int(b)) => Ok(Self::String(a.clone() + &b.to_string())),
            (Self::String(a), Self::Float(b)) => Ok(Self::String(a.clone() + &b.to_string())),
            (Self::String(a), Self::Bool(b)) => Ok(Self::String(a.clone() + &b.to_string())),
            (Self::Int(a), Self::String(b)) => Ok(Self::String(a.to_string() + b)),
            (Self::Float(a), Self::String(b)) => Ok(Self::String(a.to_string() + b)),
            (Self::Bool(a), Self::String(b)) => Ok(Self::String(a.to_string() + b)),
            _ => Err(RuntimeError::InvalidOperation {
                op: "add".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    pub fn sub(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => Ok(Self::Int(a - b)),
            (Self::Float(a), Self::Float(b)) => Ok(Self::Float(a - b)),
            (Self::Int(a), Self::Float(b)) => Ok(Self::Float(*a as f64 - b)),
            (Self::Float(a), Self::Int(b)) => Ok(Self::Float(a - *b as f64)),
            _ => Err(RuntimeError::InvalidOperation {
                op: "sub".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    pub fn mul(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => Ok(Self::Int(a * b)),
            (Self::Float(a), Self::Float(b)) => Ok(Self::Float(a * b)),
            (Self::Int(a), Self::Float(b)) => Ok(Self::Float(*a as f64 * b)),
            (Self::Float(a), Self::Int(b)) => Ok(Self::Float(a * *b as f64)),
            _ => Err(RuntimeError::InvalidOperation {
                op: "mul".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    pub fn div(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => {
                if *b == 0 {
                    Err(RuntimeError::DivisionByZero {
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    })
                } else {
                    Ok(Self::Float(*a as f64 / *b as f64))
                }
            }
            (Self::Float(a), Self::Float(b)) => {
                if *b == 0.0 {
                    Err(RuntimeError::DivisionByZero {
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    })
                } else {
                    Ok(Self::Float(a / b))
                }
            }
            _ => Err(RuntimeError::InvalidOperation {
                op: "div".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    pub fn mod_op(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => {
                if *b == 0 {
                    Err(RuntimeError::DivisionByZero {
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    })
                } else {
                    Ok(Self::Int(a % b))
                }
            }
            _ => Err(RuntimeError::InvalidOperation {
                op: "mod".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    /// Karşılaştırma işlemleri
    pub fn less(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => Ok(Self::Bool(a < b)),
            (Self::Float(a), Self::Float(b)) => Ok(Self::Bool(a < b)),
            _ => Err(RuntimeError::InvalidOperation {
                op: "less".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    pub fn greater(&self, other: &Value) -> Result<Value, RuntimeError> {
        match (self, other) {
            (Self::Int(a), Self::Int(b)) => Ok(Self::Bool(a > b)),
            (Self::Float(a), Self::Float(b)) => Ok(Self::Bool(a > b)),
            _ => Err(RuntimeError::InvalidOperation {
                op: "greater".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    /// Mantıksal işlemler
    pub fn and_op(&self, other: &Value) -> Result<Value, RuntimeError> {
        Ok(Self::Bool(self.is_truthy() && other.is_truthy()))
    }

    pub fn or_op(&self, other: &Value) -> Result<Value, RuntimeError> {
        Ok(Self::Bool(self.is_truthy() || other.is_truthy()))
    }

    pub fn not_op(&self) -> Result<Value, RuntimeError> {
        Ok(Self::Bool(!self.is_truthy()))
    }

    /// Negate işlemi
    pub fn negate(&self) -> Result<Value, RuntimeError> {
        match self {
            Self::Int(i) => Ok(Self::Int(-i)),
            Self::Float(f) => Ok(Self::Float(-f)),
            _ => Err(RuntimeError::InvalidOperation {
                op: "negate".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }
}

impl FunctionValue {
    pub fn new(chunk_index: usize, name: String, arity: u8) -> Self {
        Self {
            chunk_index,
            name,
            arity,
        }
    }
}

impl NativeFunction {
    pub fn new(name: String, arity: u8, func: Box<dyn Fn(&[Value]) -> Result<Value, RuntimeError>>) -> Self {
        Self { name, arity, func: Rc::from(func) }
    }
}


impl From<String> for Value {
    fn from(s: String) -> Self {
        Value::String(s)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_value_kind() {
        assert_eq!(Value::Int(42).kind(), ValueKind::Int);
        assert_eq!(Value::String("hello".to_string()).kind(), ValueKind::String);
        assert_eq!(Value::Bool(true).kind(), ValueKind::Bool);
    }

    #[test]
    fn test_value_truthiness() {
        assert!(Value::Int(1).is_truthy());
        assert!(!Value::Int(0).is_truthy());
        assert!(Value::Bool(true).is_truthy());
        assert!(!Value::Bool(false).is_truthy());
        assert!(Value::String("hello".to_string()).is_truthy());
        assert!(!Value::String("".to_string()).is_truthy());
        assert!(!Value::Null.is_truthy());
    }

    #[test]
    fn test_arithmetic_operations() {
        let a = Value::Int(10);
        let b = Value::Int(5);
        
        assert_eq!(a.add(&b).unwrap(), Value::Int(15));
        assert_eq!(a.sub(&b).unwrap(), Value::Int(5));
        assert_eq!(a.mul(&b).unwrap(), Value::Int(50));
        assert_eq!(a.div(&b).unwrap(), Value::Float(2.0));
        assert_eq!(a.mod_op(&b).unwrap(), Value::Int(0));
    }

    #[test]
    fn test_string_concatenation() {
        let a = Value::String("hello".to_string());
        let b = Value::String(" world".to_string());
        
        assert_eq!(a.add(&b).unwrap(), Value::String("hello world".to_string()));
    }

    #[test]
    fn test_comparison_operations() {
        let a = Value::Int(5);
        let b = Value::Int(10);
        
        assert_eq!(a.less(&b).unwrap(), Value::Bool(true));
        assert_eq!(a.greater(&b).unwrap(), Value::Bool(false));
    }

    #[test]
    fn test_logical_operations() {
        let a = Value::Bool(true);
        let b = Value::Bool(false);
        
        assert_eq!(a.and_op(&b).unwrap(), Value::Bool(false));
        assert_eq!(a.or_op(&b).unwrap(), Value::Bool(true));
        assert_eq!(a.not_op().unwrap(), Value::Bool(false));
    }

    #[test]
    fn test_division_by_zero() {
        let a = Value::Int(10);
        let b = Value::Int(0);
        
        assert!(a.div(&b).is_err());
        assert!(a.mod_op(&b).is_err());
    }
}
