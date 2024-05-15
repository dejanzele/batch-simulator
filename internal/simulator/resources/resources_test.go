package resources

import (
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestNewPodSpec(t *testing.T) {
	t.Skip("temporary")
	t.Parallel()

	t.Run("should return a valid PodSpec with default envvars if randomEnvVars is false", func(t *testing.T) {
		t.Parallel()

		podSpec := newPodSpec(false)
		assert.Len(t, podSpec.Containers[0].Env, defaultEnvVarCount)
		assert.Equal(t, medium, podSpec.Containers[0].Env)
	})

	t.Run("should return a valid PodSpec with random envvars if randomEnvVars is true", func(t *testing.T) {
		t.Parallel()

		podSpec := newPodSpec(true)
		assert.Len(t, podSpec.Containers[0].Env, defaultEnvVarCount)
	})
}

// TestSetDefaultEnvVarsType tests the SetDefaultEnvVarsType function.
func TestSetDefaultEnvVarsType(t *testing.T) {
	t.Parallel()

	// Define the test cases
	testCases := []struct {
		name         string
		envVarType   string
		expectedType []corev1.EnvVar // Assuming EnvVarsType is the type of DefaultEnvVarsType
	}{
		{"nano type", "nano", nano},
		{"micro type", "micro", micro},
		{"xsmall type", "xsmall", xsmall},
		{"small type", "small", small},
		{"medium type", "medium", medium},
		{"large type", "large", large},
		{"xlarge type", "xlarge", xlarge},
		{"xlarge2 type", "xlarge2", xlarge2},
		{"default type", "unknown", medium}, // Test for an unknown type
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			SetDefaultEnvVarsType(tc.envVarType)
			assert.Equal(t, tc.expectedType, DefaultEnvVarsType)
		})
	}
}
