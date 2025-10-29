/**
 * GPU Pro - Alert Management System
 * Handles real-time alerts, notifications, and threshold configuration
 */

// Alert State Management
const AlertManager = {
    alerts: [],
    activeAlerts: new Set(),
    snoozedAlerts: new Map(), // alertId -> snoozeUntil timestamp
    acknowledgedAlerts: new Set(),
    thresholds: {
        temp_warning: 75,
        temp_critical: 85,
        memory_warning: 85,
        memory_critical: 95,
        power_warning: 90,
        power_critical: 98
    },
    settings: {
        enableBrowserNotifications: true,
        enableSoundAlerts: false,
        enableBanner: true,
        autoAcknowledgeResolved: true,
        cooldownPeriod: 300 // 5 minutes between duplicate alerts
    },
    lastAlertTime: new Map(), // alertKey -> timestamp
    notificationPermission: 'default'
};

// Alert Severity Levels
const AlertSeverity = {
    INFO: { level: 'info', icon: '🔵', color: '--info', label: 'INFO' },
    WARNING: { level: 'warning', icon: '🟡', color: '--warning', label: 'WARNING' },
    CRITICAL: { level: 'critical', icon: '🔴', color: '--danger', label: 'CRITICAL' }
};

// Alert Types
const AlertTypes = {
    TEMPERATURE: 'Temperature',
    MEMORY: 'Memory',
    POWER: 'Power',
    UTILIZATION: 'Utilization',
    FAN: 'Fan Speed'
};

/**
 * Initialize Alert System
 */
function initAlertSystem() {
    console.log('Initializing Alert Management System...');

    // Request notification permission
    if ('Notification' in window) {
        Notification.requestPermission().then(permission => {
            AlertManager.notificationPermission = permission;
            console.log('Notification permission:', permission);
        });
    }

    // Load thresholds from backend
    loadThresholds();

    // Load alert history
    loadAlertHistory();

    // Setup periodic cleanup
    setInterval(cleanupExpiredAlerts, 5000); // Every 5 seconds

    // Setup snooze timer updates
    setInterval(updateSnoozeTimers, 1000); // Every second

    console.log('Alert system initialized');
}

/**
 * Load thresholds from backend
 */
async function loadThresholds() {
    try {
        const response = await fetch('/api/alert-thresholds');
        if (response.ok) {
            const thresholds = await response.json();
            AlertManager.thresholds = thresholds;
            console.log('Loaded alert thresholds:', thresholds);
            updateThresholdUI();
        }
    } catch (error) {
        console.warn('Could not load thresholds, using defaults:', error);
    }
}

/**
 * Save thresholds to backend
 */
async function saveThresholds() {
    try {
        const response = await fetch('/api/alert-thresholds', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(AlertManager.thresholds)
        });

        if (response.ok) {
            showToast('Alert thresholds saved successfully', 'success');
        } else {
            showToast('Failed to save thresholds', 'error');
        }
    } catch (error) {
        console.error('Error saving thresholds:', error);
        showToast('Error saving thresholds', 'error');
    }
}

/**
 * Load alert history from backend
 */
async function loadAlertHistory() {
    try {
        const response = await fetch('/api/alert-history?limit=50');
        if (response.ok) {
            const history = await response.json();
            AlertManager.alerts = history.alerts || [];
            updateAlertHistoryUI();
            updateAlertBadges();
        }
    } catch (error) {
        console.warn('Could not load alert history:', error);
    }
}

/**
 * Process incoming alert from WebSocket
 */
function processAlert(alertData) {
    const alert = {
        id: generateAlertId(),
        gpuId: alertData.gpu_id,
        gpuName: alertData.gpu_name || `GPU ${alertData.gpu_id}`,
        type: alertData.type, // 'Temperature', 'Memory', 'Power'
        severity: alertData.severity, // 'warning', 'critical'
        value: alertData.value,
        threshold: alertData.threshold,
        timestamp: Date.now(),
        message: alertData.message,
        state: 'active' // 'active', 'snoozed', 'acknowledged', 'resolved'
    };

    // Debug logging
    console.log('Processing alert:', alert);

    // Check if this is a duplicate alert within cooldown period
    const alertKey = `${alert.gpuId}-${alert.type}-${alert.severity}`;
    const lastTime = AlertManager.lastAlertTime.get(alertKey);

    if (lastTime && (Date.now() - lastTime) < AlertManager.settings.cooldownPeriod * 1000) {
        console.log('Alert suppressed (cooldown):', alertKey);
        return;
    }

    // Add to history
    AlertManager.alerts.unshift(alert);
    if (AlertManager.alerts.length > 100) {
        AlertManager.alerts = AlertManager.alerts.slice(0, 100);
    }

    // Mark as active
    AlertManager.activeAlerts.add(alert.id);
    AlertManager.lastAlertTime.set(alertKey, Date.now());

    // Update UI
    updateAlertBanner();
    updateAlertHistoryUI();
    updateAlertBadges();

    // Show browser notification
    if (AlertManager.settings.enableBrowserNotifications) {
        showBrowserNotification(alert);
    }

    // Play sound
    if (AlertManager.settings.enableSoundAlerts) {
        playAlertSound(alert.severity);
    }

    // Animate the GPU card that triggered the alert
    highlightGPUCard(alert.gpuId);
}

/**
 * Check for alerts based on current metrics
 */
function checkAlerts(gpuData) {
    // Safety check - ensure we have valid GPU data
    if (!gpuData || typeof gpuData !== 'object') {
        return;
    }

    const alerts = [];

    // Get GPU identifier (use index from data.gpus key)
    const gpuId = gpuData.index !== undefined ? gpuData.index : 0;
    const gpuName = gpuData.name || `GPU ${gpuId}`;

    // Debug logging (can be removed later)
    console.log('Checking alerts for GPU:', {
        gpuId,
        gpuName,
        temperature: gpuData.temperature,
        memory_used: gpuData.memory_used,
        memory_total: gpuData.memory_total,
        power_draw: gpuData.power_draw,
        power_limit: gpuData.power_limit
    });

    // Temperature alerts
    if (gpuData.temperature !== undefined && gpuData.temperature >= AlertManager.thresholds.temp_critical) {
        alerts.push({
            gpu_id: gpuId,
            gpu_name: gpuName,
            type: AlertTypes.TEMPERATURE,
            severity: 'critical',
            value: gpuData.temperature,
            threshold: AlertManager.thresholds.temp_critical,
            message: `Temperature critical: ${gpuData.temperature.toFixed(1)}°C`
        });
    } else if (gpuData.temperature !== undefined && gpuData.temperature >= AlertManager.thresholds.temp_warning) {
        alerts.push({
            gpu_id: gpuId,
            gpu_name: gpuName,
            type: AlertTypes.TEMPERATURE,
            severity: 'warning',
            value: gpuData.temperature,
            threshold: AlertManager.thresholds.temp_warning,
            message: `Temperature elevated: ${gpuData.temperature.toFixed(1)}°C`
        });
    }

    // Memory alerts (use snake_case field names from backend)
    const memoryUsed = gpuData.memory_used || 0;
    const memoryTotal = gpuData.memory_total || 1;
    const memoryPercent = (memoryUsed / memoryTotal) * 100;

    if (memoryPercent >= AlertManager.thresholds.memory_critical) {
        alerts.push({
            gpu_id: gpuId,
            gpu_name: gpuName,
            type: AlertTypes.MEMORY,
            severity: 'critical',
            value: memoryPercent,
            threshold: AlertManager.thresholds.memory_critical,
            message: `Memory critical: ${memoryPercent.toFixed(1)}%`
        });
    } else if (memoryPercent >= AlertManager.thresholds.memory_warning) {
        alerts.push({
            gpu_id: gpuId,
            gpu_name: gpuName,
            type: AlertTypes.MEMORY,
            severity: 'warning',
            value: memoryPercent,
            threshold: AlertManager.thresholds.memory_warning,
            message: `Memory high: ${memoryPercent.toFixed(1)}%`
        });
    }

    // Power alerts (use snake_case field names from backend)
    const powerDraw = gpuData.power_draw || 0;
    const powerLimit = gpuData.power_limit || 0;

    if (powerLimit > 0) {
        const powerPercent = (powerDraw / powerLimit) * 100;
        if (powerPercent >= AlertManager.thresholds.power_critical) {
            alerts.push({
                gpu_id: gpuId,
                gpu_name: gpuName,
                type: AlertTypes.POWER,
                severity: 'critical',
                value: powerPercent,
                threshold: AlertManager.thresholds.power_critical,
                message: `Power critical: ${powerPercent.toFixed(1)}%`
            });
        } else if (powerPercent >= AlertManager.thresholds.power_warning) {
            alerts.push({
                gpu_id: gpuId,
                gpu_name: gpuName,
                type: AlertTypes.POWER,
                severity: 'warning',
                value: powerPercent,
                threshold: AlertManager.thresholds.power_warning,
                message: `Power high: ${powerPercent.toFixed(1)}%`
            });
        }
    }

    // Process each alert
    alerts.forEach(alertData => processAlert(alertData));
}

/**
 * Update Alert Banner
 */
function updateAlertBanner() {
    if (!AlertManager.settings.enableBanner) {
        document.getElementById('alert-banner')?.classList.add('hidden');
        return;
    }

    const activeAlerts = getActiveAlerts();
    const banner = document.getElementById('alert-banner');

    if (!banner) return;

    if (activeAlerts.length === 0) {
        banner.classList.add('hidden');
        return;
    }

    // Count by severity
    const critical = activeAlerts.filter(a => a.severity === 'critical').length;
    const warning = activeAlerts.filter(a => a.severity === 'warning').length;

    // Update banner content
    const bannerContent = document.getElementById('alert-banner-content');
    bannerContent.innerHTML = `
        <div style="display: flex; align-items: center; gap: 1rem;">
            <span style="font-size: 1.25rem;">🚨</span>
            <div>
                <strong>${activeAlerts.length} Active Alert${activeAlerts.length > 1 ? 's' : ''}</strong>
                ${critical > 0 ? `<span class="alert-badge alert-badge-critical">${critical} Critical</span>` : ''}
                ${warning > 0 ? `<span class="alert-badge alert-badge-warning">${warning} Warning</span>` : ''}
            </div>
        </div>
        <div style="display: flex; gap: 0.5rem;">
            <button class="alert-banner-btn" onclick="openAlertPanel()">View Alerts</button>
            <button class="alert-banner-btn alert-banner-btn-secondary" onclick="acknowledgeAllAlerts()">Acknowledge All</button>
            <button class="alert-banner-btn-icon" onclick="closeAlertBanner()">&times;</button>
        </div>
    `;

    banner.classList.remove('hidden');
}

/**
 * Update Alert History UI
 */
function updateAlertHistoryUI() {
    const container = document.getElementById('alert-history-list');
    if (!container) return;

    if (AlertManager.alerts.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <div class="empty-state-icon">🎉</div>
                <div class="empty-state-text">No alerts recorded</div>
                <div class="empty-state-subtext">All systems operating normally</div>
            </div>
        `;
        return;
    }

    const html = AlertManager.alerts.map(alert => {
        const severityInfo = alert.severity === 'critical'
            ? AlertSeverity.CRITICAL
            : AlertSeverity.WARNING;

        const isActive = AlertManager.activeAlerts.has(alert.id);
        const isSnoozed = AlertManager.snoozedAlerts.has(alert.id);
        const isAcknowledged = AlertManager.acknowledgedAlerts.has(alert.id);

        let badge = '';
        if (isSnoozed) {
            const snoozeUntil = AlertManager.snoozedAlerts.get(alert.id);
            const remainingMin = Math.ceil((snoozeUntil - Date.now()) / 60000);
            badge = `<span class="alert-status-badge alert-status-snoozed">💤 ${remainingMin}m</span>`;
        } else if (isAcknowledged) {
            badge = `<span class="alert-status-badge alert-status-acknowledged">✓ ACK</span>`;
        } else if (alert.state === 'resolved') {
            badge = `<span class="alert-status-badge alert-status-resolved">✓ RESOLVED</span>`;
        }

        return `
            <div class="alert-item ${isActive && !isSnoozed && !isAcknowledged ? 'alert-item-active' : ''}" data-alert-id="${alert.id}">
                <div class="alert-item-header">
                    <div class="alert-item-left">
                        <span class="alert-icon">${severityInfo.icon}</span>
                        <span class="alert-time">${formatTime(alert.timestamp)}</span>
                        <span class="alert-severity alert-severity-${alert.severity}">${severityInfo.label}</span>
                        <span class="alert-gpu">${alert.gpuName}</span>
                        ${badge}
                    </div>
                    ${isActive && !isAcknowledged ? `
                        <div class="alert-item-actions">
                            <button class="alert-action-btn" onclick="snoozeAlert('${alert.id}', 5)" title="Snooze 5 min">💤 5m</button>
                            <button class="alert-action-btn" onclick="snoozeAlert('${alert.id}', 30)" title="Snooze 30 min">💤 30m</button>
                            <button class="alert-action-btn" onclick="acknowledgeAlert('${alert.id}')" title="Acknowledge">✓</button>
                        </div>
                    ` : ''}
                </div>
                <div class="alert-item-content">
                    <div class="alert-message">${alert.message}</div>
                    <div class="alert-details">
                        <span class="alert-detail-item">${alert.type}: <strong>${formatAlertValue(alert.type, alert.value)}</strong></span>
                        <span class="alert-detail-item">Threshold: <strong>${formatAlertValue(alert.type, alert.threshold)}</strong></span>
                    </div>
                </div>
            </div>
        `;
    }).join('');

    container.innerHTML = html;
}

/**
 * Update Alert Configuration UI
 */
function updateThresholdUI() {
    // Temperature
    document.getElementById('threshold-temp-warning').value = AlertManager.thresholds.temp_warning;
    document.getElementById('threshold-temp-critical').value = AlertManager.thresholds.temp_critical;

    // Memory
    document.getElementById('threshold-memory-warning').value = AlertManager.thresholds.memory_warning;
    document.getElementById('threshold-memory-critical').value = AlertManager.thresholds.memory_critical;

    // Power
    document.getElementById('threshold-power-warning').value = AlertManager.thresholds.power_warning;
    document.getElementById('threshold-power-critical').value = AlertManager.thresholds.power_critical;

    // Update preview values
    updateThresholdPreviews();
}

/**
 * Update threshold preview values
 */
function updateThresholdPreviews() {
    document.getElementById('preview-temp-warning').textContent =
        document.getElementById('threshold-temp-warning').value + '°C';
    document.getElementById('preview-temp-critical').textContent =
        document.getElementById('threshold-temp-critical').value + '°C';

    document.getElementById('preview-memory-warning').textContent =
        document.getElementById('threshold-memory-warning').value + '%';
    document.getElementById('preview-memory-critical').textContent =
        document.getElementById('threshold-memory-critical').value + '%';

    document.getElementById('preview-power-warning').textContent =
        document.getElementById('threshold-power-warning').value + '%';
    document.getElementById('preview-power-critical').textContent =
        document.getElementById('threshold-power-critical').value + '%';
}

/**
 * Update alert count badges
 */
function updateAlertBadges() {
    const activeAlerts = getActiveAlerts();
    const badge = document.getElementById('alert-count-badge');

    if (badge) {
        if (activeAlerts.length > 0) {
            badge.textContent = activeAlerts.length;
            badge.classList.remove('hidden');
        } else {
            badge.classList.add('hidden');
        }
    }
}

/**
 * Get active alerts (not snoozed, not acknowledged)
 */
function getActiveAlerts() {
    return AlertManager.alerts.filter(alert =>
        AlertManager.activeAlerts.has(alert.id) &&
        !AlertManager.snoozedAlerts.has(alert.id) &&
        !AlertManager.acknowledgedAlerts.has(alert.id) &&
        alert.state !== 'resolved'
    );
}

/**
 * Snooze alert for specified minutes
 */
function snoozeAlert(alertId, minutes) {
    const snoozeUntil = Date.now() + (minutes * 60 * 1000);
    AlertManager.snoozedAlerts.set(alertId, snoozeUntil);

    updateAlertBanner();
    updateAlertHistoryUI();

    showToast(`Alert snoozed for ${minutes} minute${minutes > 1 ? 's' : ''}`, 'info');
}

/**
 * Acknowledge alert
 */
function acknowledgeAlert(alertId) {
    AlertManager.acknowledgedAlerts.add(alertId);
    AlertManager.activeAlerts.delete(alertId);

    const alert = AlertManager.alerts.find(a => a.id === alertId);
    if (alert) {
        alert.state = 'acknowledged';
    }

    updateAlertBanner();
    updateAlertHistoryUI();
    updateAlertBadges();

    showToast('Alert acknowledged', 'success');
}

/**
 * Acknowledge all active alerts
 */
function acknowledgeAllAlerts() {
    const activeAlerts = getActiveAlerts();
    activeAlerts.forEach(alert => {
        AlertManager.acknowledgedAlerts.add(alert.id);
        AlertManager.activeAlerts.delete(alert.id);
        alert.state = 'acknowledged';
    });

    updateAlertBanner();
    updateAlertHistoryUI();
    updateAlertBadges();

    showToast(`${activeAlerts.length} alert${activeAlerts.length > 1 ? 's' : ''} acknowledged`, 'success');
}

/**
 * Cleanup expired snooze timers and old alerts
 */
function cleanupExpiredAlerts() {
    const now = Date.now();

    // Remove expired snoozes
    for (const [alertId, snoozeUntil] of AlertManager.snoozedAlerts.entries()) {
        if (now >= snoozeUntil) {
            AlertManager.snoozedAlerts.delete(alertId);
        }
    }

    // Remove old acknowledged alerts (after 1 hour)
    AlertManager.alerts = AlertManager.alerts.filter(alert => {
        if (alert.state === 'acknowledged') {
            return (now - alert.timestamp) < 3600000; // 1 hour
        }
        return true;
    });
}

/**
 * Update snooze timer displays
 */
function updateSnoozeTimers() {
    const now = Date.now();

    // Update snooze badges
    AlertManager.snoozedAlerts.forEach((snoozeUntil, alertId) => {
        const remainingMin = Math.ceil((snoozeUntil - now) / 60000);
        const badge = document.querySelector(`[data-alert-id="${alertId}"] .alert-status-snoozed`);
        if (badge && remainingMin > 0) {
            badge.textContent = `💤 ${remainingMin}m`;
        }
    });
}

/**
 * Show browser notification
 */
function showBrowserNotification(alert) {
    if (!('Notification' in window) || Notification.permission !== 'granted') {
        return;
    }

    const severityInfo = alert.severity === 'critical'
        ? AlertSeverity.CRITICAL
        : AlertSeverity.WARNING;

    new Notification(`GPU Pro Alert - ${severityInfo.label}`, {
        body: `${alert.gpuName}: ${alert.message}`,
        icon: '/static/favicon.svg',
        tag: `gpu-alert-${alert.id}`,
        requireInteraction: alert.severity === 'critical'
    });
}

/**
 * Play alert sound
 */
function playAlertSound(severity) {
    // Create a simple beep sound using Web Audio API
    const audioContext = new (window.AudioContext || window.webkitAudioContext)();
    const oscillator = audioContext.createOscillator();
    const gainNode = audioContext.createGain();

    oscillator.connect(gainNode);
    gainNode.connect(audioContext.destination);

    oscillator.frequency.value = severity === 'critical' ? 800 : 600;
    oscillator.type = 'sine';

    gainNode.gain.setValueAtTime(0.3, audioContext.currentTime);
    gainNode.gain.exponentialRampToValueAtTime(0.01, audioContext.currentTime + 0.5);

    oscillator.start(audioContext.currentTime);
    oscillator.stop(audioContext.currentTime + 0.5);
}

/**
 * Highlight GPU card that triggered alert
 */
function highlightGPUCard(gpuId) {
    const card = document.querySelector(`[data-gpu-id="${gpuId}"]`);
    if (card) {
        card.classList.add('gpu-card-alert');
        setTimeout(() => card.classList.remove('gpu-card-alert'), 2000);
    }
}

/**
 * Toggle Alert Panel
 */
function toggleAlertPanel() {
    const panel = document.getElementById('alert-panel');
    panel?.classList.toggle('alert-panel-open');
}

/**
 * Open Alert Panel
 */
function openAlertPanel() {
    const panel = document.getElementById('alert-panel');
    panel?.classList.add('alert-panel-open');
}

/**
 * Close Alert Panel
 */
function closeAlertPanel() {
    const panel = document.getElementById('alert-panel');
    panel?.classList.remove('alert-panel-open');
}

/**
 * Close Alert Banner
 */
function closeAlertBanner() {
    const banner = document.getElementById('alert-banner');
    banner?.classList.add('hidden');
}

/**
 * Switch Alert Tab
 */
function switchAlertTab(tabName) {
    // Update tab buttons
    const tabs = document.querySelectorAll('.alert-tab');
    tabs.forEach(tab => {
        if (tab.textContent.toLowerCase().includes(tabName)) {
            tab.classList.add('active');
        } else {
            tab.classList.remove('active');
        }
    });

    // Update tab content
    const contents = document.querySelectorAll('.alert-tab-content');
    contents.forEach(content => {
        if (content.id === `alert-tab-${tabName}`) {
            content.classList.add('active');
        } else {
            content.classList.remove('active');
        }
    });
}

/**
 * Save alert configuration
 */
function saveAlertConfig() {
    // Read threshold values
    AlertManager.thresholds.temp_warning = parseInt(document.getElementById('threshold-temp-warning').value);
    AlertManager.thresholds.temp_critical = parseInt(document.getElementById('threshold-temp-critical').value);
    AlertManager.thresholds.memory_warning = parseInt(document.getElementById('threshold-memory-warning').value);
    AlertManager.thresholds.memory_critical = parseInt(document.getElementById('threshold-memory-critical').value);
    AlertManager.thresholds.power_warning = parseInt(document.getElementById('threshold-power-warning').value);
    AlertManager.thresholds.power_critical = parseInt(document.getElementById('threshold-power-critical').value);

    // Validate thresholds
    if (AlertManager.thresholds.temp_warning >= AlertManager.thresholds.temp_critical) {
        showToast('Temperature warning must be less than critical', 'error');
        return;
    }
    if (AlertManager.thresholds.memory_warning >= AlertManager.thresholds.memory_critical) {
        showToast('Memory warning must be less than critical', 'error');
        return;
    }
    if (AlertManager.thresholds.power_warning >= AlertManager.thresholds.power_critical) {
        showToast('Power warning must be less than critical', 'error');
        return;
    }

    // Save to backend
    saveThresholds();
}

/**
 * Reset thresholds to defaults
 */
function resetThresholds() {
    if (!confirm('Reset all thresholds to default values?')) return;

    AlertManager.thresholds = {
        temp_warning: 75,
        temp_critical: 85,
        memory_warning: 85,
        memory_critical: 95,
        power_warning: 90,
        power_critical: 98
    };

    updateThresholdUI();
    saveThresholds();
}

/**
 * Clear all alerts
 */
function clearAllAlerts() {
    if (!confirm('Clear all alert history?')) return;

    AlertManager.alerts = [];
    AlertManager.activeAlerts.clear();
    AlertManager.snoozedAlerts.clear();
    AlertManager.acknowledgedAlerts.clear();

    updateAlertBanner();
    updateAlertHistoryUI();
    updateAlertBadges();

    showToast('All alerts cleared', 'success');
}

/**
 * Show toast notification
 */
function showToast(message, type = 'info') {
    const toast = document.getElementById('toast-notification');
    if (!toast) return;

    toast.textContent = message;
    toast.className = `toast toast-${type}`;
    toast.classList.add('toast-show');

    setTimeout(() => {
        toast.classList.remove('toast-show');
    }, 3000);
}

/**
 * Utility Functions
 */

function generateAlertId() {
    return `alert-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

function formatTime(timestamp) {
    const date = new Date(timestamp);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

function formatAlertValue(type, value) {
    switch (type) {
        case AlertTypes.TEMPERATURE:
            return `${value.toFixed(1)}°C`;
        case AlertTypes.MEMORY:
        case AlertTypes.POWER:
        case AlertTypes.UTILIZATION:
            return `${value.toFixed(1)}%`;
        case AlertTypes.FAN:
            return `${Math.round(value)} RPM`;
        default:
            return value.toFixed(1);
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    initAlertSystem();
});
