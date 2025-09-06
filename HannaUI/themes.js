// Theme definitions
const themes = {
    dark: {
        '--primary-color': '#6366f1',
        '--secondary-color': '#8b5cf6',
        '--accent-color': '#06b6d4',
        '--bg-primary': '#0f0f23',
        '--bg-secondary': '#1a1a2e',
        '--bg-tertiary': '#16213e',
        '--text-primary': '#e2e8f0',
        '--text-secondary': '#94a3b8',
        '--text-muted': '#64748b',
        '--border-color': '#334155',
        '--orb-color-1': '#6366f1',
        '--orb-color-2': '#8b5cf6',
        '--orb-color-3': '#06b6d4',
        '--orb-glow': 'rgba(99, 102, 241, 0.3)',
        '--chat-bg': 'rgba(26, 26, 46, 0.8)',
        '--input-bg': '#1e293b',
        '--message-user-bg': '#6366f1',
        '--message-ai-bg': '#374151'
    },
    blue: {
        '--primary-color': '#3b82f6',
        '--secondary-color': '#1d4ed8',
        '--accent-color': '#0ea5e9',
        '--bg-primary': '#0c1223',
        '--bg-secondary': '#1e293b',
        '--bg-tertiary': '#334155',
        '--text-primary': '#f1f5f9',
        '--text-secondary': '#cbd5e1',
        '--text-muted': '#64748b',
        '--border-color': '#475569',
        '--orb-color-1': '#3b82f6',
        '--orb-color-2': '#1d4ed8',
        '--orb-color-3': '#0ea5e9',
        '--orb-glow': 'rgba(59, 130, 246, 0.3)',
        '--chat-bg': 'rgba(30, 41, 59, 0.8)',
        '--input-bg': '#334155',
        '--message-user-bg': '#3b82f6',
        '--message-ai-bg': '#475569'
    },
    purple: {
        '--primary-color': '#7c3aed',
        '--secondary-color': '#5b21b6',
        '--accent-color': '#a855f7',
        '--bg-primary': '#1e1b33',
        '--bg-secondary': '#2d1b69',
        '--bg-tertiary': '#4c1d95',
        '--text-primary': '#f3f4f6',
        '--text-secondary': '#d1d5db',
        '--text-muted': '#9ca3af',
        '--border-color': '#6b21a8',
        '--orb-color-1': '#7c3aed',
        '--orb-color-2': '#5b21b6',
        '--orb-color-3': '#a855f7',
        '--orb-glow': 'rgba(124, 58, 237, 0.3)',
        '--chat-bg': 'rgba(45, 27, 105, 0.8)',
        '--input-bg': '#4c1d95',
        '--message-user-bg': '#7c3aed',
        '--message-ai-bg': '#6b21a8'
    },
    pink: {
        '--primary-color': '#ec4899',
        '--secondary-color': '#be185d',
        '--accent-color': '#f472b6',
        '--bg-primary': '#1f1726',
        '--bg-secondary': '#4a1a3d',
        '--bg-tertiary': '#701a75',
        '--text-primary': '#fdf2f8',
        '--text-secondary': '#f3e8ff',
        '--text-muted': '#c084fc',
        '--border-color': '#a21caf',
        '--orb-color-1': '#ec4899',
        '--orb-color-2': '#be185d',
        '--orb-color-3': '#f472b6',
        '--orb-glow': 'rgba(236, 72, 153, 0.3)',
        '--chat-bg': 'rgba(74, 26, 61, 0.8)',
        '--input-bg': '#701a75',
        '--message-user-bg': '#ec4899',
        '--message-ai-bg': '#a21caf'
    },
    cute: {
        '--primary-color': '#f472b6',
        '--secondary-color': '#ec4899',
        '--accent-color': '#c084fc',
        '--bg-primary': '#2d1b47',
        '--bg-secondary': '#4c1d95',
        '--bg-tertiary': '#7c2d92',
        '--text-primary': '#fdf4ff',
        '--text-secondary': '#f3e8ff',
        '--text-muted': '#d8b4fe',
        '--border-color': '#c084fc',
        '--orb-color-1': '#f472b6',
        '--orb-color-2': '#ec4899',
        '--orb-color-3': '#c084fc',
        '--orb-glow': 'rgba(244, 114, 182, 0.4)',
        '--chat-bg': 'rgba(76, 29, 149, 0.8)',
        '--input-bg': '#7c2d92',
        '--message-user-bg': '#f472b6',
        '--message-ai-bg': '#c084fc'
    }
};

// Theme management
class ThemeManager {
    constructor() {
        this.currentTheme = 'dark';
        this.initThemeSelector();
        this.loadSavedTheme();
    }

    initThemeSelector() {
        const themeButtons = document.querySelectorAll('.theme-btn');
        themeButtons.forEach(button => {
            button.addEventListener('click', () => {
                const theme = button.dataset.theme;
                this.setTheme(theme);
            });
        });
    }

    setTheme(themeName) {
        if (!themes[themeName]) return;

        this.currentTheme = themeName;
        const theme = themes[themeName];
        
        // Apply CSS variables
        Object.entries(theme).forEach(([property, value]) => {
            document.documentElement.style.setProperty(property, value);
        });

        // Update active button
        document.querySelectorAll('.theme-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-theme="${themeName}"]`).classList.add('active');

        // Save theme preference
        localStorage.setItem('hannaui-theme', themeName);

        // Trigger orb color update
        if (window.orbManager) {
            window.orbManager.updateOrbColors();
        }
    }

    loadSavedTheme() {
        const savedTheme = localStorage.getItem('hannaui-theme');
        if (savedTheme && themes[savedTheme]) {
            this.setTheme(savedTheme);
        } else {
            this.setTheme('dark');
        }
    }

    getCurrentTheme() {
        return this.currentTheme;
    }

    getThemeColors() {
        return themes[this.currentTheme];
    }
}

// Initialize theme manager
const themeManager = new ThemeManager();

// Make it globally available
window.themeManager = themeManager;
