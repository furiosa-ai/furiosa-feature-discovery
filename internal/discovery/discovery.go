package discovery

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/furiosa-ai/furiosa-feature-discovery/internal/config"
	"github.com/furiosa-ai/furiosa-feature-discovery/internal/consts"
	"github.com/furiosa-ai/furiosa-feature-discovery/internal/npu"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/rs/zerolog"
)

func LabelsToFeatureString(labels map[string]string) string {
	var result string
	for k, v := range labels {
		if result != "" {
			result += "\n"
		}
		result += fmt.Sprintf("%s=%s", k, v)
	}
	return result
}

type FeatureDiscoveryRunner struct {
	config      *config.Config
	logger      zerolog.Logger
	devices     []*npu.NPUDevice
	driverInfo  *npu.VersionInfo
	syncRunning bool
	syncMu      sync.Mutex
}

func NewFeatureDiscoveryRunner(cfg *config.Config) (*FeatureDiscoveryRunner, error) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return &FeatureDiscoveryRunner{
		config: cfg,
		logger: logger,
	}, nil
}

func (r *FeatureDiscoveryRunner) updateDevicesAndDriverInfo() error {
	devices, err := api.ListDevices()
	if err != nil {
		return err
	}
	newDevices := make([]*npu.NPUDevice, 0, len(devices))
	for i := range devices {
		newDevice := npu.NewNPUDevice(&devices[i])
		newDevices = append(newDevices, newDevice)
	}
	r.devices = newDevices

	rawDriverInfo, err := api.DriverInfo()
	if err != nil {
		return err
	}
	driverInfo := npu.NewVersionInfo(
		rawDriverInfo.Major(),
		rawDriverInfo.Minor(),
		rawDriverInfo.Patch(),
		npu.StringOrEmpty(rawDriverInfo.Metadata()),
		npu.StringOrEmpty(rawDriverInfo.Prerelease()))
	r.driverInfo = driverInfo
	return nil
}

func (r *FeatureDiscoveryRunner) checkLabels() {
	if len(r.devices) == 0 {
		return
	}

	baseFirmwareVersion := r.devices[0].FirmwareInfo
	allMatch := true

	for _, device := range r.devices[1:] {
		if device.FirmwareInfo == nil && baseFirmwareVersion == nil {
			continue
		}
		if device.FirmwareInfo == nil && baseFirmwareVersion != nil {
			allMatch = false
			break
		}
		if !device.FirmwareInfo.Compare(baseFirmwareVersion) {
			allMatch = false
			break
		}
	}

	if !allMatch {
		for _, device := range r.devices {
			device.FirmwareInfo = nil
		}
	}
}

func (r *FeatureDiscoveryRunner) ExtractLabels() map[string]string {
	r.checkLabels()
	labels := make(map[string]string)
	for _, device := range r.devices {
		deviceLabels := device.ToLabels()
		maps.Copy(labels, deviceLabels)
	}
	labels[consts.FuriosaNPUCountLabel] = strconv.Itoa(len(r.devices))
	labels[consts.FuriosaDriverVersionLabel] = r.driverInfo.String()
	labels[consts.FuriosaDriverVersionMajorLabel] = r.driverInfo.Major()
	labels[consts.FuriosaDriverVersionMinorLabel] = r.driverInfo.Minor()
	labels[consts.FuriosaDriverVersionPatchLabel] = r.driverInfo.Patch()
	labels[consts.FuriosaDriverVersionMetadataLabel] = r.driverInfo.Metadata
	labels[consts.FuriosaDriverVersionPrereleaseLabel] = r.driverInfo.Prerelease
	return labels
}

func (r *FeatureDiscoveryRunner) SyncFile(labels map[string]string) (err error) {
	if len(labels) == 0 {
		r.logger.Error().Msg("no labels to sync")
		return nil
	}

	parentDir := filepath.Dir(r.config.OutputPath)
	if parentDir == "" {
		return fmt.Errorf("output path is not a valid file: [%s]", r.config.OutputPath)
	}
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	filename := filepath.Base(r.config.OutputPath)
	if filename == "" {
		return fmt.Errorf("failed to get filename of output file %s", r.config.OutputPath)
	}

	tmpFileName := fmt.Sprintf(".%s-", filename)
	tmpFile, err := os.CreateTemp(parentDir, tmpFileName)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	tmpFilePath := tmpFile.Name()
	defer func() {
		if err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpFilePath)
		}
	}()
	if err = os.Chmod(tmpFile.Name(), 0644); err != nil {
		return fmt.Errorf("failed to set permissions on temporary file: %w", err)
	}

	content := LabelsToFeatureString(labels)
	if _, err = tmpFile.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to temporary file: %w", err)
	}

	if err = tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync temporary file: %w", err)
	}

	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}
	if err = os.Rename(tmpFile.Name(), r.config.OutputPath); err != nil {
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

func (r *FeatureDiscoveryRunner) RemoveFFDFile() error {
	return os.Remove(r.config.OutputPath)
}

func (r *FeatureDiscoveryRunner) tryStartSync() bool {
	if !r.syncMu.TryLock() {
		return false
	}
	defer r.syncMu.Unlock()
	if r.syncRunning {
		return false
	}
	r.syncRunning = true
	return true
}

func (r *FeatureDiscoveryRunner) stopSync() {
	r.syncMu.Lock()
	defer r.syncMu.Unlock()
	r.syncRunning = false
}

func (r *FeatureDiscoveryRunner) Run() error {
	syncTimeout := 30 * time.Second

	// initial sync
	r.logger.Info().Msg("start initial sync")
	ctx, cancel := context.WithTimeout(context.Background(), syncTimeout)
	if err := r.sync(ctx); err != nil {
		r.logger.Error().Err(err).Msg("initial sync failed")
	}
	cancel()

	interval := time.Duration(r.config.UpdateInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	r.logger.Info().Msgf("start feature discovery runner with interval %d seconds", r.config.UpdateInterval)
	for {
		select {
		case <-ticker.C:
			if !r.tryStartSync() {
				r.logger.Info().Msg("sync is already running, skipping")
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), syncTimeout)
			go func() {
				defer cancel()
				defer func() {
					r.stopSync()
				}()
				if err := r.sync(ctx); err != nil {
					if ctx.Err() == context.DeadlineExceeded {
						r.logger.Error().Err(err).Msg("sync operation timed out")
					} else {
						r.logger.Error().Err(err).Msg("sync failed")
					}
				}
			}()
		case sig := <-sigChan:
			r.logger.Info().Msgf("received signal: %v, shutting down", sig)
			if err := r.RemoveFFDFile(); err != nil {
				r.logger.Error().Err(err).Msg("failed to remove FFD file during shutdown")
			}
			return nil
		}
	}
}

func (r *FeatureDiscoveryRunner) sync(ctx context.Context) error {
	err := r.updateDevicesAndDriverInfo()
	if err != nil {
		return err
	}

	err = r.detectNPUDevices()
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	labelMap := r.ExtractLabels()
	err = r.SyncFile(labelMap)
	if err != nil {
		return err
	}
	return nil
}

func (r *FeatureDiscoveryRunner) detectNPUDevices() error {
	r.logger.Info().Msg("start npu device detection")
	if len(r.devices) <= 0 {
		r.logger.Info().Msg("no NPU devices found")
		return fmt.Errorf("no NPU devices found")
	}

	found := 0
	for _, device := range r.devices {
		err := device.DetectNPUDevice()
		if err != nil {
			r.logger.Err(err).Msg("failed to recognize NPU device")
			continue
		}
		found++
	}
	r.logger.Info().Msgf("found %d NPU devices", found)
	return nil
}
