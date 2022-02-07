package types

import (
	"github.com/fluxcd/source-controller/api/v1beta1"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getGitopsLabelMap(appName string) map[string]string {
	labels := map[string]string{
		ManagedByLabel: managedByWeaveGitops,
		CreatedByLabel: createdBySourceController,
	}

	if appName != "" {
		labels[PartOfLabel] = appName
	}

	return labels
}

func getSourceKind(kind string) pb.SourceRef_Kind {
	switch kind {
	case v1beta1.GitRepositoryKind:
		return pb.SourceRef_GitRepository
	case v1beta1.HelmRepositoryKind:
		return pb.SourceRef_HelmRepository
	case v1beta1.BucketKind:
		return pb.SourceRef_Bucket
	default:
		return -1
	}
}

func mapConditions(conditions []metav1.Condition) []*pb.Condition {
	out := []*pb.Condition{}

	for _, c := range conditions {
		out = append(out, &pb.Condition{
			Type:      c.Type,
			Status:    string(c.Status),
			Reason:    c.Reason,
			Message:   c.Message,
			Timestamp: c.LastTransitionTime.String(),
		})
	}

	return out
}
