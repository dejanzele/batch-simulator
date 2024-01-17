package simulator

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func newJob() *batchv1.Job {
	return &batchv1.Job{
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: "Never",
					Containers: []corev1.Container{
						{
							Name:  "simulator",
							Image: "dpejcev/batch-simulator:latest",
							Args: []string{
								"sleep",
							},
						},
					},
				},
			},
		},
	}
}
