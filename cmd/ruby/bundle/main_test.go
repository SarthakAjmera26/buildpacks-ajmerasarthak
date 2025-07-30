// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/buildpacktest"
	"github.com/GoogleCloudPlatform/buildpacks/internal/mockprocess"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
)

func TestDetect(t *testing.T) {
	testCases := []struct {
		name  string
		files map[string]string
		want  int
	}{
		{
			name: "with Gemfile",
			files: map[string]string{
				"Gemfile": "",
			},
			want: 0,
		},
		{
			name: "with gems.rb",
			files: map[string]string{
				"gems.rb": "",
			},
			want: 0,
		},
		{
			name:  "without Gemfile or gems.rb",
			files: map[string]string{},
			want:  100,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buildpacktest.TestDetect(t, detectFn, tc.name, tc.files, []string{}, tc.want)
		})
	}
}

func TestBuild(t *testing.T) {
	testCases := []struct {
		name         string
		app          string
		mockBin      []*mockprocess.Mock
		wantExit     int
		wantCommands []*mockprocess.WantCommand
		files        map[string]string
	}{
		{
			name: "good bundle",
			app:  "testdata/good_bundle",
			mockBin: []*mockprocess.Mock{
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "without", "development test"),
					mockprocess.WithStdout("bundle config without"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "path", ".bundle/gems"),
					mockprocess.WithStdout("bundle config path"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("lock", "--add-platform", "x86_64-linux"),
					mockprocess.WithStdout("bundle lock --add-platform"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("lock", "--add-platform", "ruby"),
					mockprocess.WithStdout("bundle lock --add-platform"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "deployment", "true"),
					mockprocess.WithStdout("bundle config deployment"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "frozen", "true"),
					mockprocess.WithStdout("bundle config frozen"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "without", "development test"),
					mockprocess.WithStdout("bundle config without"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "path", ".bundle/gems"),
					mockprocess.WithStdout("bundle config path"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("install"),
					mockprocess.WithStdout("bundle install"),
				),
			},
			wantCommands: []*mockprocess.WantCommand{
				mockprocess.Want("bundle", "config", "--local", "without", "development test"),
				mockprocess.Want("bundle", "config", "--local", "path", ".bundle/gems"),
				mockprocess.Want("bundle", "lock", "--add-platform", "x86_64-linux"),
				mockprocess.Want("bundle", "lock", "--add-platform", "ruby"),
				mockprocess.Want("bundle", "config", "--local", "deployment", "true"),
				mockprocess.Want("bundle", "config", "--local", "frozen", "true"),
				mockprocess.Want("bundle", "config", "--local", "without", "development test"),
				mockprocess.Want("bundle", "config", "--local", "path", ".bundle/gems"),
				mockprocess.Want("bundle", "install"),
			},
		},
		{
			name: "bundle install fails",
			app:  "testdata/good_bundle",
			mockBin: []*mockprocess.Mock{
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "without", "development test"),
					mockprocess.WithStdout("bundle config without"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "path", ".bundle/gems"),
					mockprocess.WithStdout("bundle config path"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("lock", "--add-platform", "x86_64-linux"),
					mockprocess.WithStdout("bundle lock --add-platform"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("lock", "--add-platform", "ruby"),
					mockprocess.WithStdout("bundle lock --add-platform"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "deployment", "true"),
					mockprocess.WithStdout("bundle config deployment"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "frozen", "true"),
					mockprocess.WithStdout("bundle config frozen"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "without", "development test"),
					mockprocess.WithStdout("bundle config without"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "path", ".bundle/gems"),
					mockprocess.WithStdout("bundle config path"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("install"),
					mockprocess.WithStderr("bundle install failed"),
					mockprocess.WithExitCode(1),
				),
			},
			wantExit: 1,
		},
		{
			name: "bundle with gems.rb",
			files: map[string]string{
				"gems.rb":      "",
				"gems.locked":  "",
				"main.rb":      "",
				"another.rb":   "",
				"lib/mylib.rb": "",
			},
			mockBin: []*mockprocess.Mock{
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "without", "development test"),
					mockprocess.WithStdout("bundle config without"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "path", ".bundle/gems"),
					mockprocess.WithStdout("bundle config path"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("lock", "--add-platform", "x86_64-linux"),
					mockprocess.WithStdout("bundle lock --add-platform"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("lock", "--add-platform", "ruby"),
					mockprocess.WithStdout("bundle lock --add-platform"),
				),
				mockprocess.New("bundle",
					mockprocess.WithArgs("config", "--local", "deployment", "true"),
					mockprocess.WithStdout("bundle config deployment"),
				),
				mockprocess.New("bundle",
					mockproce