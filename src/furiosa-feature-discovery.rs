use std::collections::BTreeMap;
use std::fs::File;
/// This command will be running in all nodes to detect Furiosa AI's NPU devices in Host machine.
/// It labels Kubernetes Nodes with properties obtained from detected Furiosa AI's NPU devices.
use std::fs::{create_dir_all, remove_file};
use std::io::Write;
use std::path::Path;
use std::time::Duration;

use structopt::StructOpt;
use tokio::signal::unix::{signal, SignalKind};
use tokio::time;

use crate::npu::NpuDevice;

mod npu;

macro_rules! expr {
    ($e: expr) => {
        $e
    };
}

macro_rules! defer {
    ($($data: tt)*) => (
        let _scope_call = ScopeCall {
            c: || -> () { expr!({ $($data)* }) }
        };
    )
}

struct ScopeCall<F: FnMut()> {
    c: F,
}
impl<F: FnMut()> Drop for ScopeCall<F> {
    fn drop(&mut self) {
        (self.c)();
    }
}

#[derive(Debug, StructOpt)]
struct Cli {
    /// Interval secs to update node labels
    #[structopt(long, default_value = "60")]
    interval: u64,
    #[structopt(default_value = "/etc/kubernetes/node-feature-discovery/features.d/nfd")]
    output: String,
}

#[allow(clippy::into_iter_on_ref)]
fn labels_to_feature(map: &BTreeMap<String, String>) -> String {
    map.into_iter()
        .map(|(k, v)| format!("{}={}", k, v))
        .collect::<Vec<String>>()
        .join("\n")
}

async fn write_node_labels(devices: Vec<NpuDevice>, output_path: &Path) -> anyhow::Result<()> {
    if !devices.is_empty() {
        let device_count = devices.len();

        let mut labels = devices.into_iter().flat_map(|d| d.to_labels()).fold(
            BTreeMap::new(),
            |mut acc, label| {
                acc.insert(label.0, label.1);
                acc
            },
        );

        labels.insert("furiosa.ai/npu.count".to_string(), device_count.to_string());

        let nfd = labels_to_feature(&labels);

        log::info!("Labels updated: \n{}", nfd);

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

        let mut output_file = File::create(output_path)?;
        output_file.write_all(nfd.as_bytes())?;
    } else {
        log::info!("No devices found.");
    }
    Ok(())
}

async fn detect_npu_devices() -> anyhow::Result<Vec<NpuDevice>> {
    let mut found = vec![];

    let devices = furiosa_smi_rs::list_devices()?;

    for device in &devices {
        let device_info = device.device_info()?;
        let arch = device_info.arch().to_string();
        let driver = device_info.driver_version();
        let driver_major = driver.major();
        let driver_minor = driver.minor();
        let driver_patch = driver.patch();
        let driver_metadata = driver.metadata();

        match NpuDevice::new(
            &arch,
            driver_major,
            driver_minor,
            driver_patch,
            driver_metadata,
        )
        .await
        {
            Ok(device) => found.push(device),
            Err(e) => log::error!("Failed to recognize device: {}", e),
        };
    }
    log::trace!("Found {} NPU devices", devices.len());
    Ok(found)
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    env_logger::init();

    log::info!("furiosa-feature-discovery has started");
    defer! {
        log::info!("furiosa-feature-discovery has finished")
    }

    let args = Cli::from_args();

    let interval = time::interval(Duration::from_secs(args.interval));

    let output = args.output;
    let output_path = Path::new(&output);
    log::info!("Writing labels to output file {}", output);

    run_loop(output_path, interval).await
}

async fn run_loop(output_path: &Path, mut interval: time::Interval) -> anyhow::Result<()> {
    defer! {
        let _ = remove_nfd(output_path);
    }
    log::info!("Start to write labels");
    let mut sigterm = signal(SignalKind::terminate())?;
    let mut sigint = signal(SignalKind::interrupt())?;
    let mut sigquit = signal(SignalKind::quit())?;

    loop {
        tokio::select! {
            _ = sigterm.recv() => {log::trace!("SIGTERM Shuting down"); break},
            _ = sigint.recv() => {log::trace!("SIGINT Shuting down"); break},
            _ = sigquit.recv() => {log::trace!("SIGQUIT Shuting down"); break},
            _ = interval.tick() => {
                log::info!("Start to detect npu devices");
                let detected = match detect_npu_devices().await {
                    Ok(dev) => dev,
                    Err(e) => {log::error!("Failed to get device information: {}", e); break},
                };
                match write_node_labels(detected, output_path).await {
                    Ok(()) => {}
                    Err(e) => {log::error!("Failed to write node labels: {}", e); break},
                }
            }
        }
    }
    Ok(())
}

fn remove_nfd(output_path: &Path) -> anyhow::Result<()> {
    remove_file(output_path)?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_labels_to_feature() {
        let mut labels = BTreeMap::new();
        labels.insert("furiosa.ai/npu.family".to_string(), "Warboy".to_string());
        labels.insert("furiosa.ai/npu.hwtype".to_string(), "Warboy".to_string());

        let feature = labels_to_feature(&labels);
        let expected = "furiosa.ai/npu.family=Warboy\nfuriosa.ai/npu.hwtype=Warboy".to_string();

        assert_eq!(expected, feature)
    }
}
