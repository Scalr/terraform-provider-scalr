package scalr

import (
	"context"

	scalr "github.com/scalr/go-scalr"
)

type workspaceNamesKey struct {
	environment, workspace string
}

type mockWorkspaces struct {
	workspaceNames map[workspaceNamesKey]*scalr.Workspace
}

func newMockWorkspaces() *mockWorkspaces {
	return &mockWorkspaces{
		workspaceNames: make(map[workspaceNamesKey]*scalr.Workspace),
	}
}

func (m *mockWorkspaces) List(ctx context.Context, options scalr.WorkspaceListOptions) (*scalr.WorkspaceList, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Create(ctx context.Context, options scalr.WorkspaceCreateOptions) (*scalr.Workspace, error) {
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

func (m *mockWorkspaces) Read(ctx context.Context, environment string, workspace string) (*scalr.Workspace, error) {
	w := m.workspaceNames[workspaceNamesKey{environment, workspace}]
	if w == nil {
		return nil, scalr.ErrResourceNotFound
	}

	return w, nil
}

func (m *mockWorkspaces) ReadByID(ctx context.Context, workspaceID string) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Update(ctx context.Context, workspaceID string, options scalr.WorkspaceUpdateOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Delete(ctx context.Context, workspaceID string) error {
	panic("not implemented")
}
