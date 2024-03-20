# Batch Simulator

The batchsim CLI tool is a simulator for batch scheduling scenarios, leveraging Kubernetes (k8s) and Kwok technologies. It's designed for users who need to model and understand various batch processing workflows within a k8s environment.

## Prerequisites

* [Golang](https://go.dev/doc/install) - Go is an open-source programming language that makes it easy to build simple, reliable, and efficient software.
* [kubectl](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) - kubectl is a command-line tool that allows you to run commands against Kubernetes clusters.
* [kind](https://kind.sigs.k8s.io/) - kind is a tool for running local Kubernetes clusters using Docker container “nodes”.
    * Install with `go`: `go install sigs.k8s.io/kind@v0.22.0`
For sanity checking, you can run the following commands:

```bash
$ go version
go version go1.22.0 darwin/arm64

$ kind version
kind v0.22.0 go1.21.7 darwin/arm64

$ kubectl version
Client Version: v1.28.4
Kustomize Version: v5.0.4-0.20230601165947-6ce0bf390ce3
Server Version: v1.29.2
```

## Demo

The CLI can either be downloaded from the GitHub Releases page or built from source.
1. Download the latest release from the [GitHub Releases page](https://github.com/dejanzele/batch-simulator/releases/tag/v0.2.0) for your operating system and architecture.
   (for convenience, the following instructions assume you have downloaded the binary to the `./bin` directory)
2. Run `make build` to compile the source code into an executable binary, conveniently located in `./bin/batchsim`.


### Create a Kubernetes cluster

In order to start, we need a Kubernetes cluster, so we create one using kind.
Feel free to skip this step if you already have a Kubernetes cluster.

```bash
kind create cluster --name simulator
```

Let's check the state of our current Kubernetes cluster

```bash
kubectl get nodes
```

If everything is ok, you should see something like this:

```
NAME                      STATUS   ROLES           AGE   VERSION
simulator-control-plane   Ready    control-plane   25h   v1.29.2
```

### Install simulator

First, we run batchsim install which will install the required simulator components

```bash
./bin/batchsim install
```

### Using the simulator

After that, we create some fake Nodes which we attach to our Kubernetes cluster

```bash
./bin/batchsim run --node-creator-limit 100 --node-creator-requests 10
```

Let's check did the Nodes get created

```bash
kubectl get nodes
```

Now we can run some fake Pods on our fake Nodes

```bash
./bin/batchsim run --pod-creator-limit 500 --pod-creator-requests 20
```

Let's check did the Pods get created

```bash
kubectl get pods
```

### Clean up

If we want to delete resources created by the simulator (Nodes, Pods, Jobs...), we can run the following command

```bash
./bin/batchsim clean
```

## Contributing

Pull requests are always welcome!

For major changes, please open an issue first to discuss what you would like to change.

Ideas for contributions:
* Add more simulation scenarios with various flag options (i.e. using the `--randomize-env-vars`)
* Add more tests for the simulator components
* Add support for [JobSet](https://github.com/kubernetes-sigs/jobset) CRD
