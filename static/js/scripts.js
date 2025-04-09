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

// Form editor functionality
document.addEventListener('DOMContentLoaded', function() {
    // Handle adding new fields
    const addFieldButtons = document.querySelectorAll('.add-field-button');
    addFieldButtons.forEach(button => {
        button.addEventListener('click', function(e) {
            // Prevent form submission if it's a button inside a form
            if (this.closest('form')) {
                e.preventDefault();
            }
            
            const step = this.dataset.step || this.closest('form')?.querySelector('input[name="step"]')?.value || "1";
            const modal = document.getElementById('field-modal');
            const stepInput = document.getElementById('step');
            stepInput.value = step;
            modal.classList.remove('hidden');
        });
    });

    // Handle adding new steps
    const addStepButton = document.querySelector('.tab-button[data-step="new"]');
    if (addStepButton) {
        addStepButton.addEventListener('click', function() {
            const formId = document.querySelector('input[name="form_id"]').value;
            const form = document.createElement('form');
            form.method = 'post';
            form.action = '/forms/update';
            form.enctype = 'application/x-www-form-urlencoded';
            
            const formIdInput = document.createElement('input');
            formIdInput.type = 'hidden';
            formIdInput.name = 'form_id';
            formIdInput.value = formId;
            
            const actionInput = document.createElement('input');
            actionInput.type = 'hidden';
            actionInput.name = 'action';
            actionInput.value = 'add_step';
            
            form.appendChild(formIdInput);
            form.appendChild(actionInput);
            document.body.appendChild(form);
            form.submit();
        });
    }

    // Handle tab switching
    const tabButtons = document.querySelectorAll('.tab-button:not([data-step="new"])');
    tabButtons.forEach(button => {
        button.addEventListener('click', function() {
            const step = this.dataset.step;
            // Remove active class from all tabs
            tabButtons.forEach(btn => btn.classList.remove('active'));
            // Add active class to clicked tab
            this.classList.add('active');
            // Show/hide step content
            document.querySelectorAll('.step-fields').forEach(content => {
                content.style.display = content.dataset.step === step ? 'block' : 'none';
            });
        });
    });

    // Handle modal close
    const closeModal = document.querySelector('.close-modal');
    const closeModalButton = document.querySelector('.close-modal-button');
    const modal = document.getElementById('field-modal');

    if (closeModal) {
        closeModal.addEventListener('click', function() {
            modal.classList.add('hidden');
        });
    }

    if (closeModalButton) {
        closeModalButton.addEventListener('click', function() {
            modal.classList.add('hidden');
        });
    }

    // Handle field type change to show/hide options
    const fieldTypeSelect = document.getElementById('field_type');
    if (fieldTypeSelect) {
        fieldTypeSelect.addEventListener('change', function() {
            const optionsGroup = document.querySelector('.options-group');
            const fieldTypes = ['select', 'radio', 'checkbox'];
            if (fieldTypes.includes(this.value)) {
                optionsGroup.classList.remove('hidden');
            } else {
                optionsGroup.classList.add('hidden');
            }
        });
    }
});