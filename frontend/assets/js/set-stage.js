// Set Stage JavaScript
const API_BASE_URL = window.location.origin + '/api';

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

// Load stage values
async function loadStageValues() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/stage-values`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        const stageValuesList = document.getElementById('stageValuesList');

        if (data.success && data.stage_values && data.stage_values.length > 0) {
            stageValuesList.innerHTML = `
                <table class="stage-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Device ID</th>
                            <th>Stage</th>
                            <th>Type</th>
                            <th>Input Hard Code</th>
                            <th>Column</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${data.stage_values.map(stage => `
                            <tr>
                                <td><strong>${stage.stagesetvalue_id || stage.id}</strong></td>
                                <td>${stage.id_device || '-'}</td>
                                <td><strong>${stage.stage || '-'}</strong></td>
                                <td><span class="type-badge type-${(stage.type_inputdata || stage.type || '').toLowerCase()}">${stage.type_inputdata || stage.type || '-'}</span></td>
                                <td>${stage.inputhardcode || stage.input_hard_code || '-'}</td>
                                <td><span class="column-badge">${stage.columnsdata || stage.column || '-'}</span></td>
                                <td>
                                    <div class="btn-action-group">
                                        <button class="btn-edit" onclick='editStageValue(${JSON.stringify(stage)})'>Edit</button>
                                        <button class="btn-delete" onclick="deleteStageValue(${stage.stagesetvalue_id || stage.id})">Delete</button>
                                    </div>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            `;
        } else {
            stageValuesList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">⚙️</div>
                    <h2 class="empty-state-title">No Stage Values Yet</h2>
                    <p class="empty-state-text">Click "Add Set Stage" to create your first stage value</p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Load stage values error:', error);
        Swal.fire({
            title: 'Error!',
            text: 'Failed to load stage values',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
    }
}

// Open stage modal
function openStageModal() {
    const modal = document.getElementById('stageModal');
    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
    loadDevices();
}

// Close stage modal
function closeStageModal() {
    const modal = document.getElementById('stageModal');
    modal.classList.remove('active');
    document.body.style.overflow = 'auto';

    // Reset form
    document.getElementById('stageForm').reset();
    document.querySelector('.modal-title').textContent = 'Add Stage Value';
    window.editingStageId = null;
}

// Save stage value
async function saveStageValue(event) {
    event.preventDefault();

    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    const isEditMode = window.editingStageId !== undefined && window.editingStageId !== null;

    const stageData = {
        id_device: document.getElementById('deviceSelect').value,
        stage: document.getElementById('stageInput').value.trim(),
        type_inputdata: document.getElementById('typeSelect').value,
        inputhardcode: document.getElementById('inputHardCode').value.trim(),
        columnsdata: document.getElementById('columnSelect').value
    };

    // Show loading
    Swal.fire({
        title: isEditMode ? 'Updating...' : 'Saving...',
        text: 'Please wait',
        allowOutsideClick: false,
        didOpen: () => {
            Swal.showLoading();
        },
        background: '#141414',
        color: '#ffffff'
    });

    try {
        const url = isEditMode
            ? `${API_BASE_URL}/stage-values/${window.editingStageId}`
            : `${API_BASE_URL}/stage-values`;

        const method = isEditMode ? 'PUT' : 'POST';

        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify(stageData)
        });

        const data = await response.json();

        if (data.success) {
            Swal.fire({
                title: 'Success!',
                text: isEditMode ? 'Stage value updated successfully' : 'Stage value created successfully',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });

            closeStageModal();
            loadStageValues();
        } else {
            Swal.fire({
                title: 'Error!',
                text: data.message || 'Failed to save stage value',
                icon: 'error',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });
        }
    } catch (error) {
        console.error('Save stage value error:', error);
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

// Edit stage value
function editStageValue(stage) {
    // Populate form
    document.getElementById('deviceSelect').value = stage.id_device || '';
    document.getElementById('stageInput').value = stage.stage || '';
    document.getElementById('typeSelect').value = stage.type_inputdata || stage.type || '';
    document.getElementById('inputHardCode').value = stage.inputhardcode || stage.input_hard_code || '';
    document.getElementById('columnSelect').value = stage.columnsdata || stage.column || '';

    // Change modal title
    document.querySelector('.modal-title').textContent = 'Edit Stage Value';

    // Store stage ID for update
    window.editingStageId = stage.stagesetvalue_id || stage.id;

    // Open modal
    openStageModal();
}

// Delete stage value
async function deleteStageValue(stageId) {
    const result = await Swal.fire({
        title: 'Delete Stage Value?',
        text: 'This action cannot be undone!',
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#e50914',
        cancelButtonColor: '#6c757d',
        confirmButtonText: 'Yes, delete it!',
        cancelButtonText: 'Cancel',
        background: '#141414',
        color: '#ffffff'
    });

    if (!result.isConfirmed) return;

    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/stage-values/${stageId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (data.success) {
            Swal.fire({
                title: 'Deleted!',
                text: 'Stage value has been deleted',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });

            loadStageValues();
        } else {
            Swal.fire({
                title: 'Error!',
                text: data.message || 'Failed to delete stage value',
                icon: 'error',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });
        }
    } catch (error) {
        console.error('Delete stage value error:', error);
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

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    loadStageValues();
});
