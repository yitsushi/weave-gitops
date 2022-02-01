---
title: "Securing NATS"
sidebar_position: 10
hide_title: true
---

import TierLabel from "../\_components/TierLabel";

# Securing NATS <TierLabel tiers="enterprise" />

NATS is a messaging server used by Weave Gitops Enterprise. It exposes a TCP endpoint that needs to be reachable by agents running on leaf clusters when you _Connect an external (not capi) cluster_. This guide describes how to use TLS to secure traffic between leaf clusters and the management cluster. This guide uses `cert-manager` to generate the certificate but it still applies and can be used without it. It is highly recommended to enable TLS connections for NATS \_before\_ connecting any external leaf clusters to WGE, otherwise you may need to re-connect any leaf clusters that were added prior to enabling TLS.

Setting up TLS for NATS requires the use of a certificate. This certificate can be added to the cluster either manually as a secret or provisioned automatically via `cert-manager`. The following manifest shows how to provision such a certificate automatically.

```yaml title=".wego-system/clusters/my-management-cluster/system/nats-cert.yaml"
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: nats-tls
  namespace: wego-system
spec:
  dnsNames:
    - <hostname-to-use> # e.g. my-management-cluster.example.com
  issuerRef:
    group: cert-manager.io
    kind: <issuer-type> # Issuer or ClusterIssuer
    name: <issuer-name> # the name of an existing Issuer or ClusterIssuer
  secretName: nats-tls
  usages:
    - digital signature
    - key encipherment
```

Add this manifest to the directory `./wego-system/clusters/my-management-cluster/system` of your cluster repository, then commit and push to your Git provider. The reconciliation process should apply it within a minute. Ensure that the certificate has been successfully provisioned by running the following command.

```console
kubectl get secrets -n wego-system nats-tls
```

Once the new certificate has been provisioned, we need to update the NATS configuration to use it. Update the `HelmRelease` located in `.wego-system/clusters/my-management-cluster/system/weave-gitops-enterprise.yaml` which is used for configuring the NATS subchart. Add and configure the highlighted lines below.

1. The `agentTemplate` parameters are included in the manifests that the agent runs, telling the agent how to connect to the management cluster.
2. The `nats` parameters configure the nats deployment itself, switching TLS on. This will result in exposing the TLS certificate as a mounted volume so that it is accessible under the `/etc/nats-tls` directory of the NATS container. The `extraFlags` configuration that is supplied instructs NATS to require TLS for client connections.

After this change is applied, NATS will be accessible only via TLS connections.

```yaml {13-15,17-29} title=".wego-system/clusters/my-management-cluster/system/weave-gitops-enterprise.yaml"
---
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
  name: weave-gitops-enterprise
  namespace: wego-system
spec:
  chart: ...
  intervals: ...
  values.yaml: |
    ...

    agentTemplate:
      natsScheme: nats
      natsURL: my-management-cluster.example.com:4222

    nats:
      extraFlags:
        tls: ""
        tlskey: /etc/nats-tls/tls.key
        tlscert: /etc/nats-tls/tls.crt
      extraVolumeMounts:
      - name: nats-tls
        mountPath: /etc/nats-tls/
        readOnly: true
      extraVolumes:
      - name: nats-tls
        secret:
          secretName: nats-tls
```
