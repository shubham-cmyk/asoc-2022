package provider

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

const (
	envAlicloudAccessKey = "ALIBABA_CLOUD_ACCESS_KEY_ID"
	envAlicloudSecretKey = "ALIBABA_CLOUD_ACCESS_KEY_SECRET"
	//envAlicloudRegion    = "ALICLOUD_REGION"
	envAliCloudAccountId = "ALIBABA_CLOUD_ACCOUNT_ID"
)

type AlibabaCloudCredentials struct {
	AccessKeyID     string `yaml:"accessKeyID"`
	AccessKeySecret string `yaml:"accessKeySecret"`
	AccountID       string `yaml:"accountID"`
}

func getAlibabaCredentials(secretData []byte, name, namespace, region string) (map[string]string, error) {
	var ak AlibabaCloudCredentials
	if err := yaml.Unmarshal(secretData, &ak); err != nil {
		klog.ErrorS(err, errConvertCredentials, "Name", name, "Namespace", namespace)
		return nil, errors.Wrap(err, errConvertCredentials)
	}
	return map[string]string{
		envAlicloudAccessKey: ak.AccessKeyID,
		envAlicloudSecretKey: ak.AccessKeySecret,
		//	envAlicloudRegion:    region,
		envAliCloudAccountId: ak.AccountID,
	}, nil
}
