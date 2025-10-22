// Flow Builder JavaScript
const API_BASE_URL = window.location.origin + '/api';

// Flow state
let flowData = {
    nodes: [],
    connections: [],
    flowName: '',
    deviceId: '',
    niche: ''
};

let nodeIdCounter = 1;
let draggedElement = null;
let selectedNode = null;
let zoomScale = 1;

// Load devices for dropdown
async function loadDevices() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/devices`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        const deviceSelect = document.getElementById('deviceSelect');
        deviceSelect.innerHTML = '<option value="">Select a device...</option>';

        if (data.success && data.devices && data.devices.length > 0) {
            data.devices.forEach(device => {
                const option = document.createElement('option');
                option.value = device.id_device || device.device_id;
                option.textContent = `${device.id_device || device.device_id} - ${device.provider}`;
                deviceSelect.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Load devices error:', error);
    }
}

// Load flow for selected device
async function loadFlowForDevice() {
    const deviceId = document.getElementById('deviceSelect').value;

    if (!deviceId) {
        clearCanvas();
        return;
    }

    const token = localStorage.getItem('auth_token');
    if (!token) return;

    try {
        const response = await fetch(`${API_BASE_URL}/flows/${deviceId}`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (data.success && data.flow) {
            // Load existing flow
            document.getElementById('flowNameSelect').value = data.flow.flow_name || '';
            document.getElementById('nicheInput').value = data.flow.niche || '';

            // Load nodes (if stored as JSON in backend)
            if (data.flow.nodes_data) {
                loadFlowFromData(JSON.parse(data.flow.nodes_data));
            }
        } else {
            // No existing flow, clear canvas
            document.getElementById('flowNameSelect').value = '';
            document.getElementById('nicheInput').value = '';
        }
    } catch (error) {
        console.error('Load flow error:', error);
    }
}

// Initialize drag and drop
function initializeDragAndDrop() {
    const nodeItems = document.querySelectorAll('.node-item');
    const canvas = document.getElementById('flowCanvas');

    // Make node items draggable
    nodeItems.forEach(item => {
        item.addEventListener('dragstart', (e) => {
            // Get the node-item element (in case user drags from child)
            const nodeItem = e.target.closest('.node-item');
            if (nodeItem) {
                draggedElement = {
                    type: nodeItem.getAttribute('data-node-type'),
                    label: nodeItem.querySelector('.node-label').textContent,
                    icon: nodeItem.querySelector('.node-icon').textContent
                };
                console.log('Dragging:', draggedElement);
            }
        });

        item.addEventListener('dragend', () => {
            draggedElement = null;
        });
    });

    // Canvas drop handling
    canvas.addEventListener('dragover', (e) => {
        e.preventDefault();
    });

    canvas.addEventListener('drop', (e) => {
        e.preventDefault();

        if (draggedElement) {
            const canvasContainer = canvas.parentElement;
            const rect = canvas.getBoundingClientRect();
            const x = e.clientX - rect.left + canvasContainer.scrollLeft;
            const y = e.clientY - rect.top + canvasContainer.scrollTop;

            console.log('Dropping at:', x, y);
            createFlowNode(draggedElement.type, draggedElement.label, draggedElement.icon, x, y);
        }
    });
}

// Create flow node
function createFlowNode(type, label, icon, x, y) {
    const nodeId = `node-${nodeIdCounter++}`;

    const node = document.createElement('div');
    node.className = 'flow-node';
    node.setAttribute('data-node-id', nodeId);
    node.setAttribute('data-node-type', type);
    node.style.left = `${x}px`;
    node.style.top = `${y}px`;

    node.innerHTML = `
        <div class="node-header">
            <span class="node-icon">${icon}</span>
            <span class="node-title">${label}</span>
        </div>
        <div class="node-body">
            <p>Click edit to configure</p>
        </div>
        <div class="node-connector input-connector" data-connector-type="input" data-node-id="${nodeId}"></div>
        <div class="node-connector output-connector" data-connector-type="output" data-node-id="${nodeId}"></div>
        <div class="node-edit" onclick="openNodeConfig('${nodeId}')">✏️</div>
        <div class="node-delete" onclick="deleteNode('${nodeId}')">×</div>
    `;

    // Make node draggable within canvas
    makeNodeDraggable(node);

    // Add click event for selection
    node.addEventListener('click', (e) => {
        e.stopPropagation();
        selectNode(node);
    });

    document.getElementById('flowCanvas').appendChild(node);

    // Add to flow data
    flowData.nodes.push({
        id: nodeId,
        type: type,
        label: label,
        x: x,
        y: y,
        config: {}
    });
}

// Make node draggable
function makeNodeDraggable(node) {
    let isDragging = false;
    let currentX;
    let currentY;
    let initialX;
    let initialY;

    node.addEventListener('mousedown', (e) => {
        if (e.target.classList.contains('node-connector') ||
            e.target.classList.contains('node-delete') ||
            e.target.classList.contains('node-edit')) {
            return;
        }

        isDragging = true;
        initialX = e.clientX - node.offsetLeft;
        initialY = e.clientY - node.offsetTop;
    });

    document.addEventListener('mousemove', (e) => {
        if (!isDragging) return;

        e.preventDefault();
        currentX = e.clientX - initialX;
        currentY = e.clientY - initialY;

        node.style.left = `${currentX}px`;
        node.style.top = `${currentY}px`;

        // Update flow data
        const nodeData = flowData.nodes.find(n => n.id === node.getAttribute('data-node-id'));
        if (nodeData) {
            nodeData.x = currentX;
            nodeData.y = currentY;
        }

        // Redraw connections while dragging
        drawConnections();
    });

    document.addEventListener('mouseup', () => {
        if (isDragging) {
            isDragging = false;
            drawConnections(); // Final redraw when released
        }
    });
}

// Select node
function selectNode(node) {
    // Deselect all nodes
    document.querySelectorAll('.flow-node').forEach(n => {
        n.classList.remove('selected');
    });

    // Select clicked node
    node.classList.add('selected');
    selectedNode = node;
}

// Delete node
function deleteNode(nodeId) {
    const node = document.querySelector(`[data-node-id="${nodeId}"]`);
    if (node) {
        // Show confirmation
        Swal.fire({
            title: 'Delete Node?',
            text: 'This action cannot be undone!',
            icon: 'warning',
            showCancelButton: true,
            confirmButtonColor: '#e50914',
            cancelButtonColor: '#6c757d',
            confirmButtonText: 'Yes, delete it!',
            cancelButtonText: 'Cancel',
            background: '#141414',
            color: '#ffffff'
        }).then((result) => {
            if (result.isConfirmed) {
                node.remove();

                // Remove from flow data
                flowData.nodes = flowData.nodes.filter(n => n.id !== nodeId);
                flowData.connections = flowData.connections.filter(
                    c => c.from !== nodeId && c.to !== nodeId
                );

                Swal.fire({
                    title: 'Deleted!',
                    text: 'Node has been deleted',
                    icon: 'success',
                    background: '#141414',
                    color: '#ffffff',
                    confirmButtonColor: '#e50914',
                    timer: 2000
                });
            }
        });
    }
}

// Clear canvas
function clearCanvas() {
    Swal.fire({
        title: 'Clear Canvas?',
        text: 'This will remove all nodes except the Start node!',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#e50914',
        cancelButtonColor: '#6c757d',
        confirmButtonText: 'Yes, clear it!',
        cancelButtonText: 'Cancel',
        background: '#141414',
        color: '#ffffff'
    }).then((result) => {
        if (result.isConfirmed) {
            // Remove all nodes except start node
            const nodes = document.querySelectorAll('.flow-node:not(.start-node)');
            nodes.forEach(node => node.remove());

            // Reset flow data
            flowData.nodes = [];
            flowData.connections = [];
            nodeIdCounter = 1;

            Swal.fire({
                title: 'Cleared!',
                text: 'Canvas has been cleared',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914',
                timer: 2000
            });
        }
    });
}

// Save flow
async function saveFlow() {
    const deviceId = document.getElementById('deviceSelect').value;
    const flowName = document.getElementById('flowNameSelect').value;
    const niche = document.getElementById('nicheInput').value.trim();

    if (!deviceId) {
        Swal.fire({
            title: 'Error!',
            text: 'Please select a device',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
        return;
    }

    if (!flowName) {
        Swal.fire({
            title: 'Error!',
            text: 'Please select a flow type',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
        return;
    }

    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    // Prepare flow data
    const flowPayload = {
        id_device: deviceId,
        flow_name: flowName,
        niche: niche,
        nodes_data: JSON.stringify(flowData)
    };

    Swal.fire({
        title: 'Saving...',
        text: 'Please wait',
        allowOutsideClick: false,
        didOpen: () => {
            Swal.showLoading();
        },
        background: '#141414',
        color: '#ffffff'
    });

    try {
        const response = await fetch(`${API_BASE_URL}/flows`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(flowPayload)
        });

        const data = await response.json();

        if (data.success) {
            Swal.fire({
                title: 'Success!',
                text: 'Flow saved successfully',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });
        } else {
            Swal.fire({
                title: 'Error!',
                text: data.message || 'Failed to save flow',
                icon: 'error',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });
        }
    } catch (error) {
        console.error('Save flow error:', error);
        Swal.fire({
            title: 'Error!',
            text: 'Network error. Please check your connection.',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
    }
}

// Export flow
function exportFlow() {
    const deviceId = document.getElementById('deviceSelect').value;
    const flowName = document.getElementById('flowNameSelect').value;

    if (!flowName) {
        Swal.fire({
            title: 'Error!',
            text: 'Please select a flow type before exporting',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
        return;
    }

    const exportData = {
        flowName: flowName,
        deviceId: deviceId,
        niche: document.getElementById('nicheInput').value.trim(),
        flowData: flowData,
        exportedAt: new Date().toISOString()
    };

    const dataStr = JSON.stringify(exportData, null, 2);
    const dataBlob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `flow-${flowName.replace(/\s+/g, '-').toLowerCase()}.json`;
    link.click();
    URL.revokeObjectURL(url);

    Swal.fire({
        title: 'Exported!',
        text: 'Flow exported successfully',
        icon: 'success',
        background: '#141414',
        color: '#ffffff',
        confirmButtonColor: '#e50914',
        timer: 2000
    });
}

// Import flow
function importFlow() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = '.json';

    input.onchange = (e) => {
        const file = e.target.files[0];
        if (!file) return;

        const reader = new FileReader();
        reader.onload = (event) => {
            try {
                const importedData = JSON.parse(event.target.result);

                // Set form values
                if (importedData.flowName) {
                    document.getElementById('flowNameSelect').value = importedData.flowName;
                }
                if (importedData.deviceId) {
                    document.getElementById('deviceSelect').value = importedData.deviceId;
                }
                if (importedData.niche) {
                    document.getElementById('nicheInput').value = importedData.niche;
                }

                // Load flow data
                if (importedData.flowData) {
                    loadFlowFromData(importedData.flowData);
                }

                Swal.fire({
                    title: 'Imported!',
                    text: 'Flow imported successfully',
                    icon: 'success',
                    background: '#141414',
                    color: '#ffffff',
                    confirmButtonColor: '#e50914'
                });
            } catch (error) {
                Swal.fire({
                    title: 'Error!',
                    text: 'Invalid flow file',
                    icon: 'error',
                    background: '#141414',
                    color: '#ffffff',
                    confirmButtonColor: '#e50914'
                });
            }
        };
        reader.readAsText(file);
    };

    input.click();
}

// Load flow from data
function loadFlowFromData(data) {
    // Clear existing nodes (except start)
    const nodes = document.querySelectorAll('.flow-node:not(.start-node)');
    nodes.forEach(node => node.remove());

    flowData = data;
    nodeIdCounter = 1;

    // Recreate nodes
    data.nodes.forEach(nodeData => {
        const nodeElement = document.createElement('div');
        nodeElement.className = 'flow-node';
        nodeElement.setAttribute('data-node-id', nodeData.id);
        nodeElement.setAttribute('data-node-type', nodeData.type);
        nodeElement.style.left = `${nodeData.x}px`;
        nodeElement.style.top = `${nodeData.y}px`;

        nodeElement.innerHTML = `
            <div class="node-header">
                <span class="node-icon">${getNodeIcon(nodeData.type)}</span>
                <span class="node-title">${nodeData.label}</span>
            </div>
            <div class="node-body">
                <p>Configure ${nodeData.label.toLowerCase()} settings</p>
            </div>
            <div class="node-connector input-connector" data-connector-type="input"></div>
            <div class="node-connector output-connector" data-connector-type="output"></div>
            <div class="node-delete" onclick="deleteNode('${nodeData.id}')">×</div>
        `;

        makeNodeDraggable(nodeElement);
        nodeElement.addEventListener('click', (e) => {
            e.stopPropagation();
            selectNode(nodeElement);
        });

        document.getElementById('flowCanvas').appendChild(nodeElement);

        // Update counter
        const idNum = parseInt(nodeData.id.split('-')[1]);
        if (idNum >= nodeIdCounter) {
            nodeIdCounter = idNum + 1;
        }
    });
}

// Get node icon by type
function getNodeIcon(type) {
    const icons = {
        'send_message': '💬',
        'waiting_reply': '💭',
        'waiting_times': '⏳',
        'ai_prompt': '✨',
        'stage': '🎯',
        'send_image': '🖼️',
        'send_audio': '🔊',
        'send_video': '🎥',
        'delay': '⏱️',
        'conditions': '🔀'
    };
    return icons[type] || '📦';
}

// Zoom In
function zoomIn() {
    if (zoomScale < 2) {
        zoomScale += 0.1;
        applyZoom();
    }
}

// Zoom Out
function zoomOut() {
    if (zoomScale > 0.5) {
        zoomScale -= 0.1;
        applyZoom();
    }
}

// Reset Zoom
function resetZoom() {
    zoomScale = 1;
    applyZoom();
}

// Apply Zoom
function applyZoom() {
    const canvas = document.getElementById('flowCanvas');
    canvas.style.transform = `scale(${zoomScale})`;
    canvas.style.transformOrigin = '0 0';

    // Update zoom level display
    document.getElementById('zoomLevel').textContent = `${Math.round(zoomScale * 100)}%`;
}

// Node Configuration Functions
let currentConfigNodeId = null;
let connectionStart = null;

// Open node configuration modal
function openNodeConfig(nodeId) {
    currentConfigNodeId = nodeId;
    const nodeData = flowData.nodes.find(n => n.id === nodeId);

    if (!nodeData) return;

    const modal = document.getElementById('nodeConfigModal');
    const title = document.getElementById('nodeConfigTitle');
    const fieldsContainer = document.getElementById('nodeConfigFields');

    title.textContent = `Configure ${nodeData.label}`;
    fieldsContainer.innerHTML = getConfigFieldsForType(nodeData.type, nodeData.config);

    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
}

// Get configuration fields based on node type
function getConfigFieldsForType(type, config = {}) {
    switch(type) {
        case 'send_message':
        case 'ai_prompt':
            return `
                <div class="form-group">
                    <label>${type === 'send_message' ? 'Message Content' : 'AI Prompt'} *</label>
                    <textarea id="nodeConfigText" rows="6" required style="width: 100%; padding: 0.9rem; background: rgba(51, 51, 51, 0.7); border: 2px solid rgba(255, 255, 255, 0.1); border-radius: 8px; color: white; font-family: inherit;" placeholder="Enter ${type === 'send_message' ? 'message' : 'prompt'}...">${config.text || ''}</textarea>
                </div>
            `;

        case 'stage':
            return `
                <div class="form-group">
                    <label>Stage Name *</label>
                    <input type="text" id="nodeConfigValue" required value="${config.value || ''}" style="width: 100%; padding: 0.9rem; background: rgba(51, 51, 51, 0.7); border: 2px solid rgba(255, 255, 255, 0.1); border-radius: 8px; color: white;" placeholder="Enter stage name...">
                </div>
            `;

        case 'send_image':
        case 'send_audio':
        case 'send_video':
            const mediaType = type.replace('send_', '');
            return `
                <div class="form-group">
                    <label>${mediaType.charAt(0).toUpperCase() + mediaType.slice(1)} URL *</label>
                    <input type="url" id="nodeConfigUrl" required value="${config.url || ''}" style="width: 100%; padding: 0.9rem; background: rgba(51, 51, 51, 0.7); border: 2px solid rgba(255, 255, 255, 0.1); border-radius: 8px; color: white;" placeholder="Enter ${mediaType} URL...">
                </div>
            `;

        case 'delay':
        case 'waiting_times':
            return `
                <div class="form-group">
                    <label>${type === 'delay' ? 'Delay' : 'Waiting Time'} (seconds) *</label>
                    <input type="number" id="nodeConfigDelay" required min="0" step="1" value="${config.delay || ''}" style="width: 100%; padding: 0.9rem; background: rgba(51, 51, 51, 0.7); border: 2px solid rgba(255, 255, 255, 0.1); border-radius: 8px; color: white;" placeholder="Enter time in seconds...">
                </div>
            `;

        case 'waiting_reply':
            return `
                <div class="form-group">
                    <label>Timeout (seconds) *</label>
                    <input type="number" id="nodeConfigTimeout" required min="0" step="1" value="${config.timeout || 60}" style="width: 100%; padding: 0.9rem; background: rgba(51, 51, 51, 0.7); border: 2px solid rgba(255, 255, 255, 0.1); border-radius: 8px; color: white;" placeholder="Enter timeout...">
                </div>
            `;

        case 'conditions':
            return getConditionsConfig(config.conditions || []);

        default:
            return '<p>No configuration needed</p>';
    }
}

// Get conditions configuration HTML
function getConditionsConfig(conditions) {
    let html = '<div id="conditionsContainer">';

    conditions.forEach((cond, index) => {
        html += getConditionItemHTML(cond, index);
    });

    html += '</div>';
    html += '<button type="button" class="add-condition-btn" onclick="addCondition()">+ Add Condition</button>';

    return html;
}

// Get single condition item HTML
function getConditionItemHTML(cond = {}, index) {
    return `
        <div class="condition-item" data-condition-index="${index}">
            <div class="condition-item-header">
                <span class="condition-label">Condition ${index + 1}</span>
                <button type="button" class="remove-condition-btn" onclick="removeCondition(${index})">Remove</button>
            </div>
            <div class="form-group">
                <label>Type</label>
                <select class="condition-type" style="width: 100%; padding: 0.9rem; background: rgba(51, 51, 51, 0.7); border: 2px solid rgba(255, 255, 255, 0.1); border-radius: 8px; color: white;">
                    <option value="contains" ${cond.type === 'contains' ? 'selected' : ''}>Contains</option>
                    <option value="match" ${cond.type === 'match' ? 'selected' : ''}>Match</option>
                    <option value="equal" ${cond.type === 'equal' ? 'selected' : ''}>Equal</option>
                    <option value="default" ${cond.type === 'default' ? 'selected' : ''}>Default</option>
                </select>
            </div>
            <div class="form-group">
                <label>Value</label>
                <input type="text" class="condition-value" value="${cond.value || ''}" style="width: 100%; padding: 0.9rem; background: rgba(51, 51, 51, 0.7); border: 2px solid rgba(255, 255, 255, 0.1); border-radius: 8px; color: white;" placeholder="Enter condition value...">
            </div>
        </div>
    `;
}

// Add new condition
function addCondition() {
    const container = document.getElementById('conditionsContainer');
    const index = container.children.length;
    container.insertAdjacentHTML('beforeend', getConditionItemHTML({}, index));
}

// Remove condition
function removeCondition(index) {
    const item = document.querySelector(`[data-condition-index="${index}"]`);
    if (item) item.remove();

    // Re-index remaining conditions
    document.querySelectorAll('.condition-item').forEach((item, newIndex) => {
        item.setAttribute('data-condition-index', newIndex);
        item.querySelector('.condition-label').textContent = `Condition ${newIndex + 1}`;
        item.querySelector('.remove-condition-btn').setAttribute('onclick', `removeCondition(${newIndex})`);
    });
}

// Save node configuration
function saveNodeConfig(event) {
    event.preventDefault();

    const nodeData = flowData.nodes.find(n => n.id === currentConfigNodeId);
    if (!nodeData) return;

    const type = nodeData.type;
    let config = {};

    // Get configuration based on node type
    switch(type) {
        case 'send_message':
        case 'ai_prompt':
            config.text = document.getElementById('nodeConfigText').value;
            break;

        case 'stage':
            config.value = document.getElementById('nodeConfigValue').value;
            break;

        case 'send_image':
        case 'send_audio':
        case 'send_video':
            config.url = document.getElementById('nodeConfigUrl').value;
            break;

        case 'delay':
        case 'waiting_times':
            config.delay = parseInt(document.getElementById('nodeConfigDelay').value);
            break;

        case 'waiting_reply':
            config.timeout = parseInt(document.getElementById('nodeConfigTimeout').value);
            break;

        case 'conditions':
            const conditions = [];
            document.querySelectorAll('.condition-item').forEach(item => {
                conditions.push({
                    type: item.querySelector('.condition-type').value,
                    value: item.querySelector('.condition-value').value
                });
            });
            config.conditions = conditions;
            break;
    }

    // Update node config
    nodeData.config = config;

    // Update node body to show configured state
    const nodeElement = document.querySelector(`[data-node-id="${currentConfigNodeId}"]`);
    if (nodeElement) {
        const nodeBody = nodeElement.querySelector('.node-body');
        nodeBody.innerHTML = '<p style="color: #4CAF50;">✓ Configured</p>';
    }

    closeNodeConfigModal();

    Swal.fire({
        title: 'Saved!',
        text: 'Node configuration saved',
        icon: 'success',
        background: '#141414',
        color: '#ffffff',
        confirmButtonColor: '#e50914',
        timer: 1500
    });
}

// Close node configuration modal
function closeNodeConfigModal() {
    const modal = document.getElementById('nodeConfigModal');
    modal.classList.remove('active');
    document.body.style.overflow = 'auto';
    currentConfigNodeId = null;
}

// Edge Connection Functions
function initializeConnectors() {
    document.addEventListener('click', (e) => {
        if (e.target.classList.contains('node-connector')) {
            handleConnectorClick(e.target);
        }
    });
}

function handleConnectorClick(connector) {
    const nodeId = connector.getAttribute('data-node-id');
    const connectorType = connector.getAttribute('data-connector-type');

    if (!connectionStart) {
        // Start connection from output connector
        if (connectorType === 'output') {
            connectionStart = { nodeId, connector };
            connector.style.background = 'var(--netflix-red)';
            connector.style.transform = 'translateX(-50%) scale(1.5)';
        }
    } else {
        // Complete connection to input connector
        if (connectorType === 'input' && nodeId !== connectionStart.nodeId) {
            createConnection(connectionStart.nodeId, nodeId);

            // Reset start connector style
            connectionStart.connector.style.background = '';
            connectionStart.connector.style.transform = '';
            connectionStart = null;
        } else {
            // Cancel connection
            connectionStart.connector.style.background = '';
            connectionStart.connector.style.transform = '';
            connectionStart = null;
        }
    }
}

function createConnection(fromNodeId, toNodeId) {
    // Check if connection already exists
    const exists = flowData.connections.find(c => c.from === fromNodeId && c.to === toNodeId);
    if (exists) return;

    // Add connection to data
    flowData.connections.push({ from: fromNodeId, to: toNodeId });

    // Redraw all connections
    drawConnections();
}

function drawConnections() {
    const svg = document.getElementById('connectionLayer');
    svg.innerHTML = ''; // Clear existing lines

    flowData.connections.forEach(conn => {
        const fromNode = document.querySelector(`[data-node-id="${conn.from}"]`);
        const toNode = document.querySelector(`[data-node-id="${conn.to}"]`);

        if (!fromNode || !toNode) return;

        const fromConnector = fromNode.querySelector('.output-connector');
        const toConnector = toNode.querySelector('.input-connector');

        // Get positions
        const fromRect = fromConnector.getBoundingClientRect();
        const toRect = toConnector.getBoundingClientRect();
        const canvasRect = document.getElementById('flowCanvas').getBoundingClientRect();

        const x1 = fromRect.left - canvasRect.left + fromRect.width / 2;
        const y1 = fromRect.top - canvasRect.top + fromRect.height / 2;
        const x2 = toRect.left - canvasRect.left + toRect.width / 2;
        const y2 = toRect.top - canvasRect.top + toRect.height / 2;

        // Create curved path
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        const midY = (y1 + y2) / 2;
        const d = `M ${x1} ${y1} C ${x1} ${midY}, ${x2} ${midY}, ${x2} ${y2}`;

        path.setAttribute('d', d);
        path.setAttribute('class', 'connection-line');
        path.setAttribute('data-from', conn.from);
        path.setAttribute('data-to', conn.to);

        svg.appendChild(path);
    });
}

// Logout function
function logout() {
    localStorage.removeItem('auth_token');
    window.location.href = '/';
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    loadDevices();
    initializeDragAndDrop();
    initializeConnectors();
});
