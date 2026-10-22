package discovery

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/furiosa-ai/furiosa-feature-discovery/internal/config"
	"github.com/furiosa-ai/furiosa-feature-discovery/internal/consts"
	"github.com/furiosa-ai/furiosa-feature-discovery/internal/npu"
	"github.com/stretchr/testify/suite"
)

type DiscoveryTestSuite struct {
	suite.Suite
}

func (s *DiscoveryTestSuite) SetupTest() {
	_ = os.Remove("test_dummy.txt")
}

func (s *DiscoveryTestSuite) TearDownTest() {
}

func (s *DiscoveryTestSuite) SetupSuite() {
}

func (s *DiscoveryTestSuite) TearDownSuite() {
}

func splitLines(s string) []string {
	if s == "" {
		return []string{}
	}
	var lines []string
	var current string
	for _, char := range s {
		if char == '\n' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func (s *DiscoveryTestSuite) createTestNPUDevice(family, product string, firmwareInfo *npu.VersionInfo) *npu.NPUDevice {
	return npu.NewTestNPUDevice(family, product, firmwareInfo)
}

func (s *DiscoveryTestSuite) TestLabelsToFeatureString() {
	tests := []struct {
		name     string
		labels   map[string]string
		expected string
	}{
		{
			name:     "empty labels",
			labels:   map[string]string{},
			expected: "",
		},
		{
			name: "single label",
			labels: map[string]string{
				"key1": "value1",
			},
			expected: "key1=value1",
		},
		{
			name: "multiple labels",
			labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "labels with empty values",
			labels: map[string]string{
				"key1": "value1",
				"key2": "",
				"key3": "value3",
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := LabelsToFeatureString(tt.labels)
			if tt.expected != "" {
				s.Equal(tt.expected, result)
			} else {
				if len(tt.labels) == 0 {
					s.Empty(result)
				} else {
					lines := make(map[string]bool)
					for k, v := range tt.labels {
						lines[k+"="+v] = true
					}
					resultLines := make(map[string]bool)
					for _, line := range splitLines(result) {
						resultLines[line] = true
					}
					s.Equal(lines, resultLines)
					s.Equal(len(tt.labels), len(resultLines))
				}
			}
		})
	}
}

func (s *DiscoveryTestSuite) TestCheckLabels() {
	tests := []struct {
		name           string
		devices        []*npu.NPUDevice
		expectedNilFw  bool
		expectedLength int
	}{
		{
			name:           "empty devices",
			devices:        []*npu.NPUDevice{},
			expectedNilFw:  false,
			expectedLength: 0,
		},
		{
			name: "single device",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(1, 2, 3, "", "")),
			},
			expectedNilFw:  false,
			expectedLength: 1,
		},
		{
			name: "matching firmware versions",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(1, 2, 3, "", "")),
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(1, 2, 3, "", "")),
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(1, 2, 3, "", "")),
			},
			expectedNilFw:  false,
			expectedLength: 3,
		},
		{
			name: "non-matching firmware versions",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(1, 2, 3, "", "")),
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(1, 2, 4, "", "")),
			},
			expectedNilFw:  true,
			expectedLength: 2,
		},
		{
			name: "nil firmware in first device",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, nil),
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(1, 2, 3, "", "")),
			},
			expectedNilFw:  true,
			expectedLength: 2,
		},
		{
			name: "nil firmware in second device",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(1, 2, 3, "", "")),
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, nil),
			},
			expectedNilFw:  true,
			expectedLength: 2,
		},
		{
			name: "all nil firmware",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, nil),
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, nil),
			},
			expectedNilFw:  false, // nil matches nil, so no change
			expectedLength: 2,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			testDevices := make([]*npu.NPUDevice, len(tt.devices))
			for i, d := range tt.devices {
				var fwInfo *npu.VersionInfo
				if d.FirmwareInfo != nil {
					major, _ := strconv.ParseUint(d.FirmwareInfo.Major(), 10, 32)
					minor, _ := strconv.ParseUint(d.FirmwareInfo.Minor(), 10, 32)
					patch, _ := strconv.ParseUint(d.FirmwareInfo.Patch(), 10, 32)
					fwInfo = npu.NewVersionInfo(
						uint32(major),
						uint32(minor),
						uint32(patch),
						d.FirmwareInfo.Metadata,
						d.FirmwareInfo.Prerelease,
					)
				}
				testDevices[i] = s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, fwInfo)
			}

			driverInfo := npu.NewVersionInfo(1, 2, 3, "", "")
			runner := &FeatureDiscoveryRunner{
				devices:    testDevices,
				driverInfo: driverInfo,
			}
			_ = runner.ExtractLabels()

			s.Equal(tt.expectedLength, len(runner.devices))
			for _, device := range runner.devices {
				if tt.expectedNilFw {
					s.Nil(device.FirmwareInfo)
				} else {
					if len(tt.devices) > 0 && tt.devices[0].FirmwareInfo != nil {
						s.NotNil(device.FirmwareInfo)
					}
				}
			}
		})
	}
}

func (s *DiscoveryTestSuite) TestExtractLabels() {
	tests := []struct {
		name          string
		devices       []*npu.NPUDevice
		expectedKeys  []string
		expectedCount string
	}{
		{
			name:          "empty devices",
			devices:       []*npu.NPUDevice{},
			expectedKeys:  []string{consts.FuriosaNPUCountLabel},
			expectedCount: "0",
		},
		{
			name: "single device",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, npu.NewVersionInfo(4, 5, 6, "", "")),
			},
			expectedKeys: []string{
				consts.FuriosaNPUCountLabel,
				consts.FuriosaNPUFamilyLabel,
				consts.FuriosaNPUProductLabel,
			},
			expectedCount: "1",
		},
		{
			name: "multiple devices",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdS, nil),
				s.createTestNPUDevice(consts.ArchRngd, consts.ArchRngdMax, nil),
			},
			expectedKeys: []string{
				consts.FuriosaNPUCountLabel,
			},
			expectedCount: "2",
		},
		{
			name: "device with different family",
			devices: []*npu.NPUDevice{
				s.createTestNPUDevice("rngd_other", consts.ArchRngdS, npu.NewVersionInfo(1, 0, 0, "", "")),
			},
			expectedKeys: []string{
				consts.FuriosaNPUCountLabel,
				consts.FuriosaNPUFamilyLabel,
				consts.FuriosaNPUProductLabel,
			},
			expectedCount: "1",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			driverInfo := npu.NewVersionInfo(1, 2, 3, "", "")
			runner := &FeatureDiscoveryRunner{
				devices:    tt.devices,
				driverInfo: driverInfo,
			}
			result := runner.ExtractLabels()
			s.NotNil(result)
			for _, key := range tt.expectedKeys {
				s.Contains(result, key)
			}
			if tt.expectedCount != "" {
				s.Equal(tt.expectedCount, result[consts.FuriosaNPUCountLabel])
			}
		})
	}
}

func (s *DiscoveryTestSuite) TestFeatureDiscoveryRunnerSyncFile() {
	tmpDir := s.T().TempDir()
	tests := []struct {
		name        string
		labels      map[string]string
		outputPath  string
		expectError bool
	}{
		{
			name:        "empty labels",
			labels:      map[string]string{},
			outputPath:  tmpDir + "/test_dummy.txt",
			expectError: false, // Returns nil error but logs warning
		},
		{
			name: "valid labels",
			labels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			outputPath:  tmpDir + "/test_dummy.txt",
			expectError: false,
		},
		{
			name: "invalid output path",
			labels: map[string]string{
				"key1": "value1",
			},
			outputPath:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.outputPath != "" {
				defer func() { _ = os.Remove(tt.outputPath) }()
			}

			cfg := &config.Config{
				OutputPath: tt.outputPath,
			}
			runner := &FeatureDiscoveryRunner{
				config: cfg,
			}

			err := runner.SyncFile(tt.labels)
			if tt.expectError {
				s.Error(err)
			} else {
				if len(tt.labels) == 0 {
					s.NoError(err)
				} else {
					s.NoError(err)
					_, err := os.Stat(tt.outputPath)
					s.NoError(err, "output file should exist")
					content, err := os.ReadFile(tt.outputPath)
					s.NoError(err)
					s.NotEmpty(content)
				}
			}
		})
	}
}

func (s *DiscoveryTestSuite) TestFeatureDiscoveryRunnerRemoveFFDFile() {
	tests := []struct {
		name        string
		setupFile   bool
		expectError bool
	}{
		{
			name:        "file exists",
			setupFile:   true,
			expectError: false,
		},
		{
			name:        "file does not exist",
			setupFile:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tmpDir := s.T().TempDir()
			outputPath := filepath.Join(tmpDir, "test_dummy.txt")

			if tt.setupFile {
				err := os.WriteFile(outputPath, []byte("test content"), 0644)
				s.NoError(err)
			}

			cfg := &config.Config{
				OutputPath: outputPath,
			}
			runner := &FeatureDiscoveryRunner{
				config: cfg,
			}

			err := runner.RemoveFFDFile()
			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				_, err := os.Stat(outputPath)
				s.Error(err, "file should not exist")
			}
		})
	}
}

func (s *DiscoveryTestSuite) TestFeatureDiscoveryRunnerDetectNPUDevices() {
	tests := []struct {
		name        string
		devices     []*npu.NPUDevice
		expectError bool
	}{
		{
			name:        "empty devices",
			devices:     []*npu.NPUDevice{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			runner := &FeatureDiscoveryRunner{
				devices: tt.devices,
			}
			err := runner.detectNPUDevices()
			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *DiscoveryTestSuite) TestFeatureDiscoveryRunnerSync() {
	tests := []struct {
		name        string
		devices     []*npu.NPUDevice
		ctxTimeout  time.Duration
		expectError bool
	}{
		{
			name:        "empty devices",
			devices:     []*npu.NPUDevice{},
			ctxTimeout:  5 * time.Second,
			expectError: true,
		},
		{
			name:        "context timeout",
			devices:     []*npu.NPUDevice{},
			ctxTimeout:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tmpDir := s.T().TempDir()
			outputPath := filepath.Join(tmpDir, "test_dummy.txt")

			cfg := &config.Config{
				OutputPath: outputPath,
			}
			driverInfo := npu.NewVersionInfo(1, 2, 3, "", "")
			runner := &FeatureDiscoveryRunner{
				config:     cfg,
				devices:    tt.devices,
				driverInfo: driverInfo,
			}

			var ctx context.Context
			var cancel context.CancelFunc
			if tt.ctxTimeout == 0 {
				ctx, cancel = context.WithTimeout(context.Background(), 0)
			} else {
				ctx, cancel = context.WithTimeout(context.Background(), tt.ctxTimeout)
			}
			defer cancel()

			err := runner.sync(ctx)
			if tt.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				_, err := os.Stat(outputPath)
				s.NoError(err, "output file should exist")
			}
		})
	}
}

func TestDiscoveryTestSuite(t *testing.T) {
	suite.Run(t, new(DiscoveryTestSuite))
}
