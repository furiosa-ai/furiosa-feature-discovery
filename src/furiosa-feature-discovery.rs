use std::collections::BTreeMap;
/// This command will be running in all nodes to detect Furiosa AI's NPU devices in Host machine.
/// It labels Kubernetes Nodes with properties obtained from detected Furiosa AI's NPU devices.
use std::fs::{create_dir_all, remove_file};
use std::io::Write;
use std::os::unix::fs::PermissionsExt;
use std::path::Path;
use std::time::Duration;

use structopt::StructOpt;
use tempfile::Builder;
use tokio::signal::unix::{signal, SignalKind};
use tokio::time;

use crate::npu::{NpuDevice, VersionInfo};

mod npu;

#[derive(Debug, StructOpt)]
struct Cli {
    /// Interval secs to update node labels
    #[structopt(long, default_value = "60")]
    interval: u64,
    #[structopt(
        long,
        default_value = "/etc/kubernetes/node-feature-discovery/features.d/ffd"
    )]
    output: String,
}

#[allow(clippy::into_iter_on_ref)]
fn labels_to_feature(map: &BTreeMap<String, String>) -> String {
    map.into_iter()
        .map(|(k, v)| format!("{}={}", k, v))
        .collect::<Vec<String>>()
        .join("\n")
}

fn check_labels(devices: Vec<NpuDevice>) -> Vec<NpuDevice> {
    let mut result = devices.clone();
    if let Some(first) = devices.first() {
        if !devices
            .iter()
            .all(|x| x.firmware_info == first.firmware_info)
        {
            log::info!("Some devices have different values for firmware version");
            result = result
                .into_iter()
                .map(|mut x| {
                    x.firmware_info = None;
                    x
                })
                .collect();
        }
        if !devices.iter().all(|x| x.pert_info == first.pert_info) {
            log::info!("Some devices have different values for pert version");
            result = result
                .into_iter()
                .map(|mut x| {
                    x.pert_info = None;
                    x
                })
                .collect();
        }
    }

    result
}

async fn extract_labels(devices: Vec<NpuDevice>) -> anyhow::Result<BTreeMap<String, String>> {
    log::info!("Start to extract node labels");
    if !devices.is_empty() {
        let device_count = devices.len();

        let devices = check_labels(devices);

        let mut labels = devices.into_iter().flat_map(|d| d.to_labels()).fold(
            BTreeMap::new(),
            |mut acc, label| {
                acc.insert(label.0, label.1);
                acc
            },
        );

        labels.insert("furiosa.ai/npu.count".to_string(), device_count.to_string());

        log::info!("Successfully extract node labels");

        Ok(labels)
    } else {
        log::info!("No devices found");
        Ok(BTreeMap::new())
    }
}

fn sync_file_atomically(
    labels: BTreeMap<String, String>,
    output_path: &Path,
) -> anyhow::Result<()> {
    if !labels.is_empty() {
        log::info!("Writing labels to output file: {}", output_path.display());
        let nfd = labels_to_feature(&labels);

        log::info!("Labels updated:\n{}", nfd);

        let parent_dir = match output_path.parent() {
            Some(dir) => dir,
            None => {
                return Err(anyhow::Error::msg(format!(
                    "failed to get parent directory of {}",
                    output_path.display()
                )))
            }
        };
        create_dir_all(parent_dir)?;

        let output_filename = match output_path.file_name() {
            Some(filename) => match filename.to_str() {
                Some(str) => str,
                None => {
                    return Err(anyhow::Error::msg(format!(
                        "failed to get filename of output file {}",
                        output_path.display()
                    )))
                }
            },
            None => {
                return Err(anyhow::Error::msg(format!(
                    "failed to get filename of output file {}",
                    output_path.display()
                )))
            }
        };

        let mut temp_file = Builder::new()
            .prefix(&format!(".{}-", output_filename))
            .permissions(std::fs::Permissions::from_mode(0o644))
            .tempfile_in(parent_dir)?;

        temp_file.write_all(nfd.as_bytes())?;

        std::fs::rename(temp_file, output_path)?;

        log::info!("Successfully write node labels");
    } else {
        log::info!("No labels found");
    }
    Ok(())
}

fn remove_ffd(output_path: &Path) -> anyhow::Result<()> {
    if output_path.is_file() {
        remove_file(output_path)?;
    }
    Ok(())
}

async fn run_loop(output_path: &Path, interval: u64) -> anyhow::Result<()> {
    log::info!("Start to write labels");

    let mut sigterm = signal(SignalKind::terminate())?;
    let mut sigint = signal(SignalKind::interrupt())?;
    let mut sigquit = signal(SignalKind::quit())?;

    let mut interval = time::interval(Duration::from_secs(interval));

    loop {
        tokio::select! {
            _ = sigterm.recv() => {log::trace!("SIGTERM Shuting down"); break},
            _ = sigint.recv() => {log::trace!("SIGINT Shuting down"); break},
            _ = sigquit.recv() => {log::trace!("SIGQUIT Shuting down"); break},
            _ = interval.tick() => match sync_label(output_path).await {
                Ok(()) => {},
                Err(_) => {
                    log::error!("Failed to write node labels");
                    break
                },
            }
        }
    }

    remove_ffd(output_path)?;

    log::info!("Finish writing labels");

    Ok(())
}

async fn sync_label(output_path: &Path) -> anyhow::Result<()> {
    let detected = match detect_npu_devices().await {
        Ok(dev) => dev,
        Err(e) => {
            log::error!("Failed to get device information: {}", e);
            return Err(e);
        }
    };

    match extract_labels(detected).await {
        Ok(labels) => match sync_file_atomically(labels, output_path) {
            Ok(()) => {}
            Err(e) => {
                log::error!("Failed to write node labels: {}", e);
                return Err(e);
            }
        },
        Err(e) => {
            log::error!("Failed to extract node labels: {}", e);
            return Err(e);
        }
    }
    Ok(())
}

async fn detect_npu_devices() -> anyhow::Result<Vec<NpuDevice>> {
    log::info!("Start to detect npu devices");
    let mut found = vec![];

    let devices = furiosa_smi_rs::list_devices()?;

    let driver = furiosa_smi_rs::driver_info()?;
    let driver_info = VersionInfo::from(driver);

    for device in &devices {
        let device_info = match device.device_info() {
            Ok(info) => info,
            Err(e) => {
                log::error!("Failed to get device information: {}", e);
                continue;
            }
        };
        let arch = device_info.arch().to_string();

        let firmware = device_info.firmware_version();
        let firmware_info = VersionInfo::from(firmware);

        let pert = device_info.pert_version();
        let pert_info = VersionInfo::from(pert);

        match NpuDevice::new(
            &arch,
            driver_info.clone(),
            Some(firmware_info),
            Some(pert_info),
        )
        .await
        {
            Ok(device) => found.push(device),
            Err(e) => log::error!("Failed to recognize device: {}", e),
        };
    }

    log::info!("Found {} NPU devices", found.len());

    Ok(found)
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    env_logger::init();

    log::info!("furiosa-feature-discovery has started");

    let args = Cli::from_args();

    let output = args.output;
    let output_path = Path::new(&output);

    furiosa_smi_rs::init()?;
    run_loop(output_path, args.interval).await?;

    log::info!("furiosa-feature-discovery has finished");

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_labels_to_feature() {
        let mut labels = BTreeMap::new();
        labels.insert("furiosa.ai/npu.family".to_string(), "Warboy".to_string());
        labels.insert("furiosa.ai/npu.product".to_string(), "Warboy".to_string());

        let feature = labels_to_feature(&labels);
        let expected = "furiosa.ai/npu.family=Warboy\nfuriosa.ai/npu.product=Warboy".to_string();

        assert_eq!(feature, expected)
    }

    #[tokio::test]
    async fn test_extract_labels() {
        let version_info = VersionInfo::new(1, 2, 3, "1a2b3c".to_string());
        let device_warboy = NpuDevice::new(
            "warboy",
            version_info.clone(),
            Some(version_info.clone()),
            Some(version_info.clone()),
        )
        .await
        .unwrap();

        let labels = extract_labels(vec![device_warboy]).await.unwrap();

        let mut expected = BTreeMap::new();
        expected.insert(
            "furiosa.ai/firmware.version".to_string(),
            "1.2.3".to_string(),
        );
        expected.insert(
            "furiosa.ai/firmware.version.major".to_string(),
            1.to_string(),
        );
        expected.insert(
            "furiosa.ai/firmware.version.minor".to_string(),
            2.to_string(),
        );
        expected.insert(
            "furiosa.ai/firmware.version.patch".to_string(),
            3.to_string(),
        );
        expected.insert(
            "furiosa.ai/firmware.version.metadata".to_string(),
            "1a2b3c".to_string(),
        );
        expected.insert("furiosa.ai/driver.version".to_string(), "1.2.3".to_string());
        expected.insert("furiosa.ai/driver.version.major".to_string(), 1.to_string());
        expected.insert("furiosa.ai/driver.version.minor".to_string(), 2.to_string());
        expected.insert("furiosa.ai/driver.version.patch".to_string(), 3.to_string());
        expected.insert(
            "furiosa.ai/driver.version.metadata".to_string(),
            "1a2b3c".to_string(),
        );
        expected.insert("furiosa.ai/pert.version".to_string(), "1.2.3".to_string());
        expected.insert("furiosa.ai/pert.version.major".to_string(), 1.to_string());
        expected.insert("furiosa.ai/pert.version.minor".to_string(), 2.to_string());
        expected.insert("furiosa.ai/pert.version.patch".to_string(), 3.to_string());
        expected.insert(
            "furiosa.ai/pert.version.metadata".to_string(),
            "1a2b3c".to_string(),
        );
        expected.insert("furiosa.ai/npu.family".to_string(), "warboy".to_string());
        expected.insert("furiosa.ai/npu.product".to_string(), "warboy".to_string());
        expected.insert("furiosa.ai/npu.count".to_string(), 1.to_string());

        assert_eq!(labels, expected);
    }
}
