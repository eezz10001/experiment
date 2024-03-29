package controllers

import (
	"bytes"
	"context"
	experimentv1 "github.com/eezz10001/experiment/api/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

const ingresstpl = `
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace}}
`

type IngressBuilder struct {
	client.Client
	experiment *experimentv1.Experiment
	ingress    *v1beta1.Ingress
	Scheme     *runtime.Scheme
	host       string
}

func NewIngressBuilder(client client.Client, experiment *experimentv1.Experiment, scheme *runtime.Scheme, host string) (*IngressBuilder, error) {
	ingress := &v1beta1.Ingress{}

	err := client.Get(context.Background(), types.NamespacedName{
		Namespace: experiment.Namespace, Name: experiment.Name}, ingress)
	if err != nil { //have no find
		ingress.Name, ingress.Namespace = experiment.Name, experiment.Namespace
		tpl, err := template.New("ingress").Parse(ingresstpl)
		var tplRet bytes.Buffer
		if err != nil {
			return nil, err
		}

		err = tpl.Execute(&tplRet, ingress)

		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(tplRet.Bytes(), ingress)
		if err != nil {
			return nil, err
		}
	}

	return &IngressBuilder{Client: client, experiment: experiment, ingress: ingress, Scheme: scheme, host: host}, nil
}

func (this *IngressBuilder) apply() *IngressBuilder {
	this.ingress.ObjectMeta.Name = this.experiment.Name
	this.ingress.ObjectMeta.Namespace = this.experiment.Namespace
	this.ingress.Labels = GetLabel(this.experiment, this.ingress.Labels)
	this.ingress.Spec.Rules = GetIngressRule(this.experiment)
	return this
}

func (this *IngressBuilder) setOwner() error {
	return controllerutil.SetControllerReference(this.experiment, this.ingress, this.Scheme)
}

func (this *IngressBuilder) Build(ctx context.Context) (status bool, err error) {
	if this.ingress.CreationTimestamp.IsZero() {
		err = this.apply().setOwner()
		if err != nil {
			return false, err
		}
		status = false
		err = this.Create(ctx, this.ingress)
		if err != nil {
			return false, err
		}
	} else {
		patch := client.MergeFrom(this.ingress.DeepCopy())

		this.apply()
		err = this.Patch(ctx, this.ingress, patch)
		if err != nil {
			return false, err
		}
	}
	if len(this.ingress.Status.LoadBalancer.Ingress) > 0 {
		if this.ingress.Status.LoadBalancer.Ingress[0].IP != "" {
			return true, nil
		}
	}
	return false, err
}
