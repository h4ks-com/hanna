#!/bin/bash

# HannaUI Quick Start Script for Docker
# This script helps you get HannaUI running quickly with Docker

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${PURPLE}🌟 HannaUI Docker Quick Start${NC}"
echo "=================================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed or not in PATH${NC}"
    echo "Please install Docker first: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not available${NC}"
    echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

echo -e "${GREEN}✅ Docker environment ready${NC}"
echo ""

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}⚠️  No .env file found${NC}"
    if [ -f ".env.example" ]; then
        echo "Creating .env from template..."
        cp .env.example .env
        echo -e "${GREEN}✅ Created .env file${NC}"
    else
        echo "Creating basic .env file..."
        cat > .env << EOF
# HannaUI Environment Configuration
HANNA_N8N_ENDPOINT=
EOF
        echo -e "${GREEN}✅ Created basic .env file${NC}"
    fi
    echo ""
    echo -e "${BLUE}💡 Edit .env file to configure your n8n endpoint:${NC}"
    echo "   nano .env"
    echo ""
fi

# Show current configuration
if [ -f ".env" ]; then
    source .env
    if [ -n "$HANNA_N8N_ENDPOINT" ]; then
        echo -e "${GREEN}🔗 n8n endpoint configured: $HANNA_N8N_ENDPOINT${NC}"
        echo -e "${GREEN}   HannaUI will connect to real AI${NC}"
    else
        echo -e "${YELLOW}🎭 No n8n endpoint configured${NC}"
        echo -e "${YELLOW}   HannaUI will run in demo mode${NC}"
    fi
    echo ""
fi

# Ask user what they want to do
echo "What would you like to do?"
echo "1) Start HannaUI (build and run)"
echo "2) Stop HannaUI"
echo "3) View logs"
echo "4) Rebuild and restart"
echo "5) Configure n8n endpoint"
echo ""
read -p "Enter your choice (1-5): " choice

case $choice in
    1)
        echo -e "${BLUE}🚀 Starting HannaUI...${NC}"
        docker-compose up -d --build
        if [ $? -eq 0 ]; then
            echo ""
            echo -e "${GREEN}✅ HannaUI is running!${NC}"
            echo ""
            echo -e "${BLUE}🌐 Access your application:${NC}"
            echo "   http://localhost:3000"
            echo ""
            echo -e "${BLUE}👤 Demo login credentials:${NC}"
            echo "   Username: admin"
            echo "   Password: admin123"
            echo ""
            echo -e "${BLUE}📊 View logs:${NC}"
            echo "   docker-compose logs -f hannaui"
        else
            echo -e "${RED}❌ Failed to start HannaUI${NC}"
        fi
        ;;
    2)
        echo -e "${YELLOW}🛑 Stopping HannaUI...${NC}"
        docker-compose down
        echo -e "${GREEN}✅ HannaUI stopped${NC}"
        ;;
    3)
        echo -e "${BLUE}📋 Showing HannaUI logs (Ctrl+C to exit):${NC}"
        docker-compose logs -f hannaui
        ;;
    4)
        echo -e "${BLUE}🔄 Rebuilding and restarting HannaUI...${NC}"
        docker-compose down
        docker-compose up -d --build
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ HannaUI rebuilt and restarted!${NC}"
            echo "   Access at: http://localhost:3000"
        else
            echo -e "${RED}❌ Failed to rebuild HannaUI${NC}"
        fi
        ;;
    5)
        echo -e "${BLUE}🔧 Configure n8n endpoint${NC}"
        echo ""
        echo "Current setting:"
        if [ -n "$HANNA_N8N_ENDPOINT" ]; then
            echo "  HANNA_N8N_ENDPOINT=$HANNA_N8N_ENDPOINT"
        else
            echo "  HANNA_N8N_ENDPOINT=(not set - demo mode)"
        fi
        echo ""
        echo "Examples:"
        echo "  http://localhost:5678/webhook/hanna-chat"
        echo "  https://your-n8n.example.com/webhook/hanna-chat"
        echo "  (leave empty for demo mode)"
        echo ""
        read -p "Enter new n8n endpoint URL (or press Enter to keep current): " new_endpoint
        
        if [ -n "$new_endpoint" ]; then
            # Update .env file
            if grep -q "HANNA_N8N_ENDPOINT=" .env; then
                sed -i "s|HANNA_N8N_ENDPOINT=.*|HANNA_N8N_ENDPOINT=$new_endpoint|" .env
            else
                echo "HANNA_N8N_ENDPOINT=$new_endpoint" >> .env
            fi
            echo -e "${GREEN}✅ Updated n8n endpoint: $new_endpoint${NC}"
        else
            echo -e "${BLUE}💡 Keeping current configuration${NC}"
        fi
        
        echo ""
        echo -e "${YELLOW}🔄 Restart HannaUI to apply changes:${NC}"
        echo "   ./start.sh (option 4)"
        ;;
    *)
        echo -e "${RED}❌ Invalid choice${NC}"
        exit 1
        ;;
esac
