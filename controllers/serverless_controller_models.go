package controllers

import (
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ServerlessConfigMapName is the CM name for serverless Input Configuration
	ServerlessConfigMapName = "serverless-configmap-%s"
	// ServerlessVariableSecret is the Secret name for variables, including credentials from Provider
	ServerlessvariableSecretName = "serverless-secret-%s"
)

// Controller Model is the Object of the Controller
type ServerlessController struct {
	Name              string
	Namespace         string
	ApplyJobName      string
	RemoveJobName     string
	ServerlessChanges bool
	// Name of ConfigMap and Secrets
	ServerlessConfigMapName string
	ServerlessCodeCM        string
	ServerelessSecretsName  string
	ServelessSecretData     map[string][]byte
	EnvChanged              bool
	// Envs Contain all the env Var that are to be loaded to the Container
	Envs        []v1.EnvVar
	Credentials map[string]string
	Region      string

	// Image that is required to start the container
	InitContainerImage string
	ContainerImage     string

	// CompleteSpec contains the marshal data of serverless spec
	CompleteSpec string
}

// ServerlessReconciler reconciles a Serverless object
type ServerlessReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	ProviderName string
	Log          logr.Logger
}
