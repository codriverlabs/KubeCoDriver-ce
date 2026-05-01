# Authentication Flow

The kubecodriver-sdk-collector uses Kubernetes TokenRequest API for secure authentication. Here's how it works:

```
Token Generation (When CoDriverJob is created):

CoDriverJob Controller                 K8s API Server
       |                                    |
       |---[1. TokenRequest]-------------->|
       |   (CoDriverJob bound,              |
       |    Duration + 10s,                |
       |    Audience: kubecodriver-sdk-collector)   |
       |                                   |
       |<--[2. Signed Token]---------------|
       |                                   |
       |--[3. Token to Pod]--------------->|
       |   (via env var)                   |


Token Validation (When collector receives data):

Profiler Pod          Collector           K8s API Server
     |                    |                    |
     |                    |                    |
     |--[1. POST Data]-->|                    |
     | Bearer Token      |                    |
     |                   |                    |
     |                   |--[2. TokenReview]->|
     |                   |                    |
     |                   |<--[3. Verified]----|
     |                   |   - Authenticated  |
     |                   |   - Not Expired    |
     |                   |   - Correct Job    |
     |                   |   - Valid Audience |
     |                   |                    |
     |<--[4. Response]---|                    |
     |                   |                    |
```

## Security Features

1. **Token Binding**: Each token is cryptographically bound to a specific CoDriverJob
2. **Limited Lifetime**: Tokens expire automatically after job duration + 10 seconds
3. **Audience Restriction**: Tokens can only be used with kubecodriver-sdk-collector
4. **Audit Logging**: All token operations are logged in Kubernetes audit logs
5. **Native Security**: Leverages Kubernetes built-in PKI infrastructure
6. **No Shared Secrets**: No manual key management required
