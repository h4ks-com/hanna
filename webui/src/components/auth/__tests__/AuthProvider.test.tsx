import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { AuthProvider, useAuth } from '../AuthProvider';

// Mock the AuthService
jest.mock('../../../lib/auth', () => ({
  AuthService: {
    getCurrentUser: jest.fn(),
    loginUser: jest.fn(),
    registerUser: jest.fn(),
    logout: jest.fn(),
    getSessionId: jest.fn(() => 'test-session-id'),
    resetSessionId: jest.fn(() => 'new-session-id'),
  },
}));

// Test component to access auth context
const TestComponent = () => {
  const { 
    user, 
    isLoading, 
    isAuthenticated, 
    login, 
    register, 
    logout, 
    getSessionId 
  } = useAuth();
  
  return (
    <div>
      <div data-testid="loading">{isLoading.toString()}</div>
      <div data-testid="authenticated">{isAuthenticated.toString()}</div>
      <div data-testid="username">{user?.username || 'none'}</div>
      <div data-testid="session-id">{getSessionId()}</div>
      <button onClick={() => login('test', 'pass')} data-testid="login-btn">
        Login
      </button>
      <button onClick={() => register('test', 'test@example.com', 'password')} data-testid="register-btn">
        Register
      </button>
      <button onClick={logout} data-testid="logout-btn">
        Logout
      </button>
    </div>
  );
};

describe('AuthProvider', () => {
  const mockAuthService = require('../../../lib/auth').AuthService;

  beforeEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  it('renders with initial loading state then false after mount', async () => {
    mockAuthService.getCurrentUser.mockReturnValue(null);

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    // After useEffect runs, loading should be false
    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
    });
  });

  it('sets user when getCurrentUser returns a user', async () => {
    const mockUser = { username: 'testuser', email: 'test@example.com' };
    mockAuthService.getCurrentUser.mockReturnValue(mockUser);

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
    });

    expect(screen.getByTestId('authenticated')).toHaveTextContent('true');
    expect(screen.getByTestId('username')).toHaveTextContent('testuser');
  });

  it('handles login success', async () => {
    mockAuthService.getCurrentUser.mockReturnValue(null);
    mockAuthService.loginUser.mockResolvedValue({
      success: true,
      user: { username: 'testuser', email: 'test@example.com' }
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
    });

    fireEvent.click(screen.getByTestId('login-btn'));

    await waitFor(() => {
      expect(screen.getByTestId('authenticated')).toHaveTextContent('true');
    });

    expect(screen.getByTestId('username')).toHaveTextContent('testuser');
  });

  it('handles login failure', async () => {
    mockAuthService.getCurrentUser.mockReturnValue(null);
    mockAuthService.loginUser.mockResolvedValue({
      success: false,
      error: 'Invalid credentials'
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
    });

    fireEvent.click(screen.getByTestId('login-btn'));

    await waitFor(() => {
      expect(screen.getByTestId('authenticated')).toHaveTextContent('false');
    });

    expect(screen.getByTestId('username')).toHaveTextContent('none');
  });

  it('handles registration success', async () => {
    mockAuthService.getCurrentUser.mockReturnValue(null);
    mockAuthService.registerUser.mockResolvedValue({ success: true });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
    });

    fireEvent.click(screen.getByTestId('register-btn'));

    await waitFor(() => {
      expect(mockAuthService.registerUser).toHaveBeenCalledWith(
        'test',
        'test@example.com',
        'password'
      );
    });
  });

  it('handles logout', async () => {
    const mockUser = { username: 'testuser', email: 'test@example.com' };
    mockAuthService.getCurrentUser.mockReturnValue(mockUser);

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('authenticated')).toHaveTextContent('true');
    });

    fireEvent.click(screen.getByTestId('logout-btn'));

    expect(mockAuthService.logout).toHaveBeenCalled();
    expect(screen.getByTestId('authenticated')).toHaveTextContent('false');
    expect(screen.getByTestId('username')).toHaveTextContent('none');
  });

  it('provides session ID', async () => {
    mockAuthService.getCurrentUser.mockReturnValue(null);

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false');
    });

    expect(screen.getByTestId('session-id')).toHaveTextContent('test-session-id');
  });
});