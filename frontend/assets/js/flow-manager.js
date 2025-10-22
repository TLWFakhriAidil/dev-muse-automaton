// Flow Manager JavaScript
const API_BASE_URL = window.location.origin + '/api';

// Load all flows on page load
document.addEventListener('DOMContentLoaded', function() {
    loadFlows();
});

// Load all flows
async function loadFlows() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/flows`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();
        console.log('Flows data:', data);

        const tbody = document.getElementById('flowsTableBody');

        if (data.success && data.flows && data.flows.length > 0) {
            tbody.innerHTML = '';

            data.flows.forEach(flow => {
                const row = createFlowRow(flow);
                tbody.appendChild(row);
            });
        } else {
            // Show empty state with icon like set-stage
            const tableContainer = document.querySelector('.table-container');
            tableContainer.innerHTML = `
                <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; min-height: 400px; padding: 3rem;">
                    <div style="font-size: 5rem; margin-bottom: 1.5rem; opacity: 0.5;">
                        📋
                    </div>
                    <h2 style="font-size: 1.8rem; font-weight: 600; color: rgba(255, 255, 255, 0.9); margin-bottom: 0.5rem;">
                        No Flows Yet
                    </h2>
                    <p style="color: rgba(255, 255, 255, 0.6); font-size: 1rem;">
                        Click "Flow Builder" in the sidebar to create your first chatbot flow
                    </p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Load flows error:', error);
        const tbody = document.getElementById('flowsTableBody');
        tbody.innerHTML = `
            <tr>
                <td colspan="7" style="text-align: center; padding: 3rem; color: #e50914;">
                    Error loading flows. Please refresh the page.
                </td>
            </tr>
        `;
    }
}

// Create table row for a flow
function createFlowRow(flow) {
    const row = document.createElement('tr');

    // Format dates
    const createdAt = new Date(flow.created_at).toLocaleString();
    const updatedAt = new Date(flow.updated_at).toLocaleString();

    row.innerHTML = `
        <td><span title="${flow.id}">${flow.id.substring(0, 8)}...</span></td>
        <td><strong>${flow.id_device || 'N/A'}</strong></td>
        <td><strong>${flow.name || 'Unnamed Flow'}</strong></td>
        <td>${flow.niche || '-'}</td>
        <td>${createdAt}</td>
        <td>${updatedAt}</td>
        <td style="display: flex; gap: 0.5rem;">
            <button class="btn-action" onclick="editFlow('${flow.id}')" title="Edit Flow">
                ✏️ Edit
            </button>
            <button class="btn-action" onclick="deleteFlow('${flow.id}', '${flow.name || 'this flow'}')" title="Delete Flow" style="background: linear-gradient(135deg, #dc2626 0%, #991b1b 100%);">
                🗑️ Delete
            </button>
        </td>
    `;

    return row;
}

// Edit flow - redirect to flow builder with flow ID
async function editFlow(flowId) {
    console.log('Editing flow:', flowId);
    // Redirect to flow builder with flow ID as query parameter
    window.location.href = `flow-builder.html?flowId=${flowId}`;
}

// Delete flow
async function deleteFlow(flowId, flowName) {
    const result = await Swal.fire({
        title: 'Delete Flow?',
        text: `Are you sure you want to delete "${flowName}"? This action cannot be undone!`,
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#e50914',
        cancelButtonColor: '#6c757d',
        confirmButtonText: 'Yes, delete it!',
        cancelButtonText: 'Cancel',
        background: '#141414',
        color: '#ffffff'
    });

    if (!result.isConfirmed) {
        return;
    }

    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    Swal.fire({
        title: 'Deleting...',
        text: 'Please wait',
        allowOutsideClick: false,
        didOpen: () => {
            Swal.showLoading();
        },
        background: '#141414',
        color: '#ffffff'
    });

    try {
        const response = await fetch(`${API_BASE_URL}/flows/${flowId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (data.success) {
            Swal.fire({
                title: 'Deleted!',
                text: 'Flow has been deleted successfully',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914',
                timer: 2000
            }).then(() => {
                // Reload flows
                loadFlows();
            });
        } else {
            Swal.fire({
                title: 'Error!',
                text: data.message || 'Failed to delete flow',
                icon: 'error',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });
        }
    } catch (error) {
        console.error('Delete flow error:', error);
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

// Logout function
function logout() {
    localStorage.removeItem('auth_token');
    window.location.href = '/';
}
