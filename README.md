# Simple Kubernetes Operator

A guide to creating and deploying a simple Kubernetes Operator using Operator SDK.

---

## 1. Operator Initialization

Initialize a new operator project:

```bash
operator-sdk init --domain=home.com --repo=github.com/kalpak44/simple-k8s-operator
```

---

## 2. Create API

Generate the API and controller for the `Backup` custom resource:

```bash
operator-sdk create api \
  --group home \
  --version v1 \
  --kind Backup \
  --resource=true \
  --controller=true
```

---

## 3. Update `Backup` CRD Spec and Status

Define the desired state for the `Backup` resource in Go:

```go
// BackupSpec defines the desired state of Backup.
type BackupSpec struct {
    // Database name to back up
    Database string `json:"database"`

    // Schedule as a cron expression
    Schedule string `json:"schedule"`
}
```

Generate deepcopy methods and CRD manifests:

```bash
make generate   # Regenerates deepcopy methods and OpenAPI schema
make manifests  # Generates CRD manifests based on your updates
```

---

## 4. Build and Deploy Operator

Build the Docker image and deploy the operator to the cluster:

```bash
make generate   # Update deepcopy and OpenAPI schema
make manifests  # Regenerate CRD manifests
export IMG=kalpak44/simple-k8s-operator:dev
make docker-build IMG=$IMG
make deploy      IMG=$IMG
```

---

## 5. Example Custom Resource

Create a sample `Backup` manifest in `backup-sample.yaml`:

```yaml
apiVersion: home.home.com/v1
kind: Backup
metadata:
  name: demo-backup
  namespace: default
spec:
  database: "postgres-db"
  schedule: "*/1 * * * *"  # Every minute
```

Apply the example:

```bash
kubectl apply -f backup-sample.yaml
```
