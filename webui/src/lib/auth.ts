import { User } from '@/types';

// Local storage keys
const USERS_KEY = 'hannaui_users';
const CURRENT_USER_KEY = 'hannaui_current_user';

// Default users
const DEFAULT_USERS = {
  admin: {
    username: 'admin',
    password: 'admin123',
    email: 'admin@hannaui.com',
    createdAt: new Date().toISOString()
  }
};

export interface StoredUser {
  username: string;
  password: string;
  email: string;
  createdAt: string;
}

export class AuthService {
  // Load users from localStorage
  static loadUsers(): Record<string, StoredUser> {
    if (typeof window === 'undefined') return {};
    
    try {
      const users = localStorage.getItem(USERS_KEY);
      if (users) {
        return JSON.parse(users);
      } else {
        // Initialize with default users
        this.saveUsers(DEFAULT_USERS);
        return DEFAULT_USERS;
      }
    } catch (error) {
      console.error('Error loading users:', error);
      return DEFAULT_USERS;
    }
  }

  // Save users to localStorage
  static saveUsers(users: Record<string, StoredUser>): void {
    if (typeof window === 'undefined') return;
    
    try {
      localStorage.setItem(USERS_KEY, JSON.stringify(users));
    } catch (error) {
      console.error('Error saving users:', error);
    }
  }

  // Get current logged-in user
  static getCurrentUser(): User | null {
    if (typeof window === 'undefined') return null;
    
    try {
      const user = localStorage.getItem(CURRENT_USER_KEY);
      return user ? JSON.parse(user) : null;
    } catch (error) {
      console.error('Error getting current user:', error);
      return null;
    }
  }

  // Set current user
  static setCurrentUser(user: User): void {
    if (typeof window === 'undefined') return;
    
    try {
      localStorage.setItem(CURRENT_USER_KEY, JSON.stringify(user));
    } catch (error) {
      console.error('Error setting current user:', error);
    }
  }

  // Clear current user (logout)
  static clearCurrentUser(): void {
    if (typeof window === 'undefined') return;
    
    try {
      localStorage.removeItem(CURRENT_USER_KEY);
    } catch (error) {
      console.error('Error clearing current user:', error);
    }
  }

  // Validate login credentials
  static validateLogin(username: string, password: string): boolean {
    const users = this.loadUsers();
    const user = users[username];
    return !!(user && user.password === password);
  }

  // Register new user
  static async registerUser(
    username: string, 
    email: string, 
    password: string
  ): Promise<{ success: boolean; error?: string }> {
    if (username.length < 3) {
      return { success: false, error: 'Username must be at least 3 characters long.' };
    }

    if (!this.isValidEmail(email)) {
      return { success: false, error: 'Please enter a valid email address.' };
    }

    if (password.length < 6) {
      return { success: false, error: 'Password must be at least 6 characters long.' };
    }

    const users = this.loadUsers();
    
    if (users[username]) {
      return { success: false, error: 'Username already exists. Please choose a different one.' };
    }

    // Add new user
    users[username] = {
      username,
      email,
      password,
      createdAt: new Date().toISOString()
    };

    this.saveUsers(users);
    return { success: true };
  }

  // Validate email format
  static isValidEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emailRegex.test(email);
  }

  // Login user
  static async loginUser(
    username: string, 
    password: string, 
    rememberMe: boolean = false
  ): Promise<{ success: boolean; user?: User; error?: string }> {
    if (!this.validateLogin(username, password)) {
      return { success: false, error: 'Invalid username or password. Please try again.' };
    }

    const users = this.loadUsers();
    const storedUser = users[username];
    
    const user: User = {
      username: storedUser.username,
      email: storedUser.email,
      rememberMe
    };

    this.setCurrentUser(user);
    return { success: true, user };
  }

  // Check if user is authenticated
  static isAuthenticated(): boolean {
    return this.getCurrentUser() !== null;
  }

  // Logout user
  static logout(): void {
    this.clearCurrentUser();
  }

  // Get session ID for chat
  static getSessionId(): string {
    if (typeof window === 'undefined') return 'server_session';
    
    let sessionId = localStorage.getItem('hanna_session_id');
    if (!sessionId) {
      sessionId = 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
      localStorage.setItem('hanna_session_id', sessionId);
    }
    return sessionId;
  }

  // Reset session ID
  static resetSessionId(): string {
    if (typeof window === 'undefined') return 'server_session';
    
    localStorage.removeItem('hanna_session_id');
    return this.getSessionId();
  }
}