package main

import (
	"strings"

	"golang.org/x/mod/semver"
)

// Version type constants
const (
	VersionTypeCaret     = "caret"
	VersionTypeTilde     = "tilde"
	VersionTypeGreaterEq = "greater-equal"
	VersionTypeLessEq    = "less-equal"
	VersionTypeGreater   = "greater"
	VersionTypeLess      = "less"
	VersionTypeRange     = "range"
	VersionTypeOr        = "or"
	VersionTypeLatest    = "latest"
	VersionTypeWildcard  = "wildcard"
	VersionTypeExact     = "exact"
	VersionTypeUnknown   = "unknown"
)

type VersionInfo struct {
	version    string
	npmPackage *NPMPackage
}

func newVersionInfo(version string, npmPackage *NPMPackage) *VersionInfo {
	return &VersionInfo{
		version:    version,
		npmPackage: npmPackage,
	}
}

func (v *VersionInfo) getVersion() string {
	version := v.version

	// Detect version type and return classification
	switch {
	case strings.HasPrefix(version, "^"):
		caretVersion := v.getVersionCaret()
		if caretVersion != "" {
			return caretVersion
		}
		return VersionTypeCaret
	case strings.HasPrefix(version, "~"):
		return VersionTypeTilde
	case strings.HasPrefix(version, ">="):
		return VersionTypeGreaterEq
	case strings.HasPrefix(version, "<="):
		return VersionTypeLessEq
	case strings.HasPrefix(version, ">"):
		return VersionTypeGreater
	case strings.HasPrefix(version, "<"):
		return VersionTypeLess
	case strings.Contains(version, " - "):
		return VersionTypeRange
	case strings.Contains(version, "||"):
		return VersionTypeOr
	case version == "*" || version == "latest":
		return VersionTypeLatest
	case strings.Contains(version, "x") || strings.Contains(version, "X"):
		return VersionTypeWildcard
	default:
		parts := strings.Split(version, ".")
		if len(parts) == 3 {
			npmVersion, exists := v.npmPackage.Versions[version]
			if exists && npmVersion.Version == version {
				return npmVersion.Version
			}

		}
		return VersionTypeUnknown
	}
}

func (v *VersionInfo) getVersionCaret() string {
	baseVersion := strings.Replace(v.version, "^", "", 1)
	v1 := "v" + baseVersion

	var bestVersion string
	var bestSemver string

	for k := range v.npmPackage.Versions {
		v2 := "v" + k
		if semver.Compare(v2, v1) >= 0 {
			majorBase := semver.Major(v1)
			majorCandidate := semver.Major(v2)

			if majorBase == majorCandidate {
				if bestSemver == "" || semver.Compare(v2, bestSemver) > 0 {
					bestVersion = k
					bestSemver = v2
				}
			}
		}
	}

	return bestVersion
}
