package resources

import (
	"fmt"
	"math/rand"
	"os"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/dejanzele/batch-simulator/internal/util"
)

const (
	LabelKeyApp          = "app"
	LabelValueFakeJob    = "fake-job"
	LabelValueFakePod    = "fake-pod"
	LabelSelectorFakePod = LabelKeyApp + "=" + LabelValueFakePod
)
var (
	// EnvVarCount is the number of envvars in a pod spec.
	EnvVarCount = 5
	// MaxEnvVarSize is the maximum size of an env var in bytes.
	MaxEnvVarSize = 10 * 1024
	// EnvVarsType is the type of env vars that should be used when creating fake pods (nano, micro, xsmall...).
	EnvVarsType   = newEnvVars(EnvVarCount, 2*1024, "SOME_ENV_VAR_MEDIUM")
)

func SetDefaultEnvVarsType(envVarType string) {
	EnvVarsType = GetEnvVars(envVarType)
}

func GetEnvVars(envVarType string) []corev1.EnvVar {
	switch envVarType {
	case "nano":
		return newEnvVars(EnvVarCount, 100, "SOME_ENV_VAR_NANO")
	case "micro":
		return newEnvVars(EnvVarCount, 200, "SOME_ENV_VAR_MICRO")
	case "xsmall":
		return newEnvVars(EnvVarCount, 500, "SOME_ENV_VAR_XSMALL")
	case "small":
		return newEnvVars(EnvVarCount, 1024, "SOME_ENV_VAR_SMALL")
	case "medium":
		return newEnvVars(EnvVarCount, 2*1024, "SOME_ENV_VAR_MEDIUM")
	case "large":
		return newEnvVars(EnvVarCount, 4*1024, "SOME_ENV_VAR_LARGE")
	case "xlarge":
		return newEnvVars(EnvVarCount, 8*1024, "SOME_ENV_VAR_XLARGE")
	case "xlarge2":
		return newEnvVars(EnvVarCount, 10*1024, "SOME_ENV_VAR_XLARGE2")
	case "xlarge8":
		return newEnvVars(EnvVarCount, 40*1024, "SOME_ENV_VAR_XLARGE8")
	default:
		return newEnvVars(EnvVarCount, 2*1024, "SOME_ENV_VAR_MEDIUM")
	}
}

func GetRandomEnvVarType() []corev1.EnvVar {
	size := 1 + rand.Intn(MaxEnvVarSize)
	return newEnvVars(EnvVarCount, size, "SOME_ENV_VAR_RANDOM")
}

// newEnvVars creates a slice of envvars with the specified count and size.
func newEnvVars(count, size int, prefix string) []corev1.EnvVar {
	envVars := make([]corev1.EnvVar, 0, count)
	for i := 0; i < count; i++ {
		envVars = append(envVars, newEnvVar(fmt.Sprintf("%s_%d", prefix, i), size))

	}
	return envVars
}

// newEnvVar creates a new envvar with the specified name and size.
func newEnvVar(name string, size int) corev1.EnvVar {
	return corev1.EnvVar{
		Name:  name,
		Value: util.RandomText(size),
	}
}

// NewFakeNode creates a fake Kubernetes Node resource, managed by KWOK, with the specified name.
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
			PodCIDR:  "10.233.1.0/24",
			PodCIDRs: []string{"10.233.1.0/24"},
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
				corev1.ResourceCPU:    resource.MustParse("20"),
				corev1.ResourceMemory: resource.MustParse("256Gi"),
				corev1.ResourcePods:   resource.MustParse("110"),
			},
			Capacity: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("20"),
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

// NewFakeJob creates a fake Kubernetes Job resource, managed by KWOK, with the specified name and namespace.
func NewFakeJob(name, namespace string, randomEnvVars bool) *batchv1.Job {
	return &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				LabelKeyApp:  LabelValueFakeJob,
				"type":       "kwok",
				"created-by": getHostname(),
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: ptr.To[int32](30),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelKeyApp:  LabelValueFakePod,
						"part-of":    LabelValueFakeJob,
						"created-by": getHostname(),
					},
				},
				Spec: newPodSpec(randomEnvVars),
			},
		},
	}
}

// NewFakePod creates a fake Kubernetes Pod resource, managed by KWOK, with the specified name and namespace.
func NewFakePod(name, namespace string, randomEnvVars bool) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				LabelKeyApp:  LabelValueFakePod,
				"type":       "kwok",
				"created-by": getHostname(),
			},
		},
		Spec: newPodSpec(randomEnvVars),
	}
}

// newPodSpec creates a new pod spec.
// If randomEnvVars is true, a random envvar slice will be used, otherwise the default (large) envvar slice will be used.
func newPodSpec(randomEnvVars bool) corev1.PodSpec {
	envVars := EnvVarsType
	if randomEnvVars {
		envVars = GetRandomEnvVarType()
	}
	podSpec := corev1.PodSpec{
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
				Env:   envVars,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1"),
					},
				},
			},
		},
	}
	return podSpec
}

// newAffinity creates a new affinity which matches nodes with the type kwok.
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

func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}
