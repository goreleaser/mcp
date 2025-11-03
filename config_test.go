package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/goreleaser/goreleaser-mcp/internal/yaml"
	"github.com/goreleaser/goreleaser-pro/v2/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestCheckDeprecated(t *testing.T) {
	tests := []struct {
		name string
		proj config.Project
		want []string
	}{
		{
			name: "no deprecated fields",
			proj: config.Project{
				ProjectName: "test",
			},
			want: []string{},
		},
		{
			name: "deprecated ko repository field",
			proj: config.Project{
				ProjectName: "test",
				Kos: []config.Ko{
					{
						ID:         "test",
						Repository: "ghcr.io/owner/repo",
					},
				},
			},
			want: []string{"kos.repository"},
		},
		{
			name: "non-deprecated fields set",
			proj: config.Project{
				ProjectName: "test",
				Builds: []config.Build{
					{
						ID:     "test",
						Binary: "mybinary",
					},
				},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDeprecated(tt.proj)
			if len(got) != len(tt.want) {
				t.Errorf("findDeprecated() returned %d fields, want %d: %v", len(got), len(tt.want), got)
			}
			for _, field := range tt.want {
				if _, ok := got[field]; !ok {
					t.Errorf("findDeprecated() missing expected field: %s", field)
				}
			}
		})
	}
}

func TestCheckDeprecated_Integration(t *testing.T) {
	input := `
project_name: test
kos:
  - id: test-ko
    repository: ghcr.io/owner/repo
`
	var proj config.Project
	if err := yaml.UnmarshalStrict([]byte(input), &proj); err != nil {
		t.Fatalf("unmarshal error = %v", err)
	}

	deprecated := findDeprecated(proj)
	if len(deprecated) == 0 {
		t.Error("findDeprecated() expected to find deprecated fields, got none")
	}

	expectedFields := []string{"kos.repository"}
	for _, field := range expectedFields {
		if _, ok := deprecated[field]; !ok {
			t.Errorf("findDeprecated() missing expected deprecated field: %s", field)
		}
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want bool
	}{
		{
			name: "zero string",
			val:  "",
			want: true,
		},
		{
			name: "non-zero string",
			val:  "test",
			want: false,
		},
		{
			name: "zero int",
			val:  0,
			want: true,
		},
		{
			name: "non-zero int",
			val:  42,
			want: false,
		},
		{
			name: "zero bool",
			val:  false,
			want: true,
		},
		{
			name: "non-zero bool",
			val:  true,
			want: false,
		},
		{
			name: "empty slice",
			val:  []string{},
			want: true,
		},
		{
			name: "non-empty slice",
			val:  []string{"test"},
			want: false,
		},
		{
			name: "nil pointer",
			val:  (*string)(nil),
			want: true,
		},
		{
			name: "non-nil pointer",
			val:  ptrString("test"),
			want: false,
		},
		{
			name: "empty map",
			val:  map[string]string{},
			want: true,
		},
		{
			name: "non-empty map",
			val:  map[string]string{"key": "value"},
			want: false,
		},
		{
			name: "zero uint",
			val:  uint(0),
			want: true,
		},
		{
			name: "non-zero uint",
			val:  uint(42),
			want: false,
		},
		{
			name: "zero float",
			val:  float64(0),
			want: true,
		},
		{
			name: "non-zero float",
			val:  float64(3.14),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.val)
			got := isZero(v)
			if got != tt.want {
				t.Errorf("isZero() = %v, want %v for %T(%v)", got, tt.want, tt.val, tt.val)
			}
		})
	}
}

func TestCheckDeprecatedFields_Nested(t *testing.T) {
	tests := []struct {
		name string
		proj config.Project
		want []string
	}{
		{
			name: "multiple ko items with deprecated repository",
			proj: config.Project{
				ProjectName: "test",
				Kos: []config.Ko{
					{
						ID:         "ko1",
						Repository: "ghcr.io/owner/repo1",
					},
					{
						ID:         "ko2",
						Repository: "ghcr.io/owner/repo2",
					},
				},
			},
			want: []string{"kos.repository"},
		},
		{
			name: "empty slice no deprecated",
			proj: config.Project{
				ProjectName: "test",
				Builds:      []config.Build{},
			},
			want: []string{},
		},
		{
			name: "slice with non-deprecated values",
			proj: config.Project{
				ProjectName: "test",
				Builds: []config.Build{
					{
						ID:     "build1",
						Binary: "test",
					},
				},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findDeprecated(tt.proj)
			if len(got) != len(tt.want) {
				t.Errorf("findDeprecated() returned %d fields, want %d: %v", len(got), len(tt.want), got)
			}
			for _, field := range tt.want {
				if _, ok := got[field]; !ok {
					t.Errorf("findDeprecated() missing expected field: %s", field)
				}
			}
		})
	}
}

func TestParse_ValidatesStructure(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		checkFunc func(config.Project) bool
	}{
		{
			name: "parses project name",
			input: `
project_name: myproject
`,
			checkFunc: func(p config.Project) bool {
				return p.ProjectName == "myproject"
			},
		},
		{
			name: "parses builds",
			input: `
project_name: test
builds:
  - id: build1
    binary: mybinary
`,
			checkFunc: func(p config.Project) bool {
				return len(p.Builds) == 1 && p.Builds[0].ID == "build1"
			},
		},
		{
			name: "parses complex config",
			input: `
project_name: test
version: 2
builds:
  - id: build1
    binary: mybinary
    goos:
      - linux
      - darwin
kos:
  - id: ko1
    repository: ghcr.io/owner/repo
`,
			checkFunc: func(p config.Project) bool {
				return p.ProjectName == "test" && len(p.Builds) == 1 && len(p.Kos) == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var proj config.Project
			if err := yaml.UnmarshalStrict([]byte(tt.input), &proj); err != nil {
				t.Fatalf("unmarshal error = %v", err)
			}
			if !tt.checkFunc(proj) {
				t.Errorf("parse() project validation failed")
			}
		})
	}
}

func TestCheckDeprecated_EmptyProject(t *testing.T) {
	proj := config.Project{}
	deprecated := findDeprecated(proj)
	if len(deprecated) != 0 {
		t.Errorf("findDeprecated() on empty project returned fields: %v", deprecated)
	}
}

func TestCheckDeprecated_NilSlices(t *testing.T) {
	proj := config.Project{
		ProjectName: "test",
		Builds:      nil,
		Kos:         nil,
	}
	deprecated := findDeprecated(proj)
	if len(deprecated) != 0 {
		t.Errorf("findDeprecated() with nil slices returned fields: %v", deprecated)
	}
}

func TestIsZero_Struct(t *testing.T) {
	type testStruct struct {
		Value string
	}

	tests := []struct {
		name string
		val  testStruct
		want bool
	}{
		{
			name: "zero struct",
			val:  testStruct{},
			want: true,
		},
		{
			name: "non-zero struct",
			val:  testStruct{Value: "test"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.val)
			got := isZero(v)
			if got != tt.want {
				t.Errorf("isZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsZero_AdditionalTypes(t *testing.T) {
	tests := []struct {
		name string
		val  any
		want bool
	}{
		{
			name: "zero int8",
			val:  int8(0),
			want: true,
		},
		{
			name: "non-zero int8",
			val:  int8(5),
			want: false,
		},
		{
			name: "zero int16",
			val:  int16(0),
			want: true,
		},
		{
			name: "non-zero int32",
			val:  int32(42),
			want: false,
		},
		{
			name: "zero int64",
			val:  int64(0),
			want: true,
		},
		{
			name: "zero uint8",
			val:  uint8(0),
			want: true,
		},
		{
			name: "zero uint16",
			val:  uint16(0),
			want: true,
		},
		{
			name: "non-zero uint32",
			val:  uint32(42),
			want: false,
		},
		{
			name: "zero uint64",
			val:  uint64(0),
			want: true,
		},
		{
			name: "zero uintptr",
			val:  uintptr(0),
			want: true,
		},
		{
			name: "zero float32",
			val:  float32(0),
			want: true,
		},
		{
			name: "non-zero float32",
			val:  float32(3.14),
			want: false,
		},
		{
			name: "empty array",
			val:  [0]int{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.val)
			got := isZero(v)
			if got != tt.want {
				t.Errorf("isZero() = %v, want %v for %T(%v)", got, tt.want, tt.val, tt.val)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}

func TestParseConfig(t *testing.T) {
	in, err := os.ReadFile("./.goreleaser.yaml")
	require.NoError(t, err)
	var proj config.Project
	require.NoError(t, yaml.UnmarshalStrict(in, &proj))
}
