import { AuthService } from '../auth';

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

describe('AuthService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.clear.mockClear();
    
    // Mock window object for Node.js environment
    Object.defineProperty(global, 'window', {
      writable: true,
      value: { localStorage: localStorageMock },
    });
  });

  describe('loadUsers', () => {
    it('returns default users when localStorage is empty', () => {
      localStorageMock.getItem.mockReturnValue(null);

      const users = AuthService.loadUsers();

      expect(users.admin).toBeDefined();
      expect(users.admin.username).toBe('admin');
      expect(users.admin.password).toBe('admin123');
      expect(users.admin.email).toBe('admin@hannaui.com');
    });

    it('returns users from localStorage when available', () => {
      const storedUsers = {
        testuser: {
          username: 'testuser',
          password: 'testpass',
          email: 'test@example.com',
          createdAt: '2024-01-01T00:00:00.000Z',
        },
      };
      
      localStorageMock.getItem.mockReturnValue(JSON.stringify(storedUsers));

      const users = AuthService.loadUsers();

      expect(users).toEqual(storedUsers);
    });

    it('handles JSON parse errors gracefully', () => {
      // Suppress console.error for this test
      const originalConsoleError = console.error;
      console.error = jest.fn();
      
      localStorageMock.getItem.mockReturnValue('invalid json');

      const users = AuthService.loadUsers();

      // Should return default users when JSON is invalid
      expect(users.admin).toBeDefined();
      
      // Restore console.error
      console.error = originalConsoleError;
    });
  });

  describe('getCurrentUser', () => {
    it('returns null when no user is stored', () => {
      localStorageMock.getItem.mockReturnValue(null);

      const user = AuthService.getCurrentUser();

      expect(user).toBeNull();
    });

    it('returns user when stored in localStorage', () => {
      const storedUser = {
        username: 'testuser',
        email: 'test@example.com',
        rememberMe: true,
      };
      
      localStorageMock.getItem.mockReturnValue(JSON.stringify(storedUser));

      const user = AuthService.getCurrentUser();

      expect(user).toEqual(storedUser);
    });

    it('handles JSON parse errors gracefully', () => {
      // Suppress console.error for this test
      const originalConsoleError = console.error;
      console.error = jest.fn();
      
      localStorageMock.getItem.mockReturnValue('invalid json');

      const user = AuthService.getCurrentUser();

      expect(user).toBeNull();
      
      // Restore console.error
      console.error = originalConsoleError;
    });
  });

  describe('validateLogin', () => {
    beforeEach(() => {
      // Mock loadUsers to return test data
      const testUsers = {
        admin: {
          username: 'admin',
          password: 'admin123',
          email: 'admin@example.com',
          createdAt: '2024-01-01T00:00:00.000Z',
        },
      };
      
      localStorageMock.getItem.mockReturnValue(JSON.stringify(testUsers));
    });

    it('returns true for valid credentials', () => {
      const isValid = AuthService.validateLogin('admin', 'admin123');

      expect(isValid).toBe(true);
    });

    it('returns false for invalid username', () => {
      const isValid = AuthService.validateLogin('nonexistent', 'admin123');

      expect(isValid).toBe(false);
    });

    it('returns false for invalid password', () => {
      const isValid = AuthService.validateLogin('admin', 'wrongpassword');

      expect(isValid).toBe(false);
    });
  });

  describe('registerUser', () => {
    beforeEach(() => {
      localStorageMock.getItem.mockReturnValue(JSON.stringify({}));
    });

    it('successfully registers a valid user', async () => {
      const result = await AuthService.registerUser(
        'newuser',
        'new@example.com',
        'password123'
      );

      expect(result.success).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('rejects username shorter than 3 characters', async () => {
      const result = await AuthService.registerUser(
        'ab',
        'new@example.com',
        'password123'
      );

      expect(result.success).toBe(false);
      expect(result.error).toBe('Username must be at least 3 characters long.');
    });

    it('rejects invalid email', async () => {
      const result = await AuthService.registerUser(
        'newuser',
        'invalid-email',
        'password123'
      );

      expect(result.success).toBe(false);
      expect(result.error).toBe('Please enter a valid email address.');
    });

    it('rejects password shorter than 6 characters', async () => {
      const result = await AuthService.registerUser(
        'newuser',
        'new@example.com',
        '123'
      );

      expect(result.success).toBe(false);
      expect(result.error).toBe('Password must be at least 6 characters long.');
    });

    it('rejects duplicate username', async () => {
      const existingUsers = {
        existing: {
          username: 'existing',
          password: 'password',
          email: 'existing@example.com',
          createdAt: '2024-01-01T00:00:00.000Z',
        },
      };
      
      localStorageMock.getItem.mockReturnValue(JSON.stringify(existingUsers));

      const result = await AuthService.registerUser(
        'existing',
        'new@example.com',
        'password123'
      );

      expect(result.success).toBe(false);
      expect(result.error).toBe('Username already exists. Please choose a different one.');
    });
  });

  describe('loginUser', () => {
    beforeEach(() => {
      const testUsers = {
        admin: {
          username: 'admin',
          password: 'admin123',
          email: 'admin@example.com',
          createdAt: '2024-01-01T00:00:00.000Z',
        },
      };
      
      localStorageMock.getItem.mockReturnValue(JSON.stringify(testUsers));
    });

    it('successfully logs in valid user', async () => {
      const result = await AuthService.loginUser('admin', 'admin123', true);

      expect(result.success).toBe(true);
      expect(result.user).toEqual({
        username: 'admin',
        email: 'admin@example.com',
        rememberMe: true,
      });
    });

    it('fails for invalid credentials', async () => {
      const result = await AuthService.loginUser('admin', 'wrongpassword');

      expect(result.success).toBe(false);
      expect(result.error).toBe('Invalid username or password. Please try again.');
    });
  });

  describe('isValidEmail', () => {
    it('returns true for valid email addresses', () => {
      expect(AuthService.isValidEmail('test@example.com')).toBe(true);
      expect(AuthService.isValidEmail('user+tag@domain.co.uk')).toBe(true);
    });

    it('returns false for invalid email addresses', () => {
      expect(AuthService.isValidEmail('invalid')).toBe(false);
      expect(AuthService.isValidEmail('@example.com')).toBe(false);
      expect(AuthService.isValidEmail('test@')).toBe(false);
      expect(AuthService.isValidEmail('test.example.com')).toBe(false);
    });
  });

  describe('getSessionId', () => {
    it('returns existing session ID from localStorage', () => {
      localStorageMock.getItem.mockReturnValue('existing-session-id');

      const sessionId = AuthService.getSessionId();

      expect(sessionId).toBe('existing-session-id');
    });

    it('generates new session ID when none exists', () => {
      localStorageMock.getItem.mockReturnValue(null);

      const sessionId = AuthService.getSessionId();

      expect(sessionId).toMatch(/^session_\d+_[a-z0-9]+$/);
      expect(localStorageMock.setItem).toHaveBeenCalledWith('hanna_session_id', sessionId);
    });
  });

  describe('resetSessionId', () => {
    it('removes old session and generates new one', () => {
      const newSessionId = AuthService.resetSessionId();

      expect(localStorageMock.removeItem).toHaveBeenCalledWith('hanna_session_id');
      expect(newSessionId).toMatch(/^session_\d+_[a-z0-9]+$/);
    });
  });
});