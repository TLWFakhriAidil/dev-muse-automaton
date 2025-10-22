// Dashboard JavaScript

// Check authentication on page load
function checkAuth() {
    const token = localStorage.getItem('auth_token');
    const userEmail = localStorage.getItem('user_email');

    if (!token) {
        // Not logged in, redirect to login
        window.location.href = '/';
        return;
    }

    // Display user email
    if (userEmail) {
        document.getElementById('userEmail').textContent = userEmail;
    }
}

// Logout function
function logout() {
    // Confirm logout
    if (confirm('Are you sure you want to logout?')) {
        // Clear localStorage
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user_id');
        localStorage.removeItem('user_email');

        // Redirect to login
        window.location.href = '/';
    }
}

// Show coming soon alert
function showComingSoon(event) {
    event.preventDefault();
    alert('This feature is coming soon! Stay tuned.');
}

// Mobile sidebar toggle (for responsive design)
function toggleSidebar() {
    const sidebar = document.querySelector('.sidebar');
    sidebar.classList.toggle('open');
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    checkAuth();

    // Add mobile menu button if needed
    if (window.innerWidth <= 768) {
        addMobileMenuButton();
    }
});

// Add mobile menu button for small screens
function addMobileMenuButton() {
    const mainContent = document.querySelector('.main-content');
    const menuButton = document.createElement('button');
    menuButton.className = 'mobile-menu-btn';
    menuButton.innerHTML = '☰';
    menuButton.onclick = toggleSidebar;
    mainContent.insertBefore(menuButton, mainContent.firstChild);
}

// Handle window resize
window.addEventListener('resize', function() {
    if (window.innerWidth > 768) {
        const sidebar = document.querySelector('.sidebar');
        sidebar.classList.remove('open');
    }
});
