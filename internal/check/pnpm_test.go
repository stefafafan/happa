package check

import "testing"

func TestParsePnpmLock(t *testing.T) {
	content := `
lockfileVersion: '9.0'

importers:
  .:
    dependencies:
      react:
        specifier: ^19.0.0
        version: 19.1.0
      '@scope/direct':
        specifier: ^2.0.0
        version: 2.1.0(peer@1.0.0)
  packages/app:
    devDependencies:
      vite:
        specifier: ^6.0.0
        version: 6.3.5

packages:
  react@19.1.0:
    resolution: {integrity: sha512-test}
  '@scope/direct@2.1.0':
    resolution: {integrity: sha512-test}
  transitive@4.5.6:
    resolution: {integrity: sha512-test}

snapshots:
  '@scope/transitive@7.8.9':
    dependencies:
      transitive: 4.5.6
`

	index := parsePnpmLock(content)

	assertVersions(t, index.Resolved["react"], "19.1.0")
	assertVersions(t, index.Resolved["@scope/direct"], "2.1.0")
	assertVersions(t, index.Resolved["transitive"], "4.5.6")
	assertVersions(t, index.Resolved["@scope/transitive"], "7.8.9")
}

func TestPackageFromLockKeyModernPnpmFormats(t *testing.T) {
	tests := []struct {
		key     string
		name    string
		version string
	}{
		{"@scope/pkg@2.0.0(peer@1.0.0)", "@scope/pkg", "2.0.0"},
		{"left-pad@1.3.0", "left-pad", "1.3.0"},
	}

	for _, test := range tests {
		name, version, ok := packageFromLockKey(test.key)
		if !ok {
			t.Fatalf("packageFromLockKey(%q) did not parse", test.key)
		}
		if name != test.name || version != test.version {
			t.Fatalf("packageFromLockKey(%q) = (%q, %q), want (%q, %q)", test.key, name, version, test.name, test.version)
		}
	}
}

func assertVersions(t *testing.T, got map[string]bool, want ...string) {
	t.Helper()
	for _, version := range want {
		if !got[version] {
			t.Fatalf("missing version %q in %#v", version, got)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("got versions %#v, want %v", got, want)
	}
}
