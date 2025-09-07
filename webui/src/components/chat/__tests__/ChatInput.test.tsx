import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ChatInput from '../ChatInput';

describe('ChatInput', () => {
  const mockOnSendMessage = jest.fn();
  const mockOnFocus = jest.fn();
  const mockOnBlur = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders input field and send button', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    expect(screen.getByPlaceholderText('Type your message...')).toBeInTheDocument();
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('calls onSendMessage when send button is clicked', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    const sendButton = screen.getByRole('button');
    
    await user.type(input, 'Hello world!');
    await user.click(sendButton);
    
    expect(mockOnSendMessage).toHaveBeenCalledWith('Hello world!');
  });

  it('calls onSendMessage when Enter key is pressed', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    
    await user.type(input, 'Hello world!');
    await user.keyboard('{Enter}');
    
    expect(mockOnSendMessage).toHaveBeenCalledWith('Hello world!');
  });

  it('does not call onSendMessage when Shift+Enter is pressed', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    
    await user.type(input, 'Hello world!');
    await user.keyboard('{Shift>}{Enter}{/Shift}');
    
    expect(mockOnSendMessage).not.toHaveBeenCalled();
  });

  it('clears input after sending message', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const input = screen.getByPlaceholderText('Type your message...') as HTMLInputElement;
    
    await user.type(input, 'Hello world!');
    expect(input.value).toBe('Hello world!');
    
    await user.click(screen.getByRole('button'));
    
    expect(input.value).toBe('');
  });

  it('does not send empty messages', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const sendButton = screen.getByRole('button');
    
    await user.click(sendButton);
    
    expect(mockOnSendMessage).not.toHaveBeenCalled();
  });

  it('does not send whitespace-only messages', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    const sendButton = screen.getByRole('button');
    
    await user.type(input, '   ');
    await user.click(sendButton);
    
    expect(mockOnSendMessage).not.toHaveBeenCalled();
  });

  it('trims message before sending', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    const sendButton = screen.getByRole('button');
    
    await user.type(input, '  Hello world!  ');
    await user.click(sendButton);
    
    expect(mockOnSendMessage).toHaveBeenCalledWith('Hello world!');
  });

  it('disables input and button when disabled prop is true', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} disabled={true} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    const sendButton = screen.getByRole('button');
    
    expect(input).toBeDisabled();
    expect(sendButton).toBeDisabled();
  });

  it('calls onFocus when input is focused', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} onFocus={mockOnFocus} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    
    await user.click(input);
    
    expect(mockOnFocus).toHaveBeenCalled();
  });

  it('calls onBlur when input loses focus', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} onBlur={mockOnBlur} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    
    await user.click(input);
    await user.tab(); // Move focus away
    
    expect(mockOnBlur).toHaveBeenCalled();
  });

  it('button is disabled when input is empty', () => {
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const sendButton = screen.getByRole('button');
    
    expect(sendButton).toBeDisabled();
  });

  it('button is enabled when input has text', async () => {
    const user = userEvent.setup();
    
    render(<ChatInput onSendMessage={mockOnSendMessage} />);
    
    const input = screen.getByPlaceholderText('Type your message...');
    const sendButton = screen.getByRole('button');
    
    expect(sendButton).toBeDisabled();
    
    await user.type(input, 'Hello');
    
    expect(sendButton).not.toBeDisabled();
  });
});