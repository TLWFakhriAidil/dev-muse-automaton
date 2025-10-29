// Billings JavaScript
const API_BASE_URL = window.location.origin + '/api';

// Check authentication on page load
document.addEventListener('DOMContentLoaded', () => {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    // Load user info
    loadUserInfo();

    // Load orders
    loadOrders();
});

// Load user info
async function loadUserInfo() {
    const token = localStorage.getItem('auth_token');
    try {
        const response = await fetch(`${API_BASE_URL}/auth/profile`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        const data = await response.json();
        if (data.user && data.user.email) {
            document.getElementById('userEmail').textContent = data.user.email;
        }
    } catch (error) {
        console.error('Error loading user info:', error);
    }
}

// Buy package function - Opens payment in NEW TAB
async function buyPackage() {
    const token = localStorage.getItem('auth_token');

    if (!token) {
        Swal.fire({
            title: 'Authentication Required',
            text: 'Please log in to make a purchase',
            icon: 'warning',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
        return;
    }

    // Show loading
    Swal.fire({
        title: 'Processing...',
        text: 'Creating your order',
        icon: 'info',
        background: '#141414',
        color: '#ffffff',
        showConfirmButton: false,
        allowOutsideClick: false,
        didOpen: () => {
            Swal.showLoading();
        }
    });

    try {
        const response = await fetch(`${API_BASE_URL}/billing/orders`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                product: 'Test Package - RM 1.00',
                method: 'billplz'
            })
        });

        const result = await response.json();

        if (result.success && result.url) {
            // Close loading dialog
            Swal.close();

            // Open payment page in NEW TAB
            window.open(result.url, '_blank');

            // Show success message
            Swal.fire({
                title: 'Order Created!',
                text: 'Payment page opened in new tab. Please complete your payment.',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            }).then(() => {
                // Reload orders after closing dialog
                loadOrders();
            });
        } else {
            throw new Error(result.message || 'Failed to create order');
        }
    } catch (error) {
        console.error('Error creating order:', error);
        Swal.fire({
            title: 'Error',
            text: error.message || 'Failed to create order. Please try again.',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
    }
}

// Load orders
async function loadOrders() {
    const token = localStorage.getItem('auth_token');

    if (!token) {
        window.location.href = '/';
        return;
    }

    // Show loading
    document.getElementById('ordersLoading').style.display = 'flex';
    document.getElementById('ordersError').style.display = 'none';
    document.getElementById('ordersEmpty').style.display = 'none';
    document.getElementById('ordersGrid').style.display = 'none';

    try {
        const response = await fetch(`${API_BASE_URL}/billing/orders`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        const result = await response.json();

        // Hide loading
        document.getElementById('ordersLoading').style.display = 'none';

        if (result.success && result.data) {
            const orders = result.data;

            if (orders.length === 0) {
                // Show empty state
                document.getElementById('ordersEmpty').style.display = 'flex';
            } else {
                // Show orders grid
                document.getElementById('ordersGrid').style.display = 'grid';
                renderOrdersGrid(orders);
            }
        } else {
            throw new Error(result.message || 'Failed to load orders');
        }
    } catch (error) {
        console.error('Error loading orders:', error);
        document.getElementById('ordersLoading').style.display = 'none';
        document.getElementById('ordersError').style.display = 'flex';
        document.getElementById('ordersErrorMessage').textContent = error.message || 'Failed to load orders';
    }
}

// Render orders as card grid (like device settings)
function renderOrdersGrid(orders) {
    const grid = document.getElementById('ordersGrid');
    grid.innerHTML = '';

    orders.forEach(order => {
        // Format date
        const date = new Date(order.created_at);
        const formattedDate = date.toLocaleDateString('en-MY', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });

        // Status badge
        const statusClass = getStatusClass(order.status);
        const statusBadge = `<span class="status-badge ${statusClass}">${order.status}</span>`;

        // Payment method badge
        const methodBadge = order.method === 'billplz'
            ? '<span style="background: rgba(59, 130, 246, 0.2); color: #3b82f6; padding: 0.4rem 0.8rem; border-radius: 6px; font-size: 0.85rem; font-weight: 600;">💳 Billplz</span>'
            : '<span style="background: rgba(34, 197, 94, 0.2); color: #22c55e; padding: 0.4rem 0.8rem; border-radius: 6px; font-size: 0.85rem; font-weight: 600;">💵 COD</span>';

        // Action button
        let actionButton = '';
        if (order.status === 'Pending' && order.url) {
            actionButton = `<button onclick="payNow('${order.url}')" class="device-action-btn" style="
                background: linear-gradient(135deg, #e50914 0%, #b00710 100%);
                border: none;
                padding: 0.6rem 1.2rem;
                border-radius: 8px;
                color: white;
                font-weight: 600;
                cursor: pointer;
                font-size: 0.9rem;
                transition: all 0.3s ease;
                box-shadow: 0 4px 15px rgba(229, 9, 20, 0.3);
            " onmouseover="this.style.transform='translateY(-2px)'; this.style.boxShadow='0 6px 20px rgba(229, 9, 20, 0.4)'" onmouseout="this.style.transform='translateY(0)'; this.style.boxShadow='0 4px 15px rgba(229, 9, 20, 0.3)'">
                💳 Pay Now
            </button>`;
        } else if (order.status === 'Success') {
            actionButton = '<span style="color: #22c55e; font-weight: 600; font-size: 1rem;">✓ Paid</span>';
        } else if (order.status === 'Failed') {
            actionButton = '<span style="color: #ef4444; font-weight: 600; font-size: 1rem;">✗ Failed</span>';
        } else {
            actionButton = '<span style="color: var(--netflix-light-gray); font-weight: 600;">Processing...</span>';
        }

        // Create card
        const card = document.createElement('div');
        card.className = 'device-card';
        card.innerHTML = `
            <div class="device-card-header">
                <div>
                    <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem;">
                        <h3 class="device-name" style="font-size: 1.1rem;">Order #${order.id}</h3>
                        ${statusBadge}
                    </div>
                    <p class="device-provider" style="color: var(--netflix-light-gray); font-size: 0.9rem;">
                        📅 ${formattedDate}
                    </p>
                </div>
            </div>
            <div class="device-info">
                <div class="info-item">
                    <span class="info-label">Product</span>
                    <span class="info-value">${order.product}</span>
                </div>
                <div class="info-item">
                    <span class="info-label">Amount</span>
                    <span class="info-value" style="color: var(--netflix-gold); font-weight: 700; font-size: 1.1rem;">
                        RM ${parseFloat(order.amount).toFixed(2)}
                    </span>
                </div>
                <div class="info-item">
                    <span class="info-label">Payment Method</span>
                    <span class="info-value">${methodBadge}</span>
                </div>
            </div>
            <div class="device-actions" style="margin-top: 1rem; display: flex; justify-content: flex-end;">
                ${actionButton}
            </div>
        `;

        grid.appendChild(card);
    });
}

// Get status class for badge
function getStatusClass(status) {
    switch (status) {
        case 'Success':
            return 'status-success';
        case 'Pending':
            return 'status-pending';
        case 'Processing':
            return 'status-processing';
        case 'Failed':
            return 'status-failed';
        default:
            return '';
    }
}

// Pay now function - Opens payment in NEW TAB
function payNow(url) {
    if (url) {
        // Open payment in new tab
        window.open(url, '_blank');

        Swal.fire({
            title: 'Payment Page Opened',
            text: 'Complete your payment in the new tab',
            icon: 'info',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        }).then(() => {
            // Reload orders after closing dialog
            loadOrders();
        });
    }
}

// Logout function
function logout() {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user_id');
    localStorage.removeItem('user_email');
    window.location.href = '/';
}
