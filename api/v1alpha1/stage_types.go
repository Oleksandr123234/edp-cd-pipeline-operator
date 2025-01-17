package v1alpha1

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StageSpec defines the desired state of Stage.
type StageSpec struct {
	// +kubebuilder:validation:MinLength=2

	// Name of a stage.
	Name string `json:"name"`

	// +kubebuilder:validation:MinLength=2

	// Name of CD pipeline which this Stage will be linked to.
	CdPipeline string `json:"cdPipeline"`

	// +kubebuilder:validation:MinLength=0

	// A description of a stage.
	Description string `json:"description"`

	// Stage provisioning trigger type. E.g. Manual, Auto
	TriggerType string `json:"triggerType"`

	// The order to lay out Stages
	Order int `json:"order"`

	// A list of quality gates to be processed
	QualityGates []QualityGate `json:"qualityGates"`

	// Specifies a source of a pipeline library which will run release
	Source Source `json:"source" valid:"Required"`

	// CD Job Provisioner for Pipeline. E.g.
	JobProvisioning string `json:"jobProvisioning"`
}

// QualityGate defines a single quality for a release.
type QualityGate struct {
	// A type of quality gate, e.g. "Manual", "Autotests"
	QualityGateType string `json:"qualityGateType" valid:"Required"`

	// +kubebuilder:validation:MinLength=2

	// Specifies a name of particular
	StepName string `json:"stepName" valid:"Required;Match(/^[A-z0-9-._]/)"`

	// A name of autotests to run with quality gate
	// +nullable
	// +optional
	AutotestName *string `json:"autotestName"`

	// A branch name to use from autotests repository
	// +nullable
	// +optional
	BranchName *string `json:"branchName"`
}

// Source defines a pipeline library.
type Source struct {
	// Type of pipeline library, e.g. default, library
	Type string `json:"type"`

	// A reference to a non default source library
	// +nullable
	// +optional
	Library Library `json:"library,omitempty"`
}

// Library defines a pipeline library for release process.
type Library struct {
	// A name of a library
	Name string `json:"name,omitempty"`

	// Branch which should be used for a library
	Branch string `json:"branch,omitempty"`
}

// StageStatus defines the observed state of Stage.
type StageStatus struct {
	// This flag indicates neither Stage are initialized and ready to work. Defaults to false.
	Available bool `json:"available"`

	// Information when  the last time the action were performed.
	LastTimeUpdated metaV1.Time `json:"last_time_updated"`

	// Specifies a current status of Stage.
	Status string `json:"status"`

	// Name of user who made a last change.
	Username string `json:"username"`

	// The last Action was performed.
	Action ActionType `json:"action"`

	// A result of an action which were performed.
	// - "success": action where performed successfully;
	// - "error": error has occurred;
	Result Result `json:"result"`

	// Detailed information regarding action result
	// which were performed
	// +optional
	DetailedMessage string `json:"detailed_message,omitempty"`

	// Specifies a current state of Stage.
	Value string `json:"value"`

	// Should update of status be handled. Defaults to false.
	// +optional
	ShouldBeHandled bool `json:"shouldBeHandled,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:deprecatedversion

// Stage is the Schema for the stages API.
type Stage struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StageSpec   `json:"spec,omitempty"`
	Status StageStatus `json:"status,omitempty"`
}

// nolint:gocritic
func (in Stage) IsFirst() bool {
	return in.Spec.Order == 0
}

// +kubebuilder:object:root=true

// StageList contains a list of Stage.
type StageList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`

	Items []Stage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Stage{}, &StageList{})
}
