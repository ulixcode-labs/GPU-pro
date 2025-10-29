/**
 * System Metrics Tab - Network I/O, Disk I/O, Connections, Files
 */

// System metrics data storage
const systemMetricsData = {
    network: {
        labels: [],
        rxData: [],
        txData: []
    },
    disk: {
        labels: [],
        readData: [],
        writeData: []
    },
    connections: {
        labels: [],
        tcpData: [],
        udpData: [],
        otherData: []
    },
    openFiles: {
        labels: [],
        data: []
    }
};

// System metrics charts
const systemMetricsCharts = {};

// Connection map
let connectionMap = null;
let connectionMarkers = [];
let lastGeoLocationsJson = null; // Track last geolocation data to avoid unnecessary updates

// Time range for system metrics charts
let systemMetricsTimeRange = 60; // Default 1 minute

// Initialize network I/O chart
function initNetworkIOChart() {
    const canvas = document.getElementById('network-io-chart');
    if (!canvas || systemMetricsCharts.network) return;

    systemMetricsCharts.network = new Chart(canvas, {
        type: 'line',
        data: {
            labels: systemMetricsData.network.labels,
            datasets: [
                {
                    label: 'Download (KB/s)',
                    data: systemMetricsData.network.rxData,
                    borderColor: '#4facfe',
                    backgroundColor: 'rgba(79, 172, 254, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: 'Upload (KB/s)',
                    data: systemMetricsData.network.txData,
                    borderColor: '#f093fb',
                    backgroundColor: 'rgba(240, 147, 251, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            layout: {
                padding: { left: 10, right: 10, top: 10, bottom: 15 }
            },
            scales: {
                x: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)'
                    },
                    ticks: {
                        color: '#94a3b8',
                        maxTicksLimit: 8
                    }
                },
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)'
                    },
                    ticks: {
                        color: '#94a3b8',
                        callback: function(value) {
                            return value.toFixed(0) + ' KB/s';
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    display: true,
                    labels: {
                        color: '#cbd5e1',
                        usePointStyle: true,
                        padding: 15
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(15, 23, 42, 0.95)',
                    titleColor: '#f1f5f9',
                    bodyColor: '#cbd5e1',
                    borderColor: 'rgba(100, 116, 139, 0.3)',
                    borderWidth: 1,
                    padding: 12,
                    displayColors: true,
                    callbacks: {
                        label: function(context) {
                            return context.dataset.label + ': ' + context.parsed.y.toFixed(2) + ' KB/s';
                        }
                    }
                }
            }
        }
    });
}

// Initialize disk I/O chart
function initDiskIOChart() {
    const canvas = document.getElementById('disk-io-chart');
    if (!canvas || systemMetricsCharts.disk) return;

    systemMetricsCharts.disk = new Chart(canvas, {
        type: 'line',
        data: {
            labels: systemMetricsData.disk.labels,
            datasets: [
                {
                    label: 'Read (KB/s)',
                    data: systemMetricsData.disk.readData,
                    borderColor: '#43e97b',
                    backgroundColor: 'rgba(67, 233, 123, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: 'Write (KB/s)',
                    data: systemMetricsData.disk.writeData,
                    borderColor: '#fa709a',
                    backgroundColor: 'rgba(250, 112, 154, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            layout: {
                padding: { left: 10, right: 10, top: 10, bottom: 15 }
            },
            scales: {
                x: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)'
                    },
                    ticks: {
                        color: '#94a3b8',
                        maxTicksLimit: 8
                    }
                },
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)'
                    },
                    ticks: {
                        color: '#94a3b8',
                        callback: function(value) {
                            return value.toFixed(0) + ' KB/s';
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    display: true,
                    labels: {
                        color: '#cbd5e1',
                        usePointStyle: true,
                        padding: 15
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(15, 23, 42, 0.95)',
                    titleColor: '#f1f5f9',
                    bodyColor: '#cbd5e1',
                    borderColor: 'rgba(100, 116, 139, 0.3)',
                    borderWidth: 1,
                    padding: 12,
                    displayColors: true,
                    callbacks: {
                        label: function(context) {
                            return context.dataset.label + ': ' + context.parsed.y.toFixed(2) + ' KB/s';
                        }
                    }
                }
            }
        }
    });
}

// Update system metrics data
function updateSystemMetrics(metricsData) {
    if (!metricsData) return;

    const now = new Date().toLocaleTimeString();
    const maxPoints = systemMetricsTimeRange * 2;

    // Update network I/O
    if (metricsData.network_io && metricsData.network_io.length > 0) {
        let totalRx = 0;
        let totalTx = 0;

        metricsData.network_io.forEach(iface => {
            totalRx += iface.rx_rate || 0;
            totalTx += iface.tx_rate || 0;
        });

        systemMetricsData.network.labels.push(now);
        systemMetricsData.network.rxData.push(totalRx / 1024); // Convert to KB/s
        systemMetricsData.network.txData.push(totalTx / 1024);

        // Trim old data
        if (systemMetricsData.network.labels.length > maxPoints) {
            systemMetricsData.network.labels.shift();
            systemMetricsData.network.rxData.shift();
            systemMetricsData.network.txData.shift();
        }

        // Update stats display
        const statsEl = document.getElementById('network-stats');
        if (statsEl) {
            statsEl.innerHTML = '<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-top: 1rem;"><div style="padding: 0.75rem; background: rgba(79, 172, 254, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">Download</div><div style="font-size: 1.25rem; font-weight: 700; color: #4facfe;">' + (totalRx / 1024).toFixed(2) + ' KB/s</div></div><div style="padding: 0.75rem; background: rgba(240, 147, 251, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">Upload</div><div style="font-size: 1.25rem; font-weight: 700; color: #f093fb;">' + (totalTx / 1024).toFixed(2) + ' KB/s</div></div></div>';
        }
    }

    // Update disk I/O
    if (metricsData.disk_io && metricsData.disk_io.length > 0) {
        let totalRead = 0;
        let totalWrite = 0;

        metricsData.disk_io.forEach(disk => {
            totalRead += disk.read_kbps || 0;
            totalWrite += disk.write_kbps || 0;
        });

        systemMetricsData.disk.labels.push(now);
        systemMetricsData.disk.readData.push(totalRead);
        systemMetricsData.disk.writeData.push(totalWrite);

        // Trim old data
        if (systemMetricsData.disk.labels.length > maxPoints) {
            systemMetricsData.disk.labels.shift();
            systemMetricsData.disk.readData.shift();
            systemMetricsData.disk.writeData.shift();
        }

        // Update stats display
        const statsEl = document.getElementById('disk-stats');
        if (statsEl) {
            statsEl.innerHTML = '<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-top: 1rem;"><div style="padding: 0.75rem; background: rgba(67, 233, 123, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">Read</div><div style="font-size: 1.25rem; font-weight: 700; color: #43e97b;">' + totalRead.toFixed(2) + ' KB/s</div></div><div style="padding: 0.75rem; background: rgba(250, 112, 154, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">Write</div><div style="font-size: 1.25rem; font-weight: 700; color: #fa709a;">' + totalWrite.toFixed(2) + ' KB/s</div></div></div>';
        }
    }

    // Update connection stats
    if (metricsData.connection_stats) {
        const connStats = metricsData.connection_stats;

        systemMetricsData.connections.labels.push(now);
        systemMetricsData.connections.tcpData.push(connStats.tcp || 0);
        systemMetricsData.connections.udpData.push(connStats.udp || 0);
        systemMetricsData.connections.otherData.push(connStats.other || 0);

        // Trim old data
        if (systemMetricsData.connections.labels.length > maxPoints) {
            systemMetricsData.connections.labels.shift();
            systemMetricsData.connections.tcpData.shift();
            systemMetricsData.connections.udpData.shift();
            systemMetricsData.connections.otherData.shift();
        }

        // Update stats display
        const connStatsEl = document.getElementById('connection-stats');
        if (connStatsEl) {
            connStatsEl.innerHTML = '<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 1rem; margin-top: 1rem;"><div style="padding: 0.75rem; background: rgba(79, 172, 254, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">TCP</div><div style="font-size: 1.25rem; font-weight: 700; color: #4facfe;">' + (connStats.tcp || 0) + '</div></div><div style="padding: 0.75rem; background: rgba(67, 233, 123, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">UDP</div><div style="font-size: 1.25rem; font-weight: 700; color: #43e97b;">' + (connStats.udp || 0) + '</div></div><div style="padding: 0.75rem; background: rgba(240, 147, 251, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">Other</div><div style="font-size: 1.25rem; font-weight: 700; color: #f093fb;">' + (connStats.other || 0) + '</div></div><div style="padding: 0.75rem; background: rgba(100, 116, 139, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">Total</div><div style="font-size: 1.25rem; font-weight: 700; color: #cbd5e1;">' + (connStats.total || 0) + '</div></div></div>';
        }
    }

    // Update open files count
    if (metricsData.open_files !== undefined) {
        systemMetricsData.openFiles.labels.push(now);
        systemMetricsData.openFiles.data.push(metricsData.open_files);

        // Trim old data
        if (systemMetricsData.openFiles.labels.length > maxPoints) {
            systemMetricsData.openFiles.labels.shift();
            systemMetricsData.openFiles.data.shift();
        }

        // Update stats display
        const filesStatsEl = document.getElementById('open-files-stats');
        if (filesStatsEl) {
            filesStatsEl.innerHTML = '<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-top: 1rem;"><div style="padding: 0.75rem; background: rgba(251, 191, 36, 0.1); border-radius: 0.5rem;"><div style="font-size: 0.75rem; color: var(--text-secondary);">System-wide Open Files</div><div style="font-size: 1.25rem; font-weight: 700; color: #fbbf24;">' + metricsData.open_files + '</div></div></div>';
        }
    }

    // Initialize charts if needed
    if (!systemMetricsCharts.network) initNetworkIOChart();
    if (!systemMetricsCharts.disk) initDiskIOChart();
    if (!systemMetricsCharts.connections) initConnectionsChart();
    if (!systemMetricsCharts.openFiles) initOpenFilesChart();

    // Update charts
    if (systemMetricsCharts.network) systemMetricsCharts.network.update('none');
    if (systemMetricsCharts.disk) systemMetricsCharts.disk.update('none');
    if (systemMetricsCharts.connections) systemMetricsCharts.connections.update('none');
    if (systemMetricsCharts.openFiles) systemMetricsCharts.openFiles.update('none');

    // Update network connections
    updateNetworkConnections(metricsData.connections || []);

    // Update largest files
    updateLargestFiles(metricsData.largest_files || []);

    // Update connection map with geolocation data
    if (metricsData.geo_locations) {
        updateConnectionMap(metricsData.geo_locations);
    }
}

// Update network connections table
function updateNetworkConnections(connections) {
    const tbody = document.getElementById('connections-tbody');
    const countEl = document.getElementById('connection-count');

    if (!tbody) return;

    if (countEl) {
        countEl.textContent = connections.length + ' active connection' + (connections.length !== 1 ? 's' : '');
    }

    if (connections.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; padding: 2rem; color: var(--text-secondary);">No active connections</td></tr>';
        return;
    }

    tbody.innerHTML = connections.slice(0, 50).map(conn => {
        // Format foreign address with external IP indicator
        let foreignAddrHtml = '<span style="font-family: monospace; font-size: 0.85rem;">';
        if (conn.is_external) {
            foreignAddrHtml += '<span style="display: inline-flex; align-items: center; gap: 0.3rem;">';
            foreignAddrHtml += '<span style="color: #fbbf24; font-weight: 600;">🌐</span>';
            foreignAddrHtml += '<span style="color: #4facfe; font-weight: 500;">' + (conn.foreign_addr || 'N/A') + '</span>';
            foreignAddrHtml += '<span style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 0.1rem 0.4rem; border-radius: 0.25rem; font-size: 0.65rem; font-weight: 600; margin-left: 0.3rem;">EXTERNAL</span>';
            foreignAddrHtml += '</span>';
        } else {
            foreignAddrHtml += (conn.foreign_addr || 'N/A');
        }
        foreignAddrHtml += '</span>';

        // Format duration with color coding
        let durationHtml = '<span style="font-family: monospace; font-size: 0.85rem; font-weight: 600;">';
        if (conn.duration_sec < 60) {
            durationHtml += '<span style="color: #43e97b;">' + (conn.duration || '0s') + '</span>'; // Green for < 1 min
        } else if (conn.duration_sec < 3600) {
            durationHtml += '<span style="color: #4facfe;">' + (conn.duration || '0s') + '</span>'; // Blue for < 1 hour
        } else if (conn.duration_sec < 86400) {
            durationHtml += '<span style="color: #fbbf24;">' + (conn.duration || '0s') + '</span>'; // Yellow for < 1 day
        } else {
            durationHtml += '<span style="color: #f093fb;">' + (conn.duration || '0s') + '</span>'; // Pink for >= 1 day
        }
        durationHtml += '</span>';

        return '<tr>' +
            '<td>' + (conn.protocol || 'N/A') + '</td>' +
            '<td style="font-family: monospace; font-size: 0.85rem;">' + (conn.local_addr || 'N/A') + '</td>' +
            '<td>' + foreignAddrHtml + '</td>' +
            '<td><span class="connection-state connection-state-' + (conn.state || 'unknown').toLowerCase() + '">' + (conn.state || 'N/A') + '</span></td>' +
            '<td>' + durationHtml + '</td>' +
            '<td>' + (conn.pid || 'N/A') + '</td>' +
            '<td style="font-weight: 500;">' + (conn.program || 'N/A') + '</td>' +
            '</tr>';
    }).join('');
}

// Update largest files list
function updateLargestFiles(files) {
    const listEl = document.getElementById('largest-files-list');
    if (!listEl) return;

    if (files.length === 0) {
        listEl.innerHTML = '<div style="text-align: center; padding: 2rem; color: var(--text-secondary);">No files found</div>';
        return;
    }

    listEl.innerHTML = files.map((file, index) => {
        // Show both sizes if file is sparse
        let sizeDisplay = file.actual_size_human;
        if (file.is_sparse) {
            sizeDisplay = '<span title="Actual: ' + file.actual_size_human + ' | Apparent: ' + file.size_human + '">' +
                         file.actual_size_human + ' <span style="color: #fbbf24; font-size: 0.7rem;">(' + file.size_human + ' sparse)</span></span>';
        }

        return '<div class="file-item">' +
               '<div class="file-rank">#' + (index + 1) + '</div>' +
               '<div class="file-details">' +
               '<div class="file-path" title="' + file.path + '">' + file.path + '</div>' +
               '<div class="file-meta">' +
               '<span class="file-size">' + sizeDisplay + '</span>' +
               '<span class="file-date">Modified: ' + file.mod_time + '</span>' +
               '</div></div></div>';
    }).join('');
}

// Refresh largest files from a specific directory
async function refreshLargestFiles() {
    const directoryInput = document.getElementById('largest-files-directory');
    const listEl = document.getElementById('largest-files-list');

    if (!directoryInput || !listEl) return;

    const directory = directoryInput.value.trim();

    // Validate directory input
    if (!directory) {
        listEl.innerHTML = '<div style="text-align: center; padding: 2rem; color: var(--text-secondary);">' +
            '<div style="font-size: 3rem; margin-bottom: 1rem;">📁</div>' +
            '<div style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem;">No directory specified</div>' +
            '<div style="font-size: 0.85rem; opacity: 0.7;">Enter a directory path and click Refresh to view largest files</div>' +
            '</div>';
        return;
    }

    // Show loading state with modern spinner
    listEl.innerHTML = '<div style="display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 3rem; color: var(--text-secondary);">' +
        '<div class="spinner" style="margin-bottom: 1.5rem; border: 3px solid rgba(0, 212, 255, 0.1); border-top: 3px solid #00d4ff; border-radius: 50%; width: 48px; height: 48px; animation: spin 0.8s linear infinite;"></div>' +
        '<div style="font-size: 1rem; font-weight: 600; color: var(--text-primary);">Loading files...</div>' +
        '<div style="font-size: 0.85rem; margin-top: 0.5rem; opacity: 0.7;">Scanning ' + directory + '</div>' +
        '</div>';

    try {
        const response = await fetch('/api/largest-files?directory=' + encodeURIComponent(directory));
        const data = await response.json();

        if (data.files && data.files.length > 0) {
            updateLargestFiles(data.files);
        } else {
            listEl.innerHTML = '<div style="text-align: center; padding: 2rem; color: var(--text-secondary);">' +
                '<div style="font-size: 3rem; margin-bottom: 1rem;">📂</div>' +
                '<div style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem;">No files found</div>' +
                '<div style="font-size: 0.85rem; opacity: 0.7;">No files found in ' + directory + '</div>' +
                '</div>';
        }
    } catch (error) {
        console.error('Error fetching largest files:', error);
        listEl.innerHTML = '<div style="text-align: center; padding: 2rem; color: #f5576c;">' +
            '<div style="font-size: 3rem; margin-bottom: 1rem;">⚠️</div>' +
            '<div style="font-size: 1rem; font-weight: 600; margin-bottom: 0.5rem;">Error loading files</div>' +
            '<div style="font-size: 0.85rem; opacity: 0.8;">' + error.message + '</div>' +
            '</div>';
    }
}

// Set time range for system metrics charts
// Deprecated: Now controlled by global time range
function setSystemMetricsTimeRange(seconds) {
    console.log('Redirecting setSystemMetricsTimeRange to global time range');
    setGlobalTimeRange(seconds);
}

// Show notification for time range change
function showSystemMetricsTimeRangeNotification(seconds) {
    let notification = document.getElementById('system-metrics-time-notification');
    if (!notification) {
        notification = document.createElement('div');
        notification.id = 'system-metrics-time-notification';
        notification.style.cssText = 'position: fixed; top: 20px; left: 50%; transform: translateX(-50%); background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 1.5rem; border-radius: 0.5rem; box-shadow: 0 10px 25px rgba(0,0,0,0.3); z-index: 10000; font-size: 0.9rem; font-weight: 600; opacity: 0; transition: opacity 0.3s ease;';
        document.body.appendChild(notification);
    }

    const minutes = seconds / 60;
    notification.textContent = 'System Metrics - Chart range: ' + minutes + ' min';
    notification.style.opacity = '1';

    setTimeout(function() {
        notification.style.opacity = '0';
        setTimeout(function() {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, 2000);
}

// Initialize network connections count chart
function initConnectionsChart() {
    const canvas = document.getElementById('connections-count-chart');
    if (!canvas || systemMetricsCharts.connections) return;

    systemMetricsCharts.connections = new Chart(canvas, {
        type: 'line',
        data: {
            labels: systemMetricsData.connections.labels,
            datasets: [
                {
                    label: 'TCP',
                    data: systemMetricsData.connections.tcpData,
                    borderColor: '#4facfe',
                    backgroundColor: 'rgba(79, 172, 254, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: 'UDP',
                    data: systemMetricsData.connections.udpData,
                    borderColor: '#43e97b',
                    backgroundColor: 'rgba(67, 233, 123, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0
                },
                {
                    label: 'Other',
                    data: systemMetricsData.connections.otherData,
                    borderColor: '#f093fb',
                    backgroundColor: 'rgba(240, 147, 251, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            layout: {
                padding: { left: 10, right: 10, top: 10, bottom: 15 }
            },
            scales: {
                x: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)'
                    },
                    ticks: {
                        color: '#94a3b8',
                        maxTicksLimit: 8
                    }
                },
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)'
                    },
                    ticks: {
                        color: '#94a3b8',
                        callback: function(value) {
                            return Math.floor(value);
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'bottom',
                    align: 'center',
                    labels: {
                        color: '#cbd5e1',
                        usePointStyle: true,
                        padding: 15,
                        boxWidth: 12,
                        boxHeight: 12
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(15, 23, 42, 0.95)',
                    titleColor: '#f1f5f9',
                    bodyColor: '#cbd5e1',
                    borderColor: 'rgba(100, 116, 139, 0.3)',
                    borderWidth: 1,
                    padding: 12,
                    displayColors: true,
                    callbacks: {
                        label: function(context) {
                            return context.dataset.label + ': ' + Math.floor(context.parsed.y) + ' connections';
                        }
                    }
                }
            }
        }
    });
}

// Initialize open files count chart
function initOpenFilesChart() {
    const canvas = document.getElementById('open-files-chart');
    if (!canvas || systemMetricsCharts.openFiles) return;

    systemMetricsCharts.openFiles = new Chart(canvas, {
        type: 'line',
        data: {
            labels: systemMetricsData.openFiles.labels,
            datasets: [
                {
                    label: 'Open Files',
                    data: systemMetricsData.openFiles.data,
                    borderColor: '#fbbf24',
                    backgroundColor: 'rgba(251, 191, 36, 0.1)',
                    borderWidth: 2,
                    tension: 0.4,
                    fill: true,
                    pointRadius: 0
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            layout: {
                padding: { left: 10, right: 10, top: 10, bottom: 15 }
            },
            scales: {
                x: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)'
                    },
                    ticks: {
                        color: '#94a3b8',
                        maxTicksLimit: 8
                    }
                },
                y: {
                    beginAtZero: true,
                    grid: {
                        color: 'rgba(255, 255, 255, 0.05)'
                    },
                    ticks: {
                        color: '#94a3b8',
                        callback: function(value) {
                            return Math.floor(value);
                        }
                    }
                }
            },
            plugins: {
                legend: {
                    display: true,
                    labels: {
                        color: '#cbd5e1',
                        usePointStyle: true,
                        padding: 15
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(15, 23, 42, 0.95)',
                    titleColor: '#f1f5f9',
                    bodyColor: '#cbd5e1',
                    borderColor: 'rgba(100, 116, 139, 0.3)',
                    borderWidth: 1,
                    padding: 12,
                    displayColors: true,
                    callbacks: {
                        label: function(context) {
                            return 'Open Files: ' + Math.floor(context.parsed.y);
                        }
                    }
                }
            }
        }
    });
}

// Initialize connection world map
function initConnectionMap() {
    const mapEl = document.getElementById('connection-map');
    if (!mapEl || connectionMap) return;

    // Initialize Leaflet map
    connectionMap = L.map('connection-map', {
        center: [20, 0],
        zoom: 2,
        minZoom: 1,
        maxZoom: 18,
        worldCopyJump: true,
        zoomControl: true
    });

    // Add dark theme tile layer
    L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
        attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
        subdomains: 'abcd',
        maxZoom: 20
    }).addTo(connectionMap);

    // Fix map size after initialization
    setTimeout(function() {
        connectionMap.invalidateSize();
        console.log('Connection map initialized and resized');
    }, 100);
}

// Update connection map with geolocation data
function updateConnectionMap(geoLocations) {
    if (!geoLocations) return;

    // Initialize map if needed
    if (!connectionMap) initConnectionMap();
    if (!connectionMap) return;

    // Check if geolocation data has changed to avoid unnecessary updates
    const currentGeoJson = JSON.stringify(geoLocations);
    if (currentGeoJson === lastGeoLocationsJson) {
        // Data hasn't changed, skip update to avoid closing popups
        return;
    }
    lastGeoLocationsJson = currentGeoJson;

    // Clear existing markers
    connectionMarkers.forEach(marker => marker.remove());
    connectionMarkers = [];

    // Count external IPs
    const ipCount = Object.keys(geoLocations).length;
    const countEl = document.getElementById('map-connection-count');
    if (countEl) {
        countEl.textContent = ipCount + ' external IP' + (ipCount !== 1 ? 's' : '');
    }

    // If no locations, reset to world view
    if (ipCount === 0) {
        connectionMap.setView([20, 0], 2);
        console.log('No locations to display on map');
        return;
    }

    // Collect marker coordinates for auto-zoom
    const bounds = [];

    // Group IPs by location to handle overlapping markers
    const locationGroups = {};
    let validLocations = 0;
    let skippedLocations = 0;

    Object.entries(geoLocations).forEach(([ip, location]) => {
        if (!location) {
            skippedLocations++;
            console.log('Skipped IP (no location):', ip);
            return;
        }

        if (location.latitude === 0 && location.longitude === 0) {
            skippedLocations++;
            console.log('Skipped IP (0,0 coordinates):', ip, location);
            return;
        }

        validLocations++;
        const locationKey = location.latitude.toFixed(4) + ',' + location.longitude.toFixed(4);

        if (!locationGroups[locationKey]) {
            locationGroups[locationKey] = {
                latLng: [location.latitude, location.longitude],
                ips: []
            };
        }

        locationGroups[locationKey].ips.push({
            ip: ip,
            location: location
        });
    });

    console.log(`Processing ${validLocations} valid locations, skipped ${skippedLocations}`);
    console.log(`Grouped into ${Object.keys(locationGroups).length} unique map positions`);

    // Add markers for each unique location
    Object.values(locationGroups).forEach(group => {
        const latLng = group.latLng;
        bounds.push(latLng);

        // Create marker
        const marker = L.marker(latLng, {
            icon: L.divIcon({
                className: 'custom-marker',
                html: '<div style="width: 12px; height: 12px; background: #4facfe; border: 2px solid #fff; border-radius: 50%; box-shadow: 0 0 10px rgba(79, 172, 254, 0.6);"></div>',
                iconSize: [16, 16],
                iconAnchor: [8, 8]
            })
        }).addTo(connectionMap);

        // Create popup content for all IPs at this location
        let popupContent = '<div style="font-family: Inter, sans-serif; min-width: 250px; max-width: 400px;">';

        if (group.ips.length > 1) {
            popupContent += '<div style="font-size: 0.85rem; color: #94a3b8; margin-bottom: 0.75rem;">' + group.ips.length + ' connections at this location</div>';
        }

        group.ips.forEach((item, index) => {
            if (index > 0) {
                popupContent += '<div style="border-top: 1px solid rgba(255,255,255,0.1); margin: 0.75rem 0;"></div>';
            }

            popupContent += '<div style="font-size: 1rem; font-weight: 700; color: #4facfe; margin-bottom: 0.5rem;">' + item.ip + '</div>';

            if (item.location.city || item.location.region) {
                popupContent += '<div style="font-size: 0.9rem; color: #cbd5e1; margin-bottom: 0.25rem;">';
                if (item.location.city) popupContent += item.location.city;
                if (item.location.city && item.location.region) popupContent += ', ';
                if (item.location.region) popupContent += item.location.region;
                popupContent += '</div>';
            }

            if (item.location.country) {
                popupContent += '<div style="font-size: 0.9rem; color: #cbd5e1; margin-bottom: 0.5rem;">' + item.location.country + '</div>';
            }

            if (item.location.isp) {
                popupContent += '<div style="font-size: 0.8rem; color: #94a3b8;">ISP: ' + item.location.isp + '</div>';
            }
        });

        popupContent += '</div>';

        // Bind popup with options to keep it open
        marker.bindPopup(popupContent, {
            closeButton: true,
            autoClose: false,
            closeOnClick: false,
            maxWidth: 400
        });

        connectionMarkers.push(marker);
    });

    // Auto-zoom to fit all markers
    if (bounds.length > 0) {
        connectionMap.fitBounds(bounds, {
            padding: [50, 50],
            maxZoom: 10
        });
    }

    console.log('Updated map with ' + ipCount + ' locations, auto-zoomed to fit');
}

// Initialize largest files with user's home directory
async function initializeLargestFiles() {
    try {
        const response = await fetch('/api/home-directory');
        const data = await response.json();

        if (data.home) {
            const directoryInput = document.getElementById('largest-files-directory');
            if (directoryInput) {
                directoryInput.value = data.home;
                directoryInput.placeholder = 'Enter directory path';
                // Don't automatically query - wait for user to click refresh
            }
        }
    } catch (error) {
        console.error('Error fetching home directory:', error);
        const directoryInput = document.getElementById('largest-files-directory');
        if (directoryInput) {
            directoryInput.placeholder = 'Enter directory path (e.g., /home/user)';
        }
    }
}

// Initialize on page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initializeLargestFiles);
} else {
    // DOM already loaded
    initializeLargestFiles();
}
