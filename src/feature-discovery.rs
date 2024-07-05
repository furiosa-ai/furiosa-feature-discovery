use std::collections::BTreeMap;
/// This command will be running in all nodes to detect Furiosa AI's NPU devices in Host machine.
/// It labels Kubernetes Nodes with properties obtained from detected Furiosa AI's NPU devices.
use std::env;
use std::time::Duration;

use anyhow::{bail, Context};
use lazy_static::lazy_static;
use structopt::StructOpt;
use tokio::fs;
use tokio::time;

use crate::npu::NpuDevice;

mod npu;

lazy_static! {
    static ref K8S_API_NODE_URL: String = format!(
        "https://{}:{}/api/v1/nodes/{}",
        env::var("KUBERNETES_SERVICE_HOST").expect("KUBERNETES_SERVICE_HOST must be set"),
        env::var("KUBERNETES_SERVICE_PORT_HTTPS")
            .expect("KUBERNETES_SERVICE_PORT_HTTPS must be set"),
        env::var("K8S_NODE_NAME").expect("K8S_NODE_NAME must be set")
    );
}

static K8S_TOKEN_PATH: &str = "/var/run/secrets/kubernetes.io/serviceaccount/token";

#[derive(Debug, StructOpt)]
struct Cli {
    /// Interval secs to update node labels
    #[structopt(long, default_value = "60")]
    interval: u64,
}

async fn update_node_labels(token: &str, devices: Vec<NpuDevice>) -> anyhow::Result<()> {
    if !devices.is_empty() {
        let client = reqwest::Client::builder()
            .danger_accept_invalid_certs(true)
            .build()
            .with_context(|| "fail to create HTTP client")?;

        let device_count = devices.len();

        let mut labels = devices.into_iter().flat_map(|d| d.to_labels()).fold(
            BTreeMap::new(),
            |mut acc, label| {
                acc.insert(label.0, label.1);
                acc
            },
        );

        labels.insert("furiosa.ai/npu.count".to_string(), device_count.to_string());

        let labels =
            &serde_json::to_string(&labels).with_context(|| "fail to serialize labels to JSON")?;
        let mut str_buf = String::from(r#"{"metadata": {"labels": "#);
        str_buf.push_str(labels);
        str_buf.push_str("}}");

        log::info!("Labels updated: \n{}", &str_buf);
        let resp = client
            .patch(K8S_API_NODE_URL.as_str())
            .bearer_auth(token)
            .header("Content-Type", "application/merge-patch+json")
            .body(str_buf)
            .send()
            .await
            .with_context(|| format!("failed to connect {}", K8S_API_NODE_URL.as_str()))?;

        if resp.status().is_success() {
            Ok(())
        } else {
            let canonical_reason = resp.status().canonical_reason().unwrap_or("Unknown");
            let status_code = resp.status().as_u16();
            let message = String::from(&resp.text().await.unwrap());
            bail!(
                "fail to update the label (status: {} ({}), message: {})",
                canonical_reason,
                status_code,
                message
            );
        }
    } else {
        Ok(())
    }
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
            Err(e) => log::error!("fail to recognize device: {}", e),
        };
    }
    log::trace!("Found {} NPU devices", devices.len());
    Ok(found)
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    env_logger::init();

    log::info!("furiosa-node-labeller has started");
    log::info!("URL: {}", K8S_API_NODE_URL.as_str());

    let args = Cli::from_args();
    let token = fs::read_to_string(&K8S_TOKEN_PATH)
        .await
        .with_context(|| format!("fail to read the kubernetes token at {}", K8S_TOKEN_PATH))?;
    let mut interval = time::interval(Duration::from_secs(args.interval));
    loop {
        interval.tick().await;
        let detected = detect_npu_devices().await?;
        update_node_labels(&token, detected).await?;
    }
}

#[test]
fn test_k8s_api_node_url() {
    std::env::set_var("KUBERNETES_SERVICE_HOST", "10.10.0.1");
    std::env::set_var("KUBERNETES_SERVICE_PORT_HTTPS", "443");
    std::env::set_var("K8S_NODE_NAME", "kube00");
    assert_eq!(
        K8S_API_NODE_URL.as_str(),
        "https://10.10.0.1:443/api/v1/nodes/kube00"
    );
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_json() {
        let mut labels = BTreeMap::new();
        labels.insert("furiosa.ai/npu.family", "warboy");
        labels.insert("furiosa.ai/npu.hwtype", "warboy");
        assert_eq!(
            r#"{"furiosa.ai/npu.family":"warboy","furiosa.ai/npu.hwtype":"warboy"}"#,
            &serde_json::to_string(&labels).unwrap()
        );
    }
}
