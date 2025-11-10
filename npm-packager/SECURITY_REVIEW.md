# Security Review Report

Based on my analysis of the code changes, I've identified **2 HIGH severity vulnerabilities** that require immediate attention:

---

# Vuln 1: Path Traversal via Malicious Package Names: `manager/manager.go:452, 527, 578`

* **Severity:** HIGH
* **Confidence:** 0.85
* **Category:** path_traversal

* **Description:** Package names received from the npm registry are used directly in file path construction without validation. A malicious package name containing path traversal sequences (`../`, absolute paths) could write files outside intended directories, potentially overwriting system files or accessing sensitive data.

* **Exploit Scenario:**
  1. Attacker publishes npm package with malicious name: `../../../../../../tmp/malicious`
  2. User installs package using `go run . add ../../../../../../tmp/malicious`
  3. Package manager constructs path: `filepath.Join("~/.config/go-npm/manifest", "../../../../../../tmp/malicious.json")` → resolves to `/tmp/malicious.json`
  4. Manifest and package files written outside cache directory
  5. Attacker achieves arbitrary file write on victim's system

* **Recommendation:**
  - Validate package names against npm's official naming rules before any file operations
  - Reject packages containing `..`, `/`, `\`, or starting with `.`
  - Use `filepath.Clean()` and verify result stays within intended directory
  - Add validation function:
    ```go
    func isValidPackageName(name string) bool {
        if strings.Contains(name, "..") || strings.ContainsAny(name, "/\\") {
            return false
        }
        if filepath.IsAbs(name) {
            return false
        }
        cleaned := filepath.Clean(name)
        return !strings.HasPrefix(cleaned, "..")
    }
    ```

---

# Vuln 2: Symlink Manipulation via Malicious Binary Names: `binlink/binlink.go:143-158`

* **Severity:** HIGH
* **Confidence:** 0.82
* **Category:** symlink_attack

* **Description:** Binary names extracted from package.json's `bin` field are used directly to create symlinks without validation. A malicious package could specify binary names with path traversal sequences to create symlinks outside the bin directory, potentially overwriting sensitive files like SSH keys or shell configuration.

* **Exploit Scenario:**
  1. Attacker publishes package with malicious package.json:
     ```json
     {
       "name": "evil-pkg",
       "bin": {
         "../../../.ssh/authorized_keys": "./backdoor_key.pub"
       }
     }
     ```
  2. User installs package globally: `go run . g evil-pkg`
  3. Symlink created at: `~/.config/go-npm/global/bin/../../../.ssh/authorized_keys` → `~/.ssh/authorized_keys`
  4. Victim's SSH authorized_keys replaced with attacker's public key
  5. Attacker gains SSH access to victim's system

* **Recommendation:**
  - Validate `binName` contains no path separators before symlink creation
  - Reject names with `/`, `\`, starting with `.`, or containing `..`
  - Add validation in `LinkPackage()` before line 105:
    ```go
    func isValidBinName(name string) bool {
        if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
            return false
        }
        if strings.HasPrefix(name, ".") {
            return false
        }
        return name == filepath.Base(name)
    }
    ```

---

**Priority:** Both vulnerabilities are directly exploitable and should be patched immediately before any public release.
