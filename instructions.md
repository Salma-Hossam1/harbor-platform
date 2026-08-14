Important for your restart procedure

Since you said Minikube restart causes this Service to return to NodePort, add this command to your recovery steps:

kubectl patch svc ingress-nginx-controller \
  -n ingress-nginx \
  -p '{"spec":{"type":"LoadBalancer"}}'

So your two recovery commands are now:

1. Restore MetalLB pool

kubectl apply -f - <<'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: config
  namespace: metallb-system
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - 192.168.49.240-192.168.49.250
EOF

2. Restore Ingress → LoadBalancer

kubectl patch svc ingress-nginx-controller \
  -n ingress-nginx \
  -p '{"spec":{"type":"LoadBalancer"}}'

That's all we need to do for this step.