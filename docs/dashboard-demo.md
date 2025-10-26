# Codedance 可视化发布管理界面演示文档

## 📋 目录

1. [概述](#概述)
2. [功能特性](#功能特性)
3. [架构设计](#架构设计)
4. [安装部署](#安装部署)
5. [界面演示](#界面演示)
6. [使用指南](#使用指南)
7. [API 接口](#api-接口)
8. [常见问题](#常见问题)

---

## 概述

Codedance 可视化发布管理界面是一个基于 Web 的实时监控和管理平台，用于可视化展示 Kubernetes 灰度发布的全过程。该界面提供了直观的发布进度追踪、实时指标监控、以及详细的配置信息展示。

### 核心价值

- **实时监控**: 实时展示灰度发布的进度和关键指标
- **直观可视化**: 通过图表和进度条直观展示发布状态
- **全面信息**: 集中展示发布配置、步骤、指标等所有相关信息
- **易于使用**: 简洁的界面设计，无需复杂配置即可使用

---

## 功能特性

### 1. 发布总览

界面顶部提供四个核心指标卡片：

- **总发布任务**: 显示当前集群中所有的灰度发布任务数量
- **已完成**: 成功完成的发布任务数量
- **进行中**: 正在执行的发布任务数量
- **已暂停**: 因指标异常或手动操作而暂停的任务数量

### 2. 发布列表

左侧边栏展示所有灰度发布任务，包含：

- **搜索功能**: 按名称或命名空间快速筛选
- **状态过滤**: 按发布状态（进行中/已暂停/已完成/失败）过滤
- **实时更新**: 每 10 秒自动刷新列表
- **状态标识**: 使用颜色编码清晰标识不同状态

### 3. 详细视图

点击任意发布任务，右侧主区域展示详细信息：

#### 3.1 发布进度

- **进度条**: 直观显示当前流量权重百分比
- **步骤信息**: 当前步骤 / 总步骤数
- **流量权重**: 实时流量分配比例

#### 3.2 发布步骤

- **时间轴视图**: 以时间轴形式展示所有发布步骤
- **步骤状态**: 
  - 已完成步骤：绿色标记
  - 当前步骤：蓝色标记
  - 待执行步骤：灰色标记
- **步骤详情**: 每个步骤的流量权重、暂停时间、指标检查配置

#### 3.3 实时指标

以卡片形式展示关键性能指标，每 5 秒自动更新：

- **成功率**: HTTP 2xx 响应占比
- **错误率**: HTTP 5xx 错误占比
- **延迟指标**: P50、P90、P99 延迟
- **请求速率**: 每秒请求数

指标卡片使用颜色编码：
- 🟢 绿色：指标正常
- 🟡 黄色：指标警告
- 🔴 红色：指标异常

#### 3.4 配置信息

- **目标部署**: 灰度发布的目标 Deployment
- **金丝雀版本**: 新版本标识
- **发布策略**: Linear/Exponential/Manual
- **自动回滚**: 回滚策略配置

#### 3.5 指标配置

- **成功率阈值**: 自动回滚的成功率下限
- **错误率阈值**: 自动回滚的错误率上限
- **延迟限制**: P99 延迟阈值
- **回滚策略**: 指标失败时的回滚配置

---

## 架构设计

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      用户浏览器                               │
│                   (Web Dashboard)                            │
└─────────────────────────┬───────────────────────────────────┘
                          │ HTTP/WebSocket
┌─────────────────────────▼───────────────────────────────────┐
│                  Dashboard API Server                        │
│                    (Go HTTP Server)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ API Handler  │  │ Static Files │  │ WebSocket    │      │
│  │              │  │ Server       │  │ Handler      │      │
│  └──────┬───────┘  └──────────────┘  └──────────────┘      │
└─────────┼──────────────────────────────────────────────────┘
          │
          │ Kubernetes API
┌─────────▼──────────────────────────────────────────────────┐
│              Kubernetes API Server                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ CRD Resources│  │ Deployments  │  │ Pods         │      │
│  │ (Canary)     │  │              │  │              │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### 技术栈

#### 后端
- **语言**: Go 1.21+
- **框架**: 标准库 net/http
- **K8s 客户端**: client-go, dynamic client
- **API 风格**: RESTful

#### 前端
- **HTML5**: 语义化结构
- **CSS3**: 响应式设计，Flexbox/Grid 布局
- **JavaScript**: 原生 ES6+，无框架依赖
- **更新机制**: 轮询（Polling）

### API 设计

#### 1. 获取发布列表
```
GET /api/canaries
```

响应示例：
```json
[
  {
    "name": "codedance-app",
    "namespace": "production",
    "phase": "Progressing",
    "currentStep": 2,
    "currentWeight": 25,
    "strategy": "linear",
    "createdAt": "2025-01-20T10:30:00Z"
  }
]
```

#### 2. 获取发布详情
```
GET /api/canaries/?namespace=production&name=codedance-app
```

响应示例：
```json
{
  "metadata": {
    "name": "codedance-app",
    "namespace": "production",
    "creationTimestamp": "2025-01-20T10:30:00Z"
  },
  "spec": {
    "targetDeployment": "codedance-app",
    "canaryVersion": "v2.0.0",
    "strategy": {
      "type": "linear",
      "steps": [...]
    },
    "metrics": {...},
    "autoRollback": {...}
  },
  "status": {
    "phase": "Progressing",
    "currentStep": 2,
    "currentWeight": 25
  }
}
```

#### 3. 获取实时指标
```
GET /api/metrics/?namespace=production&name=codedance-app
```

响应示例：
```json
{
  "timestamp": 1737369600,
  "successRate": 99.5,
  "errorRate": 0.5,
  "latencyP50": 45.2,
  "latencyP90": 89.5,
  "latencyP99": 156.8,
  "requestRate": 1250.5
}
```

---

## 安装部署

### 前置条件

1. Kubernetes 集群（v1.24+）
2. 已安装 Codedance Controller 和 CRD
3. kubectl 访问权限
4. （可选）Ingress Controller

### 本地开发运行

#### 1. 克隆代码
```bash
git clone https://github.com/codefarmer009/codedance.git
cd codedance
```

#### 2. 构建并运行
```bash
# 使用 Makefile
make dashboard

# 或直接运行
go build -o bin/codedance-dashboard cmd/dashboard/main.go
./bin/codedance-dashboard --kubeconfig=$HOME/.kube/config --port=8080
```

#### 3. 访问界面
打开浏览器访问：http://localhost:8080

### Kubernetes 部署

#### 1. 构建 Docker 镜像
```bash
docker build -f Dockerfile.dashboard -t codedance-dashboard:v1.0.0 .
```

#### 2. 推送镜像到仓库
```bash
docker tag codedance-dashboard:v1.0.0 your-registry/codedance-dashboard:v1.0.0
docker push your-registry/codedance-dashboard:v1.0.0
```

#### 3. 更新部署文件
编辑 `deploy/kubernetes/dashboard.yaml`，更新镜像地址：
```yaml
spec:
  containers:
  - name: dashboard
    image: your-registry/codedance-dashboard:v1.0.0
```

#### 4. 部署到集群
```bash
kubectl apply -f deploy/kubernetes/dashboard.yaml
```

#### 5. 验证部署
```bash
kubectl get pods -n codedance-system
kubectl get svc -n codedance-system
```

#### 6. 访问方式

**方式一：Port Forward**
```bash
kubectl port-forward -n codedance-system svc/codedance-dashboard 8080:80
```
访问：http://localhost:8080

**方式二：Ingress**

如果已配置 Ingress，更新 hosts 文件：
```bash
echo "127.0.0.1 codedance.local" | sudo tee -a /etc/hosts
```
访问：http://codedance.local

**方式三：LoadBalancer**

修改 Service 类型为 LoadBalancer：
```bash
kubectl patch svc codedance-dashboard -n codedance-system -p '{"spec":{"type":"LoadBalancer"}}'
```

---

## 界面演示

### 1. 主界面布局

```
┌──────────────────────────────────────────────────────────────┐
│          🚀 Codedance 发布管理平台                             │
│         Kubernetes 智能灰度发布系统                            │
├──────────────────────────────────────────────────────────────┤
│  ┌──────┐  ┌──────┐  ┌──────┐  ┌──────┐                     │
│  │ 📊 8 │  │ ✅ 3 │  │ ⚡ 4 │  │ ⏸️ 1 │                    │
│  │总任务│  │已完成│  │进行中│  │已暂停│                     │
│  └──────┘  └──────┘  └──────┘  └──────┘                     │
├─────────────────┬────────────────────────────────────────────┤
│ 发布列表        │ 发布详情                                    │
│ ┌─────────────┐ │ ┌────────────────────────────────────────┐ │
│ │ 🔍 搜索     │ │ │ codedance-app                          │ │
│ ├─────────────┤ │ │ production | Progressing | 2025-01-20 │ │
│ │ 📊 状态     │ │ └────────────────────────────────────────┘ │
│ ├─────────────┤ │                                            │
│ │ app-1   25% │ │ 📊 发布进度                                │
│ │ app-2   50% │ │ ▓▓▓▓▓░░░░░ 25%  步骤 3/6                  │
│ │ app-3  100% │ │                                            │
│ │ ...         │ │ 🎯 发布步骤                                │
│ │             │ │ ✅ 步骤 1: 5%                              │
│ │             │ │ ✅ 步骤 2: 10%                             │
│ │             │ │ ⚡ 步骤 3: 25%  ← 当前                    │
│ │             │ │ ○ 步骤 4: 50%                              │
│ └─────────────┘ │ ○ 步骤 5: 75%                              │
│                 │ ○ 步骤 6: 100%                             │
│                 │                                            │
│                 │ 📈 实时指标                                │
│                 │ ┌────┐ ┌────┐ ┌────┐                      │
│                 │ │99.5%│ │0.5%│ │45ms│                     │
│                 │ │成功 │ │错误│ │P50 │                     │
│                 │ └────┘ └────┘ └────┘                      │
└─────────────────┴────────────────────────────────────────────┘
```

### 2. 界面截图说明

#### 2.1 总览仪表板
![Dashboard Overview](./images/dashboard-overview.png)
- 顶部显示四个关键指标卡片
- 左侧列表展示所有发布任务
- 右侧显示欢迎消息，引导用户选择任务

#### 2.2 发布详情视图
![Deployment Detail](./images/deployment-detail.png)
- 发布进度条显示当前流量权重
- 时间轴展示所有发布步骤及其状态
- 实时指标卡片显示关键性能数据

#### 2.3 指标监控
![Metrics Monitoring](./images/metrics-monitoring.png)
- 6 个指标卡片实时更新
- 颜色编码表示指标健康状态
- 每 5 秒自动刷新

### 3. 交互流程

#### 流程 1: 查看发布进度
```
1. 用户访问仪表板 → 看到总览卡片
2. 浏览左侧发布列表 → 看到所有任务状态
3. 点击某个发布任务 → 右侧显示详细信息
4. 查看进度条和步骤 → 了解当前进度
5. 查看实时指标 → 监控系统健康状况
```

#### 流程 2: 筛选发布任务
```
1. 在搜索框输入关键词 → 列表实时过滤
2. 选择状态下拉框 → 按状态筛选
3. 清空筛选条件 → 恢复显示所有任务
```

#### 流程 3: 监控指标变化
```
1. 选择一个进行中的发布 → 查看详情
2. 观察实时指标卡片 → 每 5 秒更新
3. 指标异常时 → 卡片变为红色/黄色
4. 系统自动回滚 → 状态变为 "Failed" 或 "Paused"
```

---

## 使用指南

### 快速开始

#### 第一次使用

1. **启动仪表板**
   ```bash
   make dashboard
   ```

2. **创建示例发布任务**
   ```bash
   kubectl apply -f config/samples/example_canary.yaml
   ```

3. **访问界面**
   - 打开浏览器访问 http://localhost:8080
   - 界面会自动加载发布任务列表

4. **查看发布详情**
   - 点击左侧列表中的任意任务
   - 右侧将显示该任务的详细信息

### 常见操作

#### 1. 搜索发布任务

在搜索框中输入：
- 任务名称（如：`codedance-app`）
- 命名空间（如：`production`）

搜索是实时的，无需按回车键。

#### 2. 按状态过滤

使用状态下拉框选择：
- **全部状态**: 显示所有任务
- **进行中**: 仅显示正在执行的任务
- **已暂停**: 仅显示暂停的任务
- **已完成**: 仅显示完成的任务
- **失败**: 仅显示失败的任务

#### 3. 理解指标颜色

实时指标使用颜色编码：

| 颜色 | 含义 | 示例 |
|------|------|------|
| 🟢 绿色 | 正常 | 成功率 > 95% |
| 🟡 黄色 | 警告 | 成功率 90-95% |
| 🔴 红色 | 异常 | 成功率 < 90% |

#### 4. 查看发布历史

1. 选择一个已完成的任务
2. 查看步骤时间轴
3. 所有步骤都显示为绿色（已完成）
4. 查看最终指标和配置

### 最佳实践

#### 1. 监控策略

- **持续监控**: 在发布进行中时，保持仪表板打开
- **关注指标**: 重点关注成功率、错误率和 P99 延迟
- **及时响应**: 发现异常时，检查日志和系统状态

#### 2. 问题排查

当发布暂停或失败时：

1. **查看状态原因**
   - 在详情页面查看 `status.reason` 字段
   
2. **检查指标**
   - 查看哪个指标超过阈值
   - 对比当前值和配置的阈值

3. **查看日志**
   ```bash
   kubectl logs -n codedance-system deployment/codedance-controller
   ```

4. **手动回滚**（如果需要）
   ```bash
   kubectl delete canarydeployment <name> -n <namespace>
   ```

#### 3. 性能优化

- 仪表板每 10 秒自动刷新列表
- 实时指标每 5 秒更新一次
- 如果任务过多，使用搜索和筛选功能
- 关闭不需要监控的任务详情页

---

## API 接口

### 完整 API 文档

#### 1. 获取发布列表

**请求**
```http
GET /api/canaries HTTP/1.1
Host: localhost:8080
```

**响应**
```json
[
  {
    "name": "codedance-app",
    "namespace": "production",
    "phase": "Progressing",
    "currentStep": 2,
    "currentWeight": 25,
    "strategy": "linear",
    "createdAt": "2025-01-20T10:30:00Z"
  }
]
```

**状态码**
- `200 OK`: 成功
- `500 Internal Server Error`: 服务器错误

#### 2. 获取发布详情

**请求**
```http
GET /api/canaries/?namespace=production&name=codedance-app HTTP/1.1
Host: localhost:8080
```

**查询参数**
- `namespace` (必需): K8s 命名空间
- `name` (必需): CanaryDeployment 名称

**响应**
```json
{
  "apiVersion": "deploy.codedance.io/v1alpha1",
  "kind": "CanaryDeployment",
  "metadata": {
    "name": "codedance-app",
    "namespace": "production",
    "creationTimestamp": "2025-01-20T10:30:00Z"
  },
  "spec": {
    "targetDeployment": "codedance-app",
    "canaryVersion": "v2.0.0",
    "strategy": {
      "type": "linear",
      "steps": [
        {"weight": 5, "pause": "2m"},
        {"weight": 10, "pause": "3m"},
        {"weight": 25, "pause": "5m"}
      ]
    },
    "metrics": {
      "successRate": {"threshold": 0.95},
      "errorRate": {"threshold": 0.05},
      "latency": {"p99": "200ms"}
    },
    "autoRollback": {
      "enabled": true,
      "onMetricsFail": true
    }
  },
  "status": {
    "phase": "Progressing",
    "currentStep": 2,
    "currentWeight": 25
  }
}
```

**状态码**
- `200 OK`: 成功
- `400 Bad Request`: 缺少必需参数
- `404 Not Found`: 资源不存在
- `500 Internal Server Error`: 服务器错误

#### 3. 获取实时指标

**请求**
```http
GET /api/metrics/?namespace=production&name=codedance-app HTTP/1.1
Host: localhost:8080
```

**查询参数**
- `namespace` (必需): K8s 命名空间
- `name` (必需): CanaryDeployment 名称

**响应**
```json
{
  "timestamp": 1737369600,
  "successRate": 99.5,
  "errorRate": 0.5,
  "latencyP50": 45.2,
  "latencyP90": 89.5,
  "latencyP99": 156.8,
  "requestRate": 1250.5
}
```

**字段说明**
- `timestamp`: Unix 时间戳
- `successRate`: 成功率百分比 (0-100)
- `errorRate`: 错误率百分比 (0-100)
- `latencyP50`: P50 延迟（毫秒）
- `latencyP90`: P90 延迟（毫秒）
- `latencyP99`: P99 延迟（毫秒）
- `requestRate`: 每秒请求数

**状态码**
- `200 OK`: 成功
- `400 Bad Request`: 缺少必需参数
- `500 Internal Server Error`: 服务器错误

#### 4. 静态资源

**请求**
```http
GET /static/css/dashboard.css HTTP/1.1
GET /static/js/dashboard.js HTTP/1.1
Host: localhost:8080
```

**状态码**
- `200 OK`: 成功
- `404 Not Found`: 文件不存在

### 使用示例

#### curl 命令

```bash
# 获取发布列表
curl http://localhost:8080/api/canaries

# 获取发布详情
curl "http://localhost:8080/api/canaries/?namespace=production&name=codedance-app"

# 获取实时指标
curl "http://localhost:8080/api/metrics/?namespace=production&name=codedance-app"
```

#### JavaScript Fetch

```javascript
// 获取发布列表
fetch('/api/canaries')
  .then(res => res.json())
  .then(data => console.log(data));

// 获取发布详情
fetch('/api/canaries/?namespace=production&name=codedance-app')
  .then(res => res.json())
  .then(data => console.log(data));

// 获取实时指标
fetch('/api/metrics/?namespace=production&name=codedance-app')
  .then(res => res.json())
  .then(data => console.log(data));
```

---

## 常见问题

### Q1: 仪表板显示"加载失败"？

**可能原因**：
1. Dashboard 服务未启动
2. Kubernetes API 连接失败
3. RBAC 权限不足

**解决方案**：
```bash
# 检查服务状态
kubectl get pods -n codedance-system

# 检查日志
kubectl logs -n codedance-system deployment/codedance-dashboard

# 检查 RBAC
kubectl auth can-i get canarydeployments --as=system:serviceaccount:codedance-system:codedance-dashboard
```

### Q2: 指标显示不准确？

**可能原因**：
1. Prometheus 未正确配置
2. 查询语句错误
3. 标签选择器不匹配

**解决方案**：
1. 检查 Prometheus 是否可访问
2. 验证 PromQL 查询语句
3. 确认 Service 和 Pod 标签正确

### Q3: 界面不更新？

**可能原因**：
1. 浏览器缓存
2. JavaScript 错误
3. 网络连接问题

**解决方案**：
1. 硬刷新浏览器（Ctrl+Shift+R）
2. 检查浏览器控制台错误
3. 检查网络连接

### Q4: 如何自定义刷新频率？

修改 `web/static/js/dashboard.js`:

```javascript
// 修改列表刷新频率（默认 10 秒）
setInterval(loadCanaries, 10000);  // 改为需要的毫秒数

// 修改指标刷新频率（默认 5 秒）
metricsInterval = setInterval(() => loadMetrics(namespace, name), 5000);
```

### Q5: 能否集成到现有监控系统？

可以，仪表板提供 RESTful API，可以：

1. **Grafana 集成**
   - 使用 JSON API 数据源
   - 创建自定义面板

2. **Prometheus 集成**
   - 暴露 `/metrics` 端点
   - 添加 ServiceMonitor

3. **自定义集成**
   - 直接调用 API 接口
   - 使用 Webhook 通知

### Q6: 如何启用 HTTPS？

**方式一：Ingress TLS**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codedance-dashboard
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - codedance.example.com
    secretName: codedance-tls
  rules:
  - host: codedance.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: codedance-dashboard
            port:
              number: 80
```

**方式二：Nginx 反向代理**
```nginx
server {
    listen 443 ssl;
    server_name codedance.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Q7: 如何添加认证？

在 Ingress 中添加 Basic Auth：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: basic-auth
  namespace: codedance-system
type: Opaque
data:
  auth: <base64-encoded-htpasswd>
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: codedance-dashboard
  annotations:
    nginx.ingress.kubernetes.io/auth-type: basic
    nginx.ingress.kubernetes.io/auth-secret: basic-auth
    nginx.ingress.kubernetes.io/auth-realm: 'Authentication Required'
spec:
  # ... 其他配置
```

创建密码文件：
```bash
htpasswd -c auth admin
kubectl create secret generic basic-auth --from-file=auth -n codedance-system
```

---

## 附录

### A. 界面组件详解

#### 1. 总览卡片
- 实时统计数据
- 鼠标悬停效果
- 点击可筛选（未来功能）

#### 2. 发布列表
- 虚拟滚动（大量数据）
- 状态图标
- 快速搜索

#### 3. 详情视图
- 响应式布局
- 实时数据更新
- 历史数据保留

#### 4. 指标卡片
- 动态颜色
- 趋势图表（未来功能）
- 阈值警告

### B. 浏览器兼容性

| 浏览器 | 最低版本 | 推荐版本 |
|--------|----------|----------|
| Chrome | 90+ | 最新版 |
| Firefox | 88+ | 最新版 |
| Safari | 14+ | 最新版 |
| Edge | 90+ | 最新版 |

### C. 性能指标

- **首次加载**: < 2s
- **数据刷新**: < 500ms
- **交互响应**: < 100ms
- **内存占用**: < 50MB

### D. 未来路线图

- [ ] WebSocket 实时推送
- [ ] 历史数据查询
- [ ] 指标趋势图表
- [ ] 多集群支持
- [ ] 用户权限管理
- [ ] 暗色主题
- [ ] 移动端适配
- [ ] 导出报告功能

---

## 联系方式

- **项目主页**: https://github.com/codefarmer009/codedance
- **问题反馈**: https://github.com/codefarmer009/codedance/issues
- **文档**: https://github.com/codefarmer009/codedance/docs

---

*最后更新: 2025-01-20*
