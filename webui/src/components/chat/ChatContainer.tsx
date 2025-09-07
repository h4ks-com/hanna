'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { Message, ChatEndpointInfo, AnimationType } from '@/types';
import { ChatService } from '@/lib/chat';
import { useAuth } from '@/components/auth/AuthProvider';
import { marked } from 'marked';
import ChatMessage from './ChatMessage';
import ChatInput from './ChatInput';
import ConnectionStatus from './ConnectionStatus';

interface ChatContainerProps {
  onAnimationTrigger?: (animation: AnimationType) => void;
}

export default function ChatContainer({ onAnimationTrigger }: ChatContainerProps) {
  const { user, getSessionId } = useAuth();
  const [messages, setMessages] = useState<Message[]>([]);
  const [isTyping, setIsTyping] = useState(false);
  const [connectionInfo, setConnectionInfo] = useState<ChatEndpointInfo>({
    endpoint: null,
    status: 'checking',
    configured: false
  });
  
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);

  // Configure marked for markdown rendering
  useEffect(() => {
    marked.setOptions({
      breaks: true,
      gfm: true,
      //@ts-ignore - marked types issue
      headerIds: false,
      mangle: false
    });
  }, []);

  // Check connection status on mount
  useEffect(() => {
    const checkConnection = async () => {
      const info = await ChatService.checkConnection();
      setConnectionInfo(info);
    };
    
    checkConnection();
  }, []);

  // Add welcome message
  useEffect(() => {
    const username = user?.username || 'Guest';
    const welcomeMessage: Message = {
      type: 'ai',
      text: `Hello **${username}**! I'm **Hanna**, your AI assistant. ✨`,
      timestamp: new Date()
    };
    
    setMessages([welcomeMessage]);

    // Add tips after a moment
    setTimeout(() => {
      let tipMessage: Message;
      
      if (connectionInfo.status === 'connected') {
        tipMessage = {
          type: 'ai',
          text: `🌟 **Connected to live AI!** I'm ready to help you with anything you need.

💡 **Features available:**
- Full markdown support in our conversation
- Real-time AI responses
- Press **Ctrl+C** for orb animations
- Switch themes anytime

What would you like to talk about?`,
          timestamp: new Date()
        };
      } else {
        tipMessage = {
          type: 'ai',
          text: `💡 **Demo Mode Features:**
- Full **markdown** formatting support
- Interactive orb animations (**Ctrl+C**)
- Theme switching
- Beautiful chat interface

*Ask me about "n8n connection" to learn how to connect to the real Hanna AI!*`,
          timestamp: new Date()
        };
      }
      
      setMessages(prev => [...prev, tipMessage]);
    }, 2000);
  }, [user, connectionInfo.status]);

  // Auto-scroll to bottom
  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  // Send message handler
  const handleSendMessage = async (text: string) => {
    if (!text.trim() || isTyping) return;

    // Add user message
    const userMessage: Message = {
      type: 'user',
      text: text,
      timestamp: new Date()
    };
    
    setMessages(prev => [...prev, userMessage]);

    // Start typing animation
    setIsTyping(true);
    onAnimationTrigger?.('network');

    try {
      let aiResponse: string;
      
      if (connectionInfo.status === 'connected' && connectionInfo.endpoint) {
        // Try to get response from n8n endpoint
        const sessionId = getSessionId();
        const request = {
          message: text,
          sessionId,
          user: {
            username: user?.username || 'anonymous',
            email: user?.email || null,
            timestamp: new Date().toISOString()
          },
          context: {
            source: 'hannaui-web',
            version: '2.0.0'
          }
        };

        aiResponse = await ChatService.sendMessage(request);
      } else {
        // Fall back to demo mode
        await new Promise(resolve => setTimeout(resolve, 1000 + Math.random() * 1000));
        aiResponse = ChatService.generateDemoResponse(text);
      }

      const aiMessage: Message = {
        type: 'ai',
        text: aiResponse,
        timestamp: new Date()
      };
      
      setMessages(prev => [...prev, aiMessage]);

    } catch (error) {
      console.error('Error getting AI response:', error);
      
      const errorMessage: Message = {
        type: 'ai',
        text: `I'm having trouble connecting to my main systems right now. **Error:** ${(error as Error).message}

I'll try to help you anyway - what would you like to know?`,
        timestamp: new Date()
      };
      
      setMessages(prev => [...prev, errorMessage]);

      // Update connection status if this was a connection error
      if ((error as Error).message.includes('timeout') || (error as Error).message.includes('fetch')) {
        setConnectionInfo(prev => ({
          ...prev,
          status: 'error'
        }));
      }
    } finally {
      setIsTyping(false);
      onAnimationTrigger?.('idle');
    }
  };

  return (
    <div className="chat-overlay fixed inset-0 flex flex-col bg-gradient-to-br from-purple-900/20 to-blue-900/20 backdrop-blur-sm">
      {/* Chat header */}
      <div className="chat-header flex-shrink-0 p-4 bg-black/20 backdrop-blur-sm">
        <ConnectionStatus connectionInfo={connectionInfo} />
      </div>

      {/* Chat messages area */}
      <div 
        ref={messagesContainerRef}
        className="flex-1 overflow-y-auto p-4 space-y-4"
      >
        {messages.map((message, index) => (
          <ChatMessage key={index} message={message} />
        ))}
        
        {isTyping && (
          <div className="flex items-start space-x-3">
            <div className="flex-shrink-0 w-8 h-8 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full flex items-center justify-center">
              <span className="text-white text-sm">🤖</span>
            </div>
            <div className="flex-1 bg-white/10 backdrop-blur-sm rounded-lg p-3">
              <div className="typing-indicator flex items-center space-x-1 text-white/70">
                <span>Hanna is typing</span>
                <div className="flex space-x-1">
                  <div className="w-2 h-2 bg-white/50 rounded-full animate-pulse"></div>
                  <div className="w-2 h-2 bg-white/50 rounded-full animate-pulse delay-75"></div>
                  <div className="w-2 h-2 bg-white/50 rounded-full animate-pulse delay-150"></div>
                </div>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      {/* Chat input */}
      <div className="flex-shrink-0 p-4 bg-black/20 backdrop-blur-sm">
        <ChatInput 
          onSendMessage={handleSendMessage} 
          disabled={isTyping}
          onFocus={() => onAnimationTrigger?.('typing')}
          onBlur={() => !isTyping && onAnimationTrigger?.('idle')}
        />
      </div>
    </div>
  );
}