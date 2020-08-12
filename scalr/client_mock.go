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

func (m *mockWorkspaces) List(ctx context.Context, environment string, options scalr.WorkspaceListOptions) (*scalr.WorkspaceList, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Create(ctx context.Context, environment string, options scalr.WorkspaceCreateOptions) (*scalr.Workspace, error) {
	ws := &scalr.Workspace{
		ID:   options.ID,
		Name: *options.Name,
		Organization: &scalr.Organization{
			Name: environment,
		},
	}

	m.workspaceNames[workspaceNamesKey{environment, *options.Name}] = ws

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

func (m *mockWorkspaces) Update(ctx context.Context, environment string, workspace string, options scalr.WorkspaceUpdateOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) UpdateByID(ctx context.Context, workspaceID string, options scalr.WorkspaceUpdateOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Delete(ctx context.Context, environment string, workspace string) error {
	panic("not implemented")
}

func (m *mockWorkspaces) DeleteByID(ctx context.Context, workspaceID string) error {
	panic("not implemented")
}

func (m *mockWorkspaces) RemoveVCSConnection(ctx context.Context, environment string, workspace string) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) RemoveVCSConnectionByID(ctx context.Context, workspaceID string) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Lock(ctx context.Context, workspaceID string, options scalr.WorkspaceLockOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) Unlock(ctx context.Context, workspaceID string) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) ForceUnlock(ctx context.Context, workspaceID string) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) AssignSSHKey(ctx context.Context, workspaceID string, options scalr.WorkspaceAssignSSHKeyOptions) (*scalr.Workspace, error) {
	panic("not implemented")
}

func (m *mockWorkspaces) UnassignSSHKey(ctx context.Context, workspaceID string) (*scalr.Workspace, error) {
	panic("not implemented")
}
