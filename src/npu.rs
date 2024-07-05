use std::collections::BTreeMap;
// use std::io;
// use std::io::ErrorKind::Other;
// use std::path::PathBuf;

use lazy_static::lazy_static;
use regex::Regex;
// use tokio::fs::DirEntry;
// use tokio::io::ErrorKind;

lazy_static! {
    static ref U250_PATTERN: Regex =
        Regex::new(".*\\[0700\\]: Xilinx Corporation Device \\[10ee:f10a\\].*").unwrap();
    static ref AWS_F1: Regex =
        Regex::new(".*\\[0580\\]: Xilinx Corporation Device \\[10ee:f10a\\].*").unwrap();
    static ref PLDA: Regex = Regex::new(".*\\[0780\\]: PLDA Device \\[1556:1111\\].*").unwrap();
}

pub enum Family {
    Warboy,
    Rngd,
}

impl Family {
    fn name(&self) -> &'static str {
        match self {
            Family::Warboy => "warboy",
            Family::Rngd => "rngd",
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

fn recognize_family(arch: &str) -> std::io::Result<Family> {
    let family = match arch {
        "warboy" => Family::Warboy,
        "rngd" | "rngd_s" | "rngd_max" => Family::Rngd,
        _ => {
            return Err(std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("Unknown Arch: {}", arch),
            ))
        }
    };
    Ok(family)
}

fn recognize_product(arch: &str) -> std::io::Result<HwType> {
    let hwtype = match arch {
        "warboy" => HwType::Warboy,
        "rngd" => HwType::Rngd,
        "rngd_s" => HwType::RngdS,
        "rngd_max" => HwType::RngdMax,
        _ => {
            return Err(std::io::Error::new(
                std::io::ErrorKind::Other,
                format!("Unknown Arch: {}", arch),
            ))
        }
    };
    Ok(hwtype)
}

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
    ) -> std::io::Result<NpuDevice> {
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

    #[test]
    fn test_json() {
        let mut labels = BTreeMap::new();
        labels.insert("furiosa.ai/npu.family", "warboy");
        labels.insert("furiosa.ai/npu.hwtype", "Warboy");
        assert_eq!(
            r#"{"furiosa.ai/npu.family":"warboy","furiosa.ai/npu.hwtype":"Warboy"}"#,
            &serde_json::to_string(&labels).unwrap()
        );
    }
}
