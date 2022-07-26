/*
Copyright 2022.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MPIJobSpec defines the desired state of MPIJob
type MPIJobSpec struct {
	LauncherTemplate v1.PodTemplateSpec `json:"launcherTemplate"`

	WorkerTemplate v1.PodTemplateSpec `json:"workerTemplate"`

	NumWorkers *int32 `json:"numWorkers"`
}

// MPIJobStatus defines the observed state of MPIJob
type MPIJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+genclient

// MPIJob is the Schema for the mpijobs API
type MPIJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MPIJobSpec   `json:"spec,omitempty"`
	Status MPIJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MPIJobList contains a list of MPIJob
type MPIJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MPIJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MPIJob{}, &MPIJobList{})
}
