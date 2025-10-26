let currentCanary = null;
let canaries = [];
let metricsInterval = null;

document.addEventListener('DOMContentLoaded', () => {
    loadCanaries();
    setInterval(loadCanaries, 10000);
});

async function loadCanaries() {
    try {
        const response = await fetch('/api/canaries');
        canaries = await response.json();
        
        updateSummaryCards();
        renderCanaryList();
        
        if (currentCanary) {
            const updated = canaries.find(c => 
                c.name === currentCanary.name && 
                c.namespace === currentCanary.namespace
            );
            if (updated) {
                loadCanaryDetail(updated.namespace, updated.name);
            }
        }
    } catch (error) {
        console.error('Failed to load canaries:', error);
        document.getElementById('canary-list').innerHTML = 
            '<li class="loading">加载失败，请检查连接</li>';
    }
}

function updateSummaryCards() {
    const total = canaries.length;
    const completed = canaries.filter(c => c.phase === 'Completed').length;
    const progressing = canaries.filter(c => c.phase === 'Progressing').length;
    const paused = canaries.filter(c => c.phase === 'Paused').length;
    
    document.getElementById('total-canaries').textContent = total;
    document.getElementById('completed-canaries').textContent = completed;
    document.getElementById('progressing-canaries').textContent = progressing;
    document.getElementById('paused-canaries').textContent = paused;
}

function renderCanaryList() {
    const listEl = document.getElementById('canary-list');
    const searchTerm = document.getElementById('search-input').value.toLowerCase();
    const statusFilter = document.getElementById('status-filter').value;
    
    let filtered = canaries.filter(c => {
        const matchSearch = !searchTerm || 
            c.name.toLowerCase().includes(searchTerm) ||
            c.namespace.toLowerCase().includes(searchTerm);
        const matchStatus = !statusFilter || c.phase === statusFilter;
        return matchSearch && matchStatus;
    });
    
    if (filtered.length === 0) {
        listEl.innerHTML = '<li class="loading">暂无发布任务</li>';
        return;
    }
    
    listEl.innerHTML = filtered.map(canary => `
        <li class="${currentCanary && currentCanary.name === canary.name ? 'active' : ''}" 
            onclick="loadCanaryDetail('${canary.namespace}', '${canary.name}')">
            <div class="canary-item-name">
                ${canary.name}
                <span class="status-badge status-${(canary.phase || 'unknown').toLowerCase()}">
                    ${canary.phase || 'Unknown'}
                </span>
            </div>
            <div class="canary-item-info">
                命名空间: ${canary.namespace}<br>
                策略: ${canary.strategy || 'N/A'}<br>
                当前权重: ${canary.currentWeight || 0}%<br>
                步骤: ${(canary.currentStep || 0) + 1}
            </div>
        </li>
    `).join('');
}

async function loadCanaryDetail(namespace, name) {
    try {
        const response = await fetch(`/api/canaries/?namespace=${namespace}&name=${name}`);
        currentCanary = await response.json();
        
        renderCanaryDetail();
        startMetricsPolling(namespace, name);
    } catch (error) {
        console.error('Failed to load canary detail:', error);
    }
}

function renderCanaryDetail() {
    const detailView = document.getElementById('detail-view');
    const canary = currentCanary;
    
    const steps = canary.spec?.strategy?.steps || [];
    const currentStep = canary.status?.currentStep || 0;
    
    detailView.innerHTML = `
        <div class="detail-header">
            <h2 class="detail-title">${canary.metadata.name}</h2>
            <div class="detail-meta">
                <strong>命名空间:</strong> ${canary.metadata.namespace} | 
                <strong>状态:</strong> <span class="status-badge status-${(canary.status.phase || 'unknown').toLowerCase()}">${canary.status.phase || 'Unknown'}</span> | 
                <strong>创建时间:</strong> ${new Date(canary.metadata.creationTimestamp).toLocaleString('zh-CN')}
            </div>
        </div>
        
        <div class="section">
            <h3>📊 发布进度</h3>
            <div class="progress-container">
                <div class="progress-bar-wrapper">
                    <div class="progress-bar" style="width: ${canary.status.currentWeight || 0}%">
                        ${canary.status.currentWeight || 0}%
                    </div>
                </div>
                <div class="progress-info">
                    <span>当前步骤: ${currentStep + 1} / ${steps.length}</span>
                    <span>流量权重: ${canary.status.currentWeight || 0}%</span>
                </div>
            </div>
        </div>
        
        <div class="section">
            <h3>🎯 发布步骤</h3>
            <div class="steps-timeline">
                ${steps.map((step, index) => `
                    <div class="step-item ${index < currentStep ? 'completed' : ''} ${index === currentStep ? 'active' : ''}">
                        <div class="step-content">
                            <div class="step-header">步骤 ${index + 1}: ${step.weight}% 流量</div>
                            <div>暂停时间: ${step.pause || '无'}</div>
                            ${step.metrics && step.metrics.length > 0 ? `
                                <div>指标检查: ${step.metrics.map(m => m.name).join(', ')}</div>
                            ` : ''}
                        </div>
                    </div>
                `).join('')}
            </div>
        </div>
        
        <div class="section">
            <h3>📈 实时指标</h3>
            <div id="metrics-container" class="metrics-grid">
                <div class="loading">加载指标中...</div>
            </div>
        </div>
        
        <div class="section">
            <h3>⚙️ 配置信息</h3>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-label">目标部署</div>
                    <div class="info-value">${canary.spec.targetDeployment}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">金丝雀版本</div>
                    <div class="info-value">${canary.spec.canaryVersion}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">发布策略</div>
                    <div class="info-value">${canary.spec.strategy.type}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">自动回滚</div>
                    <div class="info-value">${canary.spec.autoRollback.enabled ? '已启用' : '已禁用'}</div>
                </div>
            </div>
        </div>
        
        <div class="section">
            <h3>📋 指标配置</h3>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-label">成功率阈值</div>
                    <div class="info-value">${canary.spec.metrics.successRate.threshold * 100}%</div>
                </div>
                <div class="info-item">
                    <div class="info-label">错误率阈值</div>
                    <div class="info-value">${canary.spec.metrics.errorRate.threshold * 100}%</div>
                </div>
                <div class="info-item">
                    <div class="info-label">P99 延迟限制</div>
                    <div class="info-value">${canary.spec.metrics.latency.p99}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">指标失败回滚</div>
                    <div class="info-value">${canary.spec.autoRollback.onMetricsFail ? '是' : '否'}</div>
                </div>
            </div>
        </div>
    `;
    
    renderCanaryList();
}

function startMetricsPolling(namespace, name) {
    if (metricsInterval) {
        clearInterval(metricsInterval);
    }
    
    loadMetrics(namespace, name);
    metricsInterval = setInterval(() => loadMetrics(namespace, name), 5000);
}

async function loadMetrics(namespace, name) {
    try {
        const response = await fetch(`/api/metrics/?namespace=${namespace}&name=${name}`);
        const metrics = await response.json();
        
        renderMetrics(metrics);
    } catch (error) {
        console.error('Failed to load metrics:', error);
    }
}

function renderMetrics(metrics) {
    const container = document.getElementById('metrics-container');
    if (!container) return;
    
    const getMetricClass = (value, threshold, inverse = false) => {
        if (inverse) {
            return value > threshold ? 'metric-danger' : 'metric-good';
        }
        return value >= threshold ? 'metric-good' : 'metric-warning';
    };
    
    container.innerHTML = `
        <div class="metric-card ${getMetricClass(metrics.successRate, 95)}">
            <div class="metric-value">${metrics.successRate.toFixed(2)}%</div>
            <div class="metric-label">成功率</div>
        </div>
        <div class="metric-card ${getMetricClass(metrics.errorRate, 5, true)}">
            <div class="metric-value">${metrics.errorRate.toFixed(2)}%</div>
            <div class="metric-label">错误率</div>
        </div>
        <div class="metric-card">
            <div class="metric-value">${metrics.latencyP50.toFixed(1)}ms</div>
            <div class="metric-label">P50 延迟</div>
        </div>
        <div class="metric-card">
            <div class="metric-value">${metrics.latencyP90.toFixed(1)}ms</div>
            <div class="metric-label">P90 延迟</div>
        </div>
        <div class="metric-card ${getMetricClass(metrics.latencyP99, 200, true)}">
            <div class="metric-value">${metrics.latencyP99.toFixed(1)}ms</div>
            <div class="metric-label">P99 延迟</div>
        </div>
        <div class="metric-card">
            <div class="metric-value">${metrics.requestRate.toFixed(0)}</div>
            <div class="metric-label">请求/秒</div>
        </div>
    `;
}

document.getElementById('search-input')?.addEventListener('input', renderCanaryList);
document.getElementById('status-filter')?.addEventListener('change', renderCanaryList);
