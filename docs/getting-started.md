# Codedance 快速开始指南

## 前置条件

- Kubernetes 集群 (v1.24+)
- kubectl 命令行工具
- Go 1.21+ (用于本地开发)
- Docker (用于构建镜像)
- Prometheus 监控系统
- Istio 或 Nginx Ingress Controller

## 安装步骤

### 1. 安装 CRD

```bash
kubectl apply -f config/crd/canary_deployment.yaml
```

### 2. 创建 RBAC 权限

```bash
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/rolebinding.yaml
```

### 3. 部署控制器

```bash
kubectl apply -f deploy/kubernetes/controller.yaml
```

### 4. 验证安装

```bash
kubectl get pods -n codedance-system
kubectl get crd canarydeployments.deploy.codedance.io
```

## 创建第一个灰度发布

### 1. 部署应用

```bash
kubectl apply -f config/samples/deployment_example.yaml
```

### 2. 创建 CanaryDeployment

```bash
kubectl apply -f config/samples/example_canary.yaml
```

### 3. 查看发布状态

```bash
kubectl get canarydeployment -n production
kubectl describe canarydeployment codedance-app -n production
```

## 本地开发

### 1. 克隆代码

```bash
git clone https://github.com/codefarmer009/codedance.git
cd codedance
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 运行控制器

```bash
make controller
```

或者

```bash
go run cmd/controller/main.go --kubeconfig=$HOME/.kube/config
```

## 构建和部署

### 1. 构建二进制

```bash
make build
```

### 2. 构建 Docker 镜像

```bash
make docker-build VERSION=v1.0.0
```

### 3. 推送镜像

```bash
make docker-push VERSION=v1.0.0
```

### 4. 部署到集群

```bash
make deploy
```

## 监控和调试

### 查看控制器日志

```bash
kubectl logs -f -n codedance-system -l app=codedance-controller
```

### 查看 Prometheus 指标

访问 Prometheus UI，查询应用指标：

```promql
rate(http_requests_total[5m])
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
```

### 手动暂停发布

编辑 CanaryDeployment 资源，修改 phase 为 "Paused"：

```bash
kubectl edit canarydeployment codedance-app -n production
```

### 手动回滚

删除 CanaryDeployment 资源会触发自动清理：

```bash
kubectl delete canarydeployment codedance-app -n production
```

## 常见问题

### 1. 控制器无法连接 Prometheus

检查 Prometheus URL 配置是否正确：

```bash
kubectl edit deployment codedance-controller -n codedance-system
```

### 2. 流量权重未生效

确认使用的流量管理器（Istio 或 Nginx）已正确安装和配置。

### 3. 指标查询失败

验证 Prometheus 中是否存在对应的指标：

```bash
kubectl port-forward -n monitoring svc/prometheus 9090:9090
```

然后访问 http://localhost:9090 查看指标。

## 下一步

- 阅读[架构文档](architecture.md)了解系统设计
- 查看[API 参考](api-reference.md)了解 CRD 详细配置
- 参考[最佳实践](best-practices.md)优化发布策略
