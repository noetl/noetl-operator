package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NoetlSparkSpec defines the desired state of NoetlSpark
type NoetlSparkSpec struct {
	Spark []Job `json:"spark,omitempty"`
}

// NoetlSparkStatus defines the observed state of NoetlSpark
type NoetlSparkStatus struct {
	Message string `json:"message,omitempty"`
}

// Job defines spark job.
type Job struct {
	Name      string   `json:"name,omitempty"`
	Resources string   `json:"resources,omitempty"`
	Priority  int      `json:"priority,omitempty"`
	Class     string   `json:"class,omitempty"`
	Image     string   `json:"image,omitempty"`
	Jar       string   `json:"jar,omitempty"`
	Confs     []string `json:"confs,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NoetlSpark is the Schema for the noetlsparks API
// +k8s:openapi-gen=true
type NoetlSpark struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NoetlSparkSpec   `json:"spec,omitempty"`
	Status NoetlSparkStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NoetlSparkList contains a list of NoetlSpark
type NoetlSparkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NoetlSpark `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NoetlSpark{}, &NoetlSparkList{})
}
