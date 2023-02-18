package provider

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	serverlessdevsv1 "serverless.domain/k8s-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// DefaultName is the name of Provider object
	DefaultName = "default"
	// DefaultNamespace is the namespace of Provider object
	DefaultNamespace = "default"
)

// CloudProvider is a type for mark a Cloud Provider
const (
	errConvertCredentials     = "failed to convert the credentials of Secret from Provider"
	errCredentialValid        = "Credentials are not valid"
	ErrCredentialNotRetrieved = "Credentials are not retrieved from referenced Provider"
)

type CloudProvider string

const (
	alibaba CloudProvider = "alibaba"
	aws     CloudProvider = "aws"
	gcp     CloudProvider = "gcp"
	tencent CloudProvider = "tencent"
)

// GetProviderCredentials gets provider credentials by cloud provider name
func GetProviderCredentials(ctx context.Context, k8sClient client.Client, providerObj *serverlessdevsv1.Provider) (map[string]string, error) {

	var secret corev1.Secret
	region := providerObj.Spec.Region

	// A SecretRef is a reference to a secret key that contains the credentials
	secretRef := providerObj.Spec.Credentials.SecretRef
	name := secretRef.Name
	namespace := secretRef.Namespace
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &secret); err != nil {
		errMsg := "failed to get the Secret from Provider"
		klog.ErrorS(err, errMsg, "Name", name, "Namespace", namespace)
		return nil, errors.Wrap(err, errMsg)
	}
	secretData, ok := secret.Data[secretRef.Key]
	if !ok {
		return nil, errors.Errorf("in the provider %s, the key %s not found in the referenced secret %s", providerObj.Name, secretRef.Key, name)
	}

	switch providerObj.Spec.Provider {
	case string(alibaba):
		return getAlibabaCredentials(secretData, name, namespace, region)
	case string(aws):
		return getAWSCredentials(secretData, name, namespace, region)
	case string(gcp):
		return getGCPCredentials(secretData, name, namespace, region)
	case string(tencent):
		return getTencentCloudCredentials(secretData, name, namespace, region)

	default:
		errMsg := "unsupported provider"
		klog.InfoS(errMsg, "Provider", providerObj.Spec.Provider)
		return nil, errors.New(errMsg)
	}

}

// GetProviderFromConfiguration gets provider object from Configuration
// Returns:
// 1) (nil, err): hit an issue to find the provider
// 2) (nil, nil): provider not found
// 3) (provider, nil): provider found
func GetProviderfromObject(ctx context.Context, k8sClient client.Client, namespace string, name string) (*serverlessdevsv1.Provider, error) {
	var providerObj = &serverlessdevsv1.Provider{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, providerObj); err != nil {
		if kerrors.IsNotFound(err) {
			return nil, nil
		}
		errMsg := "failed to get Provider object"
		klog.ErrorS(err, errMsg, "Name", name)
		return nil, errors.Wrap(err, errMsg)
	}
	return providerObj, nil
}
