package controllers

import (
	"context"
	"errors"

	serverlessdevsv1 "serverless.domain/k8s-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	provider "serverless.domain/k8s-operator/providers"
)

//getCredentials will get credentials from secret of the Provider
func (c *ServerlessController) getCredentials(ctx context.Context, k8sClient client.Client, providerObj *serverlessdevsv1.Provider) error {

	credentials, err := provider.GetProviderCredentials(ctx, k8sClient, providerObj)
	if err != nil {
		return err
	}
	if credentials == nil {
		return errors.New("credentials are not retrieved from referenced provider")
	}
	c.Credentials = credentials

	return nil
}
