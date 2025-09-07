'use client';

import { useTheme } from './ThemeProvider';
import { useAuth } from '@/components/auth/AuthProvider';
import { ThemeType } from '@/types';

const themeIcons: Record<ThemeType, string> = {
  dark: '🌙',
  blue: '💙',
  purple: '💜',
  pink: '🌸',
  cute: '🦄'
};

export default function ThemeSelector() {
  const { theme, setTheme, availableThemes } = useTheme();
  const { logout, isAuthenticated } = useAuth();

  const handleLogout = () => {
    if (confirm('Are you sure you want to logout?')) {
      logout();
    }
  };

  return (
    <div className="theme-selector fixed top-4 right-4 flex items-center space-x-2 z-50">
      {availableThemes.map((themeType) => (
        <button
          key={themeType}
          onClick={() => setTheme(themeType)}
          className={`theme-btn w-10 h-10 rounded-full backdrop-blur-sm transition-all duration-200 hover:scale-110 ${
            theme === themeType
              ? 'bg-white/30 ring-2 ring-white/50 shadow-lg'
              : 'bg-black/20 hover:bg-white/20'
          }`}
          title={`Switch to ${themeType} theme`}
        >
          <span className="text-lg">{themeIcons[themeType]}</span>
        </button>
      ))}
      
      {isAuthenticated && (
        <button
          onClick={handleLogout}
          className="theme-btn w-10 h-10 rounded-full bg-black/20 hover:bg-red-500/20 backdrop-blur-sm transition-all duration-200 hover:scale-110"
          title="Logout"
        >
          <span className="text-lg">🚪</span>
        </button>
      )}
    </div>
  );
}