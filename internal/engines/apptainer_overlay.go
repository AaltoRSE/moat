package engines

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/rs/zerolog/log"
)

// Overlay owner and driver values supported by the ApptainerRuntime, and
// the default managed directory for externally owned overlays.
const (
	overlayOwnerInternal = "internal"
	overlayOwnerExternal = "external"
	overlayDriverFuse    = "fuse-overlayfs"
	overlayDriverKernel  = "kernel-overlayfs"
	overlayDriverBtrfs   = "btrfs"
	defaultOverlayDir    = "$HOME/.cache/moat/overlays"
)

type apptainerOverlays struct {
	flags   []string
	command []string
	scratch string
	locks   []*os.File
}

func (o *apptainerOverlays) close() {
	for _, lock := range o.locks {
		_ = lock.Close()
	}
	if err := os.RemoveAll(o.scratch); err != nil {
		log.Warn().Err(err).Str("path", o.scratch).Msg("Could not remove temporary overlay storage")
	}
}

// prepareOverlays stages the mounts and holds their writable layers until the
// Apptainer process exits. The caller closes the returned plan after the run.
func (f *ApptainerRuntime) prepareOverlays(env types.MoatEnv, imagePath string, passEnv bool) (_ *apptainerOverlays, err error) {
	owner, driver := f.OverlayOwner, f.OverlayDriver
	if owner == "" {
		owner = overlayOwnerInternal
	}
	if driver == "" {
		driver = overlayDriverFuse
	}
	if owner == overlayOwnerInternal && driver != overlayDriverFuse {
		return nil, fmt.Errorf("overlayowner %q only supports overlaydriver %q, got %q", owner, overlayDriverFuse, driver)
	}

	var managedRoot string
	if owner == overlayOwnerExternal {
		base := f.OverlayDir
		if base == "" {
			base = defaultOverlayDir
		}
		envName := env.Name
		if envName == "" {
			envName = "default"
		}
		if !utils.CheckEnvironmentName(envName) {
			return nil, fmt.Errorf("invalid environment name for overlay storage: %q", envName)
		}
		managedRoot, err = utils.SanitizePath(base)
		if err != nil {
			return nil, err
		}
		managedRoot = filepath.Join(managedRoot, envName)
	}

	scratch, err := os.MkdirTemp("", "moat-overlay-")
	if err != nil {
		return nil, err
	}
	plan := &apptainerOverlays{scratch: scratch}
	defer func() {
		if err != nil {
			plan.close()
		}
	}()

	staging := "/" + filepath.Base(scratch)
	script := filepath.Join(scratch, "run.sh")
	if err := os.WriteFile(script, []byte(overlayRunScript), 0o600); err != nil {
		return nil, err
	}
	plan.flags = []string{"--no-eval", "--mount", mountSpec(script, staging+"/run.sh", "ro")}
	plan.command = []string{"/bin/sh", staging + "/run.sh", driver, strconv.Itoa(len(env.OverlayMounts))}
	if driver == overlayDriverFuse {
		// The lower tree must stay visible after Apptainer covers the target
		// with the FUSE mount. Bind the primary SIF partition at another path.
		id, err := f.primaryPartitionID(imagePath, passEnv)
		if err != nil {
			return nil, err
		}
		plan.flags = append(plan.flags, "--mount", mountSpec(imagePath, staging+"/image",
			"ro", "image-src=/", "id="+id))
	} else {
		plan.command = append([]string{"/usr/bin/unshare", "--map-current-user", "--keep-caps", "--mount", "--"}, plan.command...)
	}

	for i, mount := range env.OverlayMounts {
		source, dest, explicit := strings.Cut(mount, ":")
		if !explicit {
			dest = source
			if owner == overlayOwnerExternal {
				source = filepath.Join(managedRoot, fmt.Sprintf("%x", sha256.Sum256([]byte(dest))))
			} else {
				source = filepath.Join(scratch, strconv.Itoa(i))
			}
		}
		if driver == overlayDriverBtrfs {
			if err := utils.EnsureBtrfsSubvolume(source); err != nil {
				return nil, err
			}
		}
		if err := plan.prepareLayer(source); err != nil {
			return nil, err
		}

		layer := staging + "/" + strconv.Itoa(i)
		plan.flags = append(plan.flags, "--mount", mountSpec(source, layer))
		if driver == overlayDriverFuse {
			// Apptainer splits FUSE definitions at commas before parsing their arguments.
			command := fmt.Sprintf("container:fuse-overlayfs -o lowerdir=%s/image%s -o upperdir=%s/upper -o workdir=%s/work %s",
				staging, dest, layer, layer, dest)
			plan.flags = append(plan.flags, "--fusemount", command)
		}
		plan.command = append(plan.command, layer, dest)
	}
	return plan, nil
}

func (o *apptainerOverlays) prepareLayer(source string) error {
	if err := os.MkdirAll(source, 0o700); err != nil {
		return err
	}
	dir, err := os.Open(source)
	if err != nil {
		return err
	}
	if err := syscall.Flock(int(dir.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = dir.Close()
		return fmt.Errorf("lock overlay storage %q (each mount needs exclusive storage): %w", source, err)
	}
	o.locks = append(o.locks, dir)
	for _, name := range []string{"upper", "work"} {
		if err := os.MkdirAll(filepath.Join(source, name), 0o700); err != nil {
			return err
		}
	}
	return nil
}

func (f *ApptainerRuntime) primaryPartitionID(imagePath string, passEnv bool) (string, error) {
	list, err := utils.RunCapture(utils.RunArgs{
		Command: f.Executable, Args: []string{"sif", "list", imagePath}, PassEnv: passEnv,
	})
	if err != nil {
		return "", fmt.Errorf("list SIF partitions in %q: %w: %s", imagePath, err, strings.TrimSpace(list))
	}
	for line := range strings.SplitSeq(list, "\n") {
		columns := strings.Split(line, "|")
		if len(columns) != 5 || !strings.Contains(columns[4], "/*System/") {
			continue
		}
		id := strings.TrimSpace(columns[0])
		if n, err := strconv.ParseUint(id, 10, 32); err == nil && n > 0 {
			return id, nil
		}
	}
	return "", fmt.Errorf("no primary filesystem partition in %q", imagePath)
}

func mountSpec(source, dest string, options ...string) string {
	var spec strings.Builder
	writer := csv.NewWriter(&spec)
	_ = writer.Write(append([]string{"type=bind", "src=" + source, "dst=" + dest}, options...))
	writer.Flush()
	return strings.TrimSuffix(spec.String(), "\n")
}

// Apptainer exec initializes the container environment. This script prepares
// the overlays before calling the image's runscript with the original arguments.
const overlayRunScript = `
set -eu
driver=$1
remaining=$2
shift 2
cwd=$(pwd -P)
while [ "$remaining" -gt 0 ]; do
	layer=$1
	dest=$2
	shift 2
	remaining=$((remaining - 1))
	if [ "$driver" = fuse-overlayfs ]; then
		# Apptainer's initial FUSE root mode contains no permission bits.
		# Ask the driver for the real attributes before the first write.
		stat -L -- "$dest" >/dev/null
	else
		/usr/bin/mount -i -t overlay overlay -o "lowerdir=$dest,upperdir=$layer/upper,workdir=$layer/work,userxattr,nosuid,nodev" -- "$dest"
	fi
done
cd -- "$cwd"

runscript=/.singularity.d/runscript
if [ -n "${SINGULARITY_APPNAME:-}" ]; then
	runscript=/scif/apps/$SINGULARITY_APPNAME/scif/runscript
	if [ ! -x "$runscript" ]; then
		runscript=/scif/apps/singularity/scif/test
		if [ ! -x "$runscript" ]; then
			echo "no runscript for contained app: $SINGULARITY_APPNAME" >&2
			exit 1
		fi
	fi
elif [ ! -x "$runscript" ]; then
	runscript=/bin/sh
	if [ -x /apptainer ]; then
		runscript=/apptainer
	fi
fi

export APPTAINER_COMMAND=run
if [ "$driver" != fuse-overlayfs ]; then
	exec /usr/bin/setpriv --ambient-caps=-all --inh-caps=-all --bounding-set=-all -- "$runscript" "$@"
fi
exec "$runscript" "$@"
`
