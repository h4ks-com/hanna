import { NextResponse } from 'next/server';
import { serverEnv } from '@/lib/env';

export async function GET() {
  return NextResponse.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: serverEnv.APP_VERSION,
    services: {
      n8n: {
        configured: !!serverEnv.HANNA_N8N_INTERNAL_ENDPOINT,
        endpoint: serverEnv.HANNA_N8N_INTERNAL_ENDPOINT ? 'configured' : 'not configured',
      },
    },
  });
}