package controllers

import (
	experimentv1 "experiment/api/v1"
	coreV1 "k8s.io/api/core/v1"
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

func GetContainer(image string) []coreV1.Container {
	container := coreV1.Container{}
	//impala container
	container.Name = "experiment"
	container.ImagePullPolicy = coreV1.PullAlways
	container.Image = image
	return []coreV1.Container{container}
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
