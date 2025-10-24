# Codedance - Kubernetes 智能发布系统

基于 Kubernetes 的智能发布系统，支持灰度发布、实时监控、自动暂停和回滚。

## 概述

Codedance 是一个企业级的 Kubernetes 灰度发布解决方案，提供：

- **智能灰度发布**: 支持线性、指数和手动三种发布策略
- **实时健康监控**: 基于 Prometheus 的指标分析和健康评分
- **自动决策引擎**: 根据指标自动决定继续、暂停或回滚
- **多种流量管理**: 支持 Istio 和 Nginx Ingress
- **故障自动回滚**: 检测异常自动回滚到稳定版本

## 核心特性

### 灰度发布策略

- **线性递增 (Linear)**: 5% → 10% → 25% → 50% → 75% → 100%
- **指数递增 (Exponential)**: 1% → 5% → 10% → 25% → 50% → 100%
- **手动控制 (Manual)**: 用户手动控制每个阶段的流量切换

### 监控与决策

- **成功率监控**: 实时监控 HTTP 2xx 响应占比
- **延迟分析**: P50/P90/P99 延迟指标
- **错误率追踪**: HTTP 5xx 错误率监控
- **Pod 健康检查**: 自动检测 Pod 状态
- **智能决策**: 基于多维度指标自动决策

### 流量管理

- **Istio 支持**: 基于 VirtualService 和 DestinationRule
- **Nginx Ingress**: 使用 Canary Annotations
- **动态权重调整**: 平滑的流量切换

### 自动回滚

- **指标异常回滚**: 成功率、延迟、错误率超阈值自动回滚
- **Pod 崩溃回滚**: 检测到 Pod 异常自动回滚
- **手动回滚**: 支持用户手动触发回滚

## 快速开始

### 前置条件

- Kubernetes 集群 (v1.24+)
- kubectl 命令行工具
- Prometheus 监控系统
- Istio 或 Nginx Ingress Controller
- Go 1.21+ (用于开发)

### 安装

1. 克隆仓库：
```bash
git clone https://github.com/codefarmer009/codedance.git
cd codedance
```

2. 安装 CRD：
```bash
kubectl apply -f config/crd/canary_deployment.yaml
```

3. 创建 RBAC 权限：
```bash
kubectl apply -f config/rbac/
```

4. 部署控制器：
```bash
kubectl apply -f deploy/kubernetes/controller.yaml
```

5. 验证安装：
```bash
kubectl get pods -n codedance-system
```

### 使用示例

1. 部署应用：
```bash
kubectl apply -f config/samples/deployment_example.yaml
```

2. 创建灰度发布：
```bash
kubectl apply -f config/samples/example_canary.yaml
```

3. 查看发布状态：
```bash
kubectl get canarydeployment -n production
kubectl describe canarydeployment codedance-app -n production
```

## 项目结构

```
codedance/
├── cmd/                    # 主程序入口
│   ├── controller/         # 控制器主程序
│   ├── api-server/         # API 服务器
│   └── cli/                # 命令行工具
├── pkg/                    # 核心代码包
│   ├── apis/               # API 定义
│   │   └── deploy/v1alpha1/  # CRD 类型定义
│   ├── controller/         # 控制器逻辑
│   ├── metrics/            # 指标收集
│   ├── traffic/            # 流量管理
│   ├── strategy/           # 发布策略
│   └── utils/              # 工具函数
├── config/                 # 配置文件
│   ├── crd/                # CRD 定义
│   ├── rbac/               # RBAC 配置
│   └── samples/            # 示例配置
├── deploy/                 # 部署文件
│   ├── kubernetes/         # K8s 部署清单
│   └── helm/               # Helm Charts
├── docs/                   # 文档
│   ├── architecture.md     # 架构设计
│   ├── getting-started.md  # 快速开始
│   └── api-reference.md    # API 参考
├── Dockerfile              # Docker 镜像
├── Makefile                # 构建脚本
└── go.mod                  # Go 模块定义
```

## 开发指南

### 本地运行

```bash
make controller
```

或者：

```bash
go run cmd/controller/main.go --kubeconfig=$HOME/.kube/config
```

### 构建

```bash
make build
```

### 构建 Docker 镜像

```bash
make docker-build VERSION=v1.0.0
```

### 运行测试

```bash
make test
```

### 代码格式化

```bash
make fmt
```

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                    发布控制平面                                │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 发布控制器    │  │ 监控分析器    │  │ 决策引擎      │      │
│  │ (Controller) │→ │ (Analyzer)   │→ │ (Decider)    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│         ↓                 ↑                  ↓               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ 流量管理器    │  │ 指标收集器    │  │ 回滚管理器    │      │
│  │ (Traffic)    │  │ (Metrics)    │  │ (Rollback)   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│                  Kubernetes 集群                              │
│  ┌─────────────┐        ┌─────────────┐                     │
│  │  Stable     │  ←→   │  Canary     │                     │
│  │  Pods (v1)  │        │  Pods (v2)  │                     │
│  └─────────────┘        └─────────────┘                     │
└─────────────────────────────────────────────────────────────┘
```

详细架构设计请参考 [架构文档](docs/architecture.md)。

## 文档

- [快速开始指南](docs/getting-started.md)
- [架构设计文档](docs/architecture.md)
- [API 参考文档](docs/api-reference.md)

## 贡献指南

欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 开发路线图

- [x] 核心灰度发布功能
- [x] Prometheus 指标集成
- [x] Istio 流量管理
- [x] Nginx Ingress 支持
- [x] 自动回滚机制
- [ ] Web 管理界面
- [ ] CLI 命令行工具
- [ ] 多集群支持
- [ ] Helm Chart 支持
- [ ] AI 决策优化
- [ ] 成本分析功能

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

Project Maintainer: [@codefarmer009](https://github.com/codefarmer009)

Project Link: [https://github.com/codefarmer009/codedance](https://github.com/codefarmer009/codedance)

## Acknowledgments

- Thanks to all contributors who help make this project better
- Inspired by the global dance community
- Built with modern web technologies

---

Made with ❤️ by the codedance team
