## batchsim check

Check are required components installed & configured

### Synopsis

This command conducts comprehensive checks for essential components necessary for the system's operation,
including the presence of 'kubectl', 'kwok', and various stages.
It ensures that all required tools and configurations are in place and functioning correctly,
offering a quick and efficient way to validate the setup.

```
batchsim check [flags]
```

### Options

```
  -h, --help                   help for check
      --kube-api-burst int     Maximum burst for throttle while talking with Kubernetes API (default 2000)
      --kube-api-qps float32   Maximum QPS to use while talking with Kubernetes API (default 2000)
  -k, --kubeconfig string      absolute path to the kubeconfig file (default "/Users/zele/.kube/config")
  -v, --verbose                verbose output
```

### Options inherited from parent commands

```
  -d, --debug    enable debug output
      --no-gui   disable printing graphical elements
  -s, --silent   disable internal logging
```

### SEE ALSO

* [batchsim](batchsim.md)	 - kwok-based batch simulation tool

###### Auto generated by spf13/cobra on 11-Jan-2024
