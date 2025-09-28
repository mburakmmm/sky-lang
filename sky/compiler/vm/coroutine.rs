// Coroutine - Coop/Yield Semantiği
// coop function yürütümü, yield/resume semantiği

use super::{Value, CallFrame, RuntimeError};
use super::value::{CoroutineValue, CoroutineState};
use crate::compiler::diag::Span;

/// Coroutine yöneticisi
pub struct Coroutine {
    pub id: u64,
    pub state: CoroutineState,
    pub frame: Option<CallFrame>,
    pub yielded_value: Option<Value>,
}

impl Coroutine {
    pub fn new(id: u64) -> Self {
        Self {
            id,
            state: CoroutineState::Suspended,
            frame: None,
            yielded_value: None,
        }
    }

    /// Coroutine'i resume et
    pub fn resume(&mut self, args: &[Value]) -> Result<Value, RuntimeError> {
        if self.state == CoroutineState::Done {
            return Err(RuntimeError::CoroutineFinished {
                span: Span::new(0, 0, 0),
            });
        }

        self.state = CoroutineState::Running;
        
        // Gerçek coroutine execution
        if let Some(frame) = &mut self.frame {
            // Frame state'i restore et - metodları frame üzerinde çağır
            frame.restore_state()?;
            
            // Execution loop - yield noktasına kadar çalıştır
            loop {
                if frame.ip >= frame.chunk.len() {
                    // Function sona erdi
                    self.state = CoroutineState::Done;
                    return Ok(Value::String("finished".to_string()));
                }
                
                let instruction = frame.chunk.get_byte(frame.ip).unwrap_or(0);
                
                match instruction {
                    0x72 => { // YIELD opcode
                        // Yield noktasına ulaştık, frame state'i save et
                        frame.save_state()?;
                        self.state = CoroutineState::Suspended;
                        
                        // Yield edilen değeri al (stack'ten) - basit çözüm
                        return Ok(Value::String("yielded".to_string()));
                    }
                    _ => {
                        // Normal instruction, VM'de execute et
                        // Bu basit bir implementasyon - gerçek VM integration gerekli
                        frame.advance_ip();
                    }
                }
            }
        } else {
            Err(RuntimeError::CoroutineFinished {
                span: Span::new(0, 0, 0),
            })
        }
    }

    /// Coroutine'in tamamlanıp tamamlanmadığını kontrol et
    pub fn is_done(&self) -> bool {
        self.state == CoroutineState::Done
    }
    
    /// Frame state'i save et
    fn save_frame_state(&self, frame: &CallFrame) -> Result<(), RuntimeError> {
        // Stack snapshot, local variables, IP position save et
        // Gerçek implementasyonda:
        // - Stack state'i snapshot al
        // - Local variables'ları kaydet
        // - Instruction pointer'ı kaydet
        // - Frame context'i kaydet
        Ok(())
    }
    
    /// Frame state'i restore et
    fn restore_frame_state(&mut self, frame: &mut CallFrame) -> Result<(), RuntimeError> {
        // Stack snapshot, local variables, IP position restore et
        // Gerçek implementasyonda:
        // - Stack state'i restore et
        // - Local variables'ları restore et
        // - Instruction pointer'ı restore et
        // - Frame context'i restore et
        Ok(())
    }
    
    // Duplicate method removed - using the one below

    /// Coroutine'in suspended olup olmadığını kontrol et
    pub fn is_suspended(&self) -> bool {
        self.state == CoroutineState::Suspended
    }

    /// Coroutine'in running olup olmadığını kontrol et
    pub fn is_running(&self) -> bool {
        self.state == CoroutineState::Running
    }

    /// Coroutine'i sonlandır
    pub fn finish(&mut self) {
        self.state = CoroutineState::Done;
        self.frame = None;
    }

    /// Coroutine'in durumunu string olarak döndür
    pub fn status(&self) -> String {
        match self.state {
            CoroutineState::Suspended => "suspended".to_string(),
            CoroutineState::Running => "running".to_string(),
            CoroutineState::Done => "done".to_string(),
        }
    }

    /// Coroutine'in ID'sini döndür
    pub fn get_id(&self) -> u64 {
        self.id
    }

    /// Yield edilen değeri al
    pub fn get_yielded_value(&self) -> Option<&Value> {
        self.yielded_value.as_ref()
    }

    /// Yield değeri ayarla
    pub fn set_yielded_value(&mut self, value: Value) {
        self.yielded_value = Some(value);
    }

    /// Frame'i ayarla
    pub fn set_frame(&mut self, frame: CallFrame) {
        self.frame = Some(frame);
    }

    /// Frame'i al
    pub fn get_frame(&self) -> Option<&CallFrame> {
        self.frame.as_ref()
    }

    /// Frame'i mutable olarak al
    pub fn get_frame_mut(&mut self) -> Option<&mut CallFrame> {
        self.frame.as_mut()
    }
}

impl From<CoroutineValue> for Coroutine {
    fn from(value: CoroutineValue) -> Self {
        Self {
            id: value.id,
            state: value.state,
            frame: value.frame,
            yielded_value: None,
        }
    }
}

impl From<Coroutine> for CoroutineValue {
    fn from(coroutine: Coroutine) -> Self {
        Self {
            id: coroutine.id,
            state: coroutine.state,
            frame: coroutine.frame,
        }
    }
}

/// Coroutine pool - aktif coroutine'leri yönetir
pub struct CoroutinePool {
    coroutines: std::collections::HashMap<u64, Coroutine>,
    next_id: u64,
}

impl CoroutinePool {
    pub fn new() -> Self {
        Self {
            coroutines: std::collections::HashMap::new(),
            next_id: 1,
        }
    }

    /// Yeni coroutine oluştur
    pub fn create(&mut self) -> u64 {
        let id = self.next_id;
        self.next_id += 1;
        
        let coroutine = Coroutine::new(id);
        self.coroutines.insert(id, coroutine);
        
        id
    }

    /// Coroutine'i al
    pub fn get(&self, id: u64) -> Option<&Coroutine> {
        self.coroutines.get(&id)
    }

    /// Coroutine'i mutable olarak al
    pub fn get_mut(&mut self, id: u64) -> Option<&mut Coroutine> {
        self.coroutines.get_mut(&id)
    }

    /// Coroutine'i kaldır
    pub fn remove(&mut self, id: u64) -> Option<Coroutine> {
        self.coroutines.remove(&id)
    }

    /// Tüm coroutine'leri al
    pub fn all(&self) -> &std::collections::HashMap<u64, Coroutine> {
        &self.coroutines
    }

    /// Aktif coroutine sayısını al
    pub fn count(&self) -> usize {
        self.coroutines.len()
    }

    /// Tamamlanmış coroutine'leri temizle
    pub fn cleanup_finished(&mut self) -> usize {
        let finished_ids: Vec<u64> = self.coroutines
            .iter()
            .filter(|(_, coroutine)| coroutine.is_done())
            .map(|(id, _)| *id)
            .collect();

        let count = finished_ids.len();
        for id in finished_ids {
            self.coroutines.remove(&id);
        }

        count
    }

    /// Tüm coroutine'leri temizle
    pub fn clear(&mut self) {
        self.coroutines.clear();
    }

    /// Suspended coroutine'leri al
    pub fn get_suspended(&self) -> Vec<u64> {
        self.coroutines
            .iter()
            .filter(|(_, coroutine)| coroutine.is_suspended())
            .map(|(id, _)| *id)
            .collect()
    }

    /// Running coroutine'leri al
    pub fn get_running(&self) -> Vec<u64> {
        self.coroutines
            .iter()
            .filter(|(_, coroutine)| coroutine.is_running())
            .map(|(id, _)| *id)
            .collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::compiler::bytecode::Chunk;

    #[test]
    fn test_coroutine_creation() {
        let coroutine = Coroutine::new(1);
        assert_eq!(coroutine.id, 1);
        assert_eq!(coroutine.state, CoroutineState::Suspended);
        assert!(!coroutine.is_done());
        assert!(coroutine.is_suspended());
    }

    #[test]
    fn test_coroutine_resume() {
        let mut coroutine = Coroutine::new(1);
        let chunk = Chunk::new();
        let frame = CallFrame::new(chunk, 0);
        coroutine.set_frame(frame);
        
        let result = coroutine.resume(&[]).unwrap();
        assert_eq!(result, Value::Int(42));
        assert!(coroutine.is_suspended());
    }

    #[test]
    fn test_coroutine_finished_resume() {
        let mut coroutine = Coroutine::new(1);
        coroutine.finish();
        
        let result = coroutine.resume(&[]);
        assert!(result.is_err());
        if let Err(RuntimeError::CoroutineFinished { .. }) = result {
            // Expected error
        } else {
            panic!("Expected CoroutineFinished error");
        }
    }

    #[test]
    fn test_coroutine_pool() {
        let mut pool = CoroutinePool::new();
        
        let id1 = pool.create();
        let id2 = pool.create();
        
        assert_eq!(pool.count(), 2);
        assert!(pool.get(id1).is_some());
        assert!(pool.get(id2).is_some());
        
        pool.remove(id1);
        assert_eq!(pool.count(), 1);
        assert!(pool.get(id1).is_none());
    }

    #[test]
    fn test_coroutine_pool_cleanup() {
        let mut pool = CoroutinePool::new();
        
        let id1 = pool.create();
        let id2 = pool.create();
        
        // Birini tamamla
        if let Some(coroutine) = pool.get_mut(id1) {
            coroutine.finish();
        }
        
        let cleaned = pool.cleanup_finished();
        assert_eq!(cleaned, 1);
        assert_eq!(pool.count(), 1);
        assert!(pool.get(id1).is_none());
        assert!(pool.get(id2).is_some());
    }

    #[test]
    fn test_coroutine_status() {
        let coroutine = Coroutine::new(1);
        assert_eq!(coroutine.status(), "suspended");
        
        let mut coroutine = Coroutine::new(2);
        coroutine.finish();
        assert_eq!(coroutine.status(), "done");
    }

    #[test]
    fn test_coroutine_yielded_value() {
        let mut coroutine = Coroutine::new(1);
        
        assert!(coroutine.get_yielded_value().is_none());
        
        coroutine.set_yielded_value(Value::String("hello".to_string()));
        assert_eq!(coroutine.get_yielded_value(), Some(&Value::String("hello".to_string())));
    }
}
