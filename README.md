<p align="center">
  <img src="https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png" height="256" width="256" alt="cert-manager project logo" />
</p>

# ACME webhook for Duck DNS

A webhook to use [Duck DNS](https://www.duckdns.org) as a DNS01 ACME Issuer for cert-manager.

# Configuration

# Kubernetes

Install the helm chart:

```
helm repo add bretek-duckdns https://bretek.github.io/webhook-duckdns
helm repo update
helm install <release-name> bretek-duckdns/duckdns-webhook
```

The helm chart accepts the option `logLevel`, with value `debug`, `info`, `warn` or `error`. The default is `warn`.

Place your base64 encoded DuckDNS token in a kubernetes secret, with name `duckdns-token`.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: duckdns-token
data:
  token: $DUCKDNS_TOKEN
```

Example issuer:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # Change to your letsencrypt email
    email: certmaster@example.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-account-key
    solvers:
    - dns01:
        webhook:
          groupName: duckdns.bretek.duckdns.org
          solverName: duckdns
          config:
            apiKeySecretRef:
              name: duckdns-token
              key: token
```
