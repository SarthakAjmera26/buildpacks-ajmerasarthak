// Copyright 2023 Google LLC
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

// Implements nodejs/firebaseangular buildpack.
// The nodejs/firebaseangular buildpack does some prep work for angular and runs the build script.
package main

import (
	"github.com/GoogleCloudPlatform/buildpacks/pkg/buildermetadata"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/nodejs"
	"github.com/Masterminds/semver"

	"github.com/GoogleCloudPlatform/buildpacks/pkg/firebase/apphostingschema"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/firebase/faherror"
	"github.com/GoogleCloudPlatform/buildpacks/pkg/firebase/util"

	"github.com/GoogleCloudPlatform/buildpacks/pkg/env"
	gcp "github.com/GoogleCloudPlatform/buildpacks/pkg/gcpbuildpack"
)

const (
	// frameworkVersion is the version of angular that the application is using
	frameworkVersion = "FRAMEWORK_VERSION"
)

var (
	// minAngularVersion is the lowest version of angular supported by the firebase angular buildpack.
	minAngularVersion = semver.MustParse("17.2.0")
)

func main() {
	gcp.Main(detectFn, buildFn)
}

func detectFn(ctx *gcp.Context) (gcp.DetectResult, error) {
	if !env.IsFAH() {
		return gcp.OptOut("not a firebase apphosting application"), nil
	}
	appDir := util.ApplicationDirectory(ctx)
	angularJSONExists, err := ctx.FileExists(appDir, "angular.json")
	if err != nil {
		return nil, err
	}
	if angularJSONExists {
		return gcp.OptInFileFound("angular.json"), nil
	}

	// Some Angular project configurations don't require an angular.json file (e.g. Nx projects).
	// In these cases, we check if the angular core is specified as a dependency.
	nodeDeps, err := nodejs.ReadNodeDependencies(ctx, appDir)
	if err != nil {
		return nil, err
	}
	apphostingSchema, err := apphostingschema.ReadAndValidateFromFile(nodejs.ApphostingPreprocessedPathForPack)
	if err != nil {
		return nil, err
	}
	if nodejs.HasApphostingPackageOrYamlBuild(nodeDeps.PackageJSON, apphostingSchema) {
		return gcp.OptOut("apphosting build script found"), nil
	}

	version, err := nodejs.Version(nodeDeps, "@angular/core")
	if err != nil {
		ctx.Warnf("Error parsing version from lock file, defaulting to package.json version")
		if nodeDeps.PackageJSON.DevDependencies["@angular/core"] != "" {
			return gcp.OptIn("angular dependency found"), nil
		}
		return gcp.OptOut("angular dependency not found"), err
	}
	if version != "" {
		return gcp.OptIn("angular dependency found"), nil
	}
	return gcp.OptOut("angular dependency not found"), nil
}

func buildFn(ctx *gcp.Context) error {
	appDir := util.ApplicationDirectory(ctx)

	nodeDeps, err := nodejs.ReadNodeDependencies(ctx, appDir)
	if err != nil {
		return err
	}
	if nodeDeps.LockfilePath == "" {
		return gcp.UserErrorf("%w", faherror.MissingLockFileError(appDir))
	}

	// Ensure that the right version of the application builder is installed.
	builderVersion, err := nodejs.Version(nodeDeps, "@angular/core")
	if err != nil {
		ctx.Warnf("Error parsing version from lock file, defaulting to package.json version")
		builderVersion = nodeDeps.PackageJSON.DevDependencies["@angular/core"]
	}
	err = validateVersion(ctx, builderVersion)
	if err != nil {
		return err
	}

	// TODO(b/357644160) We should consider adding a validation step to double check that the adapter version works for the framework version.
	if version, exists := nodeDeps.PackageJSON.Dependencies["@apphosting/adapter-angular"]; exists {
		ctx.Logf("*** You already have @apphosting/adapter-angular@%s listed as a dependency, skipping installation ***", version)
		ctx.Logf("*** Your package.json build command will be run as is, please make sure it is set to apphosting-adapter-angular-build if you intend to build your app using the adapter ***")
		return nil
	}

	buildScript, exists := nodeDeps.PackageJSON.Scripts["build"]
	if exists && buildScript != "ng build" && buildScript != "apphosting-adapter-angular-build" {
		ctx.Warnf("*** You are using a custom build command (your build command is NOT 'ng build'), we will accept it as is but will error if output structure is not as expected ***")
	}

	al, err := ctx.Layer("npm_modules", gcp.BuildLayer, gcp.CacheLayer)
	if err != nil {
		return err
	}
	if err = nodejs.InstallAngularBuildAdapter(ctx, al); err != nil {
		return err
	}

	// pass angular version as environment variable that will configure the build for version matching
	al.BuildEnvironment.Override(frameworkVersion, builderVersion)

	// add angular and its version to the builder metadata
	buildermetadata.GlobalBuilderMetadata().SetValue(buildermetadata.FrameworkName, "angular")
	buildermetadata.GlobalBuilderMetadata().SetValue(buildermetadata.FrameworkVersion, buildermetadata.MetadataValue(builderVersion))

	// add the adapter name and the adapter version to the builder metadata
	adapterVersion := ctx.GetMetadata(al, nodejs.AngularVersionKey)
	buildermetadata.GlobalBuilderMetadata().SetValue(buildermetadata.AdapterName, "@apphosting/adapter-angular")
	buildermetadata.GlobalBuilderMetadata().SetValue(buildermetadata.AdapterVersion, buildermetadata.MetadataValue(adapterVersion))

	// This env var indicates to the package manager buildpack that a different command needs to be run
	nodejs.OverrideAngularBuildScript(al)

	return nil
}

func validateVersion(ctx *gcp.Context, depVersion string) error {
	version, err := semver.NewVersion(depVersion)
	// This should only happen in the case of an unexpected lockfile format, i.e. If there is a breaking update to a lock file schema
	if err != nil {
		ctx.Warnf("Unrecognized version of angular: %s", depVersion)
		ctx.Warnf("Consider updating your angular dependencies to >=%s", minAngularVersion.String())
		return nil
	}
	if version.LessThan(minAngularVersion) {
		ctx.Warnf("Update the angular dependencies to >=%s", minAngularVersion.String())
		return gcp.UserErrorf("%w", faherror.UnsupportedFrameworkVersionError("angular", depVersion))
	}
	return nil
}
