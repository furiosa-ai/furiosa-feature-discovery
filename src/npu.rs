use std::collections::BTreeMap;
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

#[derive(Debug, PartialEq)]
pub struct NpuDevice {
    family: String,
    product: String,
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
        let product = recognize_product(arch)?;

        Ok(NpuDevice {
            family,
            product,
            driver_major,
            driver_minor,
            driver_patch,
            driver_metadata,
        })
    }

    pub fn to_labels(&self) -> BTreeMap<String, String> {
        let labels: Vec<(String, String)> = vec![
            ("furiosa.ai/npu.family".to_string(), self.family.clone()),
            ("furiosa.ai/npu.product".to_string(), self.product.clone()),
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
            family: "warboy".to_string(),
            product: "warboy".to_string(),
            driver_major: 1,
            driver_minor: 2,
            driver_patch: 3,
            driver_metadata: "a1b2c3".to_string(),
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
