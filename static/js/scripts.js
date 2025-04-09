// Global JavaScript functionality

document.addEventListener('DOMContentLoaded', function() {
    // Automatically hide alerts after 5 seconds
    const alerts = document.querySelectorAll('.alert');
    alerts.forEach(alert => {
        setTimeout(() => {
            alert.style.opacity = '0';
            setTimeout(() => {
                alert.style.display = 'none';
            }, 500);
        }, 5000);
    });

    // Add field type change handler to show/hide options fields
    const fieldTypeSelects = document.querySelectorAll('select[name="field_type"]');
    fieldTypeSelects.forEach(select => {
        select.addEventListener('change', function() {
            const optionsGroup = this.closest('form').querySelector('.options-group');
            if (!optionsGroup) return;
            
            const fieldType = this.value;
            if (fieldType === 'select' || fieldType === 'radio' || fieldType === 'checkbox') {
                optionsGroup.classList.remove('hidden');
            } else {
                optionsGroup.classList.add('hidden');
            }
        });
    });
});

// Utility functions

// Generate a random string for IDs
function generateRandomId(length = 8) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    let result = '';
    for (let i = 0; i < length; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return result;
}

// Format a date to YYYY-MM-DD
function formatDate(date) {
    const d = new Date(date);
    const month = (d.getMonth() + 1).toString().padStart(2, '0');
    const day = d.getDate().toString().padStart(2, '0');
    const year = d.getFullYear();
    return `${year}-${month}-${day}`;
}

// Copy text to clipboard
function copyToClipboard(text) {
    const el = document.createElement('textarea');
    el.value = text;
    el.setAttribute('readonly', '');
    el.style.position = 'absolute';
    el.style.left = '-9999px';
    document.body.appendChild(el);
    el.select();
    document.execCommand('copy');
    document.body.removeChild(el);
}

// Add copy button functionality where needed
document.addEventListener('DOMContentLoaded', function() {
    const copyButtons = document.querySelectorAll('.copy-button');
    copyButtons.forEach(button => {
        button.addEventListener('click', function() {
            const textToCopy = this.getAttribute('data-copy');
            copyToClipboard(textToCopy);
            
            // Show feedback
            const originalText = this.textContent;
            this.textContent = 'Copied!';
            setTimeout(() => {
                this.textContent = originalText;
            }, 2000);
        });
    });
});