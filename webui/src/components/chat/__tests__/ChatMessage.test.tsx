import { render, screen } from '@testing-library/react';
import ChatMessage from '../ChatMessage';
import { Message } from '@/types';

// Mock marked
jest.mock('marked', () => ({
  marked: {
    parse: jest.fn((text) => `<p>${text}</p>`),
    setOptions: jest.fn(),
  },
}));

describe('ChatMessage', () => {
  const mockMessage: Message = {
    type: 'user',
    text: 'Hello world!',
    timestamp: new Date('2024-01-01T12:00:00Z'),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    
    // Mock current time for consistent time formatting
    jest.useFakeTimers();
    jest.setSystemTime(new Date('2024-01-01T12:05:00Z')); // 5 minutes later
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('renders user message correctly', () => {
    render(<ChatMessage message={mockMessage} />);
    
    expect(screen.getByText('👤')).toBeInTheDocument();
    expect(screen.getByText('5m ago')).toBeInTheDocument();
  });

  it('renders AI message correctly', () => {
    const aiMessage: Message = {
      ...mockMessage,
      type: 'ai',
    };

    render(<ChatMessage message={aiMessage} />);
    
    expect(screen.getByText('🤖')).toBeInTheDocument();
    expect(screen.getByText('5m ago')).toBeInTheDocument();
  });

  it('renders markdown content', () => {
    const markdownMessage: Message = {
      ...mockMessage,
      text: '**Bold text**',
    };

    render(<ChatMessage message={markdownMessage} />);
    
    const { marked } = require('marked');
    expect(marked.parse).toHaveBeenCalledWith('**Bold text**');
  });

  it('formats time correctly for recent messages', () => {
    jest.setSystemTime(new Date('2024-01-01T12:00:30Z')); // 30 seconds later
    
    render(<ChatMessage message={mockMessage} />);
    
    expect(screen.getByText('Now')).toBeInTheDocument();
  });

  it('formats time correctly for messages from hours ago', () => {
    jest.setSystemTime(new Date('2024-01-01T14:30:00Z')); // 2.5 hours later
    
    render(<ChatMessage message={mockMessage} />);
    
    expect(screen.getByText('2h ago')).toBeInTheDocument();
  });

  it('formats time correctly for messages from days ago', () => {
    jest.setSystemTime(new Date('2024-01-02T12:00:00Z')); // 1 day later
    
    render(<ChatMessage message={mockMessage} />);
    
    expect(screen.getByText('1/1/2024')).toBeInTheDocument();
  });

  it('applies correct CSS classes for user messages', () => {
    render(<ChatMessage message={mockMessage} />);
    
    const messageContainer = screen.getByText('👤').closest('div').parentElement;
    expect(messageContainer).toHaveClass('flex-row-reverse');
  });

  it('applies correct CSS classes for AI messages', () => {
    const aiMessage: Message = {
      ...mockMessage,
      type: 'ai',
    };

    render(<ChatMessage message={aiMessage} />);
    
    const messageContainer = screen.getByText('🤖').closest('.flex');
    expect(messageContainer).not.toHaveClass('flex-row-reverse');
  });

  it('handles empty message text', () => {
    const emptyMessage: Message = {
      ...mockMessage,
      text: '',
    };

    render(<ChatMessage message={emptyMessage} />);
    
    const { marked } = require('marked');
    expect(marked.parse).toHaveBeenCalledWith('');
  });

  it('handles markdown parsing errors gracefully', () => {
    // Suppress console.error for this test
    const originalConsoleError = console.error;
    console.error = jest.fn();
    
    const { marked } = require('marked');
    marked.parse.mockImplementation(() => {
      throw new Error('Markdown parsing error');
    });

    const message: Message = {
      ...mockMessage,
      text: 'Some text',
    };

    render(<ChatMessage message={message} />);
    
    // Should still render the component without crashing
    expect(screen.getByText('👤')).toBeInTheDocument();
    
    // Restore console.error
    console.error = originalConsoleError;
  });
});