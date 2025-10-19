package ir

import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
)

// AsyncTransformer async/await ifadelerini state machine'e dönüştürür
type AsyncTransformer struct {
	stateCounter int
	awaitPoints  []*AwaitPoint
}

// AwaitPoint bir await noktası
type AwaitPoint struct {
	stateID     int
	expression  ast.Expression
	resumeLabel string
}

// NewAsyncTransformer yeni bir transformer oluşturur
func NewAsyncTransformer() *AsyncTransformer {
	return &AsyncTransformer{
		stateCounter: 0,
		awaitPoints:  make([]*AwaitPoint, 0),
	}
}

// Transform async fonksiyonu state machine'e dönüştürür
func (at *AsyncTransformer) Transform(fn *ast.FunctionStatement) (*AsyncStateMachine, error) {
	if !fn.Async {
		return nil, fmt.Errorf("function is not async")
	}

	sm := &AsyncStateMachine{
		name:         fn.Name.Value,
		states:       make([]*State, 0),
		currentState: 0,
		parameters:   fn.Parameters,
		returnType:   fn.ReturnType,
	}

	// Fonksiyon body'sini analiz et ve await noktalarını bul
	at.findAwaitPoints(fn.Body)

	// Her await noktası için bir state oluştur
	for i, ap := range at.awaitPoints {
		state := &State{
			id:         i,
			label:      fmt.Sprintf("state_%d", i),
			awaitPoint: ap,
		}
		sm.states = append(sm.states, state)
	}

	// Final state
	sm.states = append(sm.states, &State{
		id:    len(at.awaitPoints),
		label: "state_final",
	})

	return sm, nil
}

// findAwaitPoints fonksiyondaki tüm await noktalarını bulur
func (at *AsyncTransformer) findAwaitPoints(block *ast.BlockStatement) {
	for _, stmt := range block.Statements {
		at.findAwaitInStatement(stmt)
	}
}

// findAwaitInStatement statement içindeki await'leri bulur
func (at *AsyncTransformer) findAwaitInStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		at.findAwaitInExpression(s.Expression)
	case *ast.LetStatement:
		at.findAwaitInExpression(s.Value)
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			at.findAwaitInExpression(s.ReturnValue)
		}
	case *ast.IfStatement:
		at.findAwaitInExpression(s.Condition)
		at.findAwaitPoints(s.Consequence)
		if s.Alternative != nil {
			at.findAwaitPoints(s.Alternative)
		}
	case *ast.WhileStatement:
		at.findAwaitInExpression(s.Condition)
		at.findAwaitPoints(s.Body)
	case *ast.ForStatement:
		at.findAwaitInExpression(s.Iterable)
		at.findAwaitPoints(s.Body)
	case *ast.BlockStatement:
		at.findAwaitPoints(s)
	}
}

// findAwaitInExpression expression içindeki await'leri bulur
func (at *AsyncTransformer) findAwaitInExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.AwaitExpression:
		// Await noktası bulundu!
		ap := &AwaitPoint{
			stateID:     at.stateCounter,
			expression:  e.Expression,
			resumeLabel: fmt.Sprintf("resume_%d", at.stateCounter),
		}
		at.awaitPoints = append(at.awaitPoints, ap)
		at.stateCounter++

	case *ast.InfixExpression:
		at.findAwaitInExpression(e.Left)
		at.findAwaitInExpression(e.Right)

	case *ast.PrefixExpression:
		at.findAwaitInExpression(e.Right)

	case *ast.CallExpression:
		at.findAwaitInExpression(e.Function)
		for _, arg := range e.Arguments {
			at.findAwaitInExpression(arg)
		}

	case *ast.IndexExpression:
		at.findAwaitInExpression(e.Left)
		at.findAwaitInExpression(e.Index)

	case *ast.MemberExpression:
		at.findAwaitInExpression(e.Object)
	}
}

// AsyncStateMachine async fonksiyonun state machine representasyonu
type AsyncStateMachine struct {
	name         string
	states       []*State
	currentState int
	parameters   []*ast.FunctionParameter
	returnType   ast.TypeAnnotation
	context      interface{} // Saved local variables
}

// State bir state machine state'i
type State struct {
	id         int
	label      string
	awaitPoint *AwaitPoint
	code       []ast.Statement
}

// Execute state machine'i çalıştırır
func (sm *AsyncStateMachine) Execute() (interface{}, error) {
	for sm.currentState < len(sm.states) {
		state := sm.states[sm.currentState]

		if state.awaitPoint != nil {
			// Await noktasına gelindi - suspend
			return nil, fmt.Errorf("await suspend at state %d", sm.currentState)
		}

		sm.currentState++
	}

	return nil, nil
}

// Resume state machine'i resume eder
func (sm *AsyncStateMachine) Resume(result interface{}) error {
	if sm.currentState >= len(sm.states) {
		return fmt.Errorf("state machine already completed")
	}

	// Sonucu context'e kaydet
	// Continue execution
	sm.currentState++

	return nil
}

// AsyncContext async fonksiyon execution context
type AsyncContext struct {
	stateMachine *AsyncStateMachine
	locals       map[string]interface{}
	suspended    bool
	result       interface{}
	err          error
}

// NewAsyncContext yeni bir async context oluşturur
func NewAsyncContext(sm *AsyncStateMachine) *AsyncContext {
	return &AsyncContext{
		stateMachine: sm,
		locals:       make(map[string]interface{}),
		suspended:    false,
	}
}

// Suspend execution'ı duraklat
func (ac *AsyncContext) Suspend(awaitExpr ast.Expression) {
	ac.suspended = true
}

// Resume execution'ı devam ettir
func (ac *AsyncContext) Resume(result interface{}) error {
	ac.suspended = false
	return ac.stateMachine.Resume(result)
}

// IsSuspended durduğu kontrol eder
func (ac *AsyncContext) IsSuspended() bool {
	return ac.suspended
}

// SetLocal local değişken ayarlar
func (ac *AsyncContext) SetLocal(name string, value interface{}) {
	ac.locals[name] = value
}

// GetLocal local değişken alır
func (ac *AsyncContext) GetLocal(name string) (interface{}, bool) {
	val, ok := ac.locals[name]
	return val, ok
}
