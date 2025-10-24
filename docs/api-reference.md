# API 参考文档

## CanaryDeployment CRD

### 概述

`CanaryDeployment` 是 Codedance 的核心自定义资源，用于定义灰度发布配置。

### API 版本

- **Group**: `deploy.codedance.io`
- **Version**: `v1alpha1`
- **Kind**: `CanaryDeployment`

### Spec 字段

#### targetDeployment (必需)

- **类型**: `string`
- **描述**: 目标 Deployment 的名称
- **示例**: `"codedance"`

#### canaryVersion (必需)

- **类型**: `string`
- **描述**: 灰度版本的镜像标签
- **示例**: `"codedance:v2.0.0"`

#### strategy (必需)

发布策略配置。

##### strategy.type

- **类型**: `string`
- **可选值**: `Linear`, `Exponential`, `Manual`
- **描述**: 发布策略类型
- **示例**: `"Linear"`

##### strategy.steps

- **类型**: `array`
- **描述**: 发布步骤列表

每个步骤包含：

- **weight** (必需): 流量权重 (0-100)
- **pause** (必需): 暂停时间 (如 "5m", "10s")
- **metrics** (可选): 该步骤的指标检查

#### metrics (必需)

监控指标配置。

##### metrics.successRate

- **threshold** (必需): 成功率阈值 (0-100)
- **query** (可选): 自定义 PromQL 查询

##### metrics.latency

- **p99** (必需): P99 延迟阈值 (如 "500ms")
- **query** (可选): 自定义 PromQL 查询

##### metrics.errorRate

- **threshold** (必需): 错误率阈值 (0-100)
- **query** (可选): 自定义 PromQL 查询

#### autoRollback (可选)

自动回滚配置。

- **enabled**: 是否启用自动回滚
- **onMetricsFail**: 指标异常时回滚
- **onPodCrash**: Pod 崩溃时回滚

### Status 字段

#### phase

当前发布阶段：

- `Progressing`: 发布进行中
- `Paused`: 已暂停
- `Completed`: 已完成
- `Failed`: 已失败

#### currentStep

当前执行的步骤索引。

#### currentWeight

当前灰度流量权重。

#### reason

状态原因说明。

#### lastUpdateTime

最后更新时间。

#### conditions

状态条件列表。

## 完整示例

```yaml
apiVersion: deploy.codedance.io/v1alpha1
kind: CanaryDeployment
metadata:
  name: myapp
  namespace: production
spec:
  targetDeployment: myapp
  canaryVersion: myapp:v2.0.0
  
  strategy:
    type: Linear
    steps:
      - weight: 10
        pause: 5m
        metrics:
          - name: success_rate
            threshold: 99.0
      - weight: 50
        pause: 10m
      - weight: 100
        pause: 0s
  
  metrics:
    successRate:
      threshold: 99.5
      query: |
        sum(rate(http_requests_total{app="myapp",status=~"2.."}[5m])) /
        sum(rate(http_requests_total{app="myapp"}[5m])) * 100
    
    latency:
      p99: 500ms
      query: |
        histogram_quantile(0.99,
          sum(rate(http_request_duration_seconds_bucket{app="myapp"}[5m])) by (le))
    
    errorRate:
      threshold: 0.5
      query: |
        sum(rate(http_requests_total{app="myapp",status=~"5.."}[5m])) /
        sum(rate(http_requests_total{app="myapp"}[5m])) * 100
  
  autoRollback:
    enabled: true
    onMetricsFail: true
    onPodCrash: true
```

## Kubectl 命令

### 创建灰度发布

```bash
kubectl apply -f canary.yaml
```

### 查看所有灰度发布

```bash
kubectl get canarydeployments -A
kubectl get cd -A  # 使用短名称
```

### 查看详细信息

```bash
kubectl describe canarydeployment myapp -n production
```

### 查看状态

```bash
kubectl get canarydeployment myapp -n production -o jsonpath='{.status.phase}'
```

### 删除灰度发布

```bash
kubectl delete canarydeployment myapp -n production
```

## 事件和日志

查看相关事件：

```bash
kubectl get events -n production --field-selector involvedObject.name=myapp
```

查看控制器日志：

```bash
kubectl logs -n codedance-system -l app=codedance-controller -f
```
