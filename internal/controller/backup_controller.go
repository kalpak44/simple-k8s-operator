/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	homev1 "github.com/kalpak44/simple-k8s-operator/api/v1"
)

// +kubebuilder:rbac:groups=home.home.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=home.home.com,resources=backups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=home.home.com,resources=backups/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
type BackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=home.home.com,resources=backups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=home.home.com,resources=backups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=home.home.com,resources=backups/finalizers,verbs=update
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1) Get the Backup object
	var bkp homev1.Backup
	if err := r.Get(ctx, req.NamespacedName, &bkp); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2) Form the desired CronJob
	cron := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bkp.Name + "-cron",
			Namespace: bkp.Namespace,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: bkp.Spec.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{{
								Name:  "backup",
								Image: "curlimages/curl:latest",
								Args: []string{
									"-s",
									"-X", "POST",
									"https://kalpak44.free.beeceptor.com",
									"-d", fmt.Sprintf("db=%s", bkp.Spec.Database),
								},
							}},
							RestartPolicy: corev1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}

	// 3) Set the owner for automatic cleanup
	if err := controllerutil.SetControllerReference(&bkp, cron, r.Scheme); err != nil {
		logger.Error(err, "unable to set owner reference on CronJob")
		return ctrl.Result{}, err
	}

	// 4) Create or update the CronJob in the cluster
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, cron, func() error {
		// here you can add logic to update the spec if needed
		return nil
	}); err != nil {
		logger.Error(err, "failed to create or update CronJob")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homev1.Backup{}).
		Named("backup").
		Complete(r)
}
