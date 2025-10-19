package optimizer

import (
	"encoding/json"
	"os"
	"sync"
)

// ProfileData stores profiling information
type ProfileData struct {
	HotFunctions map[string]int64            `json:"hot_functions"`
	HotPaths     map[string]int64            `json:"hot_paths"`
	TypeFeedback map[string]map[string]int64 `json:"type_feedback"`
	BranchProbs  map[string]float64          `json:"branch_probs"`
}

// PGOProfiler collects profile data during execution
type PGOProfiler struct {
	data ProfileData
	mu   sync.Mutex
}

// NewPGOProfiler creates a new PGO profiler
func NewPGOProfiler() *PGOProfiler {
	return &PGOProfiler{
		data: ProfileData{
			HotFunctions: make(map[string]int64),
			HotPaths:     make(map[string]int64),
			TypeFeedback: make(map[string]map[string]int64),
			BranchProbs:  make(map[string]float64),
		},
	}
}

// RecordFunctionCall records a function call
func (p *PGOProfiler) RecordFunctionCall(funcName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data.HotFunctions[funcName]++
}

// RecordPathExecution records a code path execution
func (p *PGOProfiler) RecordPathExecution(pathID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data.HotPaths[pathID]++
}

// RecordTypeFeedback records type information for a site
func (p *PGOProfiler) RecordTypeFeedback(site string, typeName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.data.TypeFeedback[site] == nil {
		p.data.TypeFeedback[site] = make(map[string]int64)
	}
	p.data.TypeFeedback[site][typeName]++
}

// RecordBranch records branch taken/not-taken
func (p *PGOProfiler) RecordBranch(branchID string, taken bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate running probability
	if taken {
		p.data.BranchProbs[branchID] = (p.data.BranchProbs[branchID]*0.9 + 1.0*0.1)
	} else {
		p.data.BranchProbs[branchID] = p.data.BranchProbs[branchID] * 0.9
	}
}

// SaveProfile saves profile data to file
func (p *PGOProfiler) SaveProfile(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(p.data)
}

// LoadProfile loads profile data from file
func LoadProfile(filename string) (*ProfileData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data ProfileData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	return &data, err
}

// GetHotFunctions returns top N hot functions
func (p *PGOProfiler) GetHotFunctions(n int) []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	// TODO: Sort and return top N
	result := []string{}
	for fn := range p.data.HotFunctions {
		result = append(result, fn)
		if len(result) >= n {
			break
		}
	}
	return result
}
