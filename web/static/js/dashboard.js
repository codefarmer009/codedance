let currentCanary = null;
let canaries = [];
let metricsInterval = null;

function escapeHtml(unsafe) {
    if (typeof unsafe !== 'string') return unsafe;
    return unsafe
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

document.addEventListener('DOMContentLoaded', () => {
    loadCanaries();
    setInterval(loadCanaries, 10000);
});

async function loadCanaries() {
    try {
        const response = await fetch('/api/canaries');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        canaries = await response.json();
        
        updateSummaryCards();
        renderCanaryList();
        
        if (currentCanary) {
            const updated = canaries.find(c => 
                c.name === currentCanary.metadata?.name && 
                c.namespace === currentCanary.metadata?.namespace
            );
            if (updated) {
                loadCanaryDetail(updated.namespace, updated.name);
            }
        }
    } catch (error) {
        console.error('Failed to load canaries:', error);
        const listEl = document.getElementById('canary-list');
        if (listEl) {
            listEl.textContent = '加载失败，请检查连接';
            listEl.className = 'loading';
        }
    }
}

function updateSummaryCards() {
    const total = canaries.length;
    const completed = canaries.filter(c => c.phase === 'Completed').length;
    const progressing = canaries.filter(c => c.phase === 'Progressing').length;
    const paused = canaries.filter(c => c.phase === 'Paused').length;
    
    const totalEl = document.getElementById('total-canaries');
    const completedEl = document.getElementById('completed-canaries');
    const progressingEl = document.getElementById('progressing-canaries');
    const pausedEl = document.getElementById('paused-canaries');
    
    if (totalEl) totalEl.textContent = total;
    if (completedEl) completedEl.textContent = completed;
    if (progressingEl) progressingEl.textContent = progressing;
    if (pausedEl) pausedEl.textContent = paused;
}

function renderCanaryList() {
    const listEl = document.getElementById('canary-list');
    if (!listEl) return;
    
    const searchInput = document.getElementById('search-input');
    const statusFilter = document.getElementById('status-filter');
    
    const searchTerm = searchInput ? searchInput.value.toLowerCase() : '';
    const statusValue = statusFilter ? statusFilter.value : '';
    
    let filtered = canaries.filter(c => {
        const matchSearch = !searchTerm || 
            (c.name && c.name.toLowerCase().includes(searchTerm)) ||
            (c.namespace && c.namespace.toLowerCase().includes(searchTerm));
        const matchStatus = !statusValue || c.phase === statusValue;
        return matchSearch && matchStatus;
    });
    
    if (filtered.length === 0) {
        listEl.innerHTML = '<li class="loading">暂无发布任务</li>';
        return;
    }
    
    listEl.innerHTML = filtered.map(canary => {
        const name = escapeHtml(canary.name || '');
        const namespace = escapeHtml(canary.namespace || '');
        const phase = escapeHtml(canary.phase || 'Unknown');
        const strategy = escapeHtml(canary.strategy || 'N/A');
        const weight = parseInt(canary.currentWeight) || 0;
        const step = parseInt(canary.currentStep) || 0;
        const phaseClass = (canary.phase || 'unknown').toLowerCase().replace(/[^a-z0-9-]/g, '');
        const isActive = currentCanary && currentCanary.metadata && 
                        currentCanary.metadata.name === canary.name;
        
        return `
        <li class="${isActive ? 'active' : ''}" 
            onclick="loadCanaryDetail('${escapeHtml(namespace)}', '${escapeHtml(name)}')">
            <div class="canary-item-name">
                ${name}
                <span class="status-badge status-${phaseClass}">
                    ${phase}
                </span>
            </div>
            <div class="canary-item-info">
                命名空间: ${namespace}<br>
                策略: ${strategy}<br>
                当前权重: ${weight}%<br>
                步骤: ${step + 1}
            </div>
        </li>
    `;
    }).join('');
}

async function loadCanaryDetail(namespace, name) {
    try {
        const encodedNs = encodeURIComponent(namespace);
        const encodedName = encodeURIComponent(name);
        const response = await fetch(`/api/canaries/?namespace=${encodedNs}&name=${encodedName}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        currentCanary = await response.json();
        
        renderCanaryDetail();
        startMetricsPolling(namespace, name);
    } catch (error) {
        console.error('Failed to load canary detail:', error);
        const detailView = document.getElementById('detail-view');
        if (detailView) {
            detailView.innerHTML = '<div class="loading">加载详情失败</div>';
        }
    }
}

function renderCanaryDetail() {
    const detailView = document.getElementById('detail-view');
    if (!detailView || !currentCanary) return;
    
    const canary = currentCanary;
    const steps = canary.spec?.strategy?.steps || [];
    const currentStep = canary.status?.currentStep || 0;
    
    const name = escapeHtml(canary.metadata?.name || '');
    const namespace = escapeHtml(canary.metadata?.namespace || '');
    const phase = escapeHtml(canary.status?.phase || 'Unknown');
    const phaseClass = (canary.status?.phase || 'unknown').toLowerCase().replace(/[^a-z0-9-]/g, '');
    const createdAt = canary.metadata?.creationTimestamp ? 
        new Date(canary.metadata.creationTimestamp).toLocaleString('zh-CN') : 'N/A';
    const currentWeight = parseInt(canary.status?.currentWeight) || 0;
    
    detailView.innerHTML = `
        <div class="detail-header">
            <h2 class="detail-title">${name}</h2>
            <div class="detail-meta">
                <strong>命名空间:</strong> ${namespace} | 
                <strong>状态:</strong> <span class="status-badge status-${phaseClass}">${phase}</span> | 
                <strong>创建时间:</strong> ${escapeHtml(createdAt)}
            </div>
        </div>
        
        <div class="section">
            <h3>📊 发布进度</h3>
            <div class="progress-container">
                <div class="progress-bar-wrapper">
                    <div class="progress-bar" style="width: ${currentWeight}%">
                        ${currentWeight}%
                    </div>
                </div>
                <div class="progress-info">
                    <span>当前步骤: ${currentStep + 1} / ${steps.length}</span>
                    <span>流量权重: ${currentWeight}%</span>
                </div>
            </div>
        </div>
        
        <div class="section">
            <h3>🎯 发布步骤</h3>
            <div class="steps-timeline">
                ${steps.map((step, index) => {
                    const weight = parseInt(step.weight) || 0;
                    const pause = escapeHtml(step.pause || '无');
                    const metrics = step.metrics && Array.isArray(step.metrics) ?
                        step.metrics.map(m => escapeHtml(m.name || '')).join(', ') : '';
                    
                    return `
                    <div class="step-item ${index < currentStep ? 'completed' : ''} ${index === currentStep ? 'active' : ''}">
                        <div class="step-content">
                            <div class="step-header">步骤 ${index + 1}: ${weight}% 流量</div>
                            <div>暂停时间: ${pause}</div>
                            ${metrics ? `<div>指标检查: ${metrics}</div>` : ''}
                        </div>
                    </div>
                `;
                }).join('')}
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
                    <div class="info-value">${escapeHtml(canary.spec?.targetDeployment || 'N/A')}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">金丝雀版本</div>
                    <div class="info-value">${escapeHtml(canary.spec?.canaryVersion || 'N/A')}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">发布策略</div>
                    <div class="info-value">${escapeHtml(canary.spec?.strategy?.type || 'N/A')}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">自动回滚</div>
                    <div class="info-value">${canary.spec?.autoRollback?.enabled ? '已启用' : '已禁用'}</div>
                </div>
            </div>
        </div>
        
        <div class="section">
            <h3>📋 指标配置</h3>
            <div class="info-grid">
                <div class="info-item">
                    <div class="info-label">成功率阈值</div>
                    <div class="info-value">${((canary.spec?.metrics?.successRate?.threshold || 0) * 100).toFixed(2)}%</div>
                </div>
                <div class="info-item">
                    <div class="info-label">错误率阈值</div>
                    <div class="info-value">${((canary.spec?.metrics?.errorRate?.threshold || 0) * 100).toFixed(2)}%</div>
                </div>
                <div class="info-item">
                    <div class="info-label">P99 延迟限制</div>
                    <div class="info-value">${escapeHtml(canary.spec?.metrics?.latency?.p99 || 'N/A')}</div>
                </div>
                <div class="info-item">
                    <div class="info-label">指标失败回滚</div>
                    <div class="info-value">${canary.spec?.autoRollback?.onMetricsFail ? '是' : '否'}</div>
                </div>
            </div>
        </div>
    `;
    
    renderCanaryList();
}

function startMetricsPolling(namespace, name) {
    if (metricsInterval) {
        clearInterval(metricsInterval);
        metricsInterval = null;
    }
    
    loadMetrics(namespace, name);
    metricsInterval = setInterval(() => loadMetrics(namespace, name), 5000);
}

async function loadMetrics(namespace, name) {
    try {
        const encodedNs = encodeURIComponent(namespace);
        const encodedName = encodeURIComponent(name);
        const response = await fetch(`/api/metrics/?namespace=${encodedNs}&name=${encodedName}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
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
    
    const successRate = parseFloat(metrics.successRate) || 0;
    const errorRate = parseFloat(metrics.errorRate) || 0;
    const latencyP50 = parseFloat(metrics.latencyP50) || 0;
    const latencyP90 = parseFloat(metrics.latencyP90) || 0;
    const latencyP99 = parseFloat(metrics.latencyP99) || 0;
    const requestRate = parseFloat(metrics.requestRate) || 0;
    
    container.innerHTML = `
        <div class="metric-card ${getMetricClass(successRate, 95)}">
            <div class="metric-value">${successRate.toFixed(2)}%</div>
            <div class="metric-label">成功率</div>
        </div>
        <div class="metric-card ${getMetricClass(errorRate, 5, true)}">
            <div class="metric-value">${errorRate.toFixed(2)}%</div>
            <div class="metric-label">错误率</div>
        </div>
        <div class="metric-card">
            <div class="metric-value">${latencyP50.toFixed(1)}ms</div>
            <div class="metric-label">P50 延迟</div>
        </div>
        <div class="metric-card">
            <div class="metric-value">${latencyP90.toFixed(1)}ms</div>
            <div class="metric-label">P90 延迟</div>
        </div>
        <div class="metric-card ${getMetricClass(latencyP99, 200, true)}">
            <div class="metric-value">${latencyP99.toFixed(1)}ms</div>
            <div class="metric-label">P99 延迟</div>
        </div>
        <div class="metric-card">
            <div class="metric-value">${requestRate.toFixed(0)}</div>
            <div class="metric-label">请求/秒</div>
        </div>
    `;
}

const searchInput = document.getElementById('search-input');
const statusFilter = document.getElementById('status-filter');

if (searchInput) {
    searchInput.addEventListener('input', renderCanaryList);
}

if (statusFilter) {
    statusFilter.addEventListener('change', renderCanaryList);
}
