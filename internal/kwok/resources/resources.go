package resources

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/dejanzele/batch-simulator/internal/util"
)

// NewFakeNode creates a fake Kubernetes Node resource with the specified nodeName.
func NewFakeNode(nodeName string) *corev1.Node {
	return &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Node",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
			Annotations: map[string]string{
				"node.alpha.kubernetes.io/ttl": "0",
				"kwok.x-k8s.io/node":           "fake",
			},
			Labels: map[string]string{
				"beta.kubernetes.io/arch":       "amd64",
				"beta.kubernetes.io/os":         "linux",
				"kubernetes.io/arch":            "amd64",
				"kubernetes.io/hostname":        nodeName,
				"kubernetes.io/os":              "linux",
				"kubernetes.io/role":            "agent",
				"node-role.kubernetes.io/agent": "",
				"type":                          "kwok",
			},
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "kwok.x-k8s.io/node",
					Value:  "fake",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
		Status: corev1.NodeStatus{
			Allocatable: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("32"),
				corev1.ResourceMemory: resource.MustParse("256Gi"),
				corev1.ResourcePods:   resource.MustParse("110"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("32"),
				corev1.ResourceMemory: resource.MustParse("256Gi"),
				corev1.ResourcePods:   resource.MustParse("110"),
			},
			NodeInfo: corev1.NodeSystemInfo{
				Architecture:            "amd64",
				BootID:                  "",
				ContainerRuntimeVersion: "",
				KernelVersion:           "",
				KubeProxyVersion:        "fake",
				KubeletVersion:          "fake",
				MachineID:               "",
				OperatingSystem:         "linux",
				OSImage:                 "",
				SystemUUID:              "",
			},
			Phase: corev1.NodeRunning,
		},
	}
}

func NewFakeJob(name, namespace string) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":  "fake-job",
				"type": "kwok",
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: ptr.To[int32](30),
			Template: corev1.PodTemplateSpec{
				Spec: newPodSpec(),
			},
		},
	}
}

func NewFakePod(name, namespace string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":  "fake-pod",
				"type": "kwok",
			},
		},
		Spec: newPodSpec(),
	}
}

func newPodSpec() corev1.PodSpec {
	return corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyNever,
		Affinity:      newAffinity(),
		Tolerations: []corev1.Toleration{
			{
				Key:      "kwok.x-k8s.io/node",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			},
		},
		Containers: []corev1.Container{
			{
				Name:  "fake-container",
				Image: "fake-image",
				Env:   newEnvVars(4, 20*1024),
			},
		},
	}
}

func newAffinity() *corev1.Affinity {
	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "type",
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{"kwok"},
							},
						},
					},
				},
			},
		},
	}
}

func newEnvVars(count, size int32) []corev1.EnvVar {
	envVars := make([]corev1.EnvVar, 0, count)
	for i := 0; i < int(count); i++ {
		envVars = append(envVars, corev1.EnvVar{
			Name:  fmt.Sprintf("SOME_ENV_VAR_%d", i),
			Value: util.RandomText(size),
		})

	}
	return envVars
}
