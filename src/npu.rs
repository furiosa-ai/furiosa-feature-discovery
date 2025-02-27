use std::collections::BTreeMap;
use std::fmt;
use std::io::{Error, ErrorKind, Result};

fn recognize_family(arch: &str) -> Result<String> {
    let family = match arch {
        "warboy" => "warboy",
        "rngd" | "rngd_s" | "rngd_max" => "rngd",
        _ => {
            return Err(Error::new(
                ErrorKind::Other,
                format!("Unknown Arch: {}", arch),
            ))
        }
    }
    .to_string();
    Ok(family)
}

fn recognize_product(arch: &str) -> Result<String> {
    Ok(arch.to_string())
}

#[derive(Debug, Clone, PartialEq)]
pub struct VersionInfo {
    major: u32,
    minor: u32,
    patch: u32,
    metadata: String,
}

impl VersionInfo {
    pub fn new(major: u32, minor: u32, patch: u32, metadata: String) -> VersionInfo {
        VersionInfo {
            major,
            minor,
            patch,
            metadata,
        }
    }

    pub fn major(self) -> u32 {
        self.major
    }

    pub fn minor(self) -> u32 {
        self.minor
    }

    pub fn patch(self) -> u32 {
        self.patch
    }

    pub fn metadata(self) -> String {
        self.metadata
    }
}

impl From<furiosa_smi_rs::VersionInfo> for VersionInfo {
    fn from(info: furiosa_smi_rs::VersionInfo) -> Self {
        VersionInfo::new(info.major(), info.minor(), info.patch(), info.metadata())
    }
}

impl fmt::Display for VersionInfo {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}.{}.{}", self.major, self.minor, self.patch)
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct NpuDevice {
    family: String,
    product: String,
    driver_info: VersionInfo,
    pub firmware_info: Option<VersionInfo>,
    pub pert_info: Option<VersionInfo>,
}

impl NpuDevice {
    pub async fn new(
        arch: &str,
        driver_info: VersionInfo,
        firmware_info: Option<VersionInfo>,
        pert_info: Option<VersionInfo>,
    ) -> Result<NpuDevice> {
        let family = recognize_family(arch)?;
        let product = recognize_product(arch)?;

        Ok(NpuDevice {
            family,
            product,
            driver_info,
            firmware_info,
            pert_info,
        })
    }

    pub fn to_labels(&self) -> BTreeMap<String, String> {
        let mut labels: Vec<(String, String)> = vec![
            ("furiosa.ai/npu.family".to_string(), self.family.clone()),
            ("furiosa.ai/npu.product".to_string(), self.product.clone()),
            (
                "furiosa.ai/driver.version".to_string(),
                self.driver_info.to_string(),
            ),
            (
                "furiosa.ai/driver.version.major".to_string(),
                self.driver_info.clone().major().to_string(),
            ),
            (
                "furiosa.ai/driver.version.minor".to_string(),
                self.driver_info.clone().minor().to_string(),
            ),
            (
                "furiosa.ai/driver.version.patch".to_string(),
                self.driver_info.clone().patch().to_string(),
            ),
            (
                "furiosa.ai/driver.version.metadata".to_string(),
                self.driver_info.clone().metadata().clone(),
            ),
        ];

        if let Some(firmware_info) = &self.firmware_info {
            labels.push((
                "furiosa.ai/firmware.version".to_string(),
                firmware_info.to_string(),
            ));
            labels.push((
                "furiosa.ai/firmware.version.major".to_string(),
                firmware_info.clone().major().to_string(),
            ));
            labels.push((
                "furiosa.ai/firmware.version.minor".to_string(),
                firmware_info.clone().minor().to_string(),
            ));
            labels.push((
                "furiosa.ai/firmware.version.patch".to_string(),
                firmware_info.clone().patch().to_string(),
            ));
            labels.push((
                "furiosa.ai/firmware.version.metadata".to_string(),
                firmware_info.clone().metadata().clone(),
            ));
        };

        if let Some(pert_info) = &self.pert_info {
            labels.push(("furiosa.ai/pert.version".to_string(), pert_info.to_string()));
            labels.push((
                "furiosa.ai/pert.version.major".to_string(),
                pert_info.clone().major().to_string(),
            ));
            labels.push((
                "furiosa.ai/pert.version.minor".to_string(),
                pert_info.clone().minor().to_string(),
            ));
            labels.push((
                "furiosa.ai/pert.version.patch".to_string(),
                pert_info.clone().patch().to_string(),
            ));
            labels.push((
                "furiosa.ai/pert.version.metadata".to_string(),
                pert_info.clone().metadata().clone(),
            ));
        };

        labels.into_iter().collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_npu_device_new() {
        let version_info = VersionInfo::new(1, 2, 3, "a1b2c3".to_string());
        let device = NpuDevice::new(
            "warboy",
            version_info.clone(),
            Some(version_info.clone()),
            Some(version_info.clone()),
        )
        .await;
        let expected = NpuDevice {
            family: "warboy".to_string(),
            product: "warboy".to_string(),
            driver_info: version_info.clone(),
            firmware_info: Some(version_info.clone()),
            pert_info: Some(version_info.clone()),
        };

        assert!(device.is_ok());
        assert_eq!(device.unwrap(), expected);
    }

    #[test]
    fn test_recognize_family() {
        let family_warboy = recognize_family("warboy");
        let family_rngd = recognize_family("rngd");

        assert!(family_warboy.is_ok());
        assert!(family_rngd.is_ok());

        assert_eq!(family_warboy.unwrap(), "warboy".to_string());
        assert_eq!(family_rngd.unwrap(), "rngd".to_string());
    }

    #[test]
    fn test_recognize_product() {
        let product_warboy = recognize_product("warboy");
        let product_rngd = recognize_product("rngd");

        assert!(product_warboy.is_ok());
        assert!(product_rngd.is_ok());

        assert_eq!(product_warboy.unwrap(), "warboy".to_string());
        assert_eq!(product_rngd.unwrap(), "rngd".to_string());
    }
}
