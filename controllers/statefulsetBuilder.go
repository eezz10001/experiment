package controllers

import (
	"bytes"
	"context"
	"fmt"
	experimentv1 "github.com/eezz10001/experiment/api/v1"
	appV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

const ststpl = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace}}
`

type statefulSetBuilder struct {
	client.Client
	experiment  *experimentv1.Experiment
	statefulSet *appV1.StatefulSet
	Scheme      *runtime.Scheme
}

func NewStatefulSetBuilder(client client.Client, experiment *experimentv1.Experiment, scheme *runtime.Scheme) (*statefulSetBuilder, error) {
	statefulSet := &appV1.StatefulSet{}

	err := client.Get(context.Background(), types.NamespacedName{
		Namespace: experiment.Namespace, Name: experiment.Name,
	}, statefulSet)

	if err != nil { //have no find
		statefulSet.Name, statefulSet.Namespace = experiment.Name, experiment.Namespace
		tpl, err := template.New("statefulSet").Parse(ststpl)
		var tplRet bytes.Buffer
		if err != nil {
			return nil, err
		}

		err = tpl.Execute(&tplRet, statefulSet)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(tplRet.Bytes(), statefulSet)
		if err != nil {
			return nil, err
		}
	}

	return &statefulSetBuilder{Client: client, experiment: experiment, statefulSet: statefulSet, Scheme: scheme}, nil
}

// Synchronization attributes
func (this *statefulSetBuilder) apply() *statefulSetBuilder {

	this.statefulSet.ObjectMeta.Name = this.experiment.Name
	this.statefulSet.ObjectMeta.Namespace = this.experiment.Namespace

	selectorLabel := GetLabel(this.experiment, nil)
	this.experiment.ObjectMeta.Labels = GetLabel(this.experiment, this.experiment.ObjectMeta.Labels)

	//.spec
	this.statefulSet.Spec.PodManagementPolicy = appV1.ParallelPodManagement
	this.statefulSet.Spec.Selector = &metaV1.LabelSelector{MatchLabels: selectorLabel}
	this.statefulSet.Spec.Template.ObjectMeta.Labels = selectorLabel
	this.statefulSet.Spec.ServiceName = this.statefulSet.Name

	//Containers
	this.statefulSet.Spec.Template.Spec.Containers = GetContainer(this.experiment)

	return this
}

func (this *statefulSetBuilder) setOwner() error {
	return controllerutil.SetControllerReference(this.experiment, this.statefulSet, this.Scheme)
}

func (this *statefulSetBuilder) Build(ctx context.Context) (status bool, err error) {

	if this.statefulSet.CreationTimestamp.IsZero() {
		err = this.apply().setOwner()
		if err != nil {
			return false, err
		}
		status = false
		err = this.Create(ctx, this.statefulSet)
		if err != nil {
			return false, err
		}
	} else {
		patch := client.MergeFrom(this.statefulSet.DeepCopy())
		this.apply()
		b, _ := json.Marshal(this.statefulSet)
		fmt.Println(string(b))
		fmt.Println(this.statefulSet.Status.Replicas, this.statefulSet.Status.ReadyReplicas)
		status = this.statefulSet.Status.Replicas == this.statefulSet.Status.ReadyReplicas && this.statefulSet.Status.ReadyReplicas != 0 && GetPodPhase(this.Client, this.experiment) == coreV1.PodRunning
		err = this.Patch(ctx, this.statefulSet, patch)
		if err != nil {
			return false, err
		}
	}
	return status, nil
}
