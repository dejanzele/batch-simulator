package test

import (
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

func GetKubeconfig() *string {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfig = ""
	}
	return &kubeconfig
}
