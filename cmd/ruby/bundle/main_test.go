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
	os"
	"path/filepath"
	"testing"

	"github.com/GoogleCloudPlatform/buildpacks/internal/buildpacktest"
	"github.com/GoogleCloudPlatform/buildpacks/internal/mockprocess"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
	"github.com/google/go-cmp/cmp"
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
		name              string
		rubyVersion       string
		gemfile           string
		gemfileLock       string
		expectedGemfile   string
		mockBin           []*mockprocess.Mock
		expectedExit      int
		expectedInstalled []string
	}{
		{
			name:        "ruby 3.4 with bundled gems",
			rubyVersion: "ruby 3.4.0p0 (2025-03-28 revision 12345) [x86_64-linux]",
			gemfile:     `source "https://rubygems.org"`,
			gemfileLock: `
GEM
  remote: https://rubygems.org/
  specs:

PLATFORMS
  ruby

DEPENDENCIES

BUNDLED WITH
   2.5.8`,
			expectedGemfile: `source "https://rubygems.org"
gem "csv"
gem "bigdecimal"
gem "base64"
gem "drb"
gem "getoptlong"
gem "mutex_m"
gem "nkf"
gem "observer"
gem "resolv-replace"
gem "rinda"
gem "syslog"`,
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
			expectedInstalled: []string{"csv", "bigdecimal", "base64", "drb", "getoptlong", "mutex_m", "nkf", "observer", "resolv-replace", "rinda", "syslog"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buildpacktest.TestBuild(t, buildFn,
				buildpacktest.WithRubyVersion(tc.rubyVersion),
				buildpacktest.WithMocks(tc.mockBin...),
				buildpacktest.WithFiles(map[string]string{
					"Gemfile":      tc.gemfile,
					"Gemfile.lock": tc.gemfileLock,
				}),
				buildpacktest.WithExpectedExit(tc.expectedExit),
				buildpacktest.WithVerify(func(t *testing.T, i *buildpacktest.Images) {
					gemfile := i.WorkspaceFile("Gemfile")
					if diff := cmp.Diff(tc.expectedGemfile, gemfile); diff != "" {
						t.Errorf("Unexpected Gemfile (-want +got):\n%s", diff)
					}
				}),
			)
		})
	}
}