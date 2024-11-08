tls certificate has expired or is not yet valid error
undeploy and redeploy fixed the issue

```console
Error from server: error when creating "config/samples/": admission webhook "mutation.gatekeeper.sh" denied the request: failed to resolve external data placeholders: failed to send external data request to provider external-data-provider: failed to send external data request: Post "https://external-data-provider.gatekeeper-system:8090": tls: failed to verify certificate: x509: certificate has expired or is not yet valid: current time 2024-11-08T20:07:39Z is after 2024-11-08T18:01:41Z
```
