package npu

import (
	"testing"

	"github.com/furiosa-ai/furiosa-feature-discovery/internal/consts"
	"github.com/stretchr/testify/suite"
)

type NPUTestSuite struct {
	suite.Suite
}

func (s *NPUTestSuite) SetupTest() {
}

func (s *NPUTestSuite) TearDownTest() {
}

func (s *NPUTestSuite) SetupSuite() {
}

func (s *NPUTestSuite) TearDownSuite() {
}
func (s *NPUTestSuite) TestVersionInfoCompare() {
	tests := []struct {
		name     string
		v        *VersionInfo
		base     *VersionInfo
		expected bool
	}{
		{
			name: "identical versions",
			v: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			base: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			expected: true,
		},
		{
			name: "different major version",
			v: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			base: &VersionInfo{
				major:      2,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			expected: false,
		},
		{
			name: "different minor version",
			v: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			base: &VersionInfo{
				major:      1,
				minor:      3,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			expected: false,
		},
		{
			name: "different patch version",
			v: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			base: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      4,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			expected: false,
		},
		{
			name: "different metadata",
			v: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta1",
				Prerelease: "pre",
			},
			base: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta2",
				Prerelease: "pre",
			},
			expected: false,
		},
		{
			name: "different prerelease",
			v: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre1",
			},
			base: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre2",
			},
			expected: false,
		},
		{
			name: "nil base version",
			v: &VersionInfo{
				major:      1,
				minor:      2,
				patch:      3,
				Metadata:   "meta",
				Prerelease: "pre",
			},
			base:     nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.v.Compare(tt.base)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *NPUTestSuite) TestVersionInfoString() {
	tests := []struct {
		name     string
		v        *VersionInfo
		expected string
	}{
		{
			name: "basic version",
			v: &VersionInfo{
				major: 1,
				minor: 2,
				patch: 3,
			},
			expected: "1.2.3",
		},
		{
			name: "zero version",
			v: &VersionInfo{
				major: 0,
				minor: 0,
				patch: 0,
			},
			expected: "0.0.0",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.v.String()
			s.Equal(tt.expected, result)
		})
	}
}

func (s *NPUTestSuite) TestRecognizeFamily() {
	tests := []struct {
		name        string
		arch        string
		expected    string
		expectedErr bool
	}{
		{
			name:        "rngd architecture",
			arch:        consts.ArchRngd,
			expected:    consts.ArchRngd,
			expectedErr: false,
		},
		{
			name:        "rngd_s architecture",
			arch:        consts.ArchRngdS,
			expected:    consts.ArchRngd,
			expectedErr: false,
		},
		{
			name:        "rngd_max architecture",
			arch:        consts.ArchRngdMax,
			expected:    consts.ArchRngd,
			expectedErr: false,
		},
		{
			name:        "unknown architecture",
			arch:        "unknown",
			expected:    "",
			expectedErr: true,
		},
		{
			name:        "empty architecture",
			arch:        "",
			expected:    "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := recognizeFamily(tt.arch)
			if tt.expectedErr {
				s.Error(err)
			} else {
				s.NoError(err)
			}
			s.Equal(tt.expected, result)
		})
	}
}

func (s *NPUTestSuite) TestNPUDeviceToLabels() {
	tests := []struct {
		name     string
		device   *NPUDevice
		expected map[string]string
	}{
		{
			name: "with firmware info",
			device: &NPUDevice{
				family:  consts.ArchRngd,
				product: consts.ArchRngdS,
				FirmwareInfo: &VersionInfo{
					major:      4,
					minor:      5,
					patch:      6,
					Metadata:   "fwmeta",
					Prerelease: "fwpre",
				},
			},
			expected: map[string]string{
				consts.FuriosaNPUFamilyLabel:                 consts.ArchRngd,
				consts.FuriosaNPUProductLabel:                consts.ArchRngdS,
				consts.FuriosaFirmwareVersionLabel:           "4.5.6",
				consts.FuriosaFirmwareVersionMajorLabel:      "4",
				consts.FuriosaFirmwareVersionMinorLabel:      "5",
				consts.FuriosaFirmwareVersionPatchLabel:      "6",
				consts.FuriosaFirmwareVersionMetadataLabel:   "fwmeta",
				consts.FuriosaFirmwareVersionPrereleaseLabel: "fwpre",
			},
		},
		{
			name: "without firmware info",
			device: &NPUDevice{
				family:       consts.ArchRngd,
				product:      consts.ArchRngdMax,
				FirmwareInfo: nil,
			},
			expected: map[string]string{
				consts.FuriosaNPUFamilyLabel:  consts.ArchRngd,
				consts.FuriosaNPUProductLabel: consts.ArchRngdMax,
			},
		},
		{
			name: "zero version values",
			device: &NPUDevice{
				family:  consts.ArchRngd,
				product: consts.ArchRngd,
				FirmwareInfo: &VersionInfo{
					major:      0,
					minor:      0,
					patch:      0,
					Metadata:   "",
					Prerelease: "",
				},
			},
			expected: map[string]string{
				consts.FuriosaNPUFamilyLabel:                 consts.ArchRngd,
				consts.FuriosaNPUProductLabel:                consts.ArchRngd,
				consts.FuriosaFirmwareVersionLabel:           "0.0.0",
				consts.FuriosaFirmwareVersionMajorLabel:      "0",
				consts.FuriosaFirmwareVersionMinorLabel:      "0",
				consts.FuriosaFirmwareVersionPatchLabel:      "0",
				consts.FuriosaFirmwareVersionMetadataLabel:   "",
				consts.FuriosaFirmwareVersionPrereleaseLabel: "",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := tt.device.ToLabels()
			s.Equal(len(tt.expected), len(result), "label count mismatch")
			s.Equal(tt.expected, result)
		})
	}
}

func TestNPUTestSuite(t *testing.T) {
	suite.Run(t, new(NPUTestSuite))
}
