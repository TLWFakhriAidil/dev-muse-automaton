// Chatbot AI JavaScript
const API_BASE_URL = window.location.origin + '/api';

// Store all conversations for filtering
let allConversations = [];
let filteredConversations = [];

// Load conversations from ai_whatsapp table
async function loadConversations() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    try {
        // Call single endpoint to get all conversations
        const response = await fetch(`${API_BASE_URL}/conversations/all`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (data.success && data.conversations && data.conversations.length > 0) {
            allConversations = data.conversations;
            filteredConversations = [...allConversations];

            // Populate filter dropdowns
            populateFilters();

            // Display table
            displayConversations(filteredConversations);
        } else {
            const conversationsList = document.getElementById('conversationsList');
            conversationsList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">💬</div>
                    <h2 class="empty-state-title">No Conversations Yet</h2>
                    <p class="empty-state-text">Start a conversation with your Chatbot AI to see it here</p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Load conversations error:', error);
        Swal.fire({
            title: 'Error!',
            text: 'Failed to load conversations',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
    }
}

// Populate filter dropdowns
function populateFilters() {
    // Get unique devices
    const devices = [...new Set(allConversations.map(c => c.id_device).filter(Boolean))];
    const deviceFilter = document.getElementById('deviceFilter');
    deviceFilter.innerHTML = '<option value="">All Devices</option>' +
        devices.map(d => `<option value="${d}">${d}</option>`).join('');

    // Get unique stages
    const stages = [...new Set(allConversations.map(c => c.stage || 'Welcome Message'))];
    const stageFilter = document.getElementById('stageFilter');
    stageFilter.innerHTML = '<option value="">All Stages</option>' +
        stages.map(s => `<option value="${s}">${s}</option>`).join('');
}

// Display conversations in table
function displayConversations(conversations) {
    const conversationsList = document.getElementById('conversationsList');

    if (conversations.length === 0) {
        conversationsList.innerHTML = `
            <div class="empty-state">
                <div class="empty-state-icon">🔍</div>
                <h2 class="empty-state-title">No Results Found</h2>
                <p class="empty-state-text">Try adjusting your filters or search criteria</p>
            </div>
        `;
        return;
    }

    conversationsList.innerHTML = `
        <div class="table-container">
            <table class="devices-table">
                <thead>
                    <tr>
                        <th>No</th>
                        <th>ID Device</th>
                        <th>Date</th>
                        <th>Name</th>
                        <th>Phone Number</th>
                        <th>Niche</th>
                        <th>Stage</th>
                        <th>Conversation History</th>
                        <th>Reply Status</th>
                        <th>Action</th>
                    </tr>
                </thead>
                <tbody>
                    ${conversations.map((conv, index) => {
                        // Format date to d-m-Y
                        let dateFormatted = '-';
                        if (conv.created_at) {
                            const date = new Date(conv.created_at);
                            const day = String(date.getDate()).padStart(2, '0');
                            const month = String(date.getMonth() + 1).padStart(2, '0');
                            const year = date.getFullYear();
                            dateFormatted = `${day}-${month}-${year}`;
                        }

                        // Reply Status: 0 = AI, 1 = Human
                        const replyStatus = conv.human === 1 ? 'Human' : 'AI';
                        const replyBadgeClass = conv.human === 1 ? 'badge-human' : 'badge-ai';

                        return `
                            <tr>
                                <td><strong>${index + 1}</strong></td>
                                <td>${conv.id_device || '-'}</td>
                                <td>${dateFormatted}</td>
                                <td>${conv.prospect_name || '-'}</td>
                                <td><strong>${conv.prospect_num || '-'}</strong></td>
                                <td><span class="badge badge-niche">${conv.niche || '-'}</span></td>
                                <td><span class="badge badge-stage">${conv.stage || 'Welcome Message'}</span></td>
                                <td>
                                    <button class="btn-view" onclick='viewConversation(${JSON.stringify(conv).replace(/'/g, "&#39;")})' title="View Conversation History">👁️</button>
                                </td>
                                <td><span class="badge ${replyBadgeClass}">${replyStatus}</span></td>
                                <td>
                                    <button class="btn-delete" onclick="deleteConversation('${conv.prospect_num}')" title="Delete">🗑️</button>
                                </td>
                            </tr>
                        `;
                    }).join('')}
                </tbody>
            </table>
        </div>
    `;
}

// Filter conversations
function applyFilters() {
    const deviceFilter = document.getElementById('deviceFilter').value;
    const stageFilter = document.getElementById('stageFilter').value;
    const dateFilter = document.getElementById('dateFilter').value;
    const searchInput = document.getElementById('searchInput').value.toLowerCase();

    filteredConversations = allConversations.filter(conv => {
        // Device filter
        if (deviceFilter && conv.id_device !== deviceFilter) return false;

        // Stage filter
        const convStage = conv.stage || 'Welcome Message';
        if (stageFilter && convStage !== stageFilter) return false;

        // Date filter (Y-m-d format)
        if (dateFilter && conv.created_at) {
            const convDate = new Date(conv.created_at);
            const year = convDate.getFullYear();
            const month = String(convDate.getMonth() + 1).padStart(2, '0');
            const day = String(convDate.getDate()).padStart(2, '0');
            const convDateStr = `${year}-${month}-${day}`;
            if (convDateStr !== dateFilter) return false;
        }

        // Search filter (search in name, phone, niche)
        if (searchInput) {
            const searchMatch =
                (conv.prospect_name && conv.prospect_name.toLowerCase().includes(searchInput)) ||
                (conv.prospect_num && conv.prospect_num.toLowerCase().includes(searchInput)) ||
                (conv.niche && conv.niche.toLowerCase().includes(searchInput));
            if (!searchMatch) return false;
        }

        return true;
    });

    displayConversations(filteredConversations);
}

// Reset filters
function resetFilters() {
    document.getElementById('deviceFilter').value = '';
    document.getElementById('stageFilter').value = '';
    document.getElementById('dateFilter').value = '';
    document.getElementById('searchInput').value = '';
    filteredConversations = [...allConversations];
    displayConversations(filteredConversations);
}

// Export to CSV
function exportToCSV() {
    if (filteredConversations.length === 0) {
        Swal.fire({
            title: 'No Data',
            text: 'No conversations to export',
            icon: 'warning',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
        return;
    }

    // CSV headers
    let csv = 'No,ID Device,Date,Name,Phone Number,Niche,Stage,Conversation History,Reply Status\n';

    // CSV rows
    filteredConversations.forEach((conv, index) => {
        // Format date to d-m-Y
        let dateFormatted = '-';
        if (conv.created_at) {
            const date = new Date(conv.created_at);
            const day = String(date.getDate()).padStart(2, '0');
            const month = String(date.getMonth() + 1).padStart(2, '0');
            const year = date.getFullYear();
            dateFormatted = `${day}-${month}-${year}`;
        }

        const replyStatus = conv.human === 1 ? 'Human' : 'AI';

        // Clean conversation history for CSV (remove line breaks, escape quotes)
        let convHistory = conv.conv_last || '';
        convHistory = convHistory.replace(/"/g, '""').replace(/\n/g, ' | ');

        csv += `${index + 1},"${conv.id_device || '-'}","${dateFormatted}","${conv.prospect_name || '-'}","${conv.prospect_num || '-'}","${conv.niche || '-'}","${conv.stage || 'Welcome Message'}","${convHistory}","${replyStatus}"\n`;
    });

    // Download CSV
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `chatbot-ai-conversations-${new Date().getTime()}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);

    Swal.fire({
        title: 'Success!',
        text: 'Conversations exported successfully',
        icon: 'success',
        background: '#141414',
        color: '#ffffff',
        confirmButtonColor: '#e50914',
        timer: 2000
    });
}

// View conversation details
function viewConversation(conv) {
    let convHistory = 'No conversation history';
    if (conv.conv_last) {
        convHistory = conv.conv_last.replace(/\n/g, '<br>');
    }

    Swal.fire({
        title: `Conversation History`,
        html: `
            <div style="text-align: left; color: #ffffff;">
                <p><strong>Phone:</strong> ${conv.prospect_num || '-'}</p>
                <p><strong>Name:</strong> ${conv.prospect_name || '-'}</p>
                <p><strong>Device:</strong> ${conv.id_device || '-'}</p>
                <p><strong>Niche:</strong> ${conv.niche || '-'}</p>
                <p><strong>Stage:</strong> ${conv.stage || 'Welcome Message'}</p>
                <hr style="border-color: #333;">
                <p><strong>Conversation:</strong></p>
                <div style="background: #1a1a1a; padding: 10px; border-radius: 5px; max-height: 300px; overflow-y: auto;">
                    ${convHistory}
                </div>
            </div>
        `,
        width: '700px',
        background: '#141414',
        color: '#ffffff',
        confirmButtonColor: '#e50914',
        confirmButtonText: 'Close'
    });
}

// Delete conversation
async function deleteConversation(prospectNum) {
    const result = await Swal.fire({
        title: 'Are you sure?',
        text: "You won't be able to revert this!",
        icon: 'warning',
        showCancelButton: true,
        background: '#141414',
        color: '#ffffff',
        confirmButtonColor: '#e50914',
        cancelButtonColor: '#6c757d',
        confirmButtonText: 'Yes, delete it!',
        cancelButtonText: 'Cancel'
    });

    if (!result.isConfirmed) return;

    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/conversations/${prospectNum}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        if (data.success) {
            Swal.fire({
                title: 'Deleted!',
                text: 'Conversation has been deleted.',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914',
                timer: 2000
            });

            // Reload conversations
            loadConversations();
        } else {
            Swal.fire({
                title: 'Error!',
                text: data.message || 'Failed to delete conversation',
                icon: 'error',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });
        }
    } catch (error) {
        console.error('Delete conversation error:', error);
        Swal.fire({
            title: 'Error!',
            text: 'Failed to delete conversation',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    loadConversations();
});
