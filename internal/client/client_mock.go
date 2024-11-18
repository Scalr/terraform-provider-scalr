package client

import (
	"context"

	"github.com/scalr/go-scalr"
)

type workspaceNamesKey struct {
	environment, workspace string
}

type MockWorkspaces struct {
	workspaceNames map[workspaceNamesKey]*scalr.Workspace
}

type MockVariables struct {
	ids map[string]*scalr.Variable
}

func NewMockWorkspaces() *MockWorkspaces {
	return &MockWorkspaces{
		workspaceNames: make(map[workspaceNamesKey]*scalr.Workspace),
	}
}

func NewMockVariables() *MockVariables {
	return &MockVariables{
		ids: make(map[string]*scalr.Variable),
	}
}

func (m *MockWorkspaces) List(_ context.Context, _ scalr.WorkspaceListOptions) (*scalr.WorkspaceList, error) {
	panic("not implemented")
}

func (m *MockWorkspaces) Create(_ context.Context, options scalr.WorkspaceCreateOptions) (*scalr.Workspace, error) {
	ws := &scalr.Workspace{
		ID:   options.ID,
		Name: *options.Name,
		Environment: &scalr.Environment{
			ID: options.Environment.ID,
		},
	}

	m.workspaceNames[workspaceNamesKey{options.Environment.ID, *options.Name}] = ws

	return ws, nil
}

func (m *MockVariables) Create(_ context.Context, options scalr.VariableCreateOptions) (*scalr.Variable, error) {
	variable := &scalr.Variable{
		ID: options.ID,
	}

	m.ids[options.ID] = variable

	return variable, nil
}

func (m *MockVariables) Read(_ context.Context, varID string) (*scalr.Variable, error) {
	v := m.ids[varID]

	if v == nil {
		return nil, scalr.ErrResourceNotFound
	}

	return v, nil
}

func (m *MockVariables) List(_ context.Context, _ scalr.VariableListOptions) (*scalr.VariableList, error) {
	panic("not implemented")
}

func (m *MockWorkspaces) Read(_ context.Context, environment string, workspace string) (*scalr.Workspace, error) {
	w := m.workspaceNames[workspaceNamesKey{environment, workspace}]
	if w == nil {
		return nil, scalr.ErrResourceNotFound
	}

	return w, nil
}

func (m *MockWorkspaces) ReadByID(_ context.Context, _ string) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *MockWorkspaces) Update(_ context.Context, _ string, _ scalr.WorkspaceUpdateOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *MockWorkspaces) Delete(_ context.Context, _ string) error {
	panic("not implemented")
}
func (m *MockWorkspaces) ReadOutputs(_ context.Context, _ string) ([]*scalr.Output, error) {
	panic("not implemented")
}

func (m *MockWorkspaces) SetSchedule(_ context.Context, _ string, _ scalr.WorkspaceRunScheduleOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *MockVariables) Update(_ context.Context, _ string, _ scalr.VariableUpdateOptions) (*scalr.Variable, error) {
	panic("not implemented")
}

func (m *MockVariables) Delete(_ context.Context, _ string) error {
	panic("not implemented")
}
