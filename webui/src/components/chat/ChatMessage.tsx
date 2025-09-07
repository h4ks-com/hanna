'use client';

import { Message } from '@/types';
import { marked } from 'marked';
import { useEffect, useRef } from 'react';

interface ChatMessageProps {
  message: Message;
}

export default function ChatMessage({ message }: ChatMessageProps) {
  const messageRef = useRef<HTMLDivElement>(null);

  // Apply syntax highlighting after render
  useEffect(() => {
    if (messageRef.current && typeof window !== 'undefined') {
      // Only apply highlighting if hljs is available
      if ((window as any).hljs) {
        const codeBlocks = messageRef.current.querySelectorAll('pre code');
        codeBlocks.forEach((block) => {
          (window as any).hljs.highlightElement(block as HTMLElement);
        });
      }
    }
  }, [message.text]);

  const formatTime = (timestamp: Date) => {
    const now = new Date();
    const diff = now.getTime() - timestamp.getTime();
    
    if (diff < 60000) { // Less than 1 minute
      return 'Now';
    } else if (diff < 3600000) { // Less than 1 hour
      const minutes = Math.floor(diff / 60000);
      return `${minutes}m ago`;
    } else if (diff < 86400000) { // Less than 1 day
      const hours = Math.floor(diff / 3600000);
      return `${hours}h ago`;
    } else {
      return timestamp.toLocaleDateString();
    }
  };

  const renderContent = () => {
    try {
      // Render markdown
      const html = marked.parse(message.text);
      return { __html: html };
    } catch (error) {
      console.error('Error parsing markdown:', error);
      return { __html: message.text };
    }
  };

  return (
    <div className={`flex items-start space-x-3 ${message.type === 'user' ? 'flex-row-reverse space-x-reverse' : ''}`}>
      {/* Avatar */}
      <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${
        message.type === 'ai' 
          ? 'bg-gradient-to-r from-purple-500 to-pink-500' 
          : 'bg-gradient-to-r from-blue-500 to-cyan-500'
      }`}>
        <span className="text-white text-sm">
          {message.type === 'ai' ? '🤖' : '👤'}
        </span>
      </div>
      
      {/* Message content */}
      <div className={`flex-1 max-w-[80%] ${message.type === 'user' ? 'text-right' : ''}`}>
        <div 
          ref={messageRef}
          className={`rounded-lg p-3 backdrop-blur-sm ${
            message.type === 'ai'
              ? 'bg-white/10 text-white'
              : 'bg-blue-600/80 text-white ml-auto'
          }`}
          style={{ 
            animation: 'fadeInUp 0.3s ease-out',
            maxWidth: message.type === 'user' ? 'fit-content' : '100%'
          }}
        >
          <div 
            className="message-text prose prose-invert max-w-none"
            dangerouslySetInnerHTML={renderContent()}
          />
          <div className={`text-xs mt-2 opacity-70 ${
            message.type === 'user' ? 'text-right' : 'text-left'
          }`}>
            {formatTime(message.timestamp)}
          </div>
        </div>
      </div>
    </div>
  );
}

// Add fadeInUp animation styles globally
if (typeof window !== 'undefined') {
  const style = document.createElement('style');
  style.textContent = `
    @keyframes fadeInUp {
      from {
        opacity: 0;
        transform: translateY(20px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }
    
    .prose h1, .prose h2, .prose h3, .prose h4, .prose h5, .prose h6 {
      color: inherit;
      margin-top: 1em;
      margin-bottom: 0.5em;
    }
    
    .prose p {
      margin-top: 0.5em;
      margin-bottom: 0.5em;
    }
    
    .prose ul, .prose ol {
      margin-top: 0.5em;
      margin-bottom: 0.5em;
    }
    
    .prose pre {
      background-color: rgba(0, 0, 0, 0.5);
      border-radius: 0.5rem;
      padding: 1rem;
      overflow-x: auto;
    }
    
    .prose code {
      background-color: rgba(255, 255, 255, 0.1);
      padding: 0.25rem 0.5rem;
      border-radius: 0.25rem;
      font-size: 0.9em;
    }
    
    .prose pre code {
      background-color: transparent;
      padding: 0;
    }
    
    .prose blockquote {
      border-left: 4px solid rgba(255, 255, 255, 0.3);
      padding-left: 1rem;
      margin: 1rem 0;
      font-style: italic;
    }
    
    .prose table {
      border-collapse: collapse;
      width: 100%;
      margin: 1rem 0;
    }
    
    .prose th, .prose td {
      border: 1px solid rgba(255, 255, 255, 0.2);
      padding: 0.5rem;
      text-align: left;
    }
    
    .prose th {
      background-color: rgba(255, 255, 255, 0.1);
      font-weight: 600;
    }
  `;
  document.head.appendChild(style);
}