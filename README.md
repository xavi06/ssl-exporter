# ssl-exporter

Monitor kubernetes, etcd certificates expired dates

## Deploy
docker build tag and push
```bash
make release
```
kubectl install (DaemonSet)
```bash
make install
```

## Prometheus
config
```yaml
- job_name: 'ssl-exporter'
  kubernetes_sd_configs:
    - role: pod
  relabel_configs:
    - source_labels: [__meta_kubernetes_namespace]
      regex: ops
      action: keep
    - source_labels: [__meta_kubernetes_pod_name]
      regex: ssl-exporter-(.+)
      action: keep
```

alert rules
```yaml
- name: ssl.rules
  rules:
    - alert: CertExpiredDaysLeft
      expr: certs_expired_days_left < 30
      labels:
        severity: warning
      annotations:
        summary: "kubernetes certs_expired_days_left < 30 days"
        description: "certs_expired_days_left is {{ $value }}"
```
