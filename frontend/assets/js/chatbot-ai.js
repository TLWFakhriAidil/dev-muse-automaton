// Chatbot AI JavaScript
const API_BASE_URL = window.location.origin + '/api';

// Load conversations from ai_whatsapp table
async function loadConversations() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    try {
        const response = await fetch(`${API_BASE_URL}/analytics/conversations`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        const conversationsList = document.getElementById('conversationsList');

        if (data.success && data.conversations && data.conversations.length > 0) {
            conversationsList.innerHTML = `
                <table class="stage-table">
                    <thead>
                        <tr>
                            <th>ID</th>
                            <th>Device ID</th>
                            <th>Phone Number</th>
                            <th>Name</th>
                            <th>Niche</th>
                            <th>Stage</th>
                            <th>Status</th>
                            <th>Created At</th>
                            <th>Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${data.conversations.map(conv => `
                            <tr>
                                <td><strong>${conv.id_prospect || '-'}</strong></td>
                                <td>${conv.id_device || '-'}</td>
                                <td>${conv.prospect_num || '-'}</td>
                                <td>${conv.prospect_name || '-'}</td>
                                <td>${conv.niche || '-'}</td>
                                <td><span class="column-badge">${conv.stage || 'Welcome Message'}</span></td>
                                <td><span class="type-badge type-${(conv.execution_status || 'active').toLowerCase()}">${conv.execution_status || 'active'}</span></td>
                                <td>${conv.created_at ? new Date(conv.created_at).toLocaleString() : '-'}</td>
                                <td>
                                    <div class="btn-action-group">
                                        <button class="btn-edit" onclick='viewConversation(${JSON.stringify(conv)})'>View</button>
                                    </div>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            `;
        } else {
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

// View conversation details
function viewConversation(conv) {
    let convHistory = 'No conversation history';
    if (conv.conv_last) {
        convHistory = conv.conv_last.replace(/\n/g, '<br>');
    }

    Swal.fire({
        title: `Conversation Details`,
        html: `
            <div style="text-align: left; color: #ffffff;">
                <p><strong>Phone:</strong> ${conv.prospect_num || '-'}</p>
                <p><strong>Name:</strong> ${conv.prospect_name || '-'}</p>
                <p><strong>Niche:</strong> ${conv.niche || '-'}</p>
                <p><strong>Stage:</strong> ${conv.stage || 'Welcome Message'}</p>
                <p><strong>Status:</strong> ${conv.execution_status || 'active'}</p>
                <hr style="border-color: #333;">
                <p><strong>Conversation History:</strong></p>
                <div style="background: #1a1a1a; padding: 10px; border-radius: 5px; max-height: 300px; overflow-y: auto;">
                    ${convHistory}
                </div>
            </div>
        `,
        width: '600px',
        background: '#141414',
        color: '#ffffff',
        confirmButtonColor: '#e50914',
        confirmButtonText: 'Close'
    });
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    loadConversations();
});
