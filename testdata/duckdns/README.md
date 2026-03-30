# ACME webhook for Duck DNS

To run the tests a kubernetes secret manifest must be provided in this path.
The environment variable `$DUCKDNS_TOKEN` must be replaced with a base64 encoded DuckDNS token.

```bash
$ export DUCKDNS_TOKEN=$(echo -n "<token>" | base64 -w 0)
$ envsubst < secret.yaml.example > secret.yaml
```
