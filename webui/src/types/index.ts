export interface User {
  username: string;
  email: string;
  rememberMe?: boolean;
}

export interface Message {
  type: 'user' | 'ai';
  text: string;
  timestamp: Date;
}

export interface ChatEndpointInfo {
  endpoint: string | null;
  status: 'connected' | 'offline' | 'error' | 'no-endpoint' | 'checking';
  configured: boolean;
}

export interface OrbPosition {
  x: number;
  y: number;
  vx: number;
  vy: number;
}

export interface OrbConfig {
  id: string;
  position: OrbPosition;
  color: string;
  size: number;
}

export type AnimationType = 
  | 'idle' 
  | 'typing' 
  | 'network' 
  | 'dna' 
  | 'loading' 
  | 'racetrack' 
  | 'bloom' 
  | 'pulse' 
  | 'swarm' 
  | 'dance' 
  | 'vortex' 
  | 'cascade';

export type ThemeType = 'dark' | 'blue' | 'purple' | 'pink' | 'cute';

export interface ChatRequest {
  message: string;
  sessionId: string;
  user: {
    username: string;
    email: string | null;
    timestamp: string;
  };
  context: {
    source: string;
    version: string;
  };
}

export interface ChatResponse {
  output?: string;
  message?: string;
  text?: string;
  content?: string;
  error?: string;
}