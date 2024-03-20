# batch-simulator

The batch-simulator, a robust CLI tool developed in Golang, offers a realistic simulation of Kubernetes API resource lifecycles, including Nodes and Pods, through integration with KWOK.
Designed to assist developers in understanding and testing Kubernetes environments without the overhead of real cluster deployment.

## References
* [Kubernetes](https://kubernetes.io/)
* [kind](https://kind.sigs.k8s.io/)
* [KWOK](https://kwok.sigs.k8s.io/)

## Architecture

At its core, the simulator leverages KWOK alongside the [Stages API](https://kwok.sigs.k8s.io/docs/user/stages-configuration/) to orchestrate complex Kubernetes resource lifecycles.

Users define simulation parameters such as number of nodes, number of pods, frequency of pod creation, node creation, requests per iteration...
The simulator then starts creating the Kubernetes resources and KWOK handles their lifecycle.

## Installation

To install the batch-simulator, simply execute `make build`.
This command compiles the source code into an executable binary, conveniently located in `./bin/batchsim`.

For troubleshooting installation issues, refer to the Installation FAQ section.

## Usage

For a deep dive into using the simulator, the docs/ folder contains detailed command descriptions.
Alternatively, appending --help to any command, like `./bin/batchsim run --help`, reveals usage instructions and options.

Here’s an example:
```bash
$ ./bin/batchsim --help

This command-line interface (CLI) tool facilitates the simulation of batch scheduling scenarios,
leveraging Kubernetes (k8s) and Kwok technologies.
It's designed for users who need to model and understand various batch processing workflows within a k8s environment.

Usage:
  sim [flags]
  sim [command]

Available Commands:
  check       Check are required components installed & configured
  clean       Clean deletes all resources (nodes, pods...) created by the simulator
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  install     Install required simulator components
  remove      Uninstall simulator components
  run         Run a simulation

Flags:
  -d, --debug     enable debug output
  -h, --help      help for sim
      --no-gui    disable printing graphical elements
  -s, --silent    disable internal logging
  -v, --verbose   enable verbose output

Use "sim [command] --help" for more information about a command.
```

1. Prepare a Kubernetes cluster (e.g. [kind](https://kind.sigs.k8s.io/)).
2. Run `./bin/batchsim install` to install required simulator components.
3. Run `./bin/batchsim check` to check if required components are installed & configured.
4. Run `./bin/batchsim run` to run a simulation.
5. Run `./bin/batchsim clean` to clean up all resources created by the simulator.

## Development

### Linting

Lint rules are defined in the `.gomodguard.yml` file and [golangci-lint](https://github.com/golangci/golangci-lint) is used to enforce them.

Run `make lint` to run the linter or `make lint-fix` to run the linter and fix lint errors.

### Testing

Run `make test-unit` to run unit tests and `make test-integration` to run integration tests, or run `make test` to run all tests.

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)
