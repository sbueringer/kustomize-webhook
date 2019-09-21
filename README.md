# Kustomize Webhook

[![Go Report Card](https://goreportcard.com/badge/sbueringer/kustomize-webhook)](https://goreportcard.com/report/sbueringer/kustomize-webhook)
![](https://github.com/sbueringer/kustomize-webhook/workflows/.github/workflows/main.yml/badge.svg)

A MutatingWebhook for Kubernetes based on Kustomize.

## Summary

The kustomize-webhook is a MutatingWebhook which applies kustomize patches to Pods. The MutatingWebhook receives
the Pod resource which is then patches via a kustomize patch. The patch itself is generated via go template with the 
Pod as data.

## Deployment 

An example deployment can be found in the deploy folder. Generate certs, e.g. via:

```bash
openssl req -x509 -newkey rsa:2048 -keyout tls.key -out tls.crt -days 365 -nodes -subj "/CN=kustomize-webhook.default.svc"
```

Replace the  following vars in [deploy/webhook.yaml](deploy/webhook.yaml):

````
TLS_CRT_BASE64=$(cat tls.crt | base64)
TLS_CRT=$(cat tls.crt)
TLS_KEY=$(cat tls.key)
````

adjust the kustomize patch in the `kustomize-webhook-patches` ConfigMap and deploy the YAML file:

````
kubectl apply -f deploy/webhook.yaml
````

