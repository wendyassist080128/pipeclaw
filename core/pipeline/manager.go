package pipeline

import (
	"encoding/json"
	"os"
	"sync"
)

const pipelineFile = "data/pipeline.json"

type Node struct {
	ID    string `json:"id"`
	Type  string `json:"type"` // input, session, output, llm, tool, etc.
	Label string `json:"label"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Pipeline struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Nodes    []Node `json:"nodes"`
	Edges    []Edge `json:"edges"`
	Deployed bool   `json:"deployed"`
}

type Manager struct {
	mu sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Load() (*Pipeline, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := os.ReadFile(pipelineFile)
	if err != nil {
		return nil, err
	}

	var pipeline Pipeline
	if err := json.Unmarshal(data, &pipeline); err != nil {
		return nil, err
	}

	return &pipeline, nil
}

func (m *Manager) Save(pipeline *Pipeline) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	os.MkdirAll("data", 0755)

	data, err := json.MarshalIndent(pipeline, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pipelineFile, data, 0644)
}

func (m *Manager) Deploy(pipelineID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(pipelineFile)
	if err != nil {
		return err
	}

	var pipeline Pipeline
	if err := json.Unmarshal(data, &pipeline); err != nil {
		return err
	}

	if pipeline.ID != pipelineID {
		return ErrPipelineNotFound
	}

	pipeline.Deployed = true

	newData, err := json.MarshalIndent(pipeline, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pipelineFile, newData, 0644)
}

func (m *Manager) Undeploy(pipelineID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(pipelineFile)
	if err != nil {
		return err
	}

	var pipeline Pipeline
	if err := json.Unmarshal(data, &pipeline); err != nil {
		return err
	}

	if pipeline.ID != pipelineID {
		return ErrPipelineNotFound
	}

	pipeline.Deployed = false

	newData, err := json.MarshalIndent(pipeline, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pipelineFile, newData, 0644)
}

func (m *Manager) IsDeployed(pipelineID string) (bool, error) {
	pipeline, err := m.Load()
	if err != nil {
		return false, err
	}

	return pipeline.Deployed, nil
}

var ErrPipelineNotFound = &Error{Msg: "Pipeline not found"}

type Error struct {
	Msg string
}

func (e *Error) Error() string {
	return e.Msg
}
