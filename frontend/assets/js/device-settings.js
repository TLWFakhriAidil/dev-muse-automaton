// Device Settings JavaScript
const API_BASE_URL = window.location.origin + '/api';

// Generate random device ID
function generateDeviceId() {
    const randomId = 'DEV-' + Math.random().toString(36).substring(2, 15).toUpperCase();
    document.getElementById('deviceId').value = randomId;
}

// Generate webhook URL
function generateWebhook() {
    const randomPath = Math.random().toString(36).substring(2, 15);
    const randomToken = Math.random().toString(36).substring(2, 15);
    const webhookUrl = `${API_BASE_URL}/webhook/whatsapp/${randomPath}/${randomToken}`;
    document.getElementById('webhookId').value = webhookUrl;
}

// Open device modal
function openDeviceModal() {
    const modal = document.getElementById('deviceModal');
    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
}

// Close device modal
function closeDeviceModal() {
    const modal = document.getElementById('deviceModal');
    modal.classList.remove('active');
    document.body.style.overflow = 'auto';

    // Reset form
    document.getElementById('deviceForm').reset();
    document.getElementById('deviceId').value = '';
    document.getElementById('webhookId').value = '';
}

// Save device
async function saveDevice(event) {
    event.preventDefault();

    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/';
        return;
    }

    // Get form values
    const deviceId = document.getElementById('deviceId').value.trim();
    const webhookId = document.getElementById('webhookId').value.trim();
    const apiKeyOption = document.querySelector('input[name="apiKeyOption"]:checked').value;
    const provider = document.querySelector('input[name="provider"]:checked').value;
    const apiKey = document.getElementById('apiKey').value.trim();
    const phoneNumber = document.getElementById('phoneNumber').value.trim();
    const idDevice = document.getElementById('idDevice').value.trim();
    const idErp = document.getElementById('idErp').value.trim();
    const idAdmin = document.getElementById('idAdmin').value.trim();

    // Validate required fields
    if (!deviceId) {
        Swal.fire({
            title: 'Error!',
            text: 'Please click "GENERATE DEVICE" button first',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
        return;
    }

    if (!idDevice) {
        Swal.fire({
            title: 'Error!',
            text: 'ID Device is required',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
        return;
    }

    // Validate phone number (only numbers)
    if (phoneNumber && !/^\d+$/.test(phoneNumber)) {
        Swal.fire({
            title: 'Error!',
            text: 'Phone number must contain only numbers',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
        return;
    }

    // Show loading
    Swal.fire({
        title: 'Saving...',
        text: 'Please wait while we save your device',
        allowOutsideClick: false,
        didOpen: () => {
            Swal.showLoading();
        },
        background: '#141414',
        color: '#ffffff'
    });

    try {
        const response = await fetch(`${API_BASE_URL}/devices`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                device_id: deviceId,
                webhook_url: webhookId,
                api_key_option: apiKeyOption,
                provider: provider,
                api_key: apiKey,
                phone_number: phoneNumber,
                id_device: idDevice,
                id_erp: idErp,
                id_admin: idAdmin
            })
        });

        const data = await response.json();

        if (data.success) {
            Swal.fire({
                title: 'Success!',
                text: 'Device has been saved successfully',
                icon: 'success',
                background: '#141414',
                color: '#ffffff',
                confirmButtonColor: '#e50914'
            });

            closeDeviceModal();
            loadDevices(); // Reload devices list
        } else {
            // Check if it's a duplicate ID error
            if (data.message && data.message.includes('already exists')) {
                Swal.fire({
                    title: 'Duplicate Device!',
                    text: 'A device with this ID already exists. Please use a different ID Device.',
                    icon: 'warning',
                    background: '#141414',
                    color: '#ffffff',
                    confirmButtonColor: '#e50914'
                });
            } else {
                Swal.fire({
                    title: 'Error!',
                    text: data.message || 'Failed to save device',
                    icon: 'error',
                    background: '#141414',
                    color: '#ffffff',
                    confirmButtonColor: '#e50914'
                });
            }
        }
    } catch (error) {
        console.error('Save device error:', error);
        Swal.fire({
            title: 'Error!',
            text: 'Network error. Please check your connection and try again.',
            icon: 'error',
            background: '#141414',
            color: '#ffffff',
            confirmButtonColor: '#e50914'
        });
    }
}

// Load devices list
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

        const devicesList = document.getElementById('devicesList');

        if (data.success && data.devices && data.devices.length > 0) {
            devicesList.innerHTML = data.devices.map(device => `
                <div class="device-card">
                    <h3 style="color: var(--netflix-gold); margin-bottom: 1rem;">${device.device_id}</h3>
                    <div style="color: var(--netflix-light-gray); font-size: 0.9rem;">
                        <p><strong>Provider:</strong> ${device.provider}</p>
                        <p><strong>ID Device:</strong> ${device.id_device}</p>
                        <p><strong>API Key Option:</strong> ${device.api_key_option}</p>
                    </div>
                </div>
            `).join('');
        } else {
            devicesList.innerHTML = `
                <div class="empty-state">
                    <div class="empty-state-icon">📱</div>
                    <h2 class="empty-state-title">No Devices Yet</h2>
                    <p class="empty-state-text">Click "New Device" to add your first WhatsApp device</p>
                </div>
            `;
        }
    } catch (error) {
        console.error('Load devices error:', error);
    }
}

// Only allow numbers in phone number field
function restrictToNumbers(event) {
    const input = event.target;
    input.value = input.value.replace(/[^0-9]/g, '');
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
    loadDevices();

    // Add input restriction to phone number field
    const phoneInput = document.getElementById('phoneNumber');
    if (phoneInput) {
        phoneInput.addEventListener('input', restrictToNumbers);
    }
});
