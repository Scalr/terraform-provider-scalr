package scalr

import (
	"context"
	"testing"

	scalr "github.com/scalr/go-scalr"
)

func TestFetchWorkspaceID(t *testing.T) {
	tests := map[string]struct {
		def  string
		want string
		err  bool
	}{
		"non existing environment": {
			"not-an-env/workspace",
			"",
			true,
		},
		"non existing workspace": {
			"hashicorp/not-a-workspace",
			"",
			true,
		},
		"found workspace": {
			"hashicorp/a-workspace",
			"ws-123",
			false,
		},
	}

	client := testScalrClient(t)
	name := "a-workspace"
	client.Workspaces.Create(context.Background(), scalr.WorkspaceCreateOptions{
		ID:          "ws-123",
		Name:        &name,
		Environment: &scalr.Environment{ID: "hashicorp"},
	})

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := fetchWorkspaceID(ctx, test.def, client)

			if (err != nil) != test.err {
				t.Fatalf("expected error is %t, got %v", test.err, err)
			}

			if got != test.want {
				t.Fatalf("wrong result\ngot: %#v\nwant: %#v", got, test.want)
			}
		})
	}
}

func TestUnpackWorkspaceID(t *testing.T) {
	cases := []struct {
		id   string
		env  string
		name string
		err  bool
	}{
		{
			id:   "my-env-name/my-workspace-name",
			env:  "my-env-name",
			name: "my-workspace-name",
			err:  false,
		},
		{
			id:   "my-workspace-name|my-env-name",
			env:  "my-env-name",
			name: "my-workspace-name",
			err:  false,
		},
		{
			id:   "some-invalid-id",
			env:  "",
			name: "",
			err:  true,
		},
	}

	for _, tc := range cases {
		env, name, err := unpackWorkspaceID(tc.id)
		if (err != nil) != tc.err {
			t.Fatalf("expected error is %t, got %v", tc.err, err)
		}

		if tc.env != env {
			t.Fatalf("expected environment ID %q, got %q", tc.env, env)
		}

		if tc.name != name {
			t.Fatalf("expected name %q, got %q", tc.name, name)
		}
	}
}
