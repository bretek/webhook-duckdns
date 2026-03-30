<p align="center">
  <img src="https://raw.githubusercontent.com/cert-manager/cert-manager/d53c0b9270f8cd90d908460d69502694e1838f5f/logo/logo-small.png" height="256" width="256" alt="cert-manager project logo" />
</p>

# ACME webhook for Duck DNS

A webhook to use [Duck DNS](https://www.duckdns.org) as a DNS01 ACME Issuer for cert-manager.

# Configuration

Place your base64 encoded DuckDNS token in a kubernetes secret, with name "duckdns-token".

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: duckdns-token
data:
  token: $DUCKDNS_TOKEN
```
