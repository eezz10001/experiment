/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExperimentPhase string

const (
	ImpalaPhasePending ExperimentPhase = "Pending"
	ImpalaPhaseRunning ExperimentPhase = "Running"
	ImpalaPhaseFail    ExperimentPhase = "Fail"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ExperimentSpec defines the desired state of Experiment
type ExperimentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Experiment. Edit experiment_types.go to remove/update
	//Foo string `json:"foo,omitempty"`
	Image     string                      `json:"image,omitempty" protobuf:"bytes,11,rep,name=image"`
	Host      string                      `json:"host,omitempty" protobuf:"bytes,11,rep,name=host"`
	Port      coreV1.ContainerPort        `json:"port,omitempty" patchStrategy:"merge" patchMergeKey:"port" protobuf:"bytes,1,rep,name=port"`
	Resources coreV1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
	Probe     Probe                       `json:"probe,omitempty" protobuf:"bytes,11,rep,name=probe"`
	Command   []string                    `json:"command,omitempty" protobuf:"bytes,3,rep,name=command"`
}
type Probe struct {
	Path string `json:"path,omitempty" protobuf:"bytes,11,opt,name=path"`
	Port int32  `json:"port,omitempty" protobuf:"varint,2,opt,name=port"`
}

// ExperimentStatus defines the observed state of Experiment
type ExperimentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SubResourcesStatus SubResourcesStatus `json:"subResourcesStatus,omitempty"`
	Phase              ExperimentPhase    `json:"phase,omitempty,default=NoRunning"`
	Message            string             `json:"message,omitempty"`
}

type SubResourcesStatus struct {
	Sts     bool `json:"statestore,omitempty"`
	Svc     bool `json:"statestoreService,omitempty"`
	Ingress bool `json:"statestoreLbService,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase"

// Experiment is the Schema for the experiments API
type Experiment struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExperimentSpec   `json:"spec,omitempty"`
	Status ExperimentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExperimentList contains a list of Experiment
type ExperimentList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`
	Items           []Experiment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Experiment{}, &ExperimentList{})
}
