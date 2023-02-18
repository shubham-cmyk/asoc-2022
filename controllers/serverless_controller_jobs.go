package controllers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServerlessExecutionType is the type for Serverless execution
type ServerlessExecutionType string

const (
	// serverlessApply is the name to mark `Serverless deploy`
	serverlessApply ServerlessExecutionType = "deploy"
	// ServerlessRemove is the name to mark `Serverless remove`
	serverlessRemove ServerlessExecutionType = "remove"
)

const (
	// ServerlessContainerName is the name of the container that executes the Serverless in the pod
	ServerlessContainerName     = "main"
	ServerlessInitContainerName = "prepare-configurations"
)

// Assesble and Trigger Job Entry point for the Job
func (c *ServerlessController) assembleAndTriggerJob(ctx context.Context, k8sClient client.Client, executiontype ServerlessExecutionType) error {

	//create Service Account
	if err := createServerlessServiceAccount(ctx, k8sClient, c.Namespace, ServiceAccountName); err != nil {
		return err
	}

	// create a cluster role
	if err := createServerlessClusterRole(ctx, k8sClient, ClusterRoleName); err != nil {
		return err
	}

	// create a cluster role binding
	if err := createServerlessClusterRoleBinding(ctx, k8sClient, c.Namespace, ClusterRoleName, ServiceAccountName); err != nil {
		return err
	}

	serverlesjob := c.ServerlessJob(executiontype)
	if err := k8sClient.Create(ctx, serverlesjob); err != nil {
		klog.Error("Serverless-controller Failed to Create the Job ")
		return errors.Wrap(err, "Serverless-Controller Failed to Create the Job")
	}
	return nil

}

// Creating the Serverless Job ( Exectution Type as String )
func (c *ServerlessController) ServerlessJob(executionType ServerlessExecutionType) *batchv1.Job {

	// Variable Declared for the ServerlessJob
	var (
		backOffLimit   int32 = 4
		initContainer  v1.Container
		initContainers []v1.Container
		containers     []v1.Container
		container      v1.Container
	)

	// Volumes
	// To ask if we need any volumes;
	volumes := c.assembleVolumes()

	// Init containers ( Name, Image, ImagePullPolicy, Command, VolumeMounts)
	initContainer = v1.Container{
		Name:            ServerlessInitContainerName,
		Image:           c.InitContainerImage,
		ImagePullPolicy: v1.PullIfNotPresent,
		Command: []string{
			"sh",
			"-c",
			"cp /opt/s-configuration/* /code",
			"cp /opt/s-code/* /code",
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "s-workspace",
				MountPath: "/code",
			},
			{
				Name:      "s-configuration",
				MountPath: "/opt/s-configuration",
			},
			// {
			// 	Name:      "s-code",
			// 	MountPath: "opt/s-code",
			// },
		},
		Env: c.Envs,
	}

	// Append the initContainer to the slice of initContainers
	initContainers = append(initContainers, initContainer)

	// Serverless Container ( Name, Image, ImagePullPolicy, Command, Args, VolumeMounts, Env)
	container = v1.Container{
		Name:            ServerlessContainerName,
		Image:           c.ContainerImage,
		ImagePullPolicy: v1.PullIfNotPresent,
		Command: []string{
			"bash",
			"-c",
			"s config add -a default-aliyun -kl AccountID,AccessKeyID,AccessKeySecret -il ${ALIBABA_CLOUD_ACCOUNT_ID},${ALIBABA_CLOUD_ACCESS_KEY_ID},${ALIBABA_CLOUD_ACCESS_KEY_SECRET}",
			fmt.Sprintf("s %s", executionType),
		},
		WorkingDir: "/code",
		Args: []string{
			"--use-local", "--assume-yes",
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "s-workspace",
				MountPath: "/code",
			},
			{
				Name:      "s-configuration",
				MountPath: "/opt/s-configuration",
			},
			// {
			// 	Name:      "s-code",
			// 	MountPath: "opt/s-code",
			// },
		},
		// To be looked to Specify env var from ConfigMaps
		Env: c.Envs,
	}

	// Append the container to the slice of containers
	containers = append(containers, container)

	// Jobs field
	serverlessJobSpec := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Job",
			APIVersion: "batch/v1",
		},

		ObjectMeta: metav1.ObjectMeta{
			// Fixing it to default for a while
			Name:      c.Name + "-" + string(executionType),
			Namespace: "default",
		},

		Spec: batchv1.JobSpec{
			BackoffLimit: &backOffLimit,
			Template: v1.PodTemplateSpec{

				// Putting Object Meta for temp purpose
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},

				Spec: v1.PodSpec{
					InitContainers: initContainers,
					Containers:     containers,
					Volumes:        volumes,
					RestartPolicy:  v1.RestartPolicyOnFailure,
				},
			},
		},
	}

	// Returning the Job of Serverless
	return serverlessJobSpec

}

// assesmble Volumes for the Jobs
func (c *ServerlessController) assembleVolumes() []v1.Volume {
	workingVolume := v1.Volume{Name: "s-workspace"}
	workingVolume.EmptyDir = &v1.EmptyDirVolumeSource{}
	// cmVolume := v1.Volume{Name: "s-configuration", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{}}}
	cmVolumeSource := v1.ConfigMapVolumeSource{}
	cmVolumeSource.Name = c.ServerlessConfigMapName
	cmVolume := v1.Volume{Name: "s-configuration"}
	cmVolume.ConfigMap = &cmVolumeSource

	// cmCodeVolumeSource := v1.ConfigMapVolumeSource{}
	// cmCodeVolumeSource.Name = c.ServerlessCodeCM
	// cmCodeVolume := v1.Volume{Name: "s-code"}
	// cmCodeVolume.ConfigMap = &cmCodeVolumeSource

	klog.Info("Volume have been Created Successfully")
	return []v1.Volume{workingVolume, cmVolume}

}
