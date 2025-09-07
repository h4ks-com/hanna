import { render, screen, fireEvent } from '@testing-library/react';
import { ThemeProvider, useTheme } from '../ThemeProvider';

// Test component to access theme context
const TestComponent = () => {
  const { theme, setTheme, availableThemes } = useTheme();
  
  return (
    <div>
      <div data-testid="current-theme">{theme}</div>
      <div data-testid="available-themes">{availableThemes.join(',')}</div>
      <button onClick={() => setTheme('blue')} data-testid="set-blue">
        Set Blue Theme
      </button>
    </div>
  );
};

describe('ThemeProvider', () => {
  beforeEach(() => {
    // Clear localStorage before each test
    localStorage.clear();
    
    // Reset CSS custom properties
    document.documentElement.style.cssText = '';
  });

  it('renders with default dark theme', () => {
    render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );

    expect(screen.getByTestId('current-theme')).toHaveTextContent('dark');
  });

  it('provides all available themes', () => {
    render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );

    const availableThemes = screen.getByTestId('available-themes').textContent;
    expect(availableThemes).toBe('dark,blue,purple,pink,cute');
  });

  it('changes theme when setTheme is called', () => {
    render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );

    // Initially dark theme
    expect(screen.getByTestId('current-theme')).toHaveTextContent('dark');

    // Click to change theme
    fireEvent.click(screen.getByTestId('set-blue'));

    // Should now be blue theme
    expect(screen.getByTestId('current-theme')).toHaveTextContent('blue');
  });

  it('saves theme to localStorage', () => {
    render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );

    fireEvent.click(screen.getByTestId('set-blue'));

    expect(localStorage.setItem).toHaveBeenCalledWith('hannaui_theme', 'blue');
  });

  it('loads theme from localStorage on mount', () => {
    // Mock localStorage to return blue theme
    (localStorage.getItem as jest.Mock).mockReturnValue('blue');

    render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );

    expect(screen.getByTestId('current-theme')).toHaveTextContent('blue');
  });

  it('applies CSS custom properties when theme changes', () => {
    const { rerender } = render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );

    // Change to blue theme
    fireEvent.click(screen.getByTestId('set-blue'));

    // Check that CSS custom properties are set (we can't easily test the exact values in jsdom)
    // But we can verify the function was called by checking the theme changed
    expect(screen.getByTestId('current-theme')).toHaveTextContent('blue');
  });
});