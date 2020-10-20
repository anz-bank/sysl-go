package core

import "github.com/anz-bank/sysl-go/status"

var (
	// Build metadata. These values are intended to be overridden by values
	// supplied by the build environment at link time. Values are specified
	// by passing "-X package.Varname=Value" to the linker. For example:
	//
	//  go build -ldflags="-X main.Name=TestApp -X main.Version=1.0" ...etc...
	//
	// Ref: https://golang.org/cmd/link/

	// Name is set at the build time.
	Name = ""

	// Version is set at the build time.
	Version = ""

	// BuildID is set at the build time.
	BuildID = ""

	// CommitSha is set at the build time.
	CommitSha = ""

	// BranchName is set at the build time.
	BranchName = ""

	// TagName is set at the build time.
	TagName = ""
)

var buildMetadata = &status.BuildMetadata{
	Name:       Name,
	Version:    Version,
	BuildID:    BuildID,
	CommitSha:  CommitSha,
	BranchName: BranchName,
	TagName:    TagName,
}
