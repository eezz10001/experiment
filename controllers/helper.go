package controllers

import (
	"fmt"
	experimentv1 "github.com/eezz10001/experiment/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	LabelConstantKey   = "experiment.touchturing.com/v1" //constant label
	LabelConstantValue = "experiment"
	LabelInstanceKey   = "experiment.touchturing.com/instance" // ins name
	LabelNsKey         = "experiment.touchturing.com/ns"       //ns
)

func GetLabel(experiment *experimentv1.Experiment, labels map[string]string) map[string]string {
	if labels == nil {
		labels = map[string]string{}
	}

	labels[LabelConstantKey] = LabelConstantValue
	labels[LabelNsKey] = experiment.Namespace
	labels[LabelInstanceKey] = experiment.Name

	return labels
}

func GetContainer(experiment *experimentv1.Experiment) []coreV1.Container {
	container := coreV1.Container{}
	//impala container
	container.Name = "experiment"
	container.ImagePullPolicy = coreV1.PullIfNotPresent
	container.Image = experiment.Spec.Image

	container.Ports = experiment.Spec.Ports

	container.Resources = experiment.Spec.Resources
	return []coreV1.Container{container}
}

func GetServicePorts(experiment *experimentv1.Experiment) []coreV1.ServicePort {
	ret := make([]coreV1.ServicePort, 0)

	for _, port := range experiment.Spec.Ports {
		ret = append(ret, coreV1.ServicePort{
			Name:     port.Name,
			Protocol: port.Protocol,
			Port:     port.ContainerPort,
		})
	}
	return ret
}

func GetIngressRule(experiment *experimentv1.Experiment) []v1beta1.IngressRule {
	ret := make([]v1beta1.IngressRule, 0)

	for _, port := range experiment.Spec.Ports {
		ret = append(ret, v1beta1.IngressRule{
			Host: fmt.Sprintf("%s-%s.%s", port.Name, experiment.Name, experiment.Spec.Host),
			IngressRuleValue: v1beta1.IngressRuleValue{
				HTTP: &v1beta1.HTTPIngressRuleValue{
					Paths: []v1beta1.HTTPIngressPath{{
						Path: "/",
						Backend: v1beta1.IngressBackend{
							ServiceName: experiment.Name,
							ServicePort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: port.ContainerPort,
							},
						},
					}},
				},
			},
		})
	}
	return ret
}

func checkIsExperimentResource(label map[string]string) bool {
	if label == nil {
		return false
	}
	v, ok := label[LabelConstantKey]
	if ok && v == LabelConstantValue {
		return true
	}
	return false
}
