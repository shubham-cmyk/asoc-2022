package controllers

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	serverlessdevsv1 "serverless.domain/k8s-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *ServerlessController) storeServerless(ctx context.Context, k8sClient client.Client) error {
	data := c.prepareServerless()
	err := c.createOrUpdateCM(ctx, k8sClient, data)
	return err
}

func (c *ServerlessController) prepareServerless() map[string]string {

	klog.Info("Data Name of config-map is set to s.yaml")
	//Data Name is set to s.yaml
	dataName := "s.yaml"

	data := map[string]string{dataName: c.CompleteSpec}
	return data

}

func (c *ServerlessController) createOrUpdateCM(ctx context.Context, k8sClient client.Client, data map[string]string) error {

	var servelessCM v1.ConfigMap
	err := k8sClient.Get(ctx, client.ObjectKey{Name: c.ServerlessConfigMapName, Namespace: c.Namespace}, &servelessCM)
	if err != nil {
		if kerrors.IsNotFound(err) {
			cm := v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      c.ServerlessConfigMapName,
					Namespace: c.Namespace,
				},
				Data: data,
			}

			err := k8sClient.Create(ctx, &cm)
			return errors.Wrap(err, "Failed to create the configMap of Serverless")
		}

		if !reflect.DeepEqual(servelessCM.Data, data) {
			servelessCM.Data = data
			err := k8sClient.Update(ctx, &servelessCM)
			if err != nil {
				errors.Wrap(err, "Failed to update the configMap of Serverless")
			}
		}
	}

	return nil
}

func (c *ServerlessController) prepareServerlessEnvVariable(ctx context.Context, k8sClient client.Client, serverless *serverlessdevsv1.Serverless) error {
	var (
		envs []v1.EnvVar
		data = map[string][]byte{}
	)
	for k, v := range c.Credentials {
		data[k] = []byte(v)
		valueFrom := &v1.EnvVarSource{SecretKeyRef: &v1.SecretKeySelector{Key: k}}
		valueFrom.SecretKeyRef.Name = c.ServerelessSecretsName
		envs = append(envs, v1.EnvVar{Name: k, ValueFrom: valueFrom})
	}

	// Some Addtional EnvVariables For Serverless
	envs = append(envs, v1.EnvVar{Name: "s_default_fc_endpoint", Value: "http://1.dev-cluster-9.test.fc.aliyun-inc.com"})
	envs = append(envs, v1.EnvVar{Name: "s_default_enable_fc_endpoint", Value: "true"})

	c.Envs = envs
	c.ServelessSecretData = data

	return nil
}
