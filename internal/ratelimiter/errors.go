package ratelimiter

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateError is used to wrap errors that occur during creation.
type CreateError struct {
	// Err is the error that occurred during creation.
	Err error
	// APIGroup is the API group of the object that failed to be created.
	APIGroup string
	// Kind is the kind of the object that failed to be created.
	Kind string
	// Resource is the object that failed to be created.
	Resource metav1.Object
}

// NewCreateError returns a new CreateError which contains info on the failed creation of a k8s object.
func NewCreateError(err error, apiGroup, kind string, resource metav1.Object) *CreateError {
	return &CreateError{
		Err:      err,
		APIGroup: apiGroup,
		Kind:     kind,
		Resource: resource,
	}
}

func (e *CreateError) Error() string {
	return fmt.Sprintf("failed to create %s %s in %s/%s: %v", e.APIGroup, e.Kind, e.Resource.GetNamespace(), e.Resource.GetName(), e.Err)
}
