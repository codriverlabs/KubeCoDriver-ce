# TLS Setup for Collector

The kubecodriver-sdk-collector uses cert-manager for TLS certificate management and secure HTTPS communication with ephemeral containers.

## Architecture

```
cert-manager                    CoDriverJob Controller
     |                                 |
     |                                 |
[1]  |     +-----------------+        |
     +---->| Certificate CR  |        |
     |     +-----------------+        |
     |            |                   |
[2]  |     +------v------+           |
     +---->| TLS Secret  |           |
           +-------------+           |
                 |                   |
[3]            +----------------+    |
               | CA ConfigMap   |<---+
               +----------------+    |
                      |             |
[4]                   +------------->| Ephemeral Container
                                    | (CA cert via env var)
```

## Components

1. **Certificate CR**: Defines the certificate requirements
   - Common Name: kubecodriver-sdk-collector.kubecodriver-system.svc
   - DNS Names: Include cluster-local service names
   - Duration: 1 year with 30-day renewal

2. **TLS Secret**: Contains the certificate and private key
   - Used by collector for HTTPS
   - Automatically renewed by cert-manager

3. **CA ConfigMap**: Stores the CA certificate
   - Automatically updated by cert-manager
   - Used by ephemeral containers for verification

4. **Environment Variable**: Passes CA cert to containers
   - No volume mounts needed
   - Works with ephemeral containers

## Setup

1. Install cert-manager:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
```

2. Apply certificate resources:
```bash
kubectl apply -f config/certmanager/collector-certificate.yaml
```

3. Verify setup:
```bash
kubectl get certificate -n kubecodriver-system collector-cert
kubectl get secret -n kubecodriver-system collector-tls
kubectl get configmap -n kubecodriver-system collector-ca
```

## Security Considerations

- Certificates are automatically rotated
- Private keys never leave the cluster
- Each ephemeral container gets its own CA cert
- TLS termination at collector pod
