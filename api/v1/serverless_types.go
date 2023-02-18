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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServerlessSpec defines the desired state of Serverless
type ServerlessSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Edition  string `json:"edition,omitempty"`
	Name     string `json:"name,omitempty"`
	Access   string `json:"access,omitempty"`
	Vars     `json:"vars,omitempty"`
	Services map[string]ServiceSpec `json:"services,omitempty"`
}

type Vars struct {
	Region  string `json:"region,omitempty"`
	Service `json:"service,omitempty"`
}

type Service struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
type ServiceSpec struct {
	Access    string                `json:"access,omitempty"`
	Component string                `json:"component,omitempty"`
	Props     *runtime.RawExtension `json:"props,omitempty"`
	Actions   Action                `json:"actions,omitempty"`
}

type Action struct {
	PreDeploy  []Deploy `json:"pre-deploy,omitempty"`
	PostDeploy []Deploy `json:"post-deploy,omitempty"`
}

type Deploy struct {
	Run       string `json:"run,omitempty"`
	Path      string `json:"path,omitempty"`
	Component string `json:"component,omitempty"`
}

// type Property struct {
// 	Region        string          `json:"region,omitempty"`
// 	Service       string          `json:"service,omitempty"`
// 	Src           string          `json:"src,omitempty"`
// 	Url           string          `json:"url,omitempty"`
// 	Function      Function        `json:"function,omitempty"`
// 	Trigger       []Trigger       `json:"trigger,omitempty"`
// 	CustomDomains []CustomDomains `json:"customDomain,omitempty"`
// }

// type Function struct {
// 	Name                  string `json:"name,omitempty"`
// 	Description           string `json:"description,omitempty"`
// 	Runtime               string `json:"runtime,omitempty"`
// 	CodeUri               string `json:"codeUri,omitempty"`
// 	Handler               string `json:"handler,omitempty"`
// 	MemorySize            int    `json:"memorySize,omitempty"`
// 	Timeout               int    `json:"timeout,omitempty"`
// 	InitializationTimeout int    `json:"initializationTimeout,omitempty"`
// 	Initializer           string `json:"initializaer,omitempty"`
// }

// type Trigger struct {
// 	Name   string  `json:"name,omitempty"`
// 	Type   string  `json:"type,omitempty"`
// 	Config SConfig `json:"config,omitempty"`
// }
// type SConfig struct {
// 	AuthType string   `json:"authType,omitempty"`
// 	Methods  []string `json:"methods,omitempty"`
// }

// type CustomDomains struct {
// 	DomainName  string         `json:"domainName,omitempty"`
// 	Protocol    string         `json:"protocol,omitempty"`
// 	RouteConfig []SRouteConfig `json:"routeConfig,omitempty"`
// }

// type SRouteConfig struct {
// 	Path    string   `json:"path,omitempty"`
// 	Methods []string `json:"methods,omitempty"`
// }

// ServerlessStatus defines the observed state of Serverless
type ServerlessStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Serverless is the Schema for the serverlesses API
type Serverless struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServerlessSpec   `json:"spec"`
	Status            ServerlessStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServerlessList contains a list of Serverless
type ServerlessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Serverless `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Serverless{}, &ServerlessList{})
}
