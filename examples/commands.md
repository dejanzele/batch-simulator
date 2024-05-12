# Commands

This is a list of all the commands that are available in CLI.

## Nodes

Run the following command to create fake Nodes in the clsuter

```bash
sim run --node-creator-limit 1200 --node-creator-requests 100 --no-gui -v -r 
```

## Pods

Run the following command to create fake Pods in the clsuter

```bash
sim run --pod-creator-limit 1200 --pod-creator-requests 100 --no-gui -v -r 
```

## Jobs

Run the following command to create fake Jobs in the clsuter

```bash
sim run --job-creator-limit 1200 --job-creator-requests 100 --no-gui -v -r 
```

## Cleanup

Run the following command to cleanup resources in the cluster.
Use the `--resources` flag to specify the resources to cleanup: `pods`, `jobs`, `nodes` or empty for all resources.

```bash
sim cleanup --resources pods,jobs
```
