apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: resources-reader
  namespace: {{.Namespace}}
rules:
  - apiGroups: [ "" ]
    resources: [ "secrets" ]
    verbs: [ "get","create" ]
  - apiGroups: [ "kustomize.toolkit.fluxcd.io" ]
    resources: [ "kustomizations" ]
    verbs: [ "get" ]
  - apiGroups: [ "helm.toolkit.fluxcd.io" ]
    resources: [ "helmreleases" ]
    verbs: [ "get" ]
  - apiGroups: [ "source.toolkit.fluxcd.io" ]
    resources: [ "helmrepositories" ]
    verbs: [ "get", "list" ]
  - apiGroups: [ "source.toolkit.fluxcd.io" ]
    resources: [ "gitrepositories" ]
    verbs: [ "get","list" ]
