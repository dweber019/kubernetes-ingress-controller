package parser

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// TranslationFailureReasonUnknown is used when no specific reason is specified when creating a TranslationFailure.
	TranslationFailureReasonUnknown = "unknown"
)

// TranslationFailure represents an error occurring during translating Kubernetes objects into Kong ones.
// It can be associated with one or more Kubernetes objects.
type TranslationFailure struct {
	causingObjects []client.Object
	reason         string
}

// NewTranslationFailure creates a TranslationFailure with a reason that should be a human-readable explanation
// of the error reason, and a causingObjects slice that specifies what objects have caused the error.
func NewTranslationFailure(reason string, causingObjects ...client.Object) (TranslationFailure, error) {
	if reason == "" {
		reason = TranslationFailureReasonUnknown
	}
	if len(causingObjects) < 1 {
		return TranslationFailure{}, fmt.Errorf("no causing objects specified, reason: %s", reason)
	}

	for _, obj := range causingObjects {
		if obj == nil {
			return TranslationFailure{}, errors.New("one of causing objects is nil")
		}
		gvk := obj.GetObjectKind().GroupVersionKind()
		if gvk.Empty() {
			return TranslationFailure{}, errors.New("one of causing objects has an empty GVK")
		}
		if obj.GetName() == "" {
			return TranslationFailure{}, fmt.Errorf("one of causing objects (%s) has no name", gvk.String())
		}
		if obj.GetNamespace() == "" {
			return TranslationFailure{}, fmt.Errorf("one of causing objects (%s) has no namespace", gvk.String())
		}
	}

	return TranslationFailure{
		causingObjects: causingObjects,
		reason:         reason,
	}, nil
}

// CausingObjects returns a slice of objects that have caused the translation error.
func (p TranslationFailure) CausingObjects() []client.Object {
	return p.causingObjects
}

// Reason returns a human-readable reason of the failure.
func (p TranslationFailure) Reason() string {
	return p.reason
}

// TranslationFailuresCollector should be used to collect all translation failures that happen during the translation process.
type TranslationFailuresCollector struct {
	failures []TranslationFailure
	logger   logrus.FieldLogger
}

func NewTranslationFailuresCollector(logger logrus.FieldLogger) (*TranslationFailuresCollector, error) {
	if logger == nil {
		return nil, errors.New("missing logger")
	}
	return &TranslationFailuresCollector{logger: logger}, nil
}

// PushTranslationFailure registers a translation failure and logs it.
func (c *TranslationFailuresCollector) PushTranslationFailure(reason string, causingObjects ...client.Object) {
	translationErr, err := NewTranslationFailure(reason, causingObjects...)
	if err != nil {
		c.logger.WithField("translation_failure_reason", reason).Warningf("failed to create translation failure: %s", err)
		return
	}

	c.failures = append(c.failures, translationErr)
	c.logTranslationFailure(reason, causingObjects...)
}

// logTranslationFailure logs an error message signaling that a translation error has occurred along with its reason
// for every causing object.
func (c *TranslationFailuresCollector) logTranslationFailure(reason string, causingObjects ...client.Object) {
	for _, obj := range causingObjects {
		c.logger.WithFields(logrus.Fields{
			"name":      obj.GetName(),
			"namespace": obj.GetNamespace(),
			"GVK":       obj.GetObjectKind().GroupVersionKind().String(),
		}).Errorf("translation failed: %s", reason)
	}
}

// PopTranslationFailures returns all translation failures that occurred during the translation process and erases them
// in the collector. It makes the collector reusable during next translation runs.
func (c *TranslationFailuresCollector) PopTranslationFailures() []TranslationFailure {
	errs := c.failures
	c.failures = nil

	return errs
}