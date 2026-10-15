// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package npm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPnpmAllowBuildScripts runs allowBuildScripts against a fake pnpm executable that prints a fixed value for
// `pnpm config get` and records the working directory and the arguments of `pnpm config set`. pnpm >= 12 prints
// "null" for an unset setting.
func TestPnpmAllowBuildScripts(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("the fake pnpm executable is a shell script")
	}

	for _, tc := range []struct {
		name    string
		version string
		current string
		// inWorkspace lays the project out as a package of a pnpm workspace (in packages/infra) instead of as a
		// standalone project directory, in which case the config is expected to be written to the workspace root
		// and the depPath to be relative to it.
		inWorkspace bool
		path        string
		expectedSet string
	}{
		{
			name:        "pnpm 12 prints null for an unset allowBuilds",
			version:     "12.3.4",
			current:     "null",
			path:        "./sdk",
			expectedSet: `config set allowBuilds {"foo@file:./sdk":true} --location project --json`,
		},
		{
			name:        "pnpm 11 prints undefined for an unset allowBuilds",
			version:     "11.0.0",
			current:     "undefined",
			path:        "./sdk",
			expectedSet: `config set allowBuilds {"foo@file:./sdk":true} --location project --json`,
		},
		{
			name:        "existing allowBuilds entries are kept",
			version:     "11.0.0",
			current:     `{"bar@file:./bar":true}`,
			path:        "./sdk",
			expectedSet: `config set allowBuilds {"bar@file:./bar":true,"foo@file:./sdk":true} --location project --json`,
		},
		{
			name:        "pnpm 10 prints null for an unset onlyBuiltDependencies",
			version:     "10.34.2",
			current:     "null",
			path:        "./sdk",
			expectedSet: `config set onlyBuiltDependencies ["foo@file:./sdk"] --location project --json`,
		},
		{
			name:        "in a workspace the depPath is relative to the workspace root",
			version:     "11.0.0",
			current:     "undefined",
			inWorkspace: true,
			path:        "sdks/random",
			expectedSet: `config set allowBuilds {"foo@file:packages/infra/sdks/random":true} --location project --json`,
		},
		{
			name:        "in a workspace on pnpm 10 the depPath is relative to the workspace root",
			version:     "10.34.2",
			current:     "null",
			inWorkspace: true,
			path:        "sdks/random",
			expectedSet: `config set onlyBuiltDependencies ["foo@file:packages/infra/sdks/random"] --location project --json`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// t.TempDir can hand out a path with a symlinked prefix (/var on macOS), while `pwd` in the shim
			// prints the resolved one.
			dir, err := filepath.EvalSymlinks(t.TempDir())
			require.NoError(t, err)
			record := filepath.Join(dir, "set-args")
			recordDir := filepath.Join(dir, "set-dir")
			shim := filepath.Join(dir, "pnpm")
			writeFile(t, shim, `#!/bin/sh
case "$1 $2" in
	"config get") echo '`+tc.current+`' ;;
	"config set") printf '%s' "$*" > "`+record+`"; pwd > "`+recordDir+`" ;;
	*) echo "unexpected args: $*" >&2; exit 1 ;;
esac
`)
			require.NoError(t, os.Chmod(shim, 0o700))

			projectDir, expectedDir := dir, dir
			if tc.inWorkspace {
				writeFile(t, filepath.Join(dir, "pnpm-workspace.yaml"), "packages:\n  - \"packages/*\"\n")
				writeFile(t, filepath.Join(dir, "package.json"), `{"name": "root", "private": true}`)
				projectDir = filepath.Join(dir, "packages", "infra")
				writeFile(t, filepath.Join(projectDir, "package.json"), `{"name": "infra"}`)
			}

			pm := &pnpmManager{executable: shim}
			err = pm.allowBuildScripts(t.Context(), semver.MustParse(tc.version), projectDir, "foo", tc.path)
			require.NoError(t, err)

			set, err := os.ReadFile(record)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedSet, string(set))

			setDir, err := os.ReadFile(recordDir)
			require.NoError(t, err)
			assert.Equal(t, expectedDir, strings.TrimSpace(string(setDir)))
		})
	}
}
