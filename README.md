# simple-k8s-operator
simple-k8s-operator


1. Operator initialisation

```shell
operator-sdk init --domain=home.com --repo github.com/kalpak44/simple-k8s-operator
```


2. Create api

```shell
simple-k8s-operator % operator-sdk create api \
  --group=home \
  --version=v1 \
  --kind=Backup \
  --resource=true \
  --controller=true
```


