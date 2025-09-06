// Authentication system for HannaUI
class AuthManager {
    constructor() {
        this.users = this.loadUsers();
        this.currentUser = this.getCurrentUser();
        this.initEventListeners();
        this.checkAuthStatus();
    }

    // Load users from localStorage (simulating a database)
    loadUsers() {
        const users = localStorage.getItem('hannaui_users');
        if (users) {
            return JSON.parse(users);
        } else {
            // Initialize with default admin user
            const defaultUsers = {
                'admin': {
                    username: 'admin',
                    password: 'admin123',
                    email: 'admin@hannaui.com',
                    createdAt: new Date().toISOString()
                }
            };
            this.saveUsers(defaultUsers);
            return defaultUsers;
        }
    }

    // Save users to localStorage
    saveUsers(users) {
        localStorage.setItem('hannaui_users', JSON.stringify(users));
    }

    // Get current logged-in user
    getCurrentUser() {
        const user = localStorage.getItem('hannaui_current_user');
        return user ? JSON.parse(user) : null;
    }

    // Set current user
    setCurrentUser(user) {
        localStorage.setItem('hannaui_current_user', JSON.stringify(user));
        this.currentUser = user;
    }

    // Clear current user (logout)
    clearCurrentUser() {
        localStorage.removeItem('hannaui_current_user');
        this.currentUser = null;
    }

    // Check if user is authenticated
    isAuthenticated() {
        return this.currentUser !== null;
    }

    // Redirect to appropriate page based on auth status
    checkAuthStatus() {
        const currentPage = window.location.pathname.split('/').pop();
        
        if (this.isAuthenticated()) {
            // User is logged in
            if (currentPage === 'login.html' || currentPage === 'register.html') {
                // Redirect to main chat if on auth pages
                window.location.href = 'index.html';
            }
        } else {
            // User is not logged in
            if (currentPage === 'index.html' || currentPage === '' || currentPage === '/') {
                // Redirect to login if trying to access chat
                window.location.href = 'login.html';
            }
        }
    }

    // Initialize event listeners
    initEventListeners() {
        // Login form
        const loginForm = document.getElementById('loginForm');
        if (loginForm) {
            loginForm.addEventListener('submit', (e) => this.handleLogin(e));
        }

        // Register form
        const registerForm = document.getElementById('registerForm');
        if (registerForm) {
            registerForm.addEventListener('submit', (e) => this.handleRegister(e));
        }

        // Password toggle buttons
        this.initPasswordToggle('passwordToggle', 'password');
        this.initPasswordToggle('passwordToggle1', 'newPassword');
        this.initPasswordToggle('passwordToggle2', 'confirmPassword');

        // Real-time validation for registration
        const newUsername = document.getElementById('newUsername');
        const newPassword = document.getElementById('newPassword');
        const confirmPassword = document.getElementById('confirmPassword');

        if (newUsername) {
            newUsername.addEventListener('input', () => this.validateUsername());
        }
        if (newPassword) {
            newPassword.addEventListener('input', () => this.validatePassword());
        }
        if (confirmPassword) {
            confirmPassword.addEventListener('input', () => this.validatePasswordMatch());
        }
    }

    // Initialize password toggle functionality
    initPasswordToggle(buttonId, inputId) {
        const toggleBtn = document.getElementById(buttonId);
        const passwordInput = document.getElementById(inputId);
        
        if (toggleBtn && passwordInput) {
            toggleBtn.addEventListener('click', () => {
                const isPassword = passwordInput.type === 'password';
                passwordInput.type = isPassword ? 'text' : 'password';
                
                // Update icon
                const svg = toggleBtn.querySelector('svg');
                if (isPassword) {
                    svg.innerHTML = `
                        <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
                        <line x1="1" y1="1" x2="23" y2="23"></line>
                    `;
                } else {
                    svg.innerHTML = `
                        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                        <circle cx="12" cy="12" r="3"></circle>
                    `;
                }
            });
        }
    }

    // Handle login form submission
    async handleLogin(e) {
        e.preventDefault();
        
        const username = document.getElementById('username').value.trim();
        const password = document.getElementById('password').value;
        const rememberMe = document.getElementById('rememberMe').checked;
        const loginBtn = document.getElementById('loginBtn');
        
        // Show loading state
        this.setButtonLoading(loginBtn, true);
        this.hideMessage('errorMessage');
        
        // Simulate network delay
        await this.delay(1000);
        
        // Validate credentials
        if (this.validateLogin(username, password)) {
            const user = this.users[username];
            this.setCurrentUser({
                username: user.username,
                email: user.email,
                rememberMe: rememberMe
            });
            
            // Redirect to chat
            window.location.href = 'index.html';
        } else {
            this.showError('Invalid username or password. Please try again.');
            this.setButtonLoading(loginBtn, false);
        }
    }

    // Handle registration form submission
    async handleRegister(e) {
        e.preventDefault();
        
        const username = document.getElementById('newUsername').value.trim();
        const email = document.getElementById('email').value.trim();
        const password = document.getElementById('newPassword').value;
        const confirmPassword = document.getElementById('confirmPassword').value;
        const acceptTerms = document.getElementById('acceptTerms').checked;
        const registerBtn = document.getElementById('registerBtn');
        
        // Validate form
        if (!this.validateRegistrationForm(username, email, password, confirmPassword, acceptTerms)) {
            return;
        }
        
        // Show loading state
        this.setButtonLoading(registerBtn, true);
        this.hideMessage('errorMessage');
        
        // Simulate network delay
        await this.delay(1500);
        
        // Check if user already exists
        if (this.users[username]) {
            this.showError('Username already exists. Please choose a different one.');
            this.setButtonLoading(registerBtn, false);
            return;
        }
        
        // Create new user
        this.users[username] = {
            username: username,
            email: email,
            password: password,
            createdAt: new Date().toISOString()
        };
        
        this.saveUsers(this.users);
        
        // Show success message
        this.showSuccess('Account created successfully! Redirecting to login...');
        
        // Redirect to login after delay
        setTimeout(() => {
            window.location.href = 'login.html';
        }, 2000);
    }

    // Validate login credentials
    validateLogin(username, password) {
        const user = this.users[username];
        return user && user.password === password;
    }

    // Validate registration form
    validateRegistrationForm(username, email, password, confirmPassword, acceptTerms) {
        let isValid = true;
        
        // Reset previous errors
        this.hideMessage('errorMessage');
        
        // Validate username
        if (username.length < 3) {
            this.showError('Username must be at least 3 characters long.');
            isValid = false;
        }
        
        // Validate email
        if (!this.isValidEmail(email)) {
            this.showError('Please enter a valid email address.');
            isValid = false;
        }
        
        // Validate password
        if (password.length < 6) {
            this.showError('Password must be at least 6 characters long.');
            isValid = false;
        }
        
        // Validate password match
        if (password !== confirmPassword) {
            this.showError('Passwords do not match.');
            isValid = false;
        }
        
        // Validate terms acceptance
        if (!acceptTerms) {
            this.showError('You must accept the Terms of Service.');
            isValid = false;
        }
        
        return isValid;
    }

    // Real-time username validation
    validateUsername() {
        const input = document.getElementById('newUsername');
        const value = input.value.trim();
        
        if (value.length === 0) {
            this.setInputState(input, 'default');
        } else if (value.length < 3) {
            this.setInputState(input, 'error');
        } else if (this.users[value]) {
            this.setInputState(input, 'error');
        } else {
            this.setInputState(input, 'success');
        }
    }

    // Real-time password validation
    validatePassword() {
        const input = document.getElementById('newPassword');
        const value = input.value;
        
        if (value.length === 0) {
            this.setInputState(input, 'default');
        } else if (value.length < 6) {
            this.setInputState(input, 'error');
        } else {
            this.setInputState(input, 'success');
        }
        
        // Also validate password match if confirm password has value
        const confirmPassword = document.getElementById('confirmPassword');
        if (confirmPassword && confirmPassword.value) {
            this.validatePasswordMatch();
        }
    }

    // Real-time password match validation
    validatePasswordMatch() {
        const password = document.getElementById('newPassword').value;
        const confirmPassword = document.getElementById('confirmPassword');
        const value = confirmPassword.value;
        
        if (value.length === 0) {
            this.setInputState(confirmPassword, 'default');
        } else if (value !== password) {
            this.setInputState(confirmPassword, 'error');
        } else {
            this.setInputState(confirmPassword, 'success');
        }
    }

    // Set input validation state
    setInputState(input, state) {
        input.classList.remove('error', 'success');
        if (state !== 'default') {
            input.classList.add(state);
        }
    }

    // Validate email format
    isValidEmail(email) {
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return emailRegex.test(email);
    }

    // Show error message
    showError(message) {
        const errorElement = document.getElementById('errorMessage');
        if (errorElement) {
            errorElement.querySelector('.error-text').textContent = message;
            errorElement.style.display = 'flex';
        }
    }

    // Show success message
    showSuccess(message) {
        const successElement = document.getElementById('successMessage');
        if (successElement) {
            successElement.querySelector('.success-text').textContent = message;
            successElement.style.display = 'flex';
        }
    }

    // Hide message
    hideMessage(elementId) {
        const element = document.getElementById(elementId);
        if (element) {
            element.style.display = 'none';
        }
    }

    // Set button loading state
    setButtonLoading(button, loading) {
        if (loading) {
            button.classList.add('loading');
            button.disabled = true;
        } else {
            button.classList.remove('loading');
            button.disabled = false;
        }
    }

    // Utility function for delays
    delay(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    // Logout function (can be called from main app)
    logout() {
        this.clearCurrentUser();
        window.location.href = 'login.html';
    }

    // Get user info for display
    getUserInfo() {
        return this.currentUser;
    }
}

// Initialize authentication manager
const authManager = new AuthManager();

// Make it globally available
window.authManager = authManager;

// Add logout functionality to main chat if needed
if (window.location.pathname.includes('index.html') || window.location.pathname === '/') {
    // Add logout option to chat interface
    setTimeout(() => {
        const themeSelector = document.querySelector('.theme-selector');
        if (themeSelector && authManager.isAuthenticated()) {
            const logoutBtn = document.createElement('button');
            logoutBtn.className = 'theme-btn';
            logoutBtn.title = 'Logout';
            logoutBtn.innerHTML = '🚪';
            logoutBtn.addEventListener('click', () => {
                if (confirm('Are you sure you want to logout?')) {
                    authManager.logout();
                }
            });
            themeSelector.appendChild(logoutBtn);
        }
    }, 1000);
}
