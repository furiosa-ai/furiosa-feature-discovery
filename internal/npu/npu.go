package npu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/furiosa-ai/furiosa-feature-discovery/internal/consts"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

type VersionInfo struct {
	major      uint32
	minor      uint32
	patch      uint32
	Metadata   string
	Prerelease string
}

func NewVersionInfo(major, minor, patch uint32, metadata, prerelease string) *VersionInfo {
	return &VersionInfo{
		major:      major,
		minor:      minor,
		patch:      patch,
		Metadata:   metadata,
		Prerelease: prerelease,
	}
}

func (v *VersionInfo) Compare(baseFirmwareVersion *VersionInfo) bool {
	if baseFirmwareVersion == nil {
		return false
	}
	return v.major == baseFirmwareVersion.major &&
		v.minor == baseFirmwareVersion.minor &&
		v.patch == baseFirmwareVersion.patch &&
		v.Metadata == baseFirmwareVersion.Metadata &&
		v.Prerelease == baseFirmwareVersion.Prerelease
}

func (v *VersionInfo) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

func (v *VersionInfo) Major() string {
	return strconv.FormatUint(uint64(v.major), 10)
}

func (v *VersionInfo) Minor() string {
	return strconv.FormatUint(uint64(v.minor), 10)
}

func (v *VersionInfo) Patch() string {
	return strconv.FormatUint(uint64(v.patch), 10)
}

// StringOrEmpty dereferences s, returning "" when it is nil.
func StringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type NPUDevice struct {
	family       string
	product      string
	smiObject    *api.Device
	FirmwareInfo *VersionInfo
}

func recognizeFamily(arch string) (string, error) {
	switch arch {
	case consts.ArchRngd, consts.ArchRngdS, consts.ArchRngdMax:
		return consts.ArchRngd, nil
	default:
		return "", fmt.Errorf("unknown architecture: %s", arch)
	}
}

func NewNPUDevice(device *api.Device) *NPUDevice {
	return &NPUDevice{
		smiObject: device,
	}
}

func (d *NPUDevice) DetectNPUDevice() error {
	archName := strings.ToLower(d.smiObject.Arch().String())
	family, err := recognizeFamily(archName)
	if err != nil {
		return err
	}
	var firmwareInfo *VersionInfo
	// An unreadable firmware version is treated as unknown (no firmware labels).
	if fw, err := d.smiObject.FirmwareVersion(); err == nil {
		firmwareInfo = NewVersionInfo(
			fw.Major(),
			fw.Minor(),
			fw.Patch(),
			StringOrEmpty(fw.Metadata()),
			StringOrEmpty(fw.Prerelease()))
	}
	d.family = family
	d.product = archName
	d.FirmwareInfo = firmwareInfo
	return nil
}

func (d *NPUDevice) ToLabels() map[string]string {
	labels := map[string]string{
		consts.FuriosaNPUFamilyLabel:  d.family,
		consts.FuriosaNPUProductLabel: d.product,
	}

	if d.FirmwareInfo != nil {
		labels[consts.FuriosaFirmwareVersionLabel] = d.FirmwareInfo.String()
		labels[consts.FuriosaFirmwareVersionMajorLabel] = d.FirmwareInfo.Major()
		labels[consts.FuriosaFirmwareVersionMinorLabel] = d.FirmwareInfo.Minor()
		labels[consts.FuriosaFirmwareVersionPatchLabel] = d.FirmwareInfo.Patch()
		labels[consts.FuriosaFirmwareVersionMetadataLabel] = d.FirmwareInfo.Metadata
		labels[consts.FuriosaFirmwareVersionPrereleaseLabel] = d.FirmwareInfo.Prerelease
	}

	return labels
}

func NewTestNPUDevice(family, product string, firmwareInfo *VersionInfo) *NPUDevice {
	return &NPUDevice{
		family:       family,
		product:      product,
		FirmwareInfo: firmwareInfo,
	}
}
