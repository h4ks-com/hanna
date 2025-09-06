// Orb management and basic animations
class OrbManager {
    constructor() {
        this.orbs = [];
        this.container = document.querySelector('.orbs-container');
        this.isMouseNear = false;
        this.mouseX = 0;
        this.mouseY = 0;
        this.currentAnimation = 'idle';
        this.animationActive = false;
        
        this.initOrbs();
        this.bindEvents();
        this.startAnimation();
    }

    initOrbs() {
        const orbElements = document.querySelectorAll('.orb');
        
        // Get actual viewport dimensions
        const viewportWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
        const viewportHeight = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0);
        
        this.orbs = Array.from(orbElements).map((element, index) => ({
            element,
            id: index,
            x: Math.random() * (viewportWidth - 100) + 50, // Keep away from edges
            y: Math.random() * (viewportHeight - 100) + 50,
            vx: (Math.random() - 0.5) * 1.5, // Much faster initial velocity
            vy: (Math.random() - 0.5) * 1.5,
            baseSpeed: 0.8 + Math.random() * 1.2, // 4-6x faster base speed
            phase: Math.random() * Math.PI * 2,
            size: 12 + Math.random() * 8,
            glowIntensity: 0.8 + Math.random() * 0.4,
            clustered: false,
            clusterTarget: null,
            trail: []
        }));

        // Position orbs initially with proper transform
        this.orbs.forEach(orb => {
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px)`;
            orb.element.style.width = orb.size + 'px';
            orb.element.style.height = orb.size + 'px';
        });
    }

    bindEvents() {
        // Mouse tracking for avoidance behavior
        document.addEventListener('mousemove', (e) => {
            this.mouseX = e.clientX;
            this.mouseY = e.clientY;
            this.checkMouseProximity();
        });

        // Window resize handling
        window.addEventListener('resize', () => {
            this.handleResize();
        });

        // Animation controls
        document.addEventListener('keydown', (e) => {
            if (e.key === 'c' && e.ctrlKey) {
                this.toggleAnimationControls();
            }
        });

        // Animation toggle button
        const animToggleBtn = document.querySelector('.animation-toggle-btn');
        if (animToggleBtn) {
            animToggleBtn.addEventListener('click', () => {
                this.toggleAnimationControls();
            });
        }
    }

    checkMouseProximity() {
        const proximityThreshold = 100;
        this.isMouseNear = this.orbs.some(orb => {
            const distance = Math.sqrt(
                Math.pow(orb.x - this.mouseX, 2) + 
                Math.pow(orb.y - this.mouseY, 2)
            );
            return distance < proximityThreshold;
        });
    }

    updateOrbColors() {
        // Called when theme changes to update orb appearances
        this.orbs.forEach((orb, index) => {
            const colorIndex = (index % 3) + 1;
            orb.element.className = `orb orb-color-${colorIndex}`;
        });
    }

    // Idle animation - firefly-style movement
    updateIdleAnimation(deltaTime) {
        // Get current viewport dimensions (cache for performance)
        const viewportWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
        const viewportHeight = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0);
        
        // Batch DOM updates
        const updates = [];
        
        this.orbs.forEach((orb, index) => {
            // Firefly behavior - erratic movement with sudden direction changes
            orb.phase = (orb.phase || 0) + deltaTime * 0.002;
            
            // Initialize velocities and firefly properties if they don't exist
            orb.vx = orb.vx || (Math.random() - 0.5) * 0.8;
            orb.vy = orb.vy || (Math.random() - 0.5) * 0.8;
            orb.dartTimer = orb.dartTimer || Math.random() * 3000;
            orb.hovering = orb.hovering || false;
            orb.hoverTimer = orb.hoverTimer || 0;
            
            // Swerving direction changes (firefly behavior)
            orb.dartTimer -= deltaTime;
            if (orb.dartTimer <= 0) {
                // Calculate current direction
                const currentAngle = Math.atan2(orb.vy, orb.vx);
                // Swerve within a 120-degree cone (not complete random direction)
                const swerveRange = Math.PI * 0.67; // 120 degrees
                const swerveAngle = currentAngle + (Math.random() - 0.5) * swerveRange;
                const swerveStrength = 1.2 + Math.random() * 1.8; // Much stronger swerves for zipping
                
                // Apply swerve as velocity adjustment rather than replacement
                orb.vx += Math.cos(swerveAngle) * swerveStrength;
                orb.vy += Math.sin(swerveAngle) * swerveStrength;
                orb.dartTimer = 600 + Math.random() * 1200; // Even more frequent zipping (0.6-1.8 seconds)
                orb.hovering = false;
            }
            
            // Random hovering behavior (fireflies sometimes pause)
            if (Math.random() < 0.0008 && !orb.hovering) {
                orb.hovering = true;
                orb.hoverTimer = 500 + Math.random() * 1500; // Hover for 0.5-2 seconds
            }
            
            if (orb.hovering) {
                orb.hoverTimer -= deltaTime;
                // Slow down significantly while hovering
                orb.vx *= 0.95;
                orb.vy *= 0.95;
                // Add slight floating motion
                orb.vx += Math.sin(orb.phase * 2) * 0.008;
                orb.vy += Math.cos(orb.phase * 2.3) * 0.008;
                
                if (orb.hoverTimer <= 0) {
                    orb.hovering = false;
                }
            } else {
                // Normal erratic movement when not hovering - preserve momentum for swerving
                // Add gentle course corrections rather than random chaos
                const currentSpeed = Math.sqrt(orb.vx * orb.vx + orb.vy * orb.vy);
                const currentAngle = Math.atan2(orb.vy, orb.vx);
                
                // Small random adjustments to current direction (swerving)
                const swerveAdjustment = (Math.random() - 0.5) * 0.25; // Larger random adjustments for zippier movement
                const newAngle = currentAngle + swerveAdjustment;
                
                // Apply stronger course corrections while maintaining speed
                orb.vx += Math.cos(newAngle) * 0.02; // Stronger continuous adjustments
                orb.vy += Math.sin(newAngle) * 0.02;
                
                // Add some wave motion for organic feel
                orb.vx += Math.sin(orb.phase) * 0.025; // Stronger wave motion
                orb.vy += Math.cos(orb.phase * 1.1) * 0.025;
                
                // Less momentum damping to maintain higher speeds
                orb.vx *= 0.999; // Reduced damping
                orb.vy *= 0.999;
            }
            
            // Mouse avoidance - fireflies are more reactive
            if (this.isMouseNear) {
                const dx = orb.x - this.mouseX;
                const dy = orb.y - this.mouseY;
                const distance = Math.sqrt(dx * dx + dy * dy);
                
                if (distance < 120 && distance > 0) {
                    const avoidForce = (120 - distance) / 120;
                    orb.vx += (dx / distance) * avoidForce * 0.04; // Stronger avoidance
                    orb.vy += (dy / distance) * avoidForce * 0.04;
                    orb.hovering = false; // Break out of hovering when mouse is near
                }
            }
            
            // Limit velocity - allow much higher speeds for fast firefly zipping
            const maxVel = orb.hovering ? orb.baseSpeed * 0.5 : orb.baseSpeed * 5.0; // Much higher max speed (5x instead of 2.5x)
            const currentVel = Math.sqrt(orb.vx * orb.vx + orb.vy * orb.vy);
            if (currentVel > maxVel) {
                orb.vx = (orb.vx / currentVel) * maxVel;
                orb.vy = (orb.vy / currentVel) * maxVel;
            }
            
            // Update position
            orb.x += orb.vx;
            orb.y += orb.vy;
            
            // Boundary wrapping with proper viewport dimensions
            if (orb.x < -orb.size) orb.x = viewportWidth + orb.size;
            if (orb.x > viewportWidth + orb.size) orb.x = -orb.size;
            if (orb.y < -orb.size) orb.y = viewportHeight + orb.size;
            if (orb.y > viewportHeight + orb.size) orb.y = -orb.size;
            
            // Clustering behavior (occasionally)
            if (Math.random() < 0.001 && !orb.clustered) {
                this.maybeFormCluster(orb, index);
            }
            
            // Store update for batching
            updates.push({ element: orb.element, x: orb.x, y: orb.y });
        });
        
        // Apply all DOM updates at once
        updates.forEach(update => {
            update.element.style.transform = `translate3d(${update.x}px, ${update.y}px, 0)`;
        });
    }

    maybeFormCluster(orb, orbIndex) {
        // Find nearby orbs to cluster with
        const nearbyOrbs = this.orbs.filter((otherOrb, otherIndex) => {
            if (otherIndex === orbIndex || otherOrb.clustered) return false;
            
            const distance = Math.sqrt(
                Math.pow(orb.x - otherOrb.x, 2) + 
                Math.pow(orb.y - otherOrb.y, 2)
            );
            return distance < 150;
        });

        if (nearbyOrbs.length > 0) {
            const targetOrb = nearbyOrbs[Math.floor(Math.random() * nearbyOrbs.length)];
            orb.clustered = true;
            orb.clusterTarget = targetOrb;
            
            // Release from cluster after some time
            setTimeout(() => {
                orb.clustered = false;
                orb.clusterTarget = null;
            }, 2000 + Math.random() * 3000);
        }
    }

    handleResize() {
        // Get new viewport dimensions
        const newWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
        const newHeight = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0);
        const oldWidth = this.container.offsetWidth || newWidth;
        const oldHeight = this.container.offsetHeight || newHeight;
        
        // Adjust orb positions on window resize
        const widthRatio = newWidth / oldWidth;
        const heightRatio = newHeight / oldHeight;
        
        this.orbs.forEach(orb => {
            orb.x = Math.min(orb.x * widthRatio, newWidth - orb.size);
            orb.y = Math.min(orb.y * heightRatio, newHeight - orb.size);
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px)`;
        });
    }

    setAnimation(animationType) {
        if (this.currentAnimation === animationType) return;
        
        // Clean up previous animation effects regardless of what it was
        this.cleanupAllAnimationEffects();
        
        this.currentAnimation = animationType;
        this.animationActive = animationType !== 'idle';
        
        // Clear ALL animation-specific properties to start completely fresh
        this.orbs.forEach(orb => {
            // Clear typing animation states
            delete orb.typingTarget;
            delete orb.typingSpeed;
            delete orb.typingPhase;
            delete orb.typingTime;
            // Clear any other animation-specific properties
            delete orb.animationPhase;
            delete orb.animationProgress;
        });

        // Update UI
        document.querySelectorAll('.anim-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        const activeBtn = document.querySelector(`[data-animation="${animationType}"]`);
        if (activeBtn) activeBtn.classList.add('active');
    }
    
    cleanupAllAnimationEffects() {
        // Always clean up network animation effects
        if (this.animationController) {
            // Clean up data packets
            if (this.animationController.dataPackets) {
                this.animationController.dataPackets.forEach(packet => {
                    if (packet.element && packet.element.parentNode) {
                        packet.element.parentNode.removeChild(packet.element);
                    }
                });
                this.animationController.dataPackets = [];
            }
            
            // Clean up ripple effects
            if (this.animationController.ripples) {
                this.animationController.ripples.forEach(ripple => {
                    if (ripple.element && ripple.element.parentNode) {
                        ripple.element.parentNode.removeChild(ripple.element);
                    }
                });
                this.animationController.ripples = [];
            }
            
            // Clean up pulse ripple effects
            if (this.animationController.pulseRipples) {
                this.animationController.pulseRipples.forEach(ripple => {
                    if (ripple.element && ripple.element.parentNode) {
                        ripple.element.parentNode.removeChild(ripple.element);
                    }
                });
                this.animationController.pulseRipples = [];
            }
            
            // Reset network state
            delete this.animationController.networkNodes;
            delete this.animationController.packetTimer;
            
            // Reset pulse state
            delete this.animationController.pulseTimers;
            delete this.animationController.lastPulseStates;
            delete this.animationController.pulseAnimationStartTime;
            
            // Reset constellation state
            if (this.animationController.constellationLines) {
                this.animationController.constellationLines.forEach(line => {
                    if (line.element && line.element.parentNode) {
                        line.element.parentNode.removeChild(line.element);
                    }
                });
                this.animationController.constellationLines = [];
            }
            
            // Reset swarm and dance states
            delete this.animationController.swarmTargets;
            delete this.animationController.danceState;
        }
        
        // Aggressive cleanup - find and remove any lingering ripple/packet elements by class
        document.querySelectorAll('.network-ripple').forEach(element => {
            if (element.parentNode) {
                element.parentNode.removeChild(element);
            }
        });
        
        document.querySelectorAll('.pulse-ripple').forEach(element => {
            if (element.parentNode) {
                element.parentNode.removeChild(element);
            }
        });
        
        document.querySelectorAll('.data-packet').forEach(element => {
            if (element.parentNode) {
                element.parentNode.removeChild(element);
            }
        });
        
        document.querySelectorAll('.constellation-line').forEach(element => {
            if (element.parentNode) {
                element.parentNode.removeChild(element);
            }
        });
        
        // Reset orb scales and brightness effects
        this.orbs.forEach(orb => {
            orb.element.style.transform = `translate3d(${orb.x}px, ${orb.y}px, 0)`;
            orb.element.style.filter = ''; // Clear brightness filter
        });
    }

    startAnimation() {
        let lastTime = 0;
        const targetFPS = 60;
        const frameTime = 1000 / targetFPS;
        let animationId;
        
        const animate = (currentTime) => {
            const deltaTime = currentTime - lastTime;
            
            // Throttle to maintain consistent framerate
            if (deltaTime >= frameTime) {
                lastTime = currentTime;
                
                if (this.currentAnimation === 'idle') {
                    this.updateIdleAnimation(deltaTime);
                } else if (window.animationController) {
                    window.animationController.updateAnimation(this.currentAnimation, deltaTime);
                }
            }
            
            animationId = requestAnimationFrame(animate);
        };
        
        animationId = requestAnimationFrame(animate);
        
        // Store animation ID for potential cleanup
        this.animationId = animationId;
    }

    toggleAnimationControls() {
        const controls = document.querySelector('.animation-controls');
        if (controls.style.display === 'none' || !controls.style.display) {
            controls.style.display = 'flex';
        } else {
            controls.style.display = 'none';
        }
    }

    getOrbs() {
        return this.orbs;
    }

    getContainer() {
        return this.container;
    }
}

// Initialize orb manager
const orbManager = new OrbManager();

// Make it globally available
window.orbManager = orbManager;
