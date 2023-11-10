package controllers

import (
	"bytes"
	"context"
	experimentv1 "github.com/eezz10001/experiment/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"text/template"
)

const svctpl = `
apiVersion: v1
kind: Service
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace}}
`

type serviceBuilder struct {
	client.Client
	experiment *experimentv1.Experiment
	service    *coreV1.Service
	Scheme     *runtime.Scheme
}

func NewServiceBuilder(client client.Client, experiment *experimentv1.Experiment, scheme *runtime.Scheme) (*serviceBuilder, error) {
	service := &coreV1.Service{}

	err := client.Get(context.Background(), types.NamespacedName{
		Namespace: experiment.Namespace, Name: experiment.Name}, service)
	if err != nil { //have no find
		service.Name, service.Namespace = experiment.Name, experiment.Namespace
		tpl, err := template.New("service").Parse(svctpl)
		var tplRet bytes.Buffer
		if err != nil {
			return nil, err
		}

		err = tpl.Execute(&tplRet, service)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(tplRet.Bytes(), service)
		if err != nil {
			return nil, err
		}
	}

	return &serviceBuilder{Client: client, experiment: experiment, service: service, Scheme: scheme}, nil
}

// 同步属性
func (this *serviceBuilder) apply() *serviceBuilder {

	this.service.ObjectMeta.Name = this.experiment.Name
	this.service.ObjectMeta.Namespace = this.experiment.Namespace
	this.service.Spec.Type = coreV1.ServiceTypeNodePort
	this.service.Spec.Selector = GetLabel(this.experiment, nil)
	this.service.Labels = GetLabel(this.experiment, this.service.Labels)

	this.service.Spec.Ports = []coreV1.ServicePort{GetServicePorts(this.experiment)}
	return this
}

func (this *serviceBuilder) setOwner() error {
	return controllerutil.SetControllerReference(this.experiment, this.service, this.Scheme)
}

func (this *serviceBuilder) Build(ctx context.Context) (status bool, err error) {
	if this.service.CreationTimestamp.IsZero() {
		err = this.apply().setOwner()
		if err != nil {
			return false, err
		}
		status = false
		err = this.Create(ctx, this.service)
		if err != nil {
			return false, err
		}
	} else {
		status = this.service.Spec.ClusterIP != ""
	}
	return
}
