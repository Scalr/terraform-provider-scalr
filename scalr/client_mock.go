package scalr

import (
	"context"

	"github.com/scalr/go-scalr"
)

type workspaceNamesKey struct {
	environment, workspace string
}

type mockWorkspaces struct {
	workspaceNames map[workspaceNamesKey]*scalr.Workspace
}

type mockVariables struct {
	ids map[string]*scalr.Variable
}

func newMockWorkspaces() *mockWorkspaces {
	return &mockWorkspaces{
		workspaceNames: make(map[workspaceNamesKey]*scalr.Workspace),
	}
}

func newMockVariables() *mockVariables {
	return &mockVariables{
		ids: make(map[string]*scalr.Variable),
	}
}

func (m *mockWorkspaces) List(_ context.Context, _ scalr.WorkspaceListOptions) (*scalr.WorkspaceList, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Create(_ context.Context, options scalr.WorkspaceCreateOptions) (*scalr.Workspace, error) {
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

func (m *mockVariables) Create(_ context.Context, options scalr.VariableCreateOptions) (*scalr.Variable, error) {
	variable := &scalr.Variable{
		ID: options.ID,
	}

	m.ids[options.ID] = variable

	return variable, nil
}

func (m *mockVariables) Read(_ context.Context, varID string) (*scalr.Variable, error) {
	v := m.ids[varID]

	if v == nil {
		return nil, scalr.ErrResourceNotFound
	}

	return v, nil
}

func (m *mockVariables) List(_ context.Context, _ scalr.VariableListOptions) (*scalr.VariableList, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Read(_ context.Context, environment string, workspace string) (*scalr.Workspace, error) {
	w := m.workspaceNames[workspaceNamesKey{environment, workspace}]
	if w == nil {
		return nil, scalr.ErrResourceNotFound
	}

	return w, nil
}

func (m *mockWorkspaces) ReadByID(_ context.Context, _ string) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Update(_ context.Context, _ string, _ scalr.WorkspaceUpdateOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Delete(_ context.Context, _ string) error {
	panic("not implemented")
}

func (m *mockWorkspaces) SetSchedule(_ context.Context, _ string, _ scalr.WorkspaceRunScheduleOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockVariables) Update(_ context.Context, _ string, _ scalr.VariableUpdateOptions) (*scalr.Variable, error) {
	panic("not implemented")
}

func (m *mockVariables) Delete(_ context.Context, _ string) error {
	panic("not implemented")
}
