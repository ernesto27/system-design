package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestPackage(versions []string, latest string) *NPMPackage {
	pkg := &NPMPackage{
		DistTags: DistTags{
			Latest: latest,
		},
		Versions: make(map[string]Version),
	}

	for _, v := range versions {
		pkg.Versions[v] = Version{
			Version: v,
		}
	}

	return pkg
}

func TestVersionInfo_getVersion_EmptyVersion(t *testing.T) {
	vi := newVersionInfo()
	pkg := createTestPackage([]string{"1.0.0", "1.1.0", "2.0.0"}, "2.0.0")

	result := vi.getVersion("", pkg)
	assert.Equal(t, "2.0.0", result, "Empty version should return latest")
}

func TestVersionInfo_getVersion_Latest(t *testing.T) {
	vi := newVersionInfo()
	pkg := createTestPackage([]string{"1.0.0", "1.5.0", "2.3.1"}, "2.3.1")

	testCases := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "Asterisk wildcard",
			version:  "*",
			expected: "2.3.1",
		},
		{
			name:     "Latest keyword",
			version:  "latest",
			expected: "2.3.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := vi.getVersion(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_getVersion_ExactVersion(t *testing.T) {
	vi := newVersionInfo()
	pkg := createTestPackage([]string{"1.0.0", "1.2.3", "2.0.0"}, "2.0.0")

	testCases := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "Exact version exists",
			version:  "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "Exact version does not exist",
			version:  "1.2.4",
			expected: "2.0.0", // Falls back to latest
		},
		{
			name:     "Exact version with two parts only",
			version:  "1.2",
			expected: "2.0.0", // Falls back to latest (not 3 parts)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := vi.getVersion(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_getVersionCaret(t *testing.T) {
	testCases := []struct {
		name      string
		version   string
		versions  []string
		latest    string
		expected  string
		expectErr bool
	}{
		{
			name:     "Caret allows minor and patch updates - major 1",
			version:  "^1.2.3",
			versions: []string{"1.0.0", "1.2.3", "1.2.5", "1.3.0", "1.9.9", "2.0.0", "2.1.0"},
			latest:   "2.1.0",
			expected: "1.9.9", // Highest in major version 1
		},
		{
			name:     "Caret with major version 0",
			version:  "^0.2.3",
			versions: []string{"0.1.0", "0.2.3", "0.2.5", "0.3.0", "1.0.0"},
			latest:   "1.0.0",
			expected: "0.3.0", // Highest in major version 0
		},
		{
			name:     "Caret with exact match only",
			version:  "^1.0.0",
			versions: []string{"1.0.0", "2.0.0", "3.0.0"},
			latest:   "3.0.0",
			expected: "1.0.0",
		},
		{
			name:     "Caret with multiple candidates",
			version:  "^2.0.0",
			versions: []string{"1.9.9", "2.0.0", "2.0.1", "2.1.0", "2.5.7", "3.0.0"},
			latest:   "3.0.0",
			expected: "2.5.7",
		},
		{
			name:     "Caret with no matching versions",
			version:  "^5.0.0",
			versions: []string{"1.0.0", "2.0.0", "3.0.0", "4.0.0"},
			latest:   "4.0.0",
			expected: "", // No version satisfies
		},
		{
			name:     "Caret with lower base version",
			version:  "^1.0.0",
			versions: []string{"0.9.0", "1.0.0", "1.1.0", "1.2.0", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.2.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vi := newVersionInfo()
			pkg := createTestPackage(tc.versions, tc.latest)
			result := vi.getVersionCaret(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_getVersionTilde(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		versions []string
		latest   string
		expected string
	}{
		{
			name:     "Tilde allows patch updates only",
			version:  "~1.2.3",
			versions: []string{"1.0.0", "1.2.3", "1.2.5", "1.2.9", "1.3.0", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.2.9", // Highest patch in 1.2.x
		},
		{
			name:     "Tilde with exact match only",
			version:  "~1.2.3",
			versions: []string{"1.2.3", "1.3.0", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.2.3",
		},
		{
			name:     "Tilde with no higher patch version",
			version:  "~2.1.5",
			versions: []string{"2.0.0", "2.1.0", "2.1.3", "2.1.5", "2.2.0"},
			latest:   "2.2.0",
			expected: "2.1.5",
		},
		{
			name:     "Tilde with multiple patch versions",
			version:  "~3.0.0",
			versions: []string{"2.9.9", "3.0.0", "3.0.1", "3.0.5", "3.0.10", "3.1.0"},
			latest:   "3.1.0",
			expected: "3.0.10",
		},
		{
			name:     "Tilde with no matching versions",
			version:  "~5.0.0",
			versions: []string{"1.0.0", "2.0.0", "3.0.0", "4.0.0"},
			latest:   "4.0.0",
			expected: "", // No version satisfies
		},
		{
			name:     "Tilde excludes minor version changes",
			version:  "~1.2.0",
			versions: []string{"1.1.9", "1.2.0", "1.2.1", "1.3.0", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.2.1", // Does not include 1.3.0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vi := newVersionInfo()
			pkg := createTestPackage(tc.versions, tc.latest)
			result := vi.getVersionTilde(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_getVersionComplexRange(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		versions []string
		latest   string
		expected string
	}{
		{
			name:     "Range with >= and <",
			version:  ">= 2.1.2 < 3.0.0",
			versions: []string{"2.0.0", "2.1.0", "2.1.2", "2.5.0", "2.9.9", "3.0.0", "3.1.0"},
			latest:   "3.1.0",
			expected: "2.9.9",
		},
		{
			name:     "Range with >= and <= (inclusive)",
			version:  ">= 1.0.0 <= 2.0.0",
			versions: []string{"0.9.0", "1.0.0", "1.5.0", "2.0.0", "2.1.0"},
			latest:   "2.1.0",
			expected: "2.0.0",
		},
		{
			name:     "Range with > and < (exclusive)",
			version:  "> 1.0.0 < 2.0.0",
			versions: []string{"1.0.0", "1.0.1", "1.5.0", "1.9.9", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.9.9",
		},
		{
			name:     "Range with > and <= (mixed)",
			version:  "> 1.5.0 <= 2.5.0",
			versions: []string{"1.5.0", "1.6.0", "2.0.0", "2.5.0", "3.0.0"},
			latest:   "3.0.0",
			expected: "2.5.0",
		},
		{
			name:     "Range with no matching versions",
			version:  ">= 5.0.0 < 6.0.0",
			versions: []string{"1.0.0", "2.0.0", "3.0.0", "4.0.0"},
			latest:   "4.0.0",
			expected: "",
		},
		{
			name:     "Narrow range with one match",
			version:  ">= 1.2.3 < 1.2.5",
			versions: []string{"1.2.0", "1.2.3", "1.2.4", "1.2.5", "1.3.0"},
			latest:   "1.3.0",
			expected: "1.2.4",
		},
		{
			name:     "Range at boundary (lower bound inclusive)",
			version:  ">= 1.0.0 < 2.0.0",
			versions: []string{"0.9.9", "1.0.0", "1.5.0", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.5.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vi := newVersionInfo()
			pkg := createTestPackage(tc.versions, tc.latest)
			result := vi.getVersionComplexRange(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_getVersionWildcard(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		versions []string
		latest   string
		expected string
	}{
		{
			name:     "Single x returns latest",
			version:  "x",
			versions: []string{"1.0.0", "2.0.0", "3.0.0"},
			latest:   "3.0.0",
			expected: "3.0.0",
		},
		{
			name:     "Major.x matches any minor/patch in that major",
			version:  "1.x",
			versions: []string{"1.0.0", "1.2.0", "1.5.9", "2.0.0", "2.1.0"},
			latest:   "2.1.0",
			expected: "1.5.9",
		},
		{
			name:     "Major.minor.x matches any patch",
			version:  "2.1.x",
			versions: []string{"2.0.0", "2.1.0", "2.1.5", "2.1.9", "2.2.0", "3.0.0"},
			latest:   "3.0.0",
			expected: "2.1.9",
		},
		{
			name:     "Case insensitive X",
			version:  "1.X",
			versions: []string{"1.0.0", "1.3.0", "1.7.2", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.7.2",
		},
		{
			name:     "Major.X.X pattern",
			version:  "2.X.X",
			versions: []string{"1.9.9", "2.0.0", "2.1.0", "2.5.7", "3.0.0"},
			latest:   "3.0.0",
			expected: "2.5.7",
		},
		{
			name:     "No matching versions for wildcard",
			version:  "5.x",
			versions: []string{"1.0.0", "2.0.0", "3.0.0", "4.0.0"},
			latest:   "4.0.0",
			expected: "",
		},
		{
			name:     "Wildcard with exact major match",
			version:  "3.x",
			versions: []string{"3.0.0", "3.0.1", "3.1.0", "4.0.0"},
			latest:   "4.0.0",
			expected: "3.1.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vi := newVersionInfo()
			pkg := createTestPackage(tc.versions, tc.latest)
			result := vi.getVersionWildcard(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_getVersionOr(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		versions []string
		latest   string
		expected string
	}{
		{
			name:     "OR with two caret ranges",
			version:  "^1.0.0 || ^2.0.0",
			versions: []string{"1.0.0", "1.2.0", "1.9.9", "2.0.0", "2.1.0", "3.0.0"},
			latest:   "3.0.0",
			expected: "2.1.0", // Highest between 1.9.9 and 2.1.0
		},
		{
			name:     "OR with exact versions",
			version:  "1.0.0 || 2.0.0",
			versions: []string{"1.0.0", "2.0.0", "3.0.0"},
			latest:   "3.0.0",
			expected: "2.0.0", // Higher of the two
		},
		{
			name:     "OR with one matching constraint",
			version:  "^1.0.0 || ^5.0.0",
			versions: []string{"1.0.0", "1.5.0", "2.0.0", "3.0.0"},
			latest:   "3.0.0",
			expected: "1.5.0", // Only ^1.0.0 matches
		},
		{
			name:     "OR with tilde and caret",
			version:  "~1.2.3 || ^2.0.0",
			versions: []string{"1.2.3", "1.2.5", "1.3.0", "2.0.0", "2.5.0"},
			latest:   "2.5.0",
			expected: "2.5.0", // ^2.0.0 gives higher version
		},
		{
			name:     "OR with no matching constraints",
			version:  "^5.0.0 || ^6.0.0",
			versions: []string{"1.0.0", "2.0.0", "3.0.0", "4.0.0"},
			latest:   "4.0.0",
			expected: "",
		},
		{
			name:     "OR with wildcards",
			version:  "1.x || 3.x",
			versions: []string{"1.0.0", "1.5.0", "2.0.0", "3.0.0", "3.2.0"},
			latest:   "3.2.0",
			expected: "3.2.0", // Highest between 1.5.0 and 3.2.0
		},
		{
			name:     "OR with multiple constraints (3 options)",
			version:  "^1.0.0 || ^2.0.0 || ^3.0.0",
			versions: []string{"1.0.0", "1.1.0", "2.0.0", "2.2.0", "3.0.0", "3.5.0"},
			latest:   "3.5.0",
			expected: "3.5.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vi := newVersionInfo()
			pkg := createTestPackage(tc.versions, tc.latest)
			result := vi.getVersionOr(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_getVersion_UnsupportedRanges(t *testing.T) {
	vi := newVersionInfo()
	pkg := createTestPackage([]string{"1.0.0", "2.0.0", "3.0.0"}, "3.0.0")

	testCases := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "Greater than or equal (unsupported alone)",
			version:  ">=1.0.0",
			expected: "",
		},
		{
			name:     "Less than or equal (unsupported alone)",
			version:  "<=2.0.0",
			expected: "",
		},
		{
			name:     "Greater than (unsupported alone)",
			version:  ">1.0.0",
			expected: "",
		},
		{
			name:     "Less than (unsupported alone)",
			version:  "<2.0.0",
			expected: "",
		},
		{
			name:     "Hyphen range (unsupported)",
			version:  "1.0.0 - 2.0.0",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := vi.getVersion(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_getVersion_Integration(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		versions []string
		latest   string
		expected string
	}{
		{
			name:     "Caret range via getVersion",
			version:  "^1.2.0",
			versions: []string{"1.0.0", "1.2.0", "1.5.0", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.5.0",
		},
		{
			name:     "Tilde range via getVersion",
			version:  "~2.1.0",
			versions: []string{"2.0.0", "2.1.0", "2.1.5", "2.2.0"},
			latest:   "2.2.0",
			expected: "2.1.5",
		},
		{
			name:     "Complex range via getVersion",
			version:  ">= 1.0.0 < 2.0.0",
			versions: []string{"0.9.0", "1.0.0", "1.5.0", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.5.0",
		},
		{
			name:     "Wildcard via getVersion",
			version:  "1.x",
			versions: []string{"1.0.0", "1.9.0", "2.0.0"},
			latest:   "2.0.0",
			expected: "1.9.0",
		},
		{
			name:     "OR constraint via getVersion",
			version:  "^1.0.0 || ^2.0.0",
			versions: []string{"1.0.0", "1.5.0", "2.0.0", "2.3.0"},
			latest:   "2.3.0",
			expected: "2.3.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vi := newVersionInfo()
			pkg := createTestPackage(tc.versions, tc.latest)
			result := vi.getVersion(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestVersionInfo_EdgeCases(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		versions []string
		latest   string
		expected string
	}{
		{
			name:     "Package with only one version",
			version:  "^1.0.0",
			versions: []string{"1.0.0"},
			latest:   "1.0.0",
			expected: "1.0.0",
		},
		{
			name:     "Empty versions map",
			version:  "^1.0.0",
			versions: []string{},
			latest:   "",
			expected: "",
		},
		{
			name:     "Version with prerelease tag (treated as string)",
			version:  "1.0.0-beta.1",
			versions: []string{"1.0.0-beta.1", "1.0.0"},
			latest:   "1.0.0",
			expected: "1.0.0", // Falls back to latest (not 3 numeric parts)
		},
		{
			name:     "Very high version numbers",
			version:  "^100.200.300",
			versions: []string{"100.200.300", "100.200.400", "100.300.0", "200.0.0"},
			latest:   "200.0.0",
			expected: "100.300.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vi := newVersionInfo()
			pkg := createTestPackage(tc.versions, tc.latest)
			result := vi.getVersion(tc.version, pkg)
			assert.Equal(t, tc.expected, result)
		})
	}
}
