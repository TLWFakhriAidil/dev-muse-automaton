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
        // Call single endpoint to get all conversations
        const response = await fetch(`${API_BASE_URL}/conversations/all`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const data = await response.json();

        const conversationsList = document.getElementById('conversationsList');

        if (data.success && data.conversations && data.conversations.length > 0) {
            conversationsList.innerHTML = `
                <div class="table-container">
                    <table class="devices-table">
                        <thead>
                            <tr>
                                <th>No</th>
                                <th>Created At</th>
                                <th>Device</th>
                                <th>Phone Number</th>
                                <th>Name</th>
                                <th>Niche</th>
                                <th>Stage</th>
                                <th>Status</th>
                                <th>Action</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${data.conversations.map((conv, index) => `
                                <tr>
                                    <td><strong>${index + 1}</strong></td>
                                    <td>${conv.created_at ? new Date(conv.created_at).toLocaleString('en-US', {
                                        month: '2-digit',
                                        day: '2-digit',
                                        year: 'numeric',
                                        hour: '2-digit',
                                        minute: '2-digit',
                                        hour12: true
                                    }) : '-'}</td>
                                    <td>${conv.id_device || '-'}</td>
                                    <td><strong>${conv.prospect_num || '-'}</strong></td>
                                    <td>${conv.prospect_name || '-'}</td>
                                    <td><span class="badge badge-niche">${conv.niche || '-'}</span></td>
                                    <td><span class="badge badge-stage">${conv.stage || 'Welcome Message'}</span></td>
                                    <td><span class="badge ${(conv.execution_status || 'active').toLowerCase() === 'active' ? 'status-connected' : 'status-disconnected'}">${conv.execution_status || 'active'}</span></td>
                                    <td>
                                        <button class="btn-action" onclick='viewConversation(${JSON.stringify(conv).replace(/'/g, "&#39;")})'>View</button>
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>
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
                <p><strong>Device:</strong> ${conv.id_device || '-'}</p>
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
        width: '700px',
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
