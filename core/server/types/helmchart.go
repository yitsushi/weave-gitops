package types

import (
	"time"

	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ProtoToHelmChart(helmChartReq *pb.AddHelmChartReq) v1beta1.HelmChart {
	return v1beta1.HelmChart{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1beta1.HelmChartKind,
			APIVersion: v1beta1.GroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      helmChartReq.HelmChart.Name,
			Namespace: helmChartReq.Namespace,
			Labels:    getGitopsLabelMap(helmChartReq.AppName),
		},
		Spec: v1beta1.HelmChartSpec{
			Chart:   helmChartReq.HelmChart.Chart,
			Version: helmChartReq.HelmChart.Version,
			SourceRef: v1beta1.LocalHelmChartSourceReference{
				Kind: helmChartReq.HelmChart.SourceRef.Kind.String(),
				Name: helmChartReq.HelmChart.SourceRef.Name,
			},
			Interval: metav1.Duration{Duration: time.Minute * 1},
		},
		Status: v1beta1.HelmChartStatus{},
	}
}

func HelmChartToProto(helmchart *v1beta1.HelmChart) *pb.HelmChart {
	return &pb.HelmChart{
		Name:      helmchart.Name,
		Namespace: helmchart.Namespace,
		SourceRef: &pb.SourceRef{
			Kind: getSourceKind(helmchart.Spec.SourceRef.Kind),
			Name: helmchart.Name,
		},
		Chart:   helmchart.Spec.Chart,
		Version: helmchart.Spec.Version,
		Interval: &pb.Interval{
			Minutes: 1,
		},
		Conditions: mapConditions(helmchart.Status.Conditions),
	}
}
