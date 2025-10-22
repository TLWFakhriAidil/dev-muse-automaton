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
let isPanning = false;
let panStart = { x: 0, y: 0 };
let canvasOffset = { x: 0, y: 0 };

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

    // Set default body text based on node type
    let bodyText = 'Click edit to configure';
    if (type === 'waiting_reply') {
        bodyText = '<p style="color: #4CAF50;">✓ Ready (No config needed)</p>';
    }

    node.innerHTML = `
        <div class="node-header">
            <span class="node-icon">${icon}</span>
            <span class="node-title">${label}</span>
        </div>
        <div class="node-body">
            ${bodyText}
        </div>
        <div class="node-connector input-connector" data-connector-type="input" data-node-id="${nodeId}"></div>
        <div class="node-connector output-connector" data-connector-type="output" data-node-id="${nodeId}"></div>
        <div class="node-edit" data-node-id="${nodeId}">✏️</div>
        <div class="node-delete" data-node-id="${nodeId}">×</div>
    `;

    // Make node draggable within canvas
    makeNodeDraggable(node);

    // Add click event for selection
    node.addEventListener('click', (e) => {
        e.stopPropagation();
        selectNode(node);
    });

    // Add edit button click event
    const editBtn = node.querySelector('.node-edit');
    if (editBtn) {
        editBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            e.preventDefault();
            console.log('Edit button clicked for node:', nodeId);
            openNodeConfig(nodeId);
        });
    } else {
        console.error('Edit button not found for node:', nodeId);
    }

    // Add delete button click event
    const deleteBtn = node.querySelector('.node-delete');
    deleteBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        deleteNode(nodeId);
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

    // Initialize connectors for this new node
    const connectors = node.querySelectorAll('.node-connector');
    console.log('Initializing connectors for new node:', nodeId, 'Count:', connectors.length);
    connectors.forEach(connector => {
        connector.addEventListener('click', (e) => {
            e.stopPropagation();
            e.preventDefault();
            console.log('New node connector clicked');
            handleConnectorClick(connector);
        }, true);
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
    console.log('openNodeConfig called with nodeId:', nodeId);
    currentConfigNodeId = nodeId;
    const nodeData = flowData.nodes.find(n => n.id === nodeId);

    console.log('Found node data:', nodeData);
    console.log('All nodes:', flowData.nodes);

    if (!nodeData) {
        console.error('Node data not found for:', nodeId);
        return;
    }

    const modal = document.getElementById('nodeConfigModal');
    const title = document.getElementById('nodeConfigTitle');
    const fieldsContainer = document.getElementById('nodeConfigFields');

    console.log('Modal element:', modal);
    console.log('Title element:', title);
    console.log('Fields container:', fieldsContainer);

    title.textContent = `Configure ${nodeData.label}`;
    fieldsContainer.innerHTML = getConfigFieldsForType(nodeData.type, nodeData.config);

    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
    console.log('Modal should now be visible with class:', modal.className);
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
                    <p style="color: rgba(255, 255, 255, 0.7); text-align: center; padding: 2rem;">
                        This node waits for user reply without timeout.<br>
                        No configuration needed.
                    </p>
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
            // No configuration needed - just waits for user reply
            config.note = 'Waits for user reply without timeout';
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
    // Make sure start node connectors are initialized
    const startConnectors = document.querySelectorAll('.start-node .node-connector');
    console.log('Initializing start node connectors:', startConnectors.length);
    startConnectors.forEach(connector => {
        connector.setAttribute('data-node-id', 'start');
        console.log('Start connector initialized:', connector);
    });

    // Global click handler for connectors using event delegation
    document.addEventListener('click', (e) => {
        // Check if the clicked element is a connector
        if (e.target.classList.contains('node-connector')) {
            e.stopPropagation();
            e.preventDefault();
            console.log('Connector click detected via event delegation');
            handleConnectorClick(e.target);
        }
    }, true); // Use capture phase to catch events before they bubble

    // Also add direct click handlers to all existing connectors
    const allConnectors = document.querySelectorAll('.node-connector');
    console.log('Adding direct handlers to', allConnectors.length, 'connectors');
    allConnectors.forEach(connector => {
        connector.addEventListener('click', (e) => {
            e.stopPropagation();
            e.preventDefault();
            console.log('Connector click detected via direct handler');
            handleConnectorClick(connector);
        }, true);
    });
}

function handleConnectorClick(connector) {
    const nodeId = connector.getAttribute('data-node-id');
    const connectorType = connector.getAttribute('data-connector-type');

    console.log('=== Connector Clicked ===');
    console.log('Node ID:', nodeId);
    console.log('Connector Type:', connectorType);
    console.log('Connection Start:', connectionStart);

    if (!connectionStart) {
        // Start connection from output connector
        if (connectorType === 'output') {
            connectionStart = { nodeId, connector };
            connector.style.background = '#e50914';
            connector.style.boxShadow = '0 0 10px #e50914';
            connector.style.transform = 'translateX(-50%) scale(1.5)';
            connector.style.zIndex = '100';

            console.log('✓ Connection started from node:', nodeId);
        } else {
            console.log('✗ Cannot start connection from input connector');
        }
    } else {
        // Complete connection to input connector
        if (connectorType === 'input' && nodeId !== connectionStart.nodeId) {
            console.log('✓ Creating connection:', connectionStart.nodeId, '->', nodeId);
            createConnection(connectionStart.nodeId, nodeId);

            // Reset start connector style
            connectionStart.connector.style.background = '';
            connectionStart.connector.style.boxShadow = '';
            connectionStart.connector.style.transform = '';
            connectionStart.connector.style.zIndex = '';

            connectionStart = null;
            console.log('✓ Connection completed and reset');
        } else {
            // Cancel connection
            console.log('✗ Connection cancelled (same node or wrong connector type)');
            connectionStart.connector.style.background = '';
            connectionStart.connector.style.boxShadow = '';
            connectionStart.connector.style.transform = '';
            connectionStart.connector.style.zIndex = '';

            connectionStart = null;
        }
    }
}

function createConnection(fromNodeId, toNodeId) {
    // Check if connection already exists
    const exists = flowData.connections.find(c => c.from === fromNodeId && c.to === toNodeId);
    if (exists) {
        console.log('Connection already exists:', fromNodeId, '->', toNodeId);
        return;
    }

    // Add connection to data
    flowData.connections.push({ from: fromNodeId, to: toNodeId });
    console.log('✓ Connection added to flowData:', fromNodeId, '->', toNodeId);
    console.log('Total connections:', flowData.connections.length);

    // Redraw all connections
    drawConnections();
}

function deleteConnection(fromNodeId, toNodeId) {
    console.log('🗑️ deleteConnection called:', fromNodeId, '->', toNodeId);
    console.log('Current connections:', flowData.connections);

    Swal.fire({
        title: 'Delete Connection?',
        text: `Remove connection from ${fromNodeId} to ${toNodeId}?`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#e50914',
        cancelButtonColor: '#6c757d',
        confirmButtonText: 'Yes, delete it!',
        cancelButtonText: 'Cancel',
        background: '#141414',
        color: '#ffffff'
    }).then((result) => {
        console.log('Dialog result:', result);
        if (result.isConfirmed) {
            console.log('User confirmed deletion');
            // Remove connection from data
            const beforeCount = flowData.connections.length;
            flowData.connections = flowData.connections.filter(
                c => !(c.from === fromNodeId && c.to === toNodeId)
            );
            const afterCount = flowData.connections.length;
            console.log('✓ Connection deleted:', fromNodeId, '->', toNodeId);
            console.log('Connections before:', beforeCount, 'after:', afterCount);

            // Redraw connections
            drawConnections();

            Swal.fire({
                title: 'Deleted!',
                text: 'Connection has been removed',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914',
                timer: 1500
            });
        } else {
            console.log('User cancelled deletion');
        }
    });
}

function drawConnections() {
    const svg = document.getElementById('connectionLayer');
    const canvasContainer = document.querySelector('.canvas-container');
    const canvas = document.getElementById('flowCanvas');

    svg.innerHTML = ''; // Clear existing lines

    // Update SVG size to match canvas
    svg.setAttribute('width', canvas.offsetWidth);
    svg.setAttribute('height', canvas.offsetHeight);

    console.log('Drawing connections:', flowData.connections.length);

    flowData.connections.forEach(conn => {
        const fromNode = document.querySelector(`[data-node-id="${conn.from}"]`);
        const toNode = document.querySelector(`[data-node-id="${conn.to}"]`);

        if (!fromNode || !toNode) {
            console.log('Node not found:', conn.from, conn.to);
            return;
        }

        const fromConnector = fromNode.querySelector('.output-connector');
        const toConnector = toNode.querySelector('.input-connector');

        if (!fromConnector || !toConnector) {
            console.log('Connector not found for nodes:', conn.from, conn.to);
            return;
        }

        // Get absolute positions relative to canvas
        const fromRect = fromConnector.getBoundingClientRect();
        const toRect = toConnector.getBoundingClientRect();
        const containerRect = canvasContainer.getBoundingClientRect();

        const x1 = fromRect.left - containerRect.left + canvasContainer.scrollLeft + fromRect.width / 2;
        const y1 = fromRect.top - containerRect.top + canvasContainer.scrollTop + fromRect.height / 2;
        const x2 = toRect.left - containerRect.left + canvasContainer.scrollLeft + toRect.width / 2;
        const y2 = toRect.top - containerRect.top + canvasContainer.scrollTop + toRect.height / 2;

        console.log('Drawing line from', conn.from, 'to', conn.to, ':', {x1, y1, x2, y2});

        // Create curved path
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        const midY = (y1 + y2) / 2;
        const d = `M ${x1} ${y1} C ${x1} ${midY}, ${x2} ${midY}, ${x2} ${y2}`;

        path.setAttribute('d', d);
        path.setAttribute('class', 'connection-line');
        path.setAttribute('data-from', conn.from);
        path.setAttribute('data-to', conn.to);
        path.setAttribute('stroke', '#ffd700');
        path.setAttribute('stroke-width', '2');
        path.setAttribute('fill', 'none');
        path.style.cursor = 'pointer';
        path.style.pointerEvents = 'all';

        // Add click handler to delete connection
        path.addEventListener('click', (e) => {
            e.stopPropagation();
            e.preventDefault();
            console.log('🗑️ Connection line clicked! From:', conn.from, 'To:', conn.to);
            deleteConnection(conn.from, conn.to);
        }, true);

        // Add hover effect
        path.addEventListener('mouseenter', () => {
            path.setAttribute('stroke', '#e50914');
            path.setAttribute('stroke-width', '3');
        });

        path.addEventListener('mouseleave', () => {
            path.setAttribute('stroke', '#ffd700');
            path.setAttribute('stroke-width', '2');
        });

        svg.appendChild(path);
    });

    console.log('Total SVG paths created:', svg.children.length);
}

// Canvas Panning and Zoom
function initializeCanvasPan() {
    const canvasContainer = document.querySelector('.canvas-container');
    const canvas = document.getElementById('flowCanvas');

    // Mouse wheel zoom
    canvasContainer.addEventListener('wheel', (e) => {
        e.preventDefault();

        const delta = e.deltaY > 0 ? -0.1 : 0.1;
        const newScale = Math.max(0.5, Math.min(2, zoomScale + delta));

        if (newScale !== zoomScale) {
            zoomScale = newScale;
            applyZoom();
            // Redraw connections after zoom
            setTimeout(() => drawConnections(), 10);
        }
    });

    // Middle mouse button or left click on canvas background for panning
    canvasContainer.addEventListener('mousedown', (e) => {
        // Check if clicking on canvas container or the canvas itself (not nodes, not connectors)
        const isCanvasBackground = e.target === canvasContainer ||
                                   e.target.id === 'flowCanvas' ||
                                   e.target.classList.contains('flow-canvas') ||
                                   e.target.id === 'connectionLayer';

        // Middle mouse button (button 1) or left click on canvas background
        if (e.button === 1 || (e.button === 0 && isCanvasBackground)) {
            isPanning = true;
            panStart = {
                x: e.clientX,
                y: e.clientY,
                scrollLeft: canvasContainer.scrollLeft,
                scrollTop: canvasContainer.scrollTop
            };
            canvasContainer.style.cursor = 'grabbing';
            e.preventDefault();
        }
    });

    canvasContainer.addEventListener('mousemove', (e) => {
        if (!isPanning) return;

        e.preventDefault();
        const dx = e.clientX - panStart.x;
        const dy = e.clientY - panStart.y;

        canvasContainer.scrollLeft = panStart.scrollLeft - dx;
        canvasContainer.scrollTop = panStart.scrollTop - dy;
    });

    canvasContainer.addEventListener('mouseup', (e) => {
        if (isPanning) {
            isPanning = false;
            canvasContainer.style.cursor = '';
            // Redraw connections after panning
            drawConnections();
        }
    });

    canvasContainer.addEventListener('mouseleave', () => {
        if (isPanning) {
            isPanning = false;
            canvasContainer.style.cursor = '';
        }
    });

    // Redraw connections on scroll
    canvasContainer.addEventListener('scroll', () => {
        drawConnections();
    });

    // Prevent context menu on middle click
    canvasContainer.addEventListener('contextmenu', (e) => {
        if (e.button === 1) {
            e.preventDefault();
        }
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
    initializeCanvasPan();
});
