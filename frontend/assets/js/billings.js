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

// Buy package function
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
            // Show success message then redirect
            Swal.fire({
                title: 'Order Created!',
                text: 'Redirecting to payment gateway...',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914',
                timer: 2000,
                showConfirmButton: false
            }).then(() => {
                // Redirect to Billplz payment page
                window.location.href = result.url;
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
    document.getElementById('ordersLoading').style.display = 'block';
    document.getElementById('ordersError').style.display = 'none';
    document.getElementById('ordersEmpty').style.display = 'none';
    document.getElementById('ordersTable').style.display = 'none';

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
                document.getElementById('ordersEmpty').style.display = 'block';
            } else {
                // Show orders table
                document.getElementById('ordersTable').style.display = 'block';
                renderOrdersTable(orders);
            }
        } else {
            throw new Error(result.message || 'Failed to load orders');
        }
    } catch (error) {
        console.error('Error loading orders:', error);
        document.getElementById('ordersLoading').style.display = 'none';
        document.getElementById('ordersError').style.display = 'block';
        document.getElementById('ordersErrorMessage').textContent = error.message || 'Failed to load orders';
    }
}

// Render orders table
function renderOrdersTable(orders) {
    const tbody = document.getElementById('ordersTableBody');
    tbody.innerHTML = '';

    orders.forEach(order => {
        const row = document.createElement('tr');

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

        // Payment method
        const methodBadge = order.method === 'billplz'
            ? '<span style="background: rgba(59, 130, 246, 0.2); color: #3b82f6; padding: 0.25rem 0.75rem; border-radius: 4px; font-size: 0.85rem;">💳 Billplz</span>'
            : '<span style="background: rgba(34, 197, 94, 0.2); color: #22c55e; padding: 0.25rem 0.75rem; border-radius: 4px; font-size: 0.85rem;">💵 COD</span>';

        // Action button
        let actionButton = '';
        if (order.status === 'Pending' && order.url) {
            actionButton = `<button onclick="payNow('${order.url}')" style="
                background: linear-gradient(135deg, #e50914 0%, #b00710 100%);
                border: none;
                padding: 0.5rem 1rem;
                border-radius: 6px;
                color: white;
                font-weight: 600;
                cursor: pointer;
                font-size: 0.85rem;
                transition: all 0.3s ease;
            " onmouseover="this.style.transform='scale(1.05)'" onmouseout="this.style.transform='scale(1)'">
                Pay Now
            </button>`;
        } else if (order.status === 'Success') {
            actionButton = '<span style="color: var(--netflix-green);">✓ Paid</span>';
        } else if (order.status === 'Failed') {
            actionButton = '<span style="color: var(--netflix-red);">✗ Failed</span>';
        } else {
            actionButton = '<span style="color: var(--netflix-light-gray);">-</span>';
        }

        row.innerHTML = `
            <td><strong>#${order.id}</strong></td>
            <td>${order.product}</td>
            <td><strong style="color: var(--netflix-gold);">RM ${parseFloat(order.amount).toFixed(2)}</strong></td>
            <td>${methodBadge}</td>
            <td>${statusBadge}</td>
            <td>${formattedDate}</td>
            <td>${actionButton}</td>
        `;

        tbody.appendChild(row);
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

// Pay now function
function payNow(url) {
    if (url) {
        Swal.fire({
            title: 'Redirecting...',
            text: 'Taking you to payment gateway',
            icon: 'info',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914',
            timer: 1500,
            showConfirmButton: false
        }).then(() => {
            window.location.href = url;
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
