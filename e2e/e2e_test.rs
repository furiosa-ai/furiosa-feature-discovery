use assert_cmd::Command;
use k8s_openapi::{api::apps::v1::DaemonSet, api::core::v1::Node};
use kube::{api::Api, api::DeleteParams, Client};
use rand::{distributions::Alphanumeric, Rng};
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
struct Cli {
    /// Interval secs to update node labels
    #[structopt(long)]
    helm_path: String,
    #[structopt(long, default_value = "furiosa-feature-discovery")]
    release_name: String,
    #[structopt(long, default_value = "default")]
    namespace: String,
}

async fn delete_daemonset(daemonset_name: &String, namespace: &str, client: Client) {
    let daemonsets: Api<DaemonSet> = Api::namespaced(client.clone(), namespace);

    let delete_params = DeleteParams::default();

    if daemonsets
        .delete(daemonset_name, &delete_params)
        .await
        .is_ok()
    {
        println!("DaemonSet {} already exists. Delete it", daemonset_name)
    }
}

async fn deploy_helm_chart(
    release_name: &String,
    path: &String,
    namespace: &String,
    interval: u32,
) -> Result<(), std::io::Error> {
    let output = Command::new("helm")
        .arg("install")
        .arg(release_name)
        .arg(path)
        .arg("--set")
        .arg(format!("daemonSet.args.interval={}", interval))
        .arg("--namespace")
        .arg(namespace)
        .output()?;

    if !output.status.success() {
        eprintln!("Helm install failed: {:?}", output);
        return Err(std::io::Error::new(
            std::io::ErrorKind::Other,
            "Helm install failed",
        ));
    }

    println!("Helm install succeeded: {:?}", output);

    Ok(())
}

async fn uninstall_helm_release(
    release_name: &String,
    namespace: &String,
) -> Result<(), std::io::Error> {
    let output = Command::new("helm")
        .arg("uninstall")
        .arg(release_name)
        .arg("--namespace")
        .arg(namespace)
        .output()?;

    if !output.status.success() {
        eprintln!("Helm uninstall failed: {:?}", output);
        return Err(std::io::Error::new(
            std::io::ErrorKind::Other,
            "Helm uninstall failed",
        ));
    }

    println!("Helm uninstall succeeded: {:?}", output);

    Ok(())
}

async fn check_furiosa_label(client: Client) -> Result<bool, Box<dyn std::error::Error>> {
    let nodes: Api<Node> = Api::all(client);

    let node_list = nodes.list(&Default::default()).await?;

    let npu_nodes: Vec<_> = node_list
        .into_iter()
        .filter(|node| {
            if let Some(labels) = &node.metadata.labels {
                labels.iter().any(|(key, value)| {
                    key == "feature.node.kubernetes.io/pci-1200_1ed2.present" && value == "true"
                })
            } else {
                false
            }
        })
        .collect();

    if npu_nodes.is_empty() {
        return Ok(false);
    }

    Ok(npu_nodes.iter().all(|node| {
        if let Some(labels) = &node.metadata.labels {
            labels.keys().any(|label| label == "furiosa.ai/npu.count")
        } else {
            false
        }
    }))
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args = Cli::from_args();

    let helm_path = &args.helm_path;
    let namespace = &args.namespace;
    let release_name = &args.release_name;

    let client = Client::try_default().await?;

    if check_furiosa_label(client.clone()).await? {
        let err = std::io::Error::new(
            std::io::ErrorKind::Other,
            "Furiosa NPU labels already exist",
        );
        return Err(Box::new(err) as Box<dyn std::error::Error>);
    }

    delete_daemonset(release_name, namespace, client.clone()).await;
    delete_daemonset(
        &format!("{}-node-feature-discovery-worker", release_name),
        namespace,
        client.clone(),
    )
    .await;

    let random_str: String = rand::thread_rng()
        .sample_iter(&Alphanumeric)
        .take(5)
        .map(char::from)
        .collect();

    let release_name_rand = format!("{}-e2e-{}", release_name, random_str.to_lowercase());

    deploy_helm_chart(&release_name_rand, helm_path, namespace, 1).await?;
    std::thread::sleep(std::time::Duration::from_secs(120));
    uninstall_helm_release(&release_name_rand, namespace).await?;

    if !check_furiosa_label(client).await? {
        let err = std::io::Error::new(std::io::ErrorKind::Other, "No label updated.");
        return Err(Box::new(err) as Box<dyn std::error::Error>);
    }

    Ok(())
}
