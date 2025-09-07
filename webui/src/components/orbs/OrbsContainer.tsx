'use client';

import { useEffect, useRef, useState, useCallback } from 'react';
import { AnimationType, OrbConfig, OrbPosition } from '@/types';

interface OrbsContainerProps {
  animation?: AnimationType;
  onAnimationChange?: (animation: AnimationType) => void;
}

interface Orb extends OrbConfig {
  element: HTMLDivElement | null;
  vx: number;
  vy: number;
  baseSpeed: number;
  phase: number;
  glowIntensity: number;
  clustered: boolean;
  clusterTarget: OrbPosition | null;
  trail: OrbPosition[];
}

const ANIMATION_TYPES: AnimationType[] = [
  'idle', 'typing', 'network', 'dna', 'loading', 'racetrack', 
  'bloom', 'pulse', 'swarm', 'dance', 'vortex', 'cascade'
];

export default function OrbsContainer({ animation = 'idle', onAnimationChange }: OrbsContainerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const orbsRef = useRef<Orb[]>([]);
  const animationFrameRef = useRef<number>(0);
  const lastTimeRef = useRef<number>(0);
  const [showControls, setShowControls] = useState(false);
  const [isMouseNear, setIsMouseNear] = useState(false);
  const mouseRef = useRef({ x: 0, y: 0 });

  // Initialize orbs
  const initOrbs = useCallback(() => {
    if (!containerRef.current) return;

    const container = containerRef.current;
    const viewportWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
    const viewportHeight = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0);

    // Clear existing orbs
    container.innerHTML = '';
    orbsRef.current = [];

    // Create 10 orbs
    for (let i = 0; i < 10; i++) {
      const orbElement = document.createElement('div');
      orbElement.className = `orb orb-color-${(i % 3) + 1}`;
      orbElement.id = `orb-${i + 1}`;
      
      const orb: Orb = {
        id: `orb-${i + 1}`,
        element: orbElement,
        position: {
          x: Math.random() * (viewportWidth - 100) + 50,
          y: Math.random() * (viewportHeight - 100) + 50,
          vx: (Math.random() - 0.5) * 1.5,
          vy: (Math.random() - 0.5) * 1.5,
        },
        vx: (Math.random() - 0.5) * 1.5,
        vy: (Math.random() - 0.5) * 1.5,
        baseSpeed: 0.8 + Math.random() * 1.2,
        phase: Math.random() * Math.PI * 2,
        size: 12 + Math.random() * 8,
        color: `hsl(${i * 36}, 70%, 60%)`,
        glowIntensity: 0.8 + Math.random() * 0.4,
        clustered: false,
        clusterTarget: null,
        trail: []
      };

      // Set initial styles
      orbElement.style.width = orb.size + 'px';
      orbElement.style.height = orb.size + 'px';
      orbElement.style.transform = `translate(${orb.position.x}px, ${orb.position.y}px)`;
      
      container.appendChild(orbElement);
      orbsRef.current.push(orb);
    }
  }, []);

  // Update orb positions based on animation type
  const updateAnimation = useCallback((deltaTime: number) => {
    const viewportWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
    const viewportHeight = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0);
    
    orbsRef.current.forEach((orb, index) => {
      switch (animation) {
        case 'idle':
          updateIdleAnimation(orb, deltaTime, viewportWidth, viewportHeight);
          break;
        case 'typing':
          updateTypingAnimation(orb, deltaTime, index);
          break;
        case 'network':
          updateNetworkAnimation(orb, deltaTime, index);
          break;
        case 'pulse':
          updatePulseAnimation(orb, deltaTime, index);
          break;
        default:
          updateIdleAnimation(orb, deltaTime, viewportWidth, viewportHeight);
      }
      
      // Apply transform
      if (orb.element) {
        orb.element.style.transform = `translate(${orb.position.x}px, ${orb.position.y}px)`;
      }
    });
  }, [animation]);

  // Idle animation - firefly-style movement
  const updateIdleAnimation = (orb: Orb, deltaTime: number, viewportWidth: number, viewportHeight: number) => {
    const time = performance.now() * 0.001;
    
    // Gentle floating movement
    orb.position.x += Math.sin(time * orb.baseSpeed + orb.phase) * 0.5;
    orb.position.y += Math.cos(time * orb.baseSpeed * 0.7 + orb.phase) * 0.3;
    
    // Add some drift
    orb.vx += (Math.random() - 0.5) * 0.02;
    orb.vy += (Math.random() - 0.5) * 0.02;
    
    // Apply velocity with damping
    orb.position.x += orb.vx * deltaTime * 60;
    orb.position.y += orb.vy * deltaTime * 60;
    
    // Velocity damping
    orb.vx *= 0.995;
    orb.vy *= 0.995;
    
    // Mouse avoidance
    const mouseDistance = Math.sqrt(
      Math.pow(orb.position.x - mouseRef.current.x, 2) + 
      Math.pow(orb.position.y - mouseRef.current.y, 2)
    );
    
    if (mouseDistance < 100) {
      const avoidForce = (100 - mouseDistance) / 100;
      const angle = Math.atan2(orb.position.y - mouseRef.current.y, orb.position.x - mouseRef.current.x);
      orb.vx += Math.cos(angle) * avoidForce * 0.5;
      orb.vy += Math.sin(angle) * avoidForce * 0.5;
    }
    
    // Keep in bounds with soft bounce
    if (orb.position.x < 0 || orb.position.x > viewportWidth) orb.vx *= -0.8;
    if (orb.position.y < 0 || orb.position.y > viewportHeight) orb.vy *= -0.8;
    
    orb.position.x = Math.max(0, Math.min(viewportWidth, orb.position.x));
    orb.position.y = Math.max(0, Math.min(viewportHeight, orb.position.y));
  };

  // Typing animation - gentle pulsing
  const updateTypingAnimation = (orb: Orb, deltaTime: number, index: number) => {
    const time = performance.now() * 0.001;
    const pulseSpeed = 2 + index * 0.2;
    const scale = 0.8 + Math.sin(time * pulseSpeed + orb.phase) * 0.2;
    
    if (orb.element) {
      orb.element.style.transform = `translate(${orb.position.x}px, ${orb.position.y}px) scale(${scale})`;
    }
  };

  // Network animation - connection lines
  const updateNetworkAnimation = (orb: Orb, deltaTime: number, index: number) => {
    // Slow orbital movement
    const time = performance.now() * 0.0005;
    const radius = 100 + index * 20;
    const centerX = window.innerWidth / 2;
    const centerY = window.innerHeight / 2;
    
    orb.position.x = centerX + Math.cos(time + index * 0.8) * radius;
    orb.position.y = centerY + Math.sin(time + index * 0.8) * radius;
  };

  // Pulse animation - synchronized pulsing
  const updatePulseAnimation = (orb: Orb, deltaTime: number, index: number) => {
    const time = performance.now() * 0.003;
    const scale = 1 + Math.sin(time) * 0.3;
    
    if (orb.element) {
      orb.element.style.transform = `translate(${orb.position.x}px, ${orb.position.y}px) scale(${scale})`;
    }
  };

  // Animation loop
  const animate = useCallback((currentTime: number) => {
    const deltaTime = (currentTime - lastTimeRef.current) / 1000;
    lastTimeRef.current = currentTime;
    
    updateAnimation(deltaTime);
    animationFrameRef.current = requestAnimationFrame(animate);
  }, [updateAnimation]);

  // Mouse tracking
  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      mouseRef.current = { x: e.clientX, y: e.clientY };
      
      const proximityThreshold = 100;
      const mouseNear = orbsRef.current.some(orb => {
        const distance = Math.sqrt(
          Math.pow(orb.position.x - e.clientX, 2) + 
          Math.pow(orb.position.y - e.clientY, 2)
        );
        return distance < proximityThreshold;
      });
      
      setIsMouseNear(mouseNear);
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'c' && e.ctrlKey) {
        setShowControls(prev => !prev);
      }
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

  // Initialize orbs and start animation
  useEffect(() => {
    initOrbs();
    animationFrameRef.current = requestAnimationFrame(animate);

    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [initOrbs, animate]);

  // Handle window resize
  useEffect(() => {
    const handleResize = () => {
      initOrbs();
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, [initOrbs]);

  return (
    <>
      <div
        ref={containerRef}
        className="orbs-container fixed inset-0 pointer-events-none z-0"
      />
      
      {showControls && (
        <div className="animation-controls fixed top-20 right-4 bg-black/20 backdrop-blur-sm rounded-lg p-4 z-50">
          <div className="grid grid-cols-2 gap-2">
            {ANIMATION_TYPES.map((animType) => (
              <button
                key={animType}
                className={`px-3 py-1 rounded text-sm transition-colors ${
                  animation === animType
                    ? 'bg-purple-600 text-white'
                    : 'bg-white/10 text-white/70 hover:bg-white/20'
                }`}
                onClick={() => onAnimationChange?.(animType)}
              >
                {animType.charAt(0).toUpperCase() + animType.slice(1)}
              </button>
            ))}
          </div>
        </div>
      )}
      
      <button
        className="animation-toggle-btn fixed top-4 right-16 bg-black/20 backdrop-blur-sm rounded-lg p-2 text-white/70 hover:text-white z-50"
        onClick={() => setShowControls(prev => !prev)}
        title="Toggle Animation Controls (Ctrl+C)"
      >
        ✨
      </button>
    </>
  );
}