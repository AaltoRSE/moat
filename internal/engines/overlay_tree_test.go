package engines

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAutoOverlaySnapshotTriesMethodsInOrder exercises the final copy
// fallback and verifies that failed candidates leave no partial output.
func TestAutoOverlaySnapshotTriesMethodsInOrder(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("MOAT_SNAPSHOT_LOG", filepath.Join(dir, "log"))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "btrfs"), []byte(`#!/bin/sh
if [ "$2" = snapshot ]; then
    printf '%s\n' nativefs >> "$MOAT_SNAPSHOT_LOG"
fi
exit 1
`), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cp"), []byte(`#!/bin/sh
case "$3" in
  --reflink=always)
    printf '%s\n' reflink >> "$MOAT_SNAPSHOT_LOG"
    mkdir "$6"
    touch "$6/partial"
    exit 1
    ;;
  --reflink=never)
    printf '%s\n' copy >> "$MOAT_SNAPSHOT_LOG"
    test ! -e "$6" && mkdir "$6"
    ;;
  *) exit 1 ;;
esac
`), 0o755))
	target := filepath.Join(dir, "snapshot")
	require.NoError(t, snapshotOverlayTree(dir, target, "auto"))
	contents, err := os.ReadFile(filepath.Join(dir, "log"))
	require.NoError(t, err)
	assert.Equal(t, "nativefs\nreflink\ncopy\n", string(contents))
	assert.NoFileExists(t, filepath.Join(target, "partial"))
}
