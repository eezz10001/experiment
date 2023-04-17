/*
Copyright 2023.

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
	experimentv1 "github.com/eezz10001/experiment/api/v1"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	log2 "log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=experiment.touchturing.com,resources=experiments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=experiment.touchturing.com,resources=experiments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=experiment.touchturing.com,resources=experiments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Experiment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *ExperimentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ret ctrl.Result, err error) {
	_ = log.FromContext(ctx)
	fmt.Println("进入判断状态1")

	experiment := &experimentv1.Experiment{}

	if err := r.Get(ctx, req.NamespacedName, experiment); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !experiment.DeletionTimestamp.IsZero() {
		log2.Println("-------------------delete crd <" + experiment.Name + ">")
		return ctrl.Result{}, nil
	}

	if experiment.Status.SubResourcesStatus.Sts, experiment.Status.SubResourcesStatus.Svc,
		experiment.Status.SubResourcesStatus.Ingress, err = r.CreateComponent(experiment); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.JudgmentStatus(experiment, ctx); err != nil {

		log2.Println("==============Judgment status fail")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ExperimentReconciler) CreateComponent(experiment *experimentv1.Experiment) (stsStatus, svcStatus, ingressStatus bool, err error) {
	//create statefulSet
	stsStatus, err = r.CreateStatefulset(experiment)
	if err != nil {
		return
	}

	//create  svc
	svcStatus, err = r.CreateService(experiment)
	if err != nil {
		return
	}

	//create  ingress
	ingressStatus, err = r.CreateIngress(experiment)
	if err != nil {
		return
	}

	return
}

func (r *ExperimentReconciler) CreateStatefulset(experiment *experimentv1.Experiment) (bool, error) {
	stsBuilder, err := NewStatefulSetBuilder(r.Client, experiment, r.Scheme)
	if err != nil {
		return false, err
	}

	return stsBuilder.Build(context.Background())
}

func (r *ExperimentReconciler) CreateService(experiment *experimentv1.Experiment) (bool, error) {
	svcBuilder, err := NewServiceBuilder(r.Client, experiment, r.Scheme)
	if err != nil {
		return false, err
	}
	status, err := svcBuilder.Build(context.Background())
	return status, err
}

func (r *ExperimentReconciler) CreateIngress(experiment *experimentv1.Experiment) (bool, error) {
	ingressBuilder, err := NewIngressBuilder(r.Client, experiment, r.Scheme, experiment.Spec.Host)
	if err != nil {
		return false, err
	}
	return ingressBuilder.Build(context.Background())
}

func (r *ExperimentReconciler) JudgmentStatus(experiment *experimentv1.Experiment, ctx context.Context) error {

	//b, _ := json.Marshal(experiment)
	//log2.Println(string(b))
	if experiment.Status.SubResourcesStatus.Sts == true &&
		experiment.Status.SubResourcesStatus.Svc == true &&
		experiment.Status.SubResourcesStatus.Ingress == true {
		if experiment.Status.Phase != experimentv1.ImpalaPhaseRunning {
			experiment.Status.Phase = experimentv1.ImpalaPhaseRunning
			return r.Client.Status().Update(ctx, experiment)
		}
	} else {
		experiment.Status.Phase = experimentv1.ImpalaPhaseFail
		return r.Client.Status().Update(ctx, experiment)
	}
	return nil
}

// Prevent manual misoperation

func (r *ExperimentReconciler) OnObjUpdate(event event.UpdateEvent, rateLimitingInterface workqueue.RateLimitingInterface) {

	if ok := checkIsExperimentResource(event.ObjectOld.GetLabels()); !ok {
		return
	}

	rateLimitingInterface.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{Name: event.ObjectOld.GetName(),
			Namespace: event.ObjectOld.GetNamespace()}})

	return
}

func (r *ExperimentReconciler) OnObjDelete(event event.DeleteEvent, rateLimitingInterface workqueue.RateLimitingInterface) {

	if ok := checkIsExperimentResource(event.Object.GetLabels()); !ok {
		return
	}
	rateLimitingInterface.Add(reconcile.Request{
		NamespacedName: types.NamespacedName{Name: event.Object.GetName(),
			Namespace: event.Object.GetNamespace()}})

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&experimentv1.Experiment{}).
		Watches(&source.Kind{Type: &appV1.StatefulSet{}}, handler.Funcs{UpdateFunc: r.OnObjUpdate, DeleteFunc: r.OnObjDelete}).
		Watches(&source.Kind{Type: &coreV1.Service{}}, handler.Funcs{UpdateFunc: r.OnObjUpdate, DeleteFunc: r.OnObjDelete}).
		Watches(&source.Kind{Type: &v1beta1.Ingress{}}, handler.Funcs{UpdateFunc: r.OnObjUpdate, DeleteFunc: r.OnObjDelete}).
		Complete(r)
}
