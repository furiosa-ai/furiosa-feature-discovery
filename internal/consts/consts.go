package consts

const (
	DefaultLabelUpdateIntervalSeconds = 60
	DefaultOutputPath                 = "/etc/kubernetes/node-feature-discovery/features.d/ffd"

	ArchRngd    = "rngd"
	ArchRngdS   = "rngd_s"
	ArchRngdMax = "rngd_max"

	FuriosaDomain                         = "furiosa.ai"
	FuriosaNPUFamilyLabel                 = FuriosaDomain + "/npu.family"
	FuriosaNPUProductLabel                = FuriosaDomain + "/npu.product"
	FuriosaNPUCountLabel                  = FuriosaDomain + "/npu.count"
	FuriosaDriverVersionLabel             = FuriosaDomain + "/driver.version"
	FuriosaDriverVersionMajorLabel        = FuriosaDomain + "/driver.version.major"
	FuriosaDriverVersionMinorLabel        = FuriosaDomain + "/driver.version.minor"
	FuriosaDriverVersionPatchLabel        = FuriosaDomain + "/driver.version.patch"
	FuriosaDriverVersionMetadataLabel     = FuriosaDomain + "/driver.version.metadata"
	FuriosaDriverVersionPrereleaseLabel   = FuriosaDomain + "/driver.version.prerelease"
	FuriosaFirmwareVersionLabel           = FuriosaDomain + "/firmware.version"
	FuriosaFirmwareVersionMajorLabel      = FuriosaDomain + "/firmware.version.major"
	FuriosaFirmwareVersionMinorLabel      = FuriosaDomain + "/firmware.version.minor"
	FuriosaFirmwareVersionPatchLabel      = FuriosaDomain + "/firmware.version.patch"
	FuriosaFirmwareVersionMetadataLabel   = FuriosaDomain + "/firmware.version.metadata"
	FuriosaFirmwareVersionPrereleaseLabel = FuriosaDomain + "/firmware.version.prerelease"
)
