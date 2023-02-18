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

package controllers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	batchv1 "k8s.io/api/batch/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	serverlessdevsv1 "serverless.domain/k8s-operator/api/v1"
	provider "serverless.domain/k8s-operator/providers"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

//+kubebuilder:rbac:groups=serverless.devs.serverless.domain,resources=serverlesses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=serverless.devs.serverless.domain,resources=serverlesses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=serverless.devs.serverless.domain,resources=serverlesses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Serverless object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
const (
	serverlessFinalizer = "serverless.finalizers.Serverless-controller"
)

func (r *ServerlessReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// Fetch the serverless resource using our client
	var serverless serverlessdevsv1.Serverless
	if err := r.Client.Get(ctx, req.NamespacedName, &serverless); err != nil {
		if kerrors.IsNotFound(err) {
			klog.ErrorS(err, "unable to fetch serverless", "Namespace", req.Namespace, "Name", req.Name)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Load the controller
	var c = initLoadController(req, serverless)

	// var sJob batchv1.Job
	// if err := r.Client.Get(ctx, client.ObjectKey{Name: c.ApplyJobName, Namespace: c.Namespace}, &sJob); err != nil {
	// 	return ctrl.Result{}, err
	// }

	// add finalizer
	var isDeleting = !serverless.ObjectMeta.DeletionTimestamp.IsZero()
	if !isDeleting {
		if !controllerutil.ContainsFinalizer(&serverless, serverlessFinalizer) {
			controllerutil.AddFinalizer(&serverless, serverlessFinalizer)
			if err := r.Update(ctx, &serverless); err != nil {
				return ctrl.Result{RequeueAfter: 5 * time.Second}, errors.Wrap(err, "failed to add finalizer")
			}
		}
	}

	// Load the preCheck
	err := r.preCheck(ctx, &serverless, c)
	if err != nil {
		klog.Error("Error found in precheck %s", err)
		return ctrl.Result{}, err
	}

	//Serverless Remove if the time-stamp is notZero
	if isDeleting {
		// Serverless Remove
		klog.InfoS("performing serverless remove", "Namespace", req.Namespace, "Name", req.Name, "JobName", c.RemoveJobName)

		if err := r.serverlessRemove(ctx, req.Namespace, serverless, c); err != nil {
			if err.Error() == "Configuration deletion isn't completed" {
				return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, errors.Wrap(err, "continue reconciling to remove vendor resource")
		}
		var serverless serverlessdevsv1.Serverless
		if err := r.Client.Get(ctx, req.NamespacedName, &serverless); err != nil {
			klog.Error("unable to fetch Serverless Job", err)
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}

		if controllerutil.ContainsFinalizer(&serverless, serverlessFinalizer) {
			controllerutil.RemoveFinalizer(&serverless, serverlessFinalizer)
			if err := r.Update(ctx, &serverless); err != nil {
				return ctrl.Result{RequeueAfter: 5 * time.Second}, errors.Wrap(err, "failed to remove finalizer")
			}
		}
		return ctrl.Result{}, nil
	}

	// Serverless Apply ( Deploy )
	klog.InfoS("performing Serverless Apply ", "Namespace", req.Namespace, "Name", req.Name)
	if err := r.serverlessApply(ctx, req.Namespace, serverless, c); err != nil {
		if err.Error() == "cloud resources are not created completed" {
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerlessReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&serverlessdevsv1.Serverless{}).
		Complete(r)
}

// Load the Controller ( Serverless Controller Struct )
func initLoadController(req ctrl.Request, serverless serverlessdevsv1.Serverless) *ServerlessController {
	var c = &ServerlessController{
		// Controller field over here
		Name:                    req.Name,
		Namespace:               req.Namespace,
		ApplyJobName:            req.Name + "-" + string(serverlessApply),
		RemoveJobName:           req.Name + "-" + string(serverlessRemove),
		ServerlessConfigMapName: fmt.Sprintf(ServerlessConfigMapName, req.Name),
		ServerelessSecretsName:  fmt.Sprintf(ServerlessvariableSecretName, req.Name),
		ServerlessCodeCM:        "code-configmap",
	}

	return c
}

func (r *ServerlessReconciler) preCheck(ctx context.Context, serverless *serverlessdevsv1.Serverless, c *ServerlessController) error {
	var k8sClient = r.Client
	c.ContainerImage = os.Getenv("CONTAINER_IMAGE")
	if c.ContainerImage == "" {
		c.ContainerImage = "shubham192001/serverless-devs:0.1"
	}

	c.InitContainerImage = os.Getenv("INIT_CONTAINER_IMAGE")
	if c.InitContainerImage == "" {
		c.InitContainerImage = "shubham192001/serverless-devs:0.1"
	}

	// LoadAllSpec actually load the Serverless Spec to the ConfigMap that can be mounted as Volume
	if err := c.loadAllSpec(serverless); err != nil {
		return errors.Wrap(err, "Failed to Run func loadAllSpec")
	}
	// Create the ConfigMap
	if err := c.storeServerless(ctx, k8sClient); err != nil {
		return errors.Wrap(err, "Unable to Run func storeServerless The data can't be loaded to the configMap")
	}

	//Check Provider, Fixing namespace, Name for Sample for Provider as default and default-name
	p, err := provider.GetProviderfromObject(ctx, k8sClient, "default", "default-name")
	if p == nil {
		klog.InfoS("Provider is not present %s", err)
	}
	if err != nil {
		klog.ErrorS(err, "Error Faced while getting the Provider Object")
		return errors.Wrap(err, "Error Faced while getting the Provider Object")
	}
	if err := c.getCredentials(ctx, k8sClient, p); err != nil {
		return errors.Wrap(err, "Provider Found but not able to load the Credentials")
	}

	// Check whether env changes
	if err := c.prepareServerlessEnvVariable(ctx, k8sClient, serverless); err != nil {
		return errors.Wrap(err, "Failed to Load the EnvVariable")
	}

	return nil
}

func (r *ServerlessReconciler) serverlessApply(ctx context.Context, namespace string, resource serverlessdevsv1.Serverless, c *ServerlessController) error {

	klog.InfoS("serverless apply job", "Namespace", namespace, "Name", c.ApplyJobName)
	var (
		k8sClient = r.Client
		sJob      batchv1.Job
	)

	if err := k8sClient.Get(ctx, client.ObjectKey{Name: c.ApplyJobName, Namespace: namespace}, &sJob); err != nil {
		if kerrors.IsNotFound(err) {
			return c.assembleAndTriggerJob(ctx, k8sClient, serverlessApply)
		}
	}

	return nil

}

func (r *ServerlessReconciler) serverlessRemove(ctx context.Context, namespace string, resource serverlessdevsv1.Serverless, c *ServerlessController) error {

	klog.InfoS("serverless remove job", "Namespace", namespace, "Name", c.RemoveJobName)
	var (
		k8sClient = r.Client
		sJob      batchv1.Job
	)

	if err := k8sClient.Get(ctx, client.ObjectKey{Name: c.RemoveJobName, Namespace: namespace}, &sJob); err != nil {
		if kerrors.IsNotFound(err) {
			return c.assembleAndTriggerJob(ctx, k8sClient, serverlessRemove)
		}
	}

	return nil
}
func (c *ServerlessController) loadAllSpec(serverless *serverlessdevsv1.Serverless) error {

	rawData, err := yaml.Marshal(serverless.Spec)
	if err != nil {
		return errors.Wrap(err, "Failed to load the Raw Data from the Spec of Struct")
	}

	rawDataString := string(rawData)
	c.CompleteSpec = rawDataString

	return nil

}
