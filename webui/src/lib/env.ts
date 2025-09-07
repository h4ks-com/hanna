// Environment configuration utilities
export const env = {
  // Server-side environment variables
  HANNA_N8N_INTERNAL_ENDPOINT: process.env.HANNA_N8N_INTERNAL_ENDPOINT || '',
  
  // Client-side environment variables (prefixed with NEXT_PUBLIC_)
  NEXT_PUBLIC_HANNA_N8N_ENDPOINT: process.env.NEXT_PUBLIC_HANNA_N8N_ENDPOINT || '',
  NEXT_PUBLIC_APP_VERSION: process.env.NEXT_PUBLIC_APP_VERSION || '2.0.0',
  NEXT_PUBLIC_APP_NAME: process.env.NEXT_PUBLIC_APP_NAME || 'HannaUI',
} as const;

// Client-side environment access (safe for browser)
export const clientEnv = {
  HANNA_N8N_ENDPOINT: env.NEXT_PUBLIC_HANNA_N8N_ENDPOINT,
  APP_VERSION: env.NEXT_PUBLIC_APP_VERSION,
  APP_NAME: env.NEXT_PUBLIC_APP_NAME,
} as const;

// Server-side environment access (includes sensitive values)
export const serverEnv = {
  ...clientEnv,
  HANNA_N8N_INTERNAL_ENDPOINT: env.HANNA_N8N_INTERNAL_ENDPOINT,
} as const;

// Validation function
export function validateEnv() {
  const errors: string[] = [];
  
  if (!env.NEXT_PUBLIC_APP_VERSION) {
    errors.push('NEXT_PUBLIC_APP_VERSION is required');
  }
  
  if (errors.length > 0) {
    throw new Error(`Environment validation failed:\n${errors.join('\n')}`);
  }
}

// Get endpoint for chat API (decides between internal and external)
export function getChatEndpoint(isServer = false): string {
  if (isServer && env.HANNA_N8N_INTERNAL_ENDPOINT) {
    return env.HANNA_N8N_INTERNAL_ENDPOINT;
  }
  
  return env.NEXT_PUBLIC_HANNA_N8N_ENDPOINT || 'http://localhost:5678/webhook/hanna-chat';
}