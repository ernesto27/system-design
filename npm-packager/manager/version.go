package manager

import (
	"strings"

	"golang.org/x/mod/semver"
)

type VersionInfo struct {
}

func newVersionInfo() *VersionInfo {
	return &VersionInfo{}
}

func (v *VersionInfo) getVersion(version string, npmPackage *NPMPackage) string {

	if version == "" {
		return npmPackage.DistTags.Latest
	}

	switch {
	case strings.Contains(version, "||"):
		orVersion := v.getVersionOr(version, npmPackage)
		return orVersion
	case strings.HasPrefix(version, "^"):
		caretVersion := v.getVersionCaret(version, npmPackage)
		return caretVersion
	case strings.HasPrefix(version, "~"):
		tildeVersion := v.getVersionTilde(version, npmPackage)
		return tildeVersion
	case strings.Contains(version, ">=") && (strings.Contains(version, "<") || strings.Contains(version, "<=")):
		complexVersion := v.getVersionComplexRange(version, npmPackage)
		return complexVersion
	case strings.HasPrefix(version, ">="):
		return ""
	case strings.HasPrefix(version, "<="):
		return ""
	case strings.HasPrefix(version, ">"):
		return ""
	case strings.HasPrefix(version, "<"):
		return ""
	case strings.Contains(version, " - "):
		return ""
	case version == "*" || version == "latest":
		return npmPackage.DistTags.Latest
	case strings.Contains(version, "x") || strings.Contains(version, "X"):
		wildcardVersion := v.getVersionWildcard(version, npmPackage)
		return wildcardVersion
	default:
		parts := strings.Split(version, ".")
		if len(parts) == 3 {
			npmVersion, exists := npmPackage.Versions[version]
			if exists && npmVersion.Version == version {
				return npmVersion.Version
			}

		}
		return npmPackage.DistTags.Latest
	}
}

func (v *VersionInfo) getVersionCaret(version string, npmPackage *NPMPackage) string {
	baseVersion := strings.Replace(version, "^", "", 1)
	v1 := "v" + baseVersion

	var bestVersion string
	var bestSemver string

	for k := range npmPackage.Versions {
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

func (v *VersionInfo) getVersionTilde(version string, npmPackage *NPMPackage) string {
	baseVersion := strings.Replace(version, "~", "", 1)
	v1 := "v" + baseVersion

	var bestVersion string
	var bestSemver string

	for k := range npmPackage.Versions {
		v2 := "v" + k
		if semver.Compare(v2, v1) >= 0 {
			// For tilde, we need to match the major and minor versions exactly
			majorBase := semver.Major(v1)
			minorBase := semver.MajorMinor(v1)
			majorCandidate := semver.Major(v2)
			minorCandidate := semver.MajorMinor(v2)

			// Tilde allows patch-level changes if minor version is specified
			// ~1.2.3 := >=1.2.3 <1.(2+1).0 := >=1.2.3 <1.3.0
			if majorBase == majorCandidate && minorBase == minorCandidate {
				if bestSemver == "" || semver.Compare(v2, bestSemver) > 0 {
					bestVersion = k
					bestSemver = v2
				}
			}
		}
	}

	return bestVersion
}

func (v *VersionInfo) getVersionComplexRange(version string, npmPackage *NPMPackage) string {

	var lowerBound, upperBound string
	var lowerInclusive, upperInclusive bool

	// Parse the complex range (e.g., ">= 2.1.2 < 3.0.0")
	parts := strings.Fields(version)

	for i := 0; i < len(parts)-1; i += 2 {
		operator := parts[i]
		versionStr := parts[i+1]

		switch operator {
		case ">=":
			lowerBound = versionStr
			lowerInclusive = true
		case ">":
			lowerBound = versionStr
			lowerInclusive = false
		case "<=":
			upperBound = versionStr
			upperInclusive = true
		case "<":
			upperBound = versionStr
			upperInclusive = false
		}
	}

	var bestVersion string
	var bestSemver string

	for k := range npmPackage.Versions {
		vCandidate := "v" + k

		// Check lower bound
		if lowerBound != "" {
			vLower := "v" + lowerBound
			comparison := semver.Compare(vCandidate, vLower)
			if lowerInclusive && comparison < 0 {
				continue
			}
			if !lowerInclusive && comparison <= 0 {
				continue
			}
		}

		// Check upper bound
		if upperBound != "" {
			vUpper := "v" + upperBound
			comparison := semver.Compare(vCandidate, vUpper)
			if upperInclusive && comparison > 0 {
				continue
			}
			if !upperInclusive && comparison >= 0 {
				continue
			}
		}

		// This version satisfies both bounds, check if it's the best one
		if bestSemver == "" || semver.Compare(vCandidate, bestSemver) > 0 {
			bestVersion = k
			bestSemver = vCandidate
		}
	}

	return bestVersion
}

func (v *VersionInfo) getVersionWildcard(version string, npmPackage *NPMPackage) string {
	normalized := strings.ToLower(version)
	parts := strings.Split(normalized, ".")

	// Handle different wildcard patterns:
	// "x" or "x.x.x" -> any version (use latest)
	// "1.x" or "1.x.x" -> any minor/patch in major 1
	// "1.2.x" -> any patch in 1.2

	if len(parts) == 1 && parts[0] == "x" {
		// "x" means any version
		return npmPackage.DistTags.Latest
	}

	var major, minor string
	var matchMinor bool

	if len(parts) >= 1 && parts[0] != "x" {
		major = parts[0]
	}
	if len(parts) >= 2 && parts[1] != "x" {
		minor = parts[1]
		matchMinor = true
	}

	var bestVersion string
	var bestSemver string

	for k := range npmPackage.Versions {
		vCandidate := "v" + k
		candidateParts := strings.Split(k, ".")

		if len(candidateParts) < 2 {
			continue
		}

		// Match major version if specified
		if major != "" && candidateParts[0] != major {
			continue
		}

		// Match minor version if specified
		if matchMinor && len(candidateParts) >= 2 && candidateParts[1] != minor {
			continue
		}

		// This version matches the pattern, check if it's the best one
		if bestSemver == "" || semver.Compare(vCandidate, bestSemver) > 0 {
			bestVersion = k
			bestSemver = vCandidate
		}
	}

	return bestVersion
}

func (v *VersionInfo) getVersionOr(version string, npmPackage *NPMPackage) string {
	// Split by || to get alternative constraints
	constraints := strings.Split(version, "||")

	var bestVersion string
	var bestSemver string

	// Try each constraint and find the highest version that satisfies any of them
	for _, constraint := range constraints {
		constraint = strings.TrimSpace(constraint)

		// Recursively resolve each constraint
		resolvedVersion := v.getVersion(constraint, npmPackage)

		if resolvedVersion != "" {
			vCandidate := "v" + resolvedVersion

			// Keep track of the highest version found
			if bestSemver == "" || semver.Compare(vCandidate, bestSemver) > 0 {
				bestVersion = resolvedVersion
				bestSemver = vCandidate
			}
		}
	}

	return bestVersion
}
