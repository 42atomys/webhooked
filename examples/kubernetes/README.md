# Atomys Webhooked on Kubernetes

The solution I personally use in my Kubernetes cluster.

In this example I will use Istio as IngressController, being the one I personally use. Of course webhooked is compatible with any type of ingress, being a proxy at layer 7.

**You can use the example as an initial configuration.**

## Workflow

First you need to apply the workload to your cluster, once the workload is installed, you can edit the configmap to configure the webhooked for your endpoints.

```sh
# Apply the example deployment files (configmap, deployment, service)
kubectl apply -f https://raw.githubusercontent.com/42Atomys/webhooked/1.0/examples/kubernetes/deployment.yaml

# Edit the configuration map to apply your redirection and configurations
kubectl edit configmap/webhooked
```

Don't forget to restart your deployment so that your webhooked takes into account the changes made to your configmap
```sh
# Restart your webhooked instance to apply the latest configuration
kubectl rollout restart deployment.apps/webhooked
```

It's all over! 🎉

Now it depends on your Ingress!

## Sugar Free: Isito Routing

If you use istio as IngressController like me, you can my virtual service (it's free)

I personally route only the prefix of version. NOTE: You can host multiple versions of configuration file with multiple virtual route ;)

```yaml
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: webhooked
spec:
  hosts:
  - atomys.codes # Change for your domain
  gateways:
  - default
  http:
  - match:
    - uri:
        prefix: /v1alpha1/webhooks
    route:
    - destination:
        port:
          number: 8080
        host: webhooked
```