apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}
  labels:
    app.kubernetes.io/part-of: weave-gitops
    app.kubernetes.io/version: {{.AppVersion}}
