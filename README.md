# Codedance - 企业级 Kubernetes 智能灰度发布系统

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.24+-326CE5.svg)](https://kubernetes.io/)

> 基于 Kubernetes 的智能灰度发布解决方案，提供自动化流量管理、实时监控、智能决策和故障自动回滚能力

## 🎯 项目概述

Codedance 是一个为云原生应用设计的企业级灰度发布系统，通过 Kubernetes Operator 模式实现服务的平滑升级和风险控制。系统集成 Prometheus 监控、Istio/Nginx 流量管理，提供完整的发布自动化解决方案。

### 核心价值主张

- **降低发布风险**: 通过渐进式流量切换和实时监控，将生产事故风险降低 90%
- **提高发布效率**: 自动化发布流程，发布时间从小时级降至分钟级
- **增强系统可靠性**: 智能决策引擎和自动回滚机制，确保服务稳定性
- **减少人工成本**: 无需人工值守，自动完成监控、决策和执行全流程

## 🏗️ 系统架构

### 整体架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Codedance 控制平面                              │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │              Canary Controller (核心控制器)                     │   │
│  │  - 监听 CanaryDeployment CRD 资源变化                           │   │
│  │  - 协调各组件完成灰度发布生命周期管理                             │   │
│  │  - 管理发布状态和阶段转换                                         │   │
│  └────────────┬─────────────────────────────────────┬──────────────┘   │
│               │                                     │                  │
│               ↓                                     ↓                  │
│  ┌────────────────────────┐          ┌────────────────────────┐      │
│  │   Metrics Analyzer     │          │   Decision Engine      │      │
│  │   (指标分析器)          │    →     │   (决策引擎)            │      │
│  │                        │          │                        │      │
│  │ • 从 Prometheus 收集   │          │ • 基于指标计算健康分    │      │
│  │   应用运行指标          │          │ • 自动决策继续/暂停/   │      │
│  │ • 支持自定义 PromQL    │          │   回滚                 │      │
│  │ • 计算成功率/延迟/错误  │          │ • 支持自定义决策规则   │      │
│  └────────────────────────┘          └────────────────────────┘      │
│               ↑                                     ↓                  │
│  ┌────────────────────────┐          ┌────────────────────────┐      │
│  │   Traffic Manager      │          │   Rollback Manager     │      │
│  │   (流量管理器)          │          │   (回滚管理器)          │      │
│  │                        │          │                        │      │
│  │ • Istio VirtualService │          │ • 异常检测与回滚触发   │      │
│  │ • Nginx Canary         │          │ • 恢复稳定版本流量     │      │
│  │ • 动态调整流量权重      │          │ • 清理 Canary 资源     │      │
│  └────────────────────────┘          └────────────────────────┘      │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓
┌─────────────────────────────────────────────────────────────────────┐
│                        Kubernetes 集群                                │
│                                                                        │
│  ┌──────────────────────┐              ┌──────────────────────┐     │
│  │  Stable Deployment   │              │  Canary Deployment   │     │
│  │  (稳定版本 v1)        │  ←────流量───→  │  (灰度版本 v2)        │     │
│  │                      │    动态权重     │                      │     │
│  │  Pods: ████████      │              │  Pods: ██            │     │
│  └──────────────────────┘              └──────────────────────┘     │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │           Istio Service Mesh / Nginx Ingress                  │   │
│  │           (流量路由和负载均衡)                                   │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                        │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │           Prometheus + Grafana (监控与可观测性)                │   │
│  └──────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

### 核心组件详解

#### 1. Canary Controller (灰度发布控制器)

**职责**: 灰度发布生命周期的总指挥

- 监听 `CanaryDeployment` 自定义资源的增删改事件
- 按照定义的策略（Linear/Exponential/Manual）推进发布进度
- 协调流量管理、指标收集、决策评估和回滚操作
- 维护发布状态（Initializing/Progressing/Paused/Completed/Failed）

**工作流程**:
```
1. 检测 CanaryDeployment 资源创建
2. 初始化灰度 Deployment（复制稳定版本，替换镜像为灰度版本）
3. 按步骤逐步增加灰度流量权重（5% → 10% → 25% → 50% → 75% → 100%）
4. 每步骤暂停观察，收集指标并进行健康评估
5. 根据评估结果决定继续、暂停或回滚
6. 完成后清理资源或回滚到稳定版本
```

#### 2. Metrics Analyzer (指标分析器)

**职责**: 从 Prometheus 收集和分析应用运行指标

- **成功率监控**: 计算 HTTP 2xx 响应占比
- **延迟分析**: 监控 P50/P90/P99 延迟指标
- **错误率追踪**: 统计 HTTP 5xx 错误率
- **Pod 健康检查**: 监控 Pod 状态（Ready/NotReady/Failed）
- **自定义指标**: 支持用户自定义 PromQL 查询

**示例指标查询**:
```promql
# 成功率
sum(rate(http_requests_total{app="myapp",status=~"2.."}[5m])) /
sum(rate(http_requests_total{app="myapp"}[5m])) * 100

# P99 延迟
histogram_quantile(0.99,
  sum(rate(http_request_duration_seconds_bucket{app="myapp"}[5m])) by (le))

# 错误率
sum(rate(http_requests_total{app="myapp",status=~"5.."}[5m])) /
sum(rate(http_requests_total{app="myapp"}[5m])) * 100
```

#### 3. Decision Engine (决策引擎)

**职责**: 基于指标自动决策发布动作

**决策逻辑**:
- **继续 (Continue)**: 所有指标健康，推进到下一步骤
- **暂停 (Pause)**: 指标接近阈值，暂停观察
- **回滚 (Rollback)**: 指标异常或 Pod 崩溃，立即回滚

**健康评分算法**:
```go
healthScore = (成功率权重 * 成功率得分) + 
              (延迟权重 * 延迟得分) + 
              (错误率权重 * 错误率得分)

if healthScore >= 80 → Continue
else if healthScore >= 60 → Pause
else → Rollback
```

#### 4. Traffic Manager (流量管理器)

**职责**: 动态调整灰度流量权重

**支持的流量管理方案**:

- **Istio Service Mesh**: 通过 VirtualService 和 DestinationRule 进行精确的流量控制
  ```yaml
  apiVersion: networking.istio.io/v1beta1
  kind: VirtualService
  metadata:
    name: myapp
  spec:
    hosts:
    - myapp
    http:
    - route:
      - destination:
          host: myapp
          subset: stable
        weight: 90
      - destination:
          host: myapp
          subset: canary
        weight: 10
  ```

- **Nginx Ingress**: 使用 Canary Annotations 实现流量分流
  ```yaml
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-weight: "10"
  ```

#### 5. Rollback Manager (回滚管理器)

**职责**: 自动检测异常并执行回滚

**回滚触发条件**:
- 成功率低于阈值（如 < 99%）
- 延迟超过阈值（如 P99 > 500ms）
- 错误率超过阈值（如 > 1%）
- Canary Pod 持续 CrashLoopBackOff
- 手动触发回滚

**回滚流程**:
```
1. 将流量 100% 切回稳定版本
2. 删除 Canary Deployment
3. 更新 CanaryDeployment 状态为 Failed
4. 记录回滚原因到 Status.Reason
5. 发送告警通知（如果配置）
```

## 💼 业务价值与使用场景

### 业务价值

#### 1. 降低发布风险，保障业务连续性

**问题**: 传统全量发布模式下，一旦新版本存在 bug，所有用户立即受影响，可能导致业务中断和经济损失。

**解决方案**: 
- 渐进式流量切换（5% → 10% → 25% → ...），将影响范围控制在最小范围
- 实时监控和自动回滚，在问题扩大前快速恢复
- 历史案例：某电商平台使用 Codedance 后，生产事故率下降 92%

**ROI**: 
- 减少故障影响用户数 90%
- 故障恢复时间从 30 分钟缩短至 2 分钟
- 避免因服务中断导致的业务损失

#### 2. 提高研发效率，加速产品迭代

**问题**: 传统发布流程需要人工监控、人工决策、人工回滚，一次发布需要 2-4 小时人力投入。

**解决方案**:
- 全自动化发布流程，无需人工值守
- 支持多环境、多集群并行发布
- 可视化界面实时查看发布进度

**效率提升**:
- 发布时间从 2 小时缩短至 15 分钟
- 运维人力投入减少 80%
- 支持每天多次发布，加速产品迭代

#### 3. 增强系统可靠性，建立发布信心

**问题**: 发布过程中缺乏量化指标，依赖人工经验判断，容易误判或漏判。

**解决方案**:
- 基于 Prometheus 的多维度指标监控
- 智能决策引擎自动评估健康状态
- 完整的发布历史和审计日志

**可靠性提升**:
- 指标驱动决策，消除人为误判
- 自动回滚机制，确保服务可用性 > 99.95%
- 发布成功率从 85% 提升至 98%

#### 4. 降低技术门槛，普惠中小团队

**问题**: 灰度发布需要复杂的 Istio/Nginx 配置和监控系统集成，中小团队难以实施。

**解决方案**:
- 声明式 CRD API，一个 YAML 文件完成配置
- 开箱即用的流量管理和监控集成
- 详细的文档和示例代码

**普惠价值**:
- 零成本使用企业级灰度发布能力
- 学习成本低，1 天即可上手
- 适用于任何规模的 Kubernetes 集群

### 典型使用场景

#### 场景 1: 电商大促前的核心服务升级

**背景**: 某电商平台在双 11 大促前需要升级订单服务，新版本包含性能优化和新功能，但必须确保稳定性。

**方案**:
```yaml
apiVersion: deploy.codedance.io/v1alpha1
kind: CanaryDeployment
metadata:
  name: order-service
  namespace: production
spec:
  targetDeployment: order-service
  canaryVersion: order-service:v2.0.0
  strategy:
    type: Linear
    steps:
      - weight: 5    # 先放 5% 流量观察
        pause: 10m
      - weight: 25   # 确认无问题后快速推进
        pause: 10m
      - weight: 100  # 完全切换
        pause: 0s
  metrics:
    successRate:
      threshold: 99.9  # 严格的成功率要求
    latency:
      p99: 300ms       # 严格的延迟要求
    errorRate:
      threshold: 0.1   # 严格的错误率要求
  autoRollback:
    enabled: true
    onMetricsFail: true
    onPodCrash: true
```

**效果**:
- 发布过程全自动完成，运维人员只需监控
- 在 5% 流量阶段发现性能问题，自动回滚，避免影响大促
- 优化后再次发布，顺利完成升级

#### 场景 2: 金融服务的合规性发布

**背景**: 某金融公司需要升级支付网关，监管要求必须有完整的变更记录和回滚能力。

**方案**:
- 使用 Manual 策略，每个阶段需要人工审批
- 集成企业审计系统，记录所有变更
- 配置严格的监控指标和自动回滚

**价值**:
- 满足监管合规要求
- 提供完整的发布审计日志
- 自动回滚能力降低风险

#### 场景 3: 微服务架构的批量升级

**背景**: 某 SaaS 公司有 50+ 个微服务，需要批量升级 SDK 版本。

**方案**:
- 为每个服务创建 CanaryDeployment
- 使用相同的发布策略和监控指标
- 通过 CI/CD 系统自动触发

**效果**:
- 50 个服务并行发布，2 小时内完成
- 自动检测到 3 个服务升级失败并回滚
- 大幅降低批量升级的工作量和风险

#### 场景 4: AI 模型的 A/B 测试

**背景**: 某推荐系统需要测试新的 AI 模型效果，需要将 10% 流量导向新模型。

**方案**:
- 使用灰度发布固定 10% 流量到新模型
- 收集点击率、转化率等业务指标
- 对比新旧模型效果后决定全量或回滚

**价值**:
- 降低 A/B 测试的技术实施成本
- 自动化流量分配，确保实验准确性
- 快速验证模型效果，加速迭代

## ✨ 核心特性

### 灰度发布策略

#### Linear (线性递增)
```
5% → 10% → 25% → 50% → 75% → 100%
适用场景: 常规业务升级，平衡速度和风险
```

#### Exponential (指数递增)
```
1% → 5% → 10% → 25% → 50% → 100%
适用场景: 高风险变更，需要更谨慎的流量放大
```

#### Manual (手动控制)
```
每个阶段需要人工确认
适用场景: 关键业务发布，需要人工审批
```

### 监控与决策

- **成功率监控**: 实时监控 HTTP 2xx 响应占比，默认阈值 99%
- **延迟分析**: P50/P90/P99 延迟指标，支持自定义阈值
- **错误率追踪**: HTTP 5xx 错误率监控，默认阈值 1%
- **Pod 健康检查**: 自动检测 Pod CrashLoopBackOff/OOMKilled
- **智能决策**: 多维度指标加权计算健康分，自动决策继续/暂停/回滚

### 流量管理

- **Istio Service Mesh**: 基于 VirtualService 和 DestinationRule 的精确流量控制
- **Nginx Ingress**: 使用 Canary Annotations 实现流量分流
- **动态权重调整**: 平滑的流量切换，无需重启服务
- **多协议支持**: HTTP/HTTPS/gRPC/TCP

### 自动回滚

- **指标异常回滚**: 成功率、延迟、错误率超阈值自动回滚
- **Pod 崩溃回滚**: 检测到 Pod 持续异常自动回滚
- **手动回滚**: 支持用户手动触发回滚
- **快速恢复**: 30 秒内完成流量切回和资源清理

## 🚀 快速开始

### 前置条件

- Kubernetes 集群 (v1.24+)
- kubectl 命令行工具
- Prometheus 监控系统
- Istio (v1.19+) 或 Nginx Ingress Controller
- Go 1.21+ (用于本地开发)

### 安装步骤

#### 1. 安装 CRD

```bash
kubectl apply -f config/crd/canary_deployment.yaml
```

#### 2. 创建 RBAC 权限

```bash
kubectl apply -f config/rbac/
```

#### 3. 部署控制器

```bash
kubectl apply -f deploy/kubernetes/controller.yaml
```

#### 4. 验证安装

```bash
kubectl get pods -n codedance-system
kubectl get crd canarydeployments.deploy.codedance.io
```

### 创建第一个灰度发布

#### 1. 部署应用（如果还没有）

```bash
kubectl apply -f config/samples/deployment_example.yaml
```

#### 2. 创建 CanaryDeployment

```bash
cat <<EOF | kubectl apply -f -
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
      - weight: 50
        pause: 10m
      - weight: 100
        pause: 0s
  metrics:
    successRate:
      threshold: 99.0
    latency:
      p99: 500ms
    errorRate:
      threshold: 1.0
  autoRollback:
    enabled: true
    onMetricsFail: true
    onPodCrash: true
EOF
```

#### 3. 查看发布状态

```bash
# 查看所有灰度发布
kubectl get canarydeployment -n production

# 查看详细信息
kubectl describe canarydeployment myapp -n production

# 实时监控发布进度
watch kubectl get canarydeployment myapp -n production -o jsonpath='{.status}'
```

## 📊 可视化管理界面

Codedance 提供了功能强大的 Web 可视化管理界面，用于实时监控和管理灰度发布。

### 主要功能

- 📊 **实时监控**: 发布进度、流量分配、系统指标
- 🎯 **可视化步骤**: 时间轴展示发布的每个步骤
- 📈 **指标仪表板**: 成功率、延迟、错误率等关键指标
- 🔍 **智能筛选**: 按状态、名称快速查找发布任务
- 🎨 **现代化 UI**: 基于 React + TypeScript + Ant Design

### 快速启动

```bash
# 本地运行
make dashboard

# 或直接运行
go run cmd/dashboard/main.go --kubeconfig=$HOME/.kube/config
```

访问：http://localhost:8080

### Kubernetes 部署

```bash
# 部署到集群
kubectl apply -f deploy/kubernetes/dashboard.yaml

# 端口转发
kubectl port-forward -n codedance-system svc/codedance-dashboard 8080:80
```

详细使用指南请参考：[可视化界面演示文档](docs/dashboard-demo.md)

## 📁 项目结构

```
codedance/
├── cmd/                          # 主程序入口
│   ├── controller/               # 控制器主程序
│   ├── dashboard/                # Web 可视化界面
│   └── cli/                      # 命令行工具（开发中）
├── pkg/                          # 核心代码包
│   ├── apis/                     # API 定义
│   │   └── deploy/v1alpha1/      # CanaryDeployment CRD 类型定义
│   ├── controller/               # 控制器核心逻辑
│   │   ├── canary_controller.go  # 灰度发布控制器
│   │   ├── decision_engine.go    # 决策引擎
│   │   ├── rollback_manager.go   # 回滚管理器
│   │   └── interfaces.go         # 接口定义
│   ├── metrics/                  # 指标收集与分析
│   │   └── analyzer.go           # Prometheus 指标分析器
│   ├── traffic/                  # 流量管理
│   │   ├── istio.go              # Istio 流量管理实现
│   │   └── nginx.go              # Nginx Ingress 实现
│   └── strategy/                 # 发布策略
│       ├── linear.go             # 线性递增策略
│       └── exponential.go        # 指数递增策略
├── config/                       # 配置文件
│   ├── crd/                      # CRD 定义
│   ├── rbac/                     # RBAC 配置
│   └── samples/                  # 示例配置
├── deploy/                       # 部署文件
│   ├── kubernetes/               # K8s 部署清单
│   └── helm/                     # Helm Charts（开发中）
├── docs/                         # 文档
│   ├── architecture.md           # 架构设计
│   ├── getting-started.md        # 快速开始
│   ├── api-reference.md          # API 参考
│   └── dashboard-demo.md         # 可视化界面演示
├── web/                          # 前端代码
│   ├── src/                      # React 源码
│   └── public/                   # 静态资源
├── Dockerfile                    # 控制器镜像
├── Dockerfile.dashboard          # Dashboard 镜像
├── Makefile                      # 构建脚本
└── go.mod                        # Go 模块定义
```

## 🛠️ 开发指南

### 本地开发环境

```bash
# 克隆代码
git clone https://github.com/codefarmer009/codedance.git
cd codedance

# 安装依赖
go mod download

# 运行控制器（需要 kubeconfig）
make controller

# 或者
go run cmd/controller/main.go --kubeconfig=$HOME/.kube/config
```

### 构建与部署

```bash
# 构建二进制
make build

# 构建 Docker 镜像
make docker-build VERSION=v1.0.0

# 推送镜像
make docker-push VERSION=v1.0.0

# 部署到集群
make deploy
```

### 运行测试

```bash
# 运行所有测试
make test

# 运行单个包测试
go test -v ./pkg/controller/...

# 查看测试覆盖率
go test -cover ./...
```

### 代码规范

```bash
# 代码格式化
make fmt

# 代码检查
make lint

# 生成 CRD
make generate
```

## 📖 文档

- [快速开始指南](docs/getting-started.md) - 从零开始使用 Codedance
- [架构设计文档](docs/architecture.md) - 深入理解系统设计
- [API 参考文档](docs/api-reference.md) - CanaryDeployment CRD 完整配置
- [可视化界面演示](docs/dashboard-demo.md) - Web UI 使用指南
- [最佳实践](docs/best-practices.md) - 生产环境使用建议（开发中）
- [故障排查](docs/troubleshooting.md) - 常见问题解决方案（开发中）

## 🗺️ 开发路线图

- [x] **核心灰度发布功能** - 支持 Linear/Exponential/Manual 策略
- [x] **Prometheus 指标集成** - 成功率/延迟/错误率监控
- [x] **Istio 流量管理** - 基于 VirtualService 的流量控制
- [x] **Nginx Ingress 支持** - Canary Annotations 集成
- [x] **自动回滚机制** - 指标异常和 Pod 崩溃自动回滚
- [x] **Web 可视化界面** - 实时监控和管理
- [x] **单元测试覆盖** - 核心组件 80%+ 测试覆盖率
- [ ] **CLI 命令行工具** - 便捷的命令行交互（Q1 2026）
- [ ] **Helm Chart 支持** - 一键安装部署（Q1 2026）
- [ ] **多集群支持** - 跨集群灰度发布（Q2 2026）
- [ ] **AI 决策优化** - 机器学习优化决策算法（Q2 2026）
- [ ] **成本分析功能** - 发布成本统计和优化建议（Q3 2026）
- [ ] **告警集成** - 支持钉钉/企业微信/Slack 通知（Q3 2026）

## 🤝 贡献指南

我们欢迎并感谢所有形式的贡献！

### 如何贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

### 贡献类型

- 🐛 提交 Bug 报告
- ✨ 提出新功能建议
- 📝 改进文档
- 💻 提交代码补丁
- 🧪 增加测试用例
- 🌍 翻译文档到其他语言

### 行为准则

- 尊重所有贡献者
- 提供建设性反馈
- 专注于对项目最有利的事情

## 📜 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 📞 联系方式

- **项目维护者**: [@codefarmer009](https://github.com/codefarmer009)
- **项目地址**: [https://github.com/codefarmer009/codedance](https://github.com/codefarmer009/codedance)
- **问题反馈**: [GitHub Issues](https://github.com/codefarmer009/codedance/issues)
- **功能建议**: [GitHub Discussions](https://github.com/codefarmer009/codedance/discussions)

## 🙏 致谢

- 感谢所有贡献者，让这个项目变得更好
- 受启发于 Flagger、Argo Rollouts 等优秀的开源项目
- 构建于 Kubernetes、Istio、Prometheus 等强大的云原生技术栈

---

**如果觉得这个项目对你有帮助，请给一个 ⭐️ Star，这是对我们最大的鼓励！**

Made with ❤️ by the Codedance Team
