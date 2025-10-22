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
            document.getElementById('flowNameInput').value = data.flow.flow_name || '';
            document.getElementById('nicheInput').value = data.flow.niche || '';

            // Load nodes (if stored as JSON in backend)
            if (data.flow.nodes_data) {
                loadFlowFromData(JSON.parse(data.flow.nodes_data));
            }
        } else {
            // No existing flow, clear canvas
            document.getElementById('flowNameInput').value = '';
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
            draggedElement = {
                type: e.target.getAttribute('data-node-type'),
                label: e.target.querySelector('.node-label').textContent,
                icon: e.target.querySelector('.node-icon').textContent
            };
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
            const rect = canvas.getBoundingClientRect();
            const x = e.clientX - rect.left + canvas.scrollLeft;
            const y = e.clientY - rect.top + canvas.scrollTop;

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
            <p>Configure ${label.toLowerCase()} settings</p>
        </div>
        <div class="node-connector input-connector" data-connector-type="input"></div>
        <div class="node-connector output-connector" data-connector-type="output"></div>
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
            e.target.classList.contains('node-delete')) {
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
    });

    document.addEventListener('mouseup', () => {
        isDragging = false;
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
    const flowName = document.getElementById('flowNameInput').value.trim();
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
            text: 'Please enter a flow name',
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
    const flowName = document.getElementById('flowNameInput').value.trim();

    if (!flowName) {
        Swal.fire({
            title: 'Error!',
            text: 'Please enter a flow name before exporting',
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
                    document.getElementById('flowNameInput').value = importedData.flowName;
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
        'manual_response': '✋',
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

// Logout function
function logout() {
    localStorage.removeItem('auth_token');
    window.location.href = '/';
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    loadDevices();
    initializeDragAndDrop();
});
