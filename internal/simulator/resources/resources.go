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
	defaultEnvVarCount   = 5
	LabelKeyApp          = "app"
	LabelValueFakeJob    = "fake-job"
	LabelValueFakePod    = "fake-pod"
	LabelSelectorFakePod = LabelKeyApp + "=" + LabelValueFakePod
)

var (
	// nano equals to 100 bytes x 5 envvars = 500 bytes
	nano = newEnvVars(defaultEnvVarCount, 100, "SOME_ENV_VAR_NANO")
	// micro equals to 200 bytes x 5 envvars = 1000 bytes
	micro = newEnvVars(defaultEnvVarCount, 200, "SOME_ENV_VAR_MICRO")
	// xsmall equals to 500 bytes x 5 envvars = 2500 bytes
	xsmall = newEnvVars(defaultEnvVarCount, 500, "SOME_ENV_VAR_XSMALL")
	// small equals to 1KB x 5 envvars = 5KB
	small = newEnvVars(defaultEnvVarCount, 1024, "SOME_ENV_VAR_SMALL")
	// medium equals to 2KB x 5 envvars = 10KB
	medium = newEnvVars(defaultEnvVarCount, 2*1024, "SOME_ENV_VAR_MEDIUM")
	// large equals to 4KB x 5 envvars = 20KB
	large = newEnvVars(defaultEnvVarCount, 4*1024, "SOME_ENV_VAR_LARGE")
	// xlarge equals to 8KB x 5 envvars = 40KB
	xlarge = newEnvVars(defaultEnvVarCount, 8*1024, "SOME_ENV_VAR_XLARGE")
	// xlarge2 equals to 10KB x 5 envvars = 50KB
	xlarge2 = newEnvVars(defaultEnvVarCount, 10*1024, "SOME_ENV_VAR_XLARGE2")
	// xlarge8 equals to 40KB x 5 envvars = 200KB
	xlarge8 = newEnvVars(defaultEnvVarCount, 40*1024, "SOME_ENV_VAR_XLARGE8")
)

// DefaultEnvVarsType is the default envvar slice type.
var DefaultEnvVarsType = medium

// SetDefaultEnvVarsType sets the default envvar slice type.
func SetDefaultEnvVarsType(envVarType string) {
	switch envVarType {
	case "nano":
		DefaultEnvVarsType = nano
	case "micro":
		DefaultEnvVarsType = micro
	case "xsmall":
		DefaultEnvVarsType = xsmall
	case "small":
		DefaultEnvVarsType = small
	case "medium":
		DefaultEnvVarsType = medium
	case "large":
		DefaultEnvVarsType = large
	case "xlarge":
		DefaultEnvVarsType = xlarge
	case "xlarge2":
		DefaultEnvVarsType = xlarge2
	case "xlarge8":
		DefaultEnvVarsType = xlarge8
	default:
		DefaultEnvVarsType = medium
	}
}

// envVarsByType is a slice of different envvar slice types.
var envVarsByType = [][]corev1.EnvVar{nano, micro, xsmall, small, medium, large, xlarge, xlarge2}

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
	envVars := DefaultEnvVarsType
	if randomEnvVars {
		envVars = getRandomEnvVarType()
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

// getRandomEnvVarType returns a random envvar slice from the envVarsByType.
func getRandomEnvVarType() []corev1.EnvVar {
	return envVarsByType[rand.Intn(len(envVarsByType))]
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
