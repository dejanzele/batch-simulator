package simulator

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/util"
)

const (
	defaultTTLSecondsAfterFinished = 300
)

func NewSimulatorJob(args []string) *batchv1.Job {
	fullArgs := make([]string, 0, len(args)+1)
	fullArgs = append(fullArgs, "run")
	fullArgs = append(fullArgs, args...)
	name := fmt.Sprintf("simulator-job-%s", util.RandomRFC1123Name(5))
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: ptr.To[int32](defaultTTLSecondsAfterFinished),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: serviceAccountName,
					RestartPolicy:      corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "simulator",
							Image: util.CreateImageString(config.SimulatorImage, config.SimulatorTag),
							Args:  fullArgs,
						},
					},
				},
			},
		},
	}

}
