// API Configuration
const API_BASE_URL = 'http://localhost:8080/api';
let autoScroll = true;
let consoleWebSocket = null;

// Initialize dashboard
document.addEventListener('DOMContentLoaded', function() {
    initializeDashboard();
    loadServerInfo();
    startAutoRefresh();
});

// Dashboard initialization
function initializeDashboard() {
    // Tab navigation
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const targetId = this.getAttribute('href').substring(1);
            
            // Update active tab
            document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
            
            this.classList.add('active');
            document.getElementById(targetId).classList.add('active');
            
            // Load tab-specific data
            switch(targetId) {
                case 'players':
                    loadPlayers();
                    break;
                case 'world':
                    loadWorldInfo();
                    break;
                case 'console':
                    connectConsole();
                    break;
                case 'plugins':
                    loadPlugins();
                    break;
            }
        });
    });
    
    // Console input handler
    document.getElementById('console-input').addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            sendCommand();
        }
    });
    
    // Player search
    document.getElementById('player-search').addEventListener('input', function() {
        filterPlayers(this.value);
    });
    
    // Modal close
    document.querySelector('.close').addEventListener('click', function() {
        document.getElementById('player-modal').style.display = 'none';
    });
}

// API Helper Functions
async function apiCall(endpoint, options = {}) {
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': 'garuda-admin-key' // In production, use proper auth
            },
            ...options
        });
        
        const data = await response.json();
        return data;
    } catch (error) {
        console.error('API call failed:', error);
        showNotification('Failed to connect to server', 'error');
        return { success: false, error: 'Connection failed' };
    }
}

// Server Information
async function loadServerInfo() {
    const data = await apiCall('/server/info');
    if (data.success) {
        updateServerStats(data.data);
    }
}

async function loadServerStats() {
    const data = await apiCall('/server/stats');
    if (data.success) {
        updatePerformanceStats(data.data);
    }
}

function updateServerStats(stats) {
    document.getElementById('online-players').textContent = stats.players_online;
    document.getElementById('server-uptime').textContent = stats.uptime;
    document.getElementById('max-players').textContent = stats.max_players;
}

function updatePerformanceStats(stats) {
    document.getElementById('memory-usage').textContent = 
        Math.round(stats.memory_allocated / 1024 / 1024) + 'MB';
    document.getElementById('tps').textContent = '20'; // Placeholder
}

// Player Management
async function loadPlayers() {
    const data = await apiCall('/players');
    if (data.success) {
        updatePlayersTable(data.data.players);
        updatePlayersList(data.data.players);
    }
}

function updatePlayersTable(players) {
    const tbody = document.getElementById('players-table-body');
    tbody.innerHTML = '';
    
    players.forEach(player => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>
                <div class="player-info">
                    <div class="player-avatar">
                        ${player.username.charAt(0).toUpperCase()}
                    </div>
                    <div>
                        <div>${player.username}</div>
                        <small>ID: ${player.entity_id}</small>
                    </div>
                </div>
            </td>
            <td>
                <span class="status-indicator online"></span>
                Online
            </td>
            <td>${Math.round(player.health)}/${Math.round(player.max_health)}</td>
            <td>${getGameModeName(player.game_mode)}</td>
            <td>${formatPosition(player.position)}</td>
            <td>
                <button class="btn btn-secondary btn-sm" onclick="showPlayerModal('${player.username}')">
                    <i class="fas fa-cog"></i>
                </button>
                <button class="btn btn-warning btn-sm" onclick="kickPlayer('${player.username}')">
                    <i class="fas fa-sign-out-alt"></i>
                </button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

function updatePlayersList(players) {
    const container = document.getElementById('players-list');
    container.innerHTML = '';
    
    players.slice(0, 5).forEach(player => {
        const item = document.createElement('div');
        item.className = 'player-item';
        item.innerHTML = `
            <div class="player-info">
                <div class="player-avatar">
                    ${player.username.charAt(0).toUpperCase()}
                </div>
                <div class="player-details">
                    <h4>${player.username}</h4>
                    <p>Health: ${Math.round(player.health)} | Mode: ${getGameModeName(player.game_mode)}</p>
                </div>
            </div>
            <div class="player-position">
                ${formatPosition(player.position)}
            </div>
        `;
        container.appendChild(item);
    });
}

function filterPlayers(searchTerm) {
    const rows = document.querySelectorAll('#players-table-body tr');
    rows.forEach(row => {
        const username = row.cells[0].textContent.toLowerCase();
        if (username.includes(searchTerm.toLowerCase())) {
            row.style.display = '';
        } else {
            row.style.display = 'none';
        }
    });
}

// World Management
async function loadWorldInfo() {
    const data = await apiCall('/world/info');
    if (data.success) {
        document.getElementById('world-name').textContent = data.data.name;
        document.getElementById('world-seed').textContent = data.data.seed;
        document.getElementById('world-time').textContent = formatTime(data.data.time);
    }
    
    const weatherData = await apiCall('/world/weather');
    if (weatherData.success) {
        document.getElementById('world-weather').textContent = 
            weatherData.data.weather.charAt(0).toUpperCase() + weatherData.data.weather.slice(1);
    }
}

async function setWorldTime() {
    const timeSelect = document.getElementById('time-set');
    const timeValue = timeSelect.value;
    
    const data = await apiCall('/world/time', {
        method: 'POST',
        body: JSON.stringify({ time: parseInt(timeValue) })
    });
    
    if (data.success) {
        showNotification(data.message, 'success');
        loadWorldInfo();
    } else {
        showNotification(data.error, 'error');
    }
}

async function setWorldWeather() {
    const weatherSelect = document.getElementById('weather-set');
    const weatherValue = weatherSelect.value;
    
    const data = await apiCall('/world/weather', {
        method: 'POST',
        body: JSON.stringify({ 
            weather: weatherValue,
            duration: 6000 // 5 minutes
        })
    });
    
    if (data.success) {
        showNotification(data.message, 'success');
        loadWorldInfo();
    } else {
        showNotification(data.error, 'error');
    }
}

// Console Management
function connectConsole() {
    // WebSocket connection for real-time console
    consoleWebSocket = new WebSocket(`ws://localhost:8080/api/console/stream`);
    
    consoleWebSocket.onmessage = function(event) {
        const data = JSON.parse(event.data);
        addConsoleMessage(data.message, data.type);
    };
    
    consoleWebSocket.onclose = function() {
        addConsoleMessage('Console connection closed', 'system');
    };
}

function addConsoleMessage(message, type = 'log') {
    const consoleOutput = document.getElementById('console-output');
    const messageElement = document.createElement('div');
    
    const timestamp = new Date().toLocaleTimeString();
    messageElement.innerHTML = `
        <span class="console-timestamp">[${timestamp}]</span>
        <span class="console-type">[${type.toUpperCase()}]</span>
        <span class="console-message">${escapeHtml(message)}</span>
    `;
    
    consoleOutput.appendChild(messageElement);
    
    if (autoScroll) {
        consoleOutput.scrollTop = consoleOutput.scrollHeight;
    }
}

function sendCommand() {
    const input = document.getElementById('console-input');
    const command = input.value.trim();
    
    if (command) {
        addConsoleMessage(`> ${command}`, 'command');
        
        apiCall('/command', {
            method: 'POST',
            body: JSON.stringify({ command: command })
        }).then(data => {
            if (data.success) {
                addConsoleMessage(data.message, 'success');
            } else {
                addConsoleMessage(data.error, 'error');
            }
        });
        
        input.value = '';
    }
}

function clearConsole() {
    document.getElementById('console-output').innerHTML = '';
}

function toggleAutoScroll() {
    autoScroll = !autoScroll;
    const button = event.target;
    button.innerHTML = autoScroll ? 
        '<i class="fas fa-arrow-down"></i> Auto Scroll' : 
        '<i class="fas fa-pause"></i> Manual Scroll';
}

// Server Control
async function saveWorld() {
    const data = await apiCall('/server/save', { method: 'POST' });
    if (data.success) {
        showNotification('World saved successfully', 'success');
    } else {
        showNotification('Failed to save world', 'error');
    }
}

async function stopServer() {
    if (confirm('Are you sure you want to stop the server?')) {
        const data = await apiCall('/server/stop', { method: 'POST' });
        if (data.success) {
            showNotification('Server stopping...', 'warning');
        }
    }
}

async function restartServer() {
    if (confirm('Are you sure you want to restart the server?')) {
        showNotification('Restart functionality coming soon...', 'info');
    }
}

// Player Actions
async function kickPlayer(username) {
    const reason = prompt(`Enter kick reason for ${username}:`, 'Kicked by administrator');
    if (reason !== null) {
        const data = await apiCall(`/players/${username}/kick`, {
            method: 'POST',
            body: JSON.stringify({ reason: reason })
        });
        
        if (data.success) {
            showNotification(data.message, 'success');
            loadPlayers();
        } else {
            showNotification(data.error, 'error');
        }
    }
}

async function banPlayer(username) {
    const reason = prompt(`Enter ban reason for ${username}:`, 'Banned by administrator');
    if (reason !== null) {
        const data = await apiCall(`/players/${username}/ban`, {
            method: 'POST',
            body: JSON.stringify({ reason: reason })
        });
        
        if (data.success) {
            showNotification(data.message, 'success');
            loadPlayers();
        } else {
            showNotification(data.error, 'error');
        }
    }
}

async function showPlayerModal(username) {
    const data = await apiCall(`/players/${username}`);
    if (data.success) {
        const player = data.data;
        const modalContent = document.getElementById('player-modal-content');
        
        modalContent.innerHTML = `
            <div class="player-modal-info">
                <div class="info-item">
                    <label>Username:</label>
                    <span>${player.username}</span>
                </div>
                <div class="info-item">
                    <label>Health:</label>
                    <span>${player.health}/${player.max_health}</span>
                </div>
                <div class="info-item">
                    <label>Position:</label>
                    <span>${formatPosition(player.position)}</span>
                </div>
                <div class="info-item">
                    <label>Game Mode:</label>
                    <span>${getGameModeName(player.game_mode)}</span>
                </div>
                <div class="info-item">
                    <label>Operator:</label>
                    <span>${player.op ? 'Yes' : 'No'}</span>
                </div>
            </div>
            <div class="player-modal-actions">
                <button class="btn btn-warning" onclick="kickPlayer('${username}')">
                    <i class="fas fa-sign-out-alt"></i> Kick
                </button>
                <button class="btn btn-danger" onclick="banPlayer('${username}')">
                    <i class="fas fa-ban"></i> Ban
                </button>
                ${player.op ? `
                <button class="btn btn-secondary" onclick="deopPlayer('${username}')">
                    <i class="fas fa-user-times"></i> Deop
                </button>
                ` : `
                <button class="btn btn-primary" onclick="opPlayer('${username}')">
                                        <i class="fas fa-user-shield"></i> Op
                </button>
                `}
            </div>
        `;
        
        document.getElementById('player-modal').style.display = 'block';
    }
}

// Plugin Management
async function loadPlugins() {
    // Placeholder - implement based on your plugin API
    const pluginsList = document.getElementById('plugins-list');
    pluginsList.innerHTML = `
        <div class="plugin-card">
            <div class="plugin-header">
                <h4>WorldEdit</h4>
                <span class="plugin-version">v7.2.12</span>
            </div>
            <div class="plugin-description">
                Powerful in-game map editing tool
            </div>
            <div class="plugin-status enabled">
                <i class="fas fa-check-circle"></i>
                Enabled
            </div>
        </div>
        <div class="plugin-card">
            <div class="plugin-header">
                <h4>EssentialsX</h4>
                <span class="plugin-version">v2.19.7</span>
            </div>
            <div class="plugin-description">
                Essential commands and features
            </div>
            <div class="plugin-status enabled">
                <i class="fas fa-check-circle"></i>
                Enabled
            </div>
        </div>
    `;
}

// Settings Management
async function saveSettings() {
    const settings = {
        motd: document.getElementById('motd-setting').value,
        maxPlayers: document.getElementById('max-players-setting').value,
        viewDistance: document.getElementById('view-distance-setting').value
    };
    
    showNotification('Settings saved successfully', 'success');
}

// Utility Functions
function getGameModeName(mode) {
    const modes = {
        0: 'Survival',
        1: 'Creative',
        2: 'Adventure',
        3: 'Spectator'
    };
    return modes[mode] || 'Unknown';
}

function formatPosition(position) {
    if (!position) return 'N/A';
    return `X:${Math.round(position.x)} Y:${Math.round(position.y)} Z:${Math.round(position.z)}`;
}

function formatTime(time) {
    // Convert Minecraft time to readable format
    const hours = Math.floor((time / 1000 + 6) % 24);
    const minutes = Math.floor((time % 1000) / 1000 * 60);
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}`;
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <div class="notification-content">
            <i class="fas fa-${getNotificationIcon(type)}"></i>
            <span>${message}</span>
        </div>
    `;
    
    // Add to page
    document.body.appendChild(notification);
    
    // Animate in
    setTimeout(() => notification.classList.add('show'), 100);
    
    // Remove after delay
    setTimeout(() => {
        notification.classList.remove('show');
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

function getNotificationIcon(type) {
    const icons = {
        success: 'check-circle',
        error: 'exclamation-circle',
        warning: 'exclamation-triangle',
        info: 'info-circle'
    };
    return icons[type] || 'info-circle';
}

// Auto-refresh system
function startAutoRefresh() {
    // Refresh server info every 5 seconds
    setInterval(() => {
        loadServerInfo();
        loadServerStats();
        
        // Only refresh players if on players tab
        if (document.getElementById('players').classList.contains('active')) {
            loadPlayers();
        }
        
        // Only refresh world info if on world tab
        if (document.getElementById('world').classList.contains('active')) {
            loadWorldInfo();
        }
    }, 5000);
}

// Health check monitoring
async function checkServerHealth() {
    try {
        const response = await fetch(`${API_BASE_URL}/health`);
        const statusIndicator = document.querySelector('.status-indicator');
        const statusText = document.querySelector('.server-status span:last-child');
        
        if (response.ok) {
            statusIndicator.className = 'status-indicator online';
            statusText.textContent = 'Online';
        } else {
            statusIndicator.className = 'status-indicator offline';
            statusText.textContent = 'Offline';
        }
    } catch (error) {
        const statusIndicator = document.querySelector('.status-indicator');
        const statusText = document.querySelector('.server-status span:last-child');
        statusIndicator.className = 'status-indicator offline';
        statusText.textContent = 'Offline';
    }
}

// Command shortcuts for common operations
async function opPlayer(username) {
    const data = await apiCall('/command', {
        method: 'POST',
        body: JSON.stringify({ command: `op ${username}` })
    });
    
    if (data.success) {
        showNotification(`Made ${username} operator`, 'success');
        document.getElementById('player-modal').style.display = 'none';
    } else {
        showNotification(data.error, 'error');
    }
}

async function deopPlayer(username) {
    const data = await apiCall('/command', {
        method: 'POST',
        body: JSON.stringify({ command: `deop ${username}` })
    });
    
    if (data.success) {
        showNotification(`Removed operator from ${username}`, 'success');
        document.getElementById('player-modal').style.display = 'none';
    } else {
        showNotification(data.error, 'error');
    }
}

// Export functionality for backup
async function exportServerData() {
    const data = {
        players: await apiCall('/players'),
        world: await apiCall('/world/info'),
        server: await apiCall('/server/info')
    };
    
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `garuda-server-backup-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

// Keyboard shortcuts
document.addEventListener('keydown', function(e) {
    // Ctrl+K to focus command input
    if (e.ctrlKey && e.key === 'k') {
        e.preventDefault();
        const consoleInput = document.getElementById('console-input');
        if (consoleInput) {
            consoleInput.focus();
        }
    }
    
    // Ctrl+S to save world
    if (e.ctrlKey && e.key === 's') {
        e.preventDefault();
        saveWorld();
    }
});

// Initialize health check
setInterval(checkServerHealth, 10000);
checkServerHealth();

// Error handling for WebSocket
window.addEventListener('beforeunload', function() {
    if (consoleWebSocket) {
        consoleWebSocket.close();
    }
});