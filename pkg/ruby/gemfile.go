// Copyright 2022 Google LLC
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

package ruby

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
	"github.com/Masterminds/semver"
)

// Match against ruby string example: ruby 2.6.7p450
var rubyVersionRe = regexp.MustCompile(`^\s*ruby\s+([^p^\s]+)(p\d+)?\s*// Copyright 2022 Google LLC
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

)

// MaybeAddBundledGems checks the ruby version and adds bundled gems to the Gemfile if needed.
// This is necessary for Ruby 3.4+ where some default gems are no longer part of the standard library.
func MaybeAddBundledGems(ctx *gcp.Context, gemfilePath string) error {
	rubyVersionStr := os.Getenv(RubyVersionKey)
	if rubyVersionStr == "" {
		// If ruby version is not available, we cannot decide whether to add the gems.
		// This should not happen in a normal build process.
		return nil
	}

	rubyVersion, err := semver.NewVersion(rubyVersionStr)
	if err != nil {
		return fmt.Errorf("parsing ruby version %q: %w", rubyVersionStr, err)
	}

	ruby34, _ := semver.NewVersion("3.4.0")
	if rubyVersion.LessThan(ruby34) {
		// No need to add bundled gems for ruby versions older than 3.4.
		return nil
	}

	// List of default gems that became bundled in Ruby 3.4.
	bundledGems := []string{
		"abbrev", "base64", "bigdecimal", "csv", "drb", "English", "fileutils", "find",
		"getoptlong", "logger", "mutex_m", "nkf", "observer", "open-uri", "optparse", "pp",
		"prettyprint", "resolv", "resolv-replace", "rinda", "set", "shellwords", "tempfile",
		"time", "tmpdir", "tsort", "un", "weakref",
	}

	gemfileContent, err := os.ReadFile(gemfilePath)
	if err != nil {
		return fmt.Errorf("reading Gemfile: %w", err)
	}

	gemfileLines := strings.Split(string(gemfileContent), "\n")
	var gemsToAdd []string

	for _, gem := range bundledGems {
		gemFound := false
		for _, line := range gemfileLines {
			// A simple check to see if the gem is already listed.
			// This could be improved with a more robust Gemfile parser.
			if strings.Contains(line, fmt.Sprintf("gem '%s'", gem)) || strings.Contains(line, fmt.Sprintf("gem \"%s\"", gem)) {
				gemFound = true
				break
			}
		}
		if !gemFound {
			gemsToAdd = append(gemsToAdd, gem)
		}
	}

	if len(gemsToAdd) > 0 {
		ctx.Logf("Adding bundled gems for Ruby %s: %s", rubyVersionStr, strings.Join(gemsToAdd, ", "))
		f, err := os.OpenFile(gemfilePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("opening Gemfile for appending: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString("\n# Added by Google Cloud Buildpacks for Ruby 3.4+ compatibility\n"); err != nil {
			return fmt.Errorf("writing comment to Gemfile: %w", err)
		}
		for _, gem := range gemsToAdd {
			if _, err := f.WriteString(fmt.Sprintf("gem '%s'\n", gem)); err != nil {
				return fmt.Errorf("writing gem '%s' to Gemfile: %w", gem, err)
			}
		}
	}

	return nil
}


// ParseRubyVersion extracts the version number from Gemfile.lock or gems.locked, returns an error in
// case the version string is malformed.
func ParseRubyVersion(path string) (string, error) {
	version, err := readLineAfter(path, "RUBY VERSION")
	if err != nil {
		return "", err
	}
	if version == "" {
		return "", nil
	}

	matches := rubyVersionRe.FindStringSubmatch(version)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", gcp.UserErrorf("parsing ruby version %q", version)
}

// ParseBundlerVersion extacts the version of bundler from Gemfile.lock or gems.locked,
// returns an error in case the version string is malformed.
func ParseBundlerVersion(path string) (string, error) {
	version, err := readLineAfter(path, "BUNDLED WITH")
	if err != nil {
		return "", err
	}
	if version == "" {
		return "", nil
	}

	semver, err := semver.NewVersion(strings.TrimSpace(version))
	if err != nil {
		return "", gcp.UserErrorf("parsing bundler version %q: %v", version, err)
	}

	return fmt.Sprintf("%d.%d.%d", semver.Major(), semver.Minor(), semver.Patch()), nil
}

func readLineAfter(path string, token string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == token {
			// Read the next line once the token is found
			if !scanner.Scan() {
				break
			}

			return scanner.Text(), nil
		}
	}

	return "", nil
}
