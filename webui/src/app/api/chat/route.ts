import { NextRequest, NextResponse } from 'next/server';
import { ChatRequest, ChatResponse } from '@/types';
import { serverEnv } from '@/lib/env';

export async function POST(request: NextRequest) {
  try {
    const body: ChatRequest = await request.json();
    
    // Validate request
    if (!body.message || !body.sessionId) {
      return NextResponse.json(
        { error: 'Missing required fields: message, sessionId' },
        { status: 400 }
      );
    }

    // Get the internal n8n endpoint
    const endpoint = serverEnv.HANNA_N8N_INTERNAL_ENDPOINT;
    
    if (!endpoint) {
      return NextResponse.json(
        { error: 'n8n endpoint not configured on server' },
        { status: 503 }
      );
    }

    // Forward request to n8n
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        // Forward some headers from the original request
        'User-Agent': request.headers.get('user-agent') || 'HannaUI/2.0',
        'X-Forwarded-For': request.headers.get('x-forwarded-for') || 'unknown',
      },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      console.error(`n8n request failed: ${response.status} ${response.statusText}`);
      return NextResponse.json(
        { error: `n8n request failed: ${response.status} ${response.statusText}` },
        { status: response.status }
      );
    }

    const data: ChatResponse = await response.json();
    
    // Return the response
    return NextResponse.json(data);

  } catch (error) {
    console.error('Chat API error:', error);
    
    return NextResponse.json(
      { 
        error: 'Internal server error',
        details: error instanceof Error ? error.message : 'Unknown error'
      },
      { status: 500 }
    );
  }
}

// Health check endpoint
export async function GET() {
  const endpoint = serverEnv.HANNA_N8N_INTERNAL_ENDPOINT;
  
  return NextResponse.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    n8n_configured: !!endpoint,
    endpoint: endpoint ? 'configured' : 'not configured',
  });
}