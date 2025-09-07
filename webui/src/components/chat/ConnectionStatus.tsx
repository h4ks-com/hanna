'use client';

import { ChatEndpointInfo } from '@/types';

interface ConnectionStatusProps {
  connectionInfo: ChatEndpointInfo;
}

export default function ConnectionStatus({ connectionInfo }: ConnectionStatusProps) {
  const getStatusColor = () => {
    switch (connectionInfo.status) {
      case 'connected':
        return 'bg-green-500';
      case 'offline':
        return 'bg-yellow-500';
      case 'error':
        return 'bg-red-500';
      case 'no-endpoint':
        return 'bg-gray-500';
      default:
        return 'bg-purple-500 animate-pulse';
    }
  };

  const getStatusText = () => {
    switch (connectionInfo.status) {
      case 'connected':
        return 'Online';
      case 'offline':
        return 'Offline (Demo Mode)';
      case 'error':
        return 'Connection Error';
      case 'no-endpoint':
        return 'Demo Mode';
      default:
        return 'Connecting...';
    }
  };

  return (
    <div className="ai-status flex items-center space-x-3">
      <div className="flex items-center space-x-2">
        <div className={`w-3 h-3 rounded-full ${getStatusColor()}`}></div>
        <span className="text-white font-medium">Hanna AI</span>
        <span className="text-white/70 text-sm">{getStatusText()}</span>
      </div>
      
      {connectionInfo.endpoint && (
        <div className="text-xs text-white/50 font-mono">
          {connectionInfo.endpoint.replace(/^https?:\/\//, '').split('/')[0]}
        </div>
      )}
    </div>
  );
}