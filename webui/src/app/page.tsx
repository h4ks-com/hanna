'use client';

import { useEffect, useState } from 'react';
import { useAuth } from '@/components/auth/AuthProvider';
import { useRouter } from 'next/navigation';
import { AnimationType } from '@/types';
import OrbsContainer from '@/components/orbs/OrbsContainer';
import ChatContainer from '@/components/chat/ChatContainer';
import ThemeSelector from '@/components/ui/ThemeSelector';

export default function Home() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const [currentAnimation, setCurrentAnimation] = useState<AnimationType>('idle');

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  // Handle animation changes from chat
  const handleAnimationTrigger = (animation: AnimationType) => {
    setCurrentAnimation(animation);
  };

  // Show loading while checking auth
  if (isLoading) {
    return (
      <div className="fixed inset-0 flex items-center justify-center bg-gradient-to-br from-purple-900 to-blue-900">
        <div className="text-white text-lg">Loading...</div>
      </div>
    );
  }

  // Don't render if not authenticated (will redirect)
  if (!isAuthenticated) {
    return null;
  }

  return (
    <div className="relative min-h-screen overflow-hidden fade-in">
      {/* Orbs background */}
      <OrbsContainer 
        animation={currentAnimation}
        onAnimationChange={setCurrentAnimation}
      />
      
      {/* Theme selector */}
      <ThemeSelector />
      
      {/* Chat interface */}
      <ChatContainer onAnimationTrigger={handleAnimationTrigger} />
    </div>
  );
}
