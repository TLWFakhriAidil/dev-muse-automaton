// Common JavaScript functions for all pages
const API_BASE_URL = window.location.origin + '/api';

// Show Packages tab for admin users
async function showPackagesTabForAdmin() {
    const token = localStorage.getItem('auth_token');
    if (!token) return;

    try {
        const response = await fetch(`${API_BASE_URL}/auth/profile`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });
        const data = await response.json();

        if (data.user && (data.user.email === 'Admin@gmail.com' || data.user.role === 'admin')) {
            const packagesTab = document.getElementById('packagesTab');
            if (packagesTab) {
                packagesTab.style.display = 'flex';
            }
        }
    } catch (error) {
        console.error('Error checking admin status:', error);
    }
}

// Call this on page load
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', showPackagesTabForAdmin);
} else {
    showPackagesTabForAdmin();
}
