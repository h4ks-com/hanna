// Advanced orb animations controller
class AnimationController {
    constructor() {
        this.orbs = [];
        this.centerX = window.innerWidth / 2;
        this.centerY = window.innerHeight / 2;
        this.animationTime = 0;
        
        this.initAnimationControls();
        this.updateCenterPoint();
        window.addEventListener('resize', () => this.updateCenterPoint());
    }

    resetAnimationTime() {
        this.animationTime = 0;
    }

    setOrbs(orbs) {
        this.orbs = orbs;
    }

    updateCenterPoint() {
        this.centerX = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0) / 2;
        this.centerY = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0) / 2;
    }

    initAnimationControls() {
        document.querySelectorAll('.anim-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                const animation = btn.dataset.animation;
                if (window.orbManager) {
                    window.orbManager.setAnimation(animation);
                }
            });
        });
    }

    updateAnimation(animationType, deltaTime) {
        this.animationTime += deltaTime;
        
        switch(animationType) {
            case 'typing':
                this.updateTypingAnimation(deltaTime);
                break;
            case 'network':
                this.updateNetworkAnimation(deltaTime);
                break;
            case 'dna':
                this.updateDNAAnimation(deltaTime);
                break;
            case 'loading':
                this.updateLoadingAnimation(deltaTime);
                break;
            case 'racetrack':
                this.updateRacetrackAnimation(deltaTime);
                break;
            case 'wave':
                this.updateWaveAnimation(deltaTime);
                break;
            case 'bloom':
                this.updateBloomAnimation(deltaTime);
                break;
            case 'pulse':
                this.updatePulseAnimation(deltaTime);
                break;
            case 'vortex':
                this.updateVortexAnimation(deltaTime);
                break;
            case 'swarm':
                this.updateSwarmAnimation(deltaTime);
                break;
            case 'dance':
                this.updateDanceAnimation(deltaTime);
                break;
            case 'cascade':
                this.updateCascadeAnimation(deltaTime);
                break;
            default:
                // Fall back to idle
                if (window.orbManager) {
                    window.orbManager.updateIdleAnimation(deltaTime);
                }
        }
    }

    // Typing Animation - Wave patterns for natural typing feel
    updateTypingAnimation(deltaTime) {
        const waveSpeed = 0.004;
        const waveLength = 200;
        const amplitude = 40; // Slightly smaller amplitude for typing
        const viewportWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
        const updates = [];
        
        this.orbs.forEach((orb, index) => {
            // Initialize wave target if needed
            if (!orb.typingTarget) {
                orb.typingTarget = {
                    x: (index / this.orbs.length) * viewportWidth * 0.8 + viewportWidth * 0.1, // Center 80% of screen
                    y: this.centerY
                };
                orb.typingWavePhase = 0;
            }
            
            // Update wave phase
            orb.typingWavePhase += deltaTime * waveSpeed;
            
            // Calculate wave position
            const baseX = orb.typingTarget.x;
            const targetY = this.centerY + 
                           Math.sin((baseX / waveLength) * Math.PI * 2 + orb.typingWavePhase) * amplitude;
            
            // Move smoothly toward wave positions
            const lerpSpeed = 0.06;
            orb.x += (baseX - orb.x) * lerpSpeed;
            orb.y += (targetY - orb.y) * lerpSpeed;
            
            updates.push({ element: orb.element, x: orb.x, y: orb.y });
        });
        
        // Batch DOM updates
        updates.forEach(update => {
            update.element.style.transform = `translate3d(${update.x}px, ${update.y}px, 0)`;
        });
    }

    // Network Animation - Network topology with data packets
    updateNetworkAnimation(deltaTime) {
        const viewportWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
        const viewportHeight = Math.max(document.documentElement.clientHeight || 0, window.innerHeight || 0);
        
        // Initialize network nodes if not set
        if (!this.networkNodes) {
            this.networkNodes = [];
            this.dataPackets = [];
            this.packetTimer = 0;
            
            // Create network topology - arrange orbs as network nodes
            this.orbs.forEach((orb, index) => {
                // Position nodes in a distributed network layout
                let x, y;
                if (index < 3) {
                    // Core nodes - central triangle
                    const angle = (index / 3) * Math.PI * 2;
                    x = this.centerX + Math.cos(angle) * 80;
                    y = this.centerY + Math.sin(angle) * 80;
                } else if (index < 6) {
                    // Secondary ring
                    const angle = ((index - 3) / 3) * Math.PI * 2 + Math.PI / 3;
                    x = this.centerX + Math.cos(angle) * 180;
                    y = this.centerY + Math.sin(angle) * 180;
                } else {
                    // Edge nodes - distributed around perimeter
                    const angle = ((index - 6) / (this.orbs.length - 6)) * Math.PI * 2;
                    const radius = 120 + Math.random() * 100;
                    x = this.centerX + Math.cos(angle) * radius;
                    y = this.centerY + Math.sin(angle) * radius;
                }
                
                // Keep nodes within viewport bounds
                x = Math.max(50, Math.min(viewportWidth - 50, x));
                y = Math.max(50, Math.min(viewportHeight - 50, y));
                
                this.networkNodes.push({
                    id: index,
                    targetX: x,
                    targetY: y,
                    connections: [],
                    activity: 0
                });
            });
            
            // Create network connections between nodes
            this.networkNodes.forEach((node, i) => {
                // Connect to 2-4 nearby nodes
                const connectionCount = 2 + Math.floor(Math.random() * 3);
                const possibleConnections = this.networkNodes
                    .map((other, j) => ({ node: other, distance: Math.sqrt(
                        Math.pow(node.targetX - other.targetX, 2) + 
                        Math.pow(node.targetY - other.targetY, 2)
                    ), index: j }))
                    .filter(conn => conn.index !== i)
                    .sort((a, b) => a.distance - b.distance);
                
                for (let j = 0; j < Math.min(connectionCount, possibleConnections.length); j++) {
                    const targetIndex = possibleConnections[j].index;
                    if (!node.connections.includes(targetIndex)) {
                        node.connections.push(targetIndex);
                    }
                }
            });
        }
        
        // Move orbs to their network node positions
        const updates = [];
        this.orbs.forEach((orb, index) => {
            const node = this.networkNodes[index];
            orb.x += (node.targetX - orb.x) * 0.05;
            orb.y += (node.targetY - orb.y) * 0.05;
            
            // Enhanced pulsing effect based on network activity
            node.pulseTime = (node.pulseTime || 0) + deltaTime;
            
            let scale = 1;
            if (node.activity > 0) {
                // Create a dramatic pulse wave when receiving data
                const pulsePhase = node.pulseTime * 0.008; // Pulse speed
                const pulseIntensity = node.activity;
                
                // Double pulse effect - quick burst then gentle fade
                const quickPulse = Math.exp(-pulsePhase * 3) * Math.sin(pulsePhase * 8);
                const slowPulse = Math.exp(-pulsePhase * 1.5) * 0.3;
                
                scale = 1 + (quickPulse + slowPulse) * pulseIntensity * 0.8;
                
                // Add glow intensity variation
                const glowIntensity = 1 + pulseIntensity * 2;
                orb.element.style.filter = `brightness(${glowIntensity})`;
                
                // Decay activity more gradually for visible effect
                node.activity *= 0.992;
                
                // Reset when activity is very low
                if (node.activity < 0.01) {
                    node.activity = 0;
                    orb.element.style.filter = '';
                }
            }
            
            updates.push({ 
                element: orb.element, 
                x: orb.x, 
                y: orb.y,
                scale: Math.max(0.8, scale) // Prevent going too small
            });
        });
        
        // Create data packets between connected nodes
        this.packetTimer += deltaTime;
        if (this.packetTimer > 300 + Math.random() * 500) { // Every 0.3-0.8 seconds
            this.createDataPacket();
            this.packetTimer = 0;
        }
        
        // Update data packets
        this.updateDataPackets(deltaTime);
        
        // Update ripple effects
        this.updateRipples(deltaTime);
        
        // Apply all DOM updates
        updates.forEach(update => {
            update.element.style.transform = `translate3d(${update.x}px, ${update.y}px, 0) scale(${update.scale})`;
        });
    }
    
    createDataPacket() {
        // Choose random source and destination nodes that are connected
        const sourceIndex = Math.floor(Math.random() * this.networkNodes.length);
        const sourceNode = this.networkNodes[sourceIndex];
        
        if (sourceNode.connections.length === 0) return;
        
        const targetIndex = sourceNode.connections[Math.floor(Math.random() * sourceNode.connections.length)];
        const targetNode = this.networkNodes[targetIndex];
        
        // Get actual orb positions instead of static target positions
        const sourceOrb = this.orbs[sourceIndex];
        const targetOrb = this.orbs[targetIndex];
        
        // Create packet
        const packet = {
            id: Date.now() + Math.random(),
            sourceIndex,
            targetIndex,
            x: sourceOrb.x, // Use actual current position
            y: sourceOrb.y, // Use actual current position
            progress: 0,
            speed: 0.8 + Math.random() * 1.2, // Variable speed
            element: null
        };
        
        // Create visual element for packet
        packet.element = document.createElement('div');
        packet.element.className = 'data-packet';
        packet.element.style.cssText = `
            position: fixed;
            width: 4px;
            height: 4px;
            background: var(--orb-color-1);
            border-radius: 50%;
            box-shadow: 0 0 8px var(--orb-color-1);
            pointer-events: none;
            z-index: 999;
            transform: translate3d(${packet.x}px, ${packet.y}px, 0);
        `;
        document.body.appendChild(packet.element);
        
        this.dataPackets.push(packet);
        
        // Increase activity at source node
        sourceNode.activity = Math.min(1, sourceNode.activity + 0.5);
    }
    
    updateDataPackets(deltaTime) {
        this.dataPackets = this.dataPackets.filter(packet => {
            packet.progress += deltaTime * packet.speed * 0.001;
            
            if (packet.progress >= 1) {
                // Packet reached destination - create dramatic pulse effect!
                const targetNode = this.networkNodes[packet.targetIndex];
                const targetOrb = this.orbs[packet.targetIndex];
                targetNode.activity = 1.0; // Full pulse intensity
                targetNode.pulseTime = 0; // Reset pulse timer for fresh pulse
                
                // Create expanding ripple effect using the orb's ACTUAL current position
                this.createRippleEffect(targetOrb.x, targetOrb.y, packet.targetIndex);
                
                // Remove packet element
                if (packet.element && packet.element.parentNode) {
                    packet.element.parentNode.removeChild(packet.element);
                }
                return false; // Remove from array
            }
            
            // Update packet position along path using ACTUAL orb positions
            const sourceOrb = this.orbs[packet.sourceIndex];
            const targetOrb = this.orbs[packet.targetIndex];
            
            packet.x = sourceOrb.x + (targetOrb.x - sourceOrb.x) * packet.progress;
            packet.y = sourceOrb.y + (targetOrb.y - sourceOrb.y) * packet.progress;
            
            // Update visual position
            if (packet.element) {
                packet.element.style.transform = `translate3d(${packet.x}px, ${packet.y}px, 0)`;
            }
            
            return true; // Keep in array
        });
    }
    
    createRippleEffect(x, y, orbIndex) {
        // Initialize ripples array if it doesn't exist
        if (!this.ripples) {
            this.ripples = [];
        }
        
        // Create multiple ripple rings for layered effect
        for (let i = 0; i < 3; i++) {
            const ripple = {
                id: Date.now() + Math.random() + i,
                orbIndex: orbIndex, // Track which orb this ripple belongs to
                offsetX: 0, // Will store the offset from orb center
                offsetY: 0,
                radius: 0,
                maxRadius: (16 + i * 6) * 1.2, // 20% larger - now 19.2, 26.4, 33.6px
                opacity: 0.8 - i * 0.2, // Different opacities
                speed: (12 + i * 4) * 1.2, // Proportionally faster speed - 14.4, 19.2, 24px/sec
                delay: i * 100, // Stagger the ripples
                age: -i * 100, // Start with negative age for delay
                element: null
            };
            
            // Create visual element for ripple
            ripple.element = document.createElement('div');
            ripple.element.className = 'network-ripple';
            ripple.element.style.cssText = `
                position: fixed;
                border: 2px solid var(--orb-color-1);
                border-radius: 50%;
                pointer-events: none;
                z-index: 998;
                opacity: 0;
                width: 0px;
                height: 0px;
                transform: translate3d(${x}px, ${y}px, 0);
                transform-origin: center center;
                box-shadow: 0 0 20px var(--orb-color-1);
            `;
            document.body.appendChild(ripple.element);
            
            this.ripples.push(ripple);
        }
    }
    
    updateRipples(deltaTime) {
        if (!this.ripples) return;
        
        this.ripples = this.ripples.filter(ripple => {
            ripple.age += deltaTime;
            
            // Don't start expanding until delay is over
            if (ripple.age < 0) {
                return true; // Keep but don't update visually yet
            }
            
            // Expand the ripple
            ripple.radius += deltaTime * ripple.speed * 0.001;
            
            // Calculate opacity fade
            const progress = ripple.radius / ripple.maxRadius;
            const fadeOpacity = ripple.opacity * (1 - progress) * Math.exp(-progress * 2);
            
            // Remove when fully expanded or invisible
            if (ripple.radius >= ripple.maxRadius || fadeOpacity < 0.01) {
                if (ripple.element && ripple.element.parentNode) {
                    ripple.element.parentNode.removeChild(ripple.element);
                }
                return false; // Remove from array
            }
            
            // Update visual appearance - follow the orb's current position!
            if (ripple.element && this.orbs[ripple.orbIndex]) {
                const orb = this.orbs[ripple.orbIndex];
                const size = ripple.radius * 2;
                ripple.element.style.width = size + 'px';
                ripple.element.style.height = size + 'px';
                ripple.element.style.opacity = fadeOpacity;
                
                // Account for orb's own size/positioning - orbs are centered on their x,y
                const orbHalfSize = orb.size / 2; // orb.size is the orb's width/height
                
                // Position ripple so its CENTER is on the orb's visual CENTER
                const rippleCenterX = orb.x + orbHalfSize - ripple.radius;
                const rippleCenterY = orb.y + orbHalfSize - ripple.radius;
                ripple.element.style.transform = `translate3d(${rippleCenterX}px, ${rippleCenterY}px, 0)`;
            }
            
            return true; // Keep in array
        });
    }

    // DNA Animation - Double helix pattern
    updateDNAAnimation(deltaTime) {
        const helixSpeed = 0.002;
        const helixHeight = 200;
        const helixRadius = 60;
        
        this.orbs.forEach((orb, index) => {
            const progress = (this.animationTime * helixSpeed + index * 0.5) % (Math.PI * 4);
            const isFirstStrand = index % 2 === 0;
            
            const x = this.centerX + 
                     Math.cos(progress) * helixRadius * (isFirstStrand ? 1 : -1);
            const y = this.centerY + 
                     (progress / (Math.PI * 4)) * helixHeight - helixHeight / 2;
            
            // Wrap around when reaching the end
            if (y > this.centerY + helixHeight / 2) {
                orb.animationPhase = (orb.animationPhase || 0) + Math.PI * 4;
            }
            
            orb.x += (x - orb.x) * 0.04;
            orb.y += (y - orb.y) * 0.04;
            
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px)`;
        });
    }

    // Loading Animation - Circular loading pattern
    updateLoadingAnimation(deltaTime) {
        const rotationSpeed = 0.005;
        const radius = 80;
        
        this.orbs.forEach((orb, index) => {
            const baseAngle = (index / this.orbs.length) * Math.PI * 2;
            const angle = baseAngle + this.animationTime * rotationSpeed;
            
            // Pulsing effect for loading
            const pulse = Math.sin(this.animationTime * 0.008 - index * 0.5) * 0.5 + 0.5;
            const currentRadius = radius * (0.7 + pulse * 0.3);
            
            const targetX = this.centerX + Math.cos(angle) * currentRadius;
            const targetY = this.centerY + Math.sin(angle) * currentRadius;
            
            orb.x += (targetX - orb.x) * 0.1;
            orb.y += (targetY - orb.y) * 0.1;
            
            // Fade effect based on position
            const opacity = 0.3 + pulse * 0.7;
            orb.element.style.opacity = opacity;
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px)`;
        });
    }

    // Racetrack Animation - Oval track pattern
    updateRacetrackAnimation(deltaTime) {
        const speed = 0.003;
        const radiusX = 150;
        const radiusY = 80;
        
        this.orbs.forEach((orb, index) => {
            const offset = (index / this.orbs.length) * Math.PI * 2;
            const angle = this.animationTime * speed + offset;
            
            const targetX = this.centerX + Math.cos(angle) * radiusX;
            const targetY = this.centerY + Math.sin(angle) * radiusY;
            
            orb.x += (targetX - orb.x) * 0.08;
            orb.y += (targetY - orb.y) * 0.08;
            
            // Add slight trailing effect
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px)`;
        });
    }

    // Wave Animation - Sine wave patterns
    updateWaveAnimation(deltaTime) {
        const waveSpeed = 0.004;
        const waveLength = 200;
        const amplitude = 60;
        const viewportWidth = Math.max(document.documentElement.clientWidth || 0, window.innerWidth || 0);
        
        this.orbs.forEach((orb, index) => {
            const baseX = (index / this.orbs.length) * viewportWidth;
            const wavePhase = this.animationTime * waveSpeed;
            
            const targetX = baseX;
            const targetY = this.centerY + 
                           Math.sin((baseX / waveLength) * Math.PI * 2 + wavePhase) * amplitude;
            
            orb.x += (targetX - orb.x) * 0.05;
            orb.y += (targetY - orb.y) * 0.05;
            
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px)`;
        });
    }

    // Bloom Animation - Expanding flower pattern
    updateBloomAnimation(deltaTime) {
        const bloomSpeed = 0.003;
        const maxRadius = 120;
        
        this.orbs.forEach((orb, index) => {
            const angle = (index / this.orbs.length) * Math.PI * 2;
            const bloomPhase = Math.sin(this.animationTime * bloomSpeed) * 0.5 + 0.5;
            const radius = bloomPhase * maxRadius;
            
            const targetX = this.centerX + Math.cos(angle) * radius;
            const targetY = this.centerY + Math.sin(angle) * radius;
            
            orb.x += (targetX - orb.x) * 0.06;
            orb.y += (targetY - orb.y) * 0.06;
            
            // Scale effect
            const scale = 0.5 + bloomPhase * 0.5;
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px) scale(${scale})`;
        });
    }

    // Pulse Animation - Rhythmic pulsing
    updatePulseAnimation(deltaTime) {
        const pulseSpeed = 0.006;
        const gatherDuration = 2000; // 2 seconds to gather at center
        
        // Initialize pulse animation state if needed
        if (!this.pulseAnimationStartTime) {
            this.pulseAnimationStartTime = this.animationTime;
            this.pulseTimers = this.orbs.map(() => 0);
            this.lastPulseStates = this.orbs.map(() => false);
        }
        
        const timeInAnimation = this.animationTime - this.pulseAnimationStartTime;
        const gatherProgress = Math.min(timeInAnimation / gatherDuration, 1); // 0 to 1 over 2 seconds
        const isGathering = gatherProgress < 1;
        
        this.orbs.forEach((orb, index) => {
            if (isGathering) {
                // Gathering phase - zoom all orbs to center
                const gatherEase = 1 - Math.pow(1 - gatherProgress, 3); // Ease-out cubic for smooth deceleration
                
                orb.x += (this.centerX - orb.x) * 0.08 * gatherEase;
                orb.y += (this.centerY - orb.y) * 0.08 * gatherEase;
                
                // Gradually increase brightness during gather
                const gatherBrightness = 0.3 + gatherProgress * 0.4;
                orb.element.style.opacity = gatherBrightness;
                orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px) scale(${0.8 + gatherProgress * 0.2})`;
            } else {
                // Pulsing phase - start pulsing once gathered
                const delay = index * 0.2;
                const pulse = Math.sin((timeInAnimation - gatherDuration) * pulseSpeed - delay) * 0.5 + 0.5;
                
                // Detect pulse peaks for ripple creation
                const isPulsePeak = pulse > 0.95;
                const wasPulsePeak = this.lastPulseStates[index];
                
                // Create large ripple when pulse reaches peak (only once per peak)
                if (isPulsePeak && !wasPulsePeak) {
                    this.createPulseRippleEffect(orb.x, orb.y, index);
                }
                this.lastPulseStates[index] = isPulsePeak;
                
                // Keep orbs near center with gentle movement
                const centerForce = pulse * 0.01; // Reduced force to keep them more centered
                const dx = this.centerX - orb.x;
                const dy = this.centerY - orb.y;
                
                orb.x += dx * centerForce;
                orb.y += dy * centerForce;
                
                // Pulsing glow effect
                const brightness = 0.5 + pulse * 0.5;
                orb.element.style.opacity = brightness;
                orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px) scale(${0.8 + pulse * 0.4})`;
            }
        });
        
        // Update pulse ripples (only during pulsing phase)
        if (!isGathering) {
            this.updatePulseRipples(deltaTime);
        }
    }
    
    createPulseRippleEffect(x, y, orbIndex) {
        // Initialize pulse ripples array if it doesn't exist
        if (!this.pulseRipples) {
            this.pulseRipples = [];
        }
        
        // Create larger ripple rings for pulse effect
        for (let i = 0; i < 2; i++) {
            const baseRadius = 80 + i * 40; // Base radius: 80px and 120px
            const radiusVariation = 0.8 + Math.random() * 0.4; // Random between 0.8 and 1.2 (±20%)
            const finalRadius = baseRadius * radiusVariation;
            
            const baseSpeed = 40 + i * 20; // Base speed: 40 and 60 pixels/sec
            const speedVariation = 0.8 + Math.random() * 0.4; // Same ±20% variation for speed
            const finalSpeed = baseSpeed * speedVariation;
            
            const ripple = {
                id: Date.now() + Math.random() + i,
                orbIndex: orbIndex,
                radius: 0,
                maxRadius: finalRadius, // Randomized radius
                opacity: 0.6 - i * 0.2, // 0.6 and 0.4 opacity
                speed: finalSpeed, // Randomized speed
                delay: i * 50, // Shorter stagger - 50ms (unchanged)
                age: -i * 50,
                element: null
            };
            
            // Create visual element for pulse ripple
            ripple.element = document.createElement('div');
            ripple.element.className = 'pulse-ripple';
            ripple.element.style.cssText = `
                position: fixed;
                border: 3px solid var(--orb-color-2);
                border-radius: 50%;
                pointer-events: none;
                z-index: 998;
                opacity: 0;
                width: 0px;
                height: 0px;
                transform: translate3d(${x}px, ${y}px, 0);
                transform-origin: center center;
                box-shadow: 0 0 30px var(--orb-color-2);
            `;
            document.body.appendChild(ripple.element);
            
            this.pulseRipples.push(ripple);
        }
    }
    
    updatePulseRipples(deltaTime) {
        if (!this.pulseRipples) return;
        
        this.pulseRipples = this.pulseRipples.filter(ripple => {
            ripple.age += deltaTime;
            
            // Don't start expanding until delay is over
            if (ripple.age < 0) {
                return true;
            }
            
            // Expand the ripple
            ripple.radius += deltaTime * ripple.speed * 0.001;
            
            // Calculate opacity fade - slower fade for pulse effect
            const progress = ripple.radius / ripple.maxRadius;
            const fadeOpacity = ripple.opacity * (1 - progress) * Math.exp(-progress * 1.5);
            
            // Remove when fully expanded or invisible
            if (ripple.radius >= ripple.maxRadius || fadeOpacity < 0.01) {
                if (ripple.element && ripple.element.parentNode) {
                    ripple.element.parentNode.removeChild(ripple.element);
                }
                return false;
            }
            
            // Update visual appearance - follow the orb's current position!
            if (ripple.element && this.orbs[ripple.orbIndex]) {
                const orb = this.orbs[ripple.orbIndex];
                const size = ripple.radius * 2;
                const orbHalfSize = orb.size / 2;
                
                ripple.element.style.width = size + 'px';
                ripple.element.style.height = size + 'px';
                ripple.element.style.opacity = fadeOpacity;
                
                // Position ripple centered on orb
                const rippleCenterX = orb.x + orbHalfSize - ripple.radius;
                const rippleCenterY = orb.y + orbHalfSize - ripple.radius;
                ripple.element.style.transform = `translate3d(${rippleCenterX}px, ${rippleCenterY}px, 0)`;
            }
            
            return true;
        });
    }

    // Constellation Animation - Connected star pattern
    // Vortex Animation - Two counter-rotating tilted circles creating 3D vortex effect
    updateVortexAnimation(deltaTime) {
        const rotationSpeed = 0.004; // Reduced speed by 50% (was 0.008)
        const axisRotationSpeed = 0.001; // Reduced axis rotation by 50% (was 0.002)
        const tiltAngle = Math.PI / 4; // 45-degree tilt
        const radius = 240; // Increased radius by 200% (was 80, now 240)
        const circleOffset = 30; // Distance between the two circles
        
        // Initialize vortex state if needed
        if (!this.vortexPhase) {
            this.vortexPhase = 0;
        }
        if (!this.axisRotation) {
            this.axisRotation = 0;
        }
        
        this.vortexPhase += deltaTime * rotationSpeed;
        this.axisRotation += deltaTime * axisRotationSpeed;
        
        // Split orbs into two groups for the two circles
        const circle1Count = 5; // First 5 orbs for clockwise circle
        const circle2Count = this.orbs.length - circle1Count; // Remaining orbs for counter-clockwise
        
        this.orbs.forEach((orb, index) => {
            let x, y, z, scale;
            
            if (index < circle1Count) {
                // First circle - Clockwise rotation
                const angle = (index / circle1Count) * Math.PI * 2 + this.vortexPhase;
                
                // Calculate 3D position with tilt
                const circleX = Math.cos(angle) * radius;
                const circleY = Math.sin(angle) * radius * Math.cos(tiltAngle); // Compressed by tilt
                const circleZ = Math.sin(angle) * radius * Math.sin(tiltAngle); // Z-depth from tilt
                
                // Apply axis rotation to the entire coordinate system
                const rotatedX = circleX * Math.cos(this.axisRotation) - circleZ * Math.sin(this.axisRotation);
                const rotatedZ = circleX * Math.sin(this.axisRotation) + circleZ * Math.cos(this.axisRotation);
                
                x = this.centerX + rotatedX;
                y = this.centerY + circleY;
                z = rotatedZ;
                
            } else {
                // Second circle - Counter-clockwise rotation, offset position
                const circle2Index = index - circle1Count;
                const angle = -(circle2Index / circle2Count) * Math.PI * 2 - this.vortexPhase; // Negative for counter-clockwise
                
                // Calculate 3D position with tilt, offset by circleOffset
                const circleX = Math.cos(angle) * radius;
                const circleY = Math.sin(angle) * radius * Math.cos(tiltAngle);
                const circleZ = Math.sin(angle) * radius * Math.sin(tiltAngle) + circleOffset; // Offset in Z
                
                // Apply axis rotation to the entire coordinate system
                const rotatedX = circleX * Math.cos(this.axisRotation) - circleZ * Math.sin(this.axisRotation);
                const rotatedZ = circleX * Math.sin(this.axisRotation) + circleZ * Math.cos(this.axisRotation);
                
                x = this.centerX + rotatedX + circleOffset; // Also offset slightly in X for visibility
                y = this.centerY + circleY;
                z = rotatedZ;
            }
            
            // Calculate perspective scaling based on Z-depth
            const perspective = 800; // Perspective distance
            const perspectiveScale = perspective / (perspective + z);
            scale = perspectiveScale * (0.8 + Math.abs(z) / 200); // Size varies with depth
            
            // Apply perspective to X and Y coordinates
            const perspectiveX = x + (x - this.centerX) * (1 - perspectiveScale) * 0.3;
            const perspectiveY = y + (y - this.centerY) * (1 - perspectiveScale) * 0.3;
            
            // Smooth movement to calculated position
            orb.x += (perspectiveX - orb.x) * 0.08;
            orb.y += (perspectiveY - orb.y) * 0.08;
            
            // Apply scaling and slight opacity variation based on depth
            const opacity = 0.7 + (z + 100) / 400; // Orbs further back are slightly dimmer
            const finalScale = Math.max(0.5, Math.min(1.5, scale)); // Clamp scale
            
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px) scale(${finalScale})`;
            orb.element.style.opacity = Math.max(0.4, Math.min(1.0, opacity));
            
            // Add subtle glow effect based on depth
            const glowIntensity = 1 + Math.abs(z) / 150;
            orb.element.style.filter = `brightness(${glowIntensity})`;
        });
    }

    // Swarm Animation - Chaotic buzzing like flies around something that stinks
    updateSwarmAnimation(deltaTime) {
        const swarmIntensity = 0.008;
        const swarmRadius = 60;
        
        // Initialize swarm state if needed
        if (!this.swarmTargets) {
            this.swarmTargets = this.orbs.map(() => ({
                x: this.centerX + (Math.random() - 0.5) * swarmRadius * 2,
                y: this.centerY + (Math.random() - 0.5) * swarmRadius * 2,
                changeTimer: Math.random() * 500
            }));
        }
        
        this.orbs.forEach((orb, index) => {
            const target = this.swarmTargets[index];
            
            // Update target position randomly (erratic buzzing)
            target.changeTimer -= deltaTime;
            if (target.changeTimer <= 0) {
                target.x = this.centerX + (Math.random() - 0.5) * swarmRadius * 2;
                target.y = this.centerY + (Math.random() - 0.5) * swarmRadius * 2;
                target.changeTimer = 100 + Math.random() * 400; // Change direction every 0.1-0.5 seconds
            }
            
            // Chaotic movement with lots of buzzing
            const buzzX = Math.sin(this.animationTime * swarmIntensity + index) * 15;
            const buzzY = Math.cos(this.animationTime * swarmIntensity * 1.3 + index) * 15;
            
            // Add random jittery movement
            const jitterX = (Math.random() - 0.5) * 8;
            const jitterY = (Math.random() - 0.5) * 8;
            
            const finalTargetX = target.x + buzzX + jitterX;
            const finalTargetY = target.y + buzzY + jitterY;
            
            // Quick, erratic movement to target
            orb.x += (finalTargetX - orb.x) * 0.12;
            orb.y += (finalTargetY - orb.y) * 0.12;
            
            // Rapid scale changes for buzzing effect
            const scale = 0.8 + Math.sin(this.animationTime * 0.02 + index * 0.5) * 0.3;
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px) scale(${scale})`;
        });
    }

    // Dance Animation - Pairs spinning and switching partners
    updateDanceAnimation(deltaTime) {
        const spinSpeed = 0.00125; // Reduced from 0.0025 (50% slower again)
        const partnerSwitchTime = 6750; // Increased from 4500ms (50% slower again)
        const switchDuration = 1500; // Increased from 1000ms (50% slower travel)
        
        // Initialize dance state if needed
        if (!this.danceState) {
            this.danceState = {
                currentPairs: this.createCircularDancePairs(),
                switchTimer: partnerSwitchTime,
                spinPhase: 0,
                isSwitching: false,
                switchProgress: 0,
                travelingPartners: []
            };
        }
        
        // Update partner switch timer
        this.danceState.switchTimer -= deltaTime;
        
        // Start partner switching
        if (this.danceState.switchTimer <= 0 && !this.danceState.isSwitching) {
            this.startPartnerSwitch();
            this.danceState.isSwitching = true;
            this.danceState.switchProgress = 0;
        }
        
        // Handle partner switching animation
        if (this.danceState.isSwitching) {
            this.danceState.switchProgress += deltaTime / switchDuration;
            
            if (this.danceState.switchProgress >= 1) {
                // Switch complete
                this.completePartnerSwitch();
                this.danceState.isSwitching = false;
                this.danceState.switchTimer = partnerSwitchTime;
            }
        }
        
        // Update spin phase (slower during partner switch)
        const spinMultiplier = this.danceState.isSwitching ? 0.3 : 1.0;
        this.danceState.spinPhase += deltaTime * spinSpeed * spinMultiplier;
        
        // Position orbs in dancing pairs
        this.danceState.currentPairs.forEach((pair, pairIndex) => {
            const pairAngle = (pairIndex / this.danceState.currentPairs.length) * Math.PI * 2;
            const pairCenterX = this.centerX + Math.cos(pairAngle) * 120;
            const pairCenterY = this.centerY + Math.sin(pairAngle) * 120;
            const radius = 40;
            
            pair.forEach((orbData, posInPair) => {
                if (orbData.orbIndex < this.orbs.length) {
                    const orb = this.orbs[orbData.orbIndex];
                    let targetX, targetY;
                    
                    // Check if this orb is currently traveling to a new partner
                    const travelingData = this.danceState.travelingPartners.find(t => t.orbIndex === orbData.orbIndex);
                    
                    if (travelingData && this.danceState.isSwitching) {
                        // Interpolate between current and target dance circle
                        const progress = this.easeInOutCubic(this.danceState.switchProgress);
                        targetX = travelingData.startX + (travelingData.endX - travelingData.startX) * progress;
                        targetY = travelingData.startY + (travelingData.endY - travelingData.startY) * progress;
                    } else {
                        // Normal dance position
                        const angle = this.danceState.spinPhase + (posInPair * Math.PI);
                        targetX = pairCenterX + Math.cos(angle) * radius;
                        targetY = pairCenterY + Math.sin(angle) * radius;
                    }
                    
                    // Smooth movement to dance positions
                    orb.x += (targetX - orb.x) * 0.08;
                    orb.y += (targetY - orb.y) * 0.08;
                    
                    // Add a little bounce to the dance
                    const bounce = 1 + Math.sin(this.danceState.spinPhase * 0.75 + orbData.orbIndex) * 0.2; // Reduced from *1.5 to *0.75
                    orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px) scale(${bounce})`;
                }
            });
        });
    }
    
    // Create circular arrangement of dance pairs
    createCircularDancePairs() {
        const numPairs = Math.floor(this.orbs.length / 2);
        const pairs = [];
        
        for (let i = 0; i < numPairs; i++) {
            pairs.push([
                { orbIndex: i * 2, isStaying: true },     // First orb stays in this circle
                { orbIndex: i * 2 + 1, isStaying: false } // Second orb will travel
            ]);
        }
        
        // Handle odd orb (solo dancer)
        if (this.orbs.length % 2 === 1) {
            pairs.push([{ orbIndex: this.orbs.length - 1, isStaying: true }]);
        }
        
        return pairs;
    }
    
    // Start the partner switching process
    startPartnerSwitch() {
        this.danceState.travelingPartners = [];
        
        // Find all traveling partners and their destinations
        this.danceState.currentPairs.forEach((pair, pairIndex) => {
            const travelingOrb = pair.find(orbData => !orbData.isStaying);
            if (travelingOrb && pair.length === 2) {
                const nextPairIndex = (pairIndex + 1) % this.danceState.currentPairs.length;
                
                // Calculate current and target positions
                const currentPairAngle = (pairIndex / this.danceState.currentPairs.length) * Math.PI * 2;
                const targetPairAngle = (nextPairIndex / this.danceState.currentPairs.length) * Math.PI * 2;
                
                const currentCenterX = this.centerX + Math.cos(currentPairAngle) * 120;
                const currentCenterY = this.centerY + Math.sin(currentPairAngle) * 120;
                const targetCenterX = this.centerX + Math.cos(targetPairAngle) * 120;
                const targetCenterY = this.centerY + Math.sin(targetPairAngle) * 120;
                
                this.danceState.travelingPartners.push({
                    orbIndex: travelingOrb.orbIndex,
                    startX: currentCenterX,
                    startY: currentCenterY,
                    endX: targetCenterX,
                    endY: targetCenterY,
                    sourcePair: pairIndex,
                    targetPair: nextPairIndex
                });
            }
        });
    }
    
    // Complete the partner switch
    completePartnerSwitch() {
        // Rebuild pairs with new partners
        const newPairs = [];
        
        this.danceState.currentPairs.forEach((pair, pairIndex) => {
            const stayingOrb = pair.find(orbData => orbData.isStaying);
            
            if (stayingOrb) {
                // Find the orb that's traveling to this pair
                const incomingTraveler = this.danceState.travelingPartners.find(t => t.targetPair === pairIndex);
                
                if (incomingTraveler) {
                    newPairs.push([
                        stayingOrb,
                        { orbIndex: incomingTraveler.orbIndex, isStaying: false }
                    ]);
                } else {
                    // Solo dancer
                    newPairs.push([stayingOrb]);
                }
            }
        });
        
        this.danceState.currentPairs = newPairs;
        this.danceState.travelingPartners = [];
    }
    
    // Easing function for smooth partner travel
    easeInOutCubic(t) {
        return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
    }
    
    // Helper method to create random dance pairs (keeping for compatibility)
    createDancePairs() {
        return this.createCircularDancePairs();
    }

    // Cascade Animation - Following chain movement
    updateCascadeAnimation(deltaTime) {
        const cascadeSpeed = 0.004;
        
        this.orbs.forEach((orb, index) => {
            const delay = index * 0.3;
            const time = this.animationTime * cascadeSpeed - delay;
            
            if (time > 0) {
                const progress = time % (Math.PI * 2);
                const radius = 100;
                
                const targetX = this.centerX + Math.cos(progress) * radius;
                const targetY = this.centerY + Math.sin(progress) * radius * 0.6;
                
                orb.x += (targetX - orb.x) * 0.1;
                orb.y += (targetY - orb.y) * 0.1;
            }
            
            orb.element.style.transform = `translate(${orb.x}px, ${orb.y}px)`;
        });
    }
}

// Initialize animation controller
const animationController = new AnimationController();

// Connect with orb manager when ready
if (window.orbManager) {
    animationController.setOrbs(window.orbManager.getOrbs());
}

// Make it globally available
window.animationController = animationController;
