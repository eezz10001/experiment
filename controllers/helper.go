package controllers

import (
	"context"
	"fmt"
	experimentv1 "github.com/eezz10001/experiment/api/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	container.Ports = []coreV1.ContainerPort{experiment.Spec.Port}
	container.Resources = experiment.Spec.Resources
	if experiment.Spec.Probe != (experimentv1.Probe{}) {
		container.ReadinessProbe = GetReadinessProbe(experiment)
		container.LivenessProbe = GetLivenessProbe(experiment)
	}

	container.Command = experiment.Spec.Command
	return []coreV1.Container{container}
}

func GetServicePorts(experiment *experimentv1.Experiment) coreV1.ServicePort {
	return coreV1.ServicePort{
		Name:     experiment.Spec.Port.Name,
		Protocol: experiment.Spec.Port.Protocol,
		Port:     experiment.Spec.Port.ContainerPort,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: experiment.Spec.Port.ContainerPort,
		},
	}
}

func GetReadinessProbe(experiment *experimentv1.Experiment) *coreV1.Probe {
	return &coreV1.Probe{
		ProbeHandler: coreV1.ProbeHandler{
			TCPSocket: &coreV1.TCPSocketAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: experiment.Spec.Probe.Port,
				},
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      1,
		PeriodSeconds:       10,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func GetLivenessProbe(experiment *experimentv1.Experiment) *coreV1.Probe {
	return &coreV1.Probe{
		ProbeHandler: coreV1.ProbeHandler{
			TCPSocket: &coreV1.TCPSocketAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: experiment.Spec.Probe.Port,
				},
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      3,
		PeriodSeconds:       30,
		FailureThreshold:    2,
	}
}
func GetIngressRule(experiment *experimentv1.Experiment) []v1beta1.IngressRule {
	PathType := v1beta1.PathTypePrefix
	ret := make([]v1beta1.IngressRule, 0)
	ret = append(ret, v1beta1.IngressRule{
		Host: experiment.Spec.Host,
		IngressRuleValue: v1beta1.IngressRuleValue{
			HTTP: &v1beta1.HTTPIngressRuleValue{
				Paths: []v1beta1.HTTPIngressPath{{
					PathType: &PathType,
					Path:     "/",
					Backend: v1beta1.IngressBackend{
						ServiceName: experiment.Name,
						ServicePort: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: experiment.Spec.Port.ContainerPort,
						},
					},
				}},
			},
		},
	})
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

func GetPodPhase(c client.Client, experiment *experimentv1.Experiment) (status coreV1.PodPhase) {
	pod := &coreV1.Pod{}
	err := c.Get(context.TODO(), client.ObjectKey{
		Namespace: experiment.Namespace,
		Name: fmt.Sprintf(
			"%s-0", experiment.Name),
	}, pod)

	if err != nil {
		return coreV1.PodUnknown
	}
	return pod.Status.Phase

}
