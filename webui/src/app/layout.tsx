import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import { AuthProvider } from '@/components/auth/AuthProvider';
import { ThemeProvider } from '@/components/ui/ThemeProvider';
import { validateEnv, clientEnv } from '@/lib/env';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'HannaUI - AI Chat Interface',
  description: 'Beautiful AI chat interface with orb animations and real-time streaming',
  keywords: ['AI', 'chat', 'interface', 'n8n', 'streaming'],
};

// Validate environment variables
try {
  validateEnv();
} catch (error) {
  console.error('Environment validation failed:', error);
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        {/* Markdown rendering */}
        <script src="https://cdn.jsdelivr.net/npm/marked@12.0.0/marked.min.js" async />
        {/* Syntax highlighting */}
        <link 
          rel="stylesheet" 
          href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css" 
        />
        <script 
          src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js" 
          async 
        />
      </head>
      <body className={inter.className}>
        <ThemeProvider>
          <AuthProvider>
            {children}
          </AuthProvider>
        </ThemeProvider>
        
        {/* Client-side environment for debugging */}
        <script
          dangerouslySetInnerHTML={{
            __html: `window.ENV = ${JSON.stringify(clientEnv)};`
          }}
        />
      </body>
    </html>
  );
}
