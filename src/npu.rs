use std::collections::BTreeMap;
use std::io::{Error, ErrorKind, Result};

#[derive(Debug, PartialEq)]
pub enum Family {
    Warboy,
    Rngd,
}

impl Family {
    fn name(&self) -> &'static str {
        match self {
            Family::Warboy => "Warboy",
            Family::Rngd => "Rngd",
        }
    }

    pub fn to_label(&self) -> (String, String) {
        ("furiosa.ai/npu.family".to_string(), self.name().to_string())
    }
}

#[derive(Debug, Eq, PartialEq)]
pub enum HwType {
    Warboy,
    Rngd,
    RngdS,
    RngdMax,
}

impl HwType {
    pub fn name(&self) -> &'static str {
        match self {
            HwType::Warboy => "Warboy",
            HwType::Rngd => "Rngd",
            HwType::RngdS => "RngdS",
            HwType::RngdMax => "RngdMax",
        }
    }

    pub fn to_label(&self) -> (String, String) {
        ("furiosa.ai/npu.hwtype".to_string(), self.name().to_string())
    }
}

fn recognize_family(arch: &str) -> Result<Family> {
    let family = match arch {
        "warboy" => Family::Warboy,
        "rngd" | "rngd_s" | "rngd_max" => Family::Rngd,
        _ => {
            return Err(Error::new(
                ErrorKind::Other,
                format!("Unknown Arch: {}", arch),
            ))
        }
    };
    Ok(family)
}

fn recognize_product(arch: &str) -> Result<HwType> {
    let hwtype = match arch {
        "warboy" => HwType::Warboy,
        "rngd" => HwType::Rngd,
        "rngd_s" => HwType::RngdS,
        "rngd_max" => HwType::RngdMax,
        _ => {
            return Err(Error::new(
                ErrorKind::Other,
                format!("Unknown Arch: {}", arch),
            ))
        }
    };
    Ok(hwtype)
}

#[derive(Debug, PartialEq)]
pub struct NpuDevice {
    family: Family,
    hwtype: HwType,
    driver_major: u32,
    driver_minor: u32,
    driver_patch: u32,
    driver_metadata: String,
}

impl NpuDevice {
    pub async fn new(
        arch: &str,
        driver_major: u32,
        driver_minor: u32,
        driver_patch: u32,
        driver_metadata: String,
    ) -> Result<NpuDevice> {
        let family = recognize_family(arch)?;
        let hwtype = recognize_product(arch)?;

        Ok(NpuDevice {
            family,
            hwtype,
            driver_major,
            driver_minor,
            driver_patch,
            driver_metadata,
        })
    }

    pub fn to_labels(&self) -> BTreeMap<String, String> {
        let labels: Vec<(String, String)> = vec![
            self.family.to_label(),
            self.hwtype.to_label(),
            (
                "furiosa.ai/driver.version".to_string(),
                format!(
                    "{}.{}.{}",
                    self.driver_major, self.driver_minor, self.driver_patch
                ),
            ),
            (
                "furiosa.ai/driver.version.major".to_string(),
                self.driver_major.to_string(),
            ),
            (
                "furiosa.ai/driver.version.minor".to_string(),
                self.driver_minor.to_string(),
            ),
            (
                "furiosa.ai/driver.version.patch".to_string(),
                self.driver_patch.to_string(),
            ),
            (
                "furiosa.ai/driver.version.metadata".to_string(),
                self.driver_metadata.clone(),
            ),
        ];
        labels.into_iter().collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_npu_device_new() {
        let device = NpuDevice::new("warboy", 1, 2, 3, "a1b2c3".to_string()).await;
        let expected = NpuDevice {
            family: Family::Warboy,
            hwtype: HwType::Warboy,
            driver_major: 1,
            driver_minor: 2,
            driver_patch: 3,
            driver_metadata: "a1b2c3".to_string(),
        };

        assert!(device.is_ok());
        assert_eq!(expected, device.unwrap());
    }

    #[test]
    fn test_recognize_family() {
        let family_warboy = recognize_family("warboy");
        let family_rngd = recognize_family("rngd");

        assert!(family_warboy.is_ok());
        assert!(family_rngd.is_ok());

        assert_eq!(family_warboy.unwrap(), Family::Warboy);
        assert_eq!(family_rngd.unwrap(), Family::Rngd);
    }

    #[test]
    fn test_recognize_product() {
        let product_warboy = recognize_product("warboy");
        let product_rngd = recognize_product("rngd");

        assert!(product_warboy.is_ok());
        assert!(product_rngd.is_ok());

        assert_eq!(product_warboy.unwrap(), HwType::Warboy);
        assert_eq!(product_rngd.unwrap(), HwType::Rngd);
    }

    #[test]
    fn test_family_to_label() {
        let family_warboy = recognize_family("warboy");
        let family_rngd = recognize_family("rngd");

        assert!(family_warboy.is_ok());
        assert!(family_rngd.is_ok());

        assert_eq!(
            family_warboy.unwrap().to_label(),
            ("furiosa.ai/npu.family".to_string(), "Warboy".to_string())
        );
        assert_eq!(
            family_rngd.unwrap().to_label(),
            ("furiosa.ai/npu.family".to_string(), "Rngd".to_string())
        );
    }

    #[test]
    fn test_hwtype_to_label() {
        let product_warboy = recognize_product("warboy");
        let product_rngd = recognize_product("rngd");

        assert!(product_warboy.is_ok());
        assert!(product_rngd.is_ok());

        assert_eq!(
            product_warboy.unwrap().to_label(),
            ("furiosa.ai/npu.hwtype".to_string(), "Warboy".to_string())
        );
        assert_eq!(
            product_rngd.unwrap().to_label(),
            ("furiosa.ai/npu.hwtype".to_string(), "Rngd".to_string())
        );
    }
}
