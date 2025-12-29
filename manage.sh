#!/bin/bash

# LevPay Backend Management Script
# This script manages Docker containers and system health for the LevPay platform.
# Usage: ./manage.sh [command] [options]

set -e

# Colors for output - Vibrant Palette
RED='\033[38;2;255;85;85m'
GREEN='\033[38;2;80;250;123m'
YELLOW='\033[38;2;241;250;140m'
BLUE='\033[38;2;139;233;253m'
MAGENTA='\033[38;2;255;121;198m'
CYAN='\033[38;2;139;233;253m'
ORANGE='\033[38;2;255;184;108m'
PURPLE='\033[38;2;189;147;249m'
WHITE='\033[38;2;255;255;255m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.dev.yaml"
ENV_FILE=".env"
PROFILE="full"

print_header() {
    clear
    echo -e "${PURPLE}${BOLD}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}${BOLD}║${NC}  ${CYAN}${BOLD}LevPay${NC} - ${WHITE}Premium Backend Manager${NC}                     ${PURPLE}${BOLD}║${NC}"
    echo -e "${PURPLE}${BOLD}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_info() {
    echo -e "${BLUE}✨ $1${NC}"
}

print_success() {
    echo -e "${GREEN}🚀 $1${NC}"
}

print_error() {
    echo -e "${RED}💢 $1${NC}"
}

print_warning() {
    echo -e "${ORANGE}⚠️  $1${NC}"
}

print_step() {
    echo -e "${MAGENTA}${BOLD}⚡ $1${NC}"
}

# Check directory
check_directory() {
    if [ ! -f "go.mod" ] || [ ! -f "$COMPOSE_FILE" ]; then
        print_error "Error: Please run this script from the LevPay project root directory."
        exit 1
    fi
}

show_usage() {
    print_header
    echo -e "${BOLD}Usage:${NC} $0 [command] [options]"
    echo ""
    echo -e "${CYAN}${BOLD}💎 DOCKER OPERATIONS${NC}"
    echo "  up              Start all services (Full System)"
    echo "  up-minimal      Start core services only"
    echo "  down            Stop all services"
    echo "  clean           Stop services and wipe all volumes (Fresh Start)"
    echo "  restart         Restart the entire system"
    echo "  restart <svc>   Restart a specific service"
    echo "  rebuild         Rebuild all images and restart"
    echo "  rebuild <svc>   Rebuild a specific service"
    echo ""
    echo -e "${ORANGE}${BOLD}🔍 MONITORING & TOOLS${NC}"
    echo "  status          Display real-time system status"
    echo "  logs            Stream all logs"
    echo "  logs <svc>      Stream logs for a specific service"
    echo "  health          Perform a deep system health check"
    echo "  shell <svc>     Open an interactive shell in a container"
    echo ""
    echo -e "${GREEN}${BOLD}👤 IDENTITY & ACCESS${NC}"
    echo "  register        Interactively create a new user/merchant"
    echo ""
    echo -e "${BLUE}${BOLD}❓ HELP${NC}"
    echo "  help            Display this luxury menu"
    echo ""
}

# Start services
start_services() {
    local profile="${1:-$PROFILE}"
    print_step "Launching LevPay Ecosystem (Profile: $profile)..."
    
    # Check for .env
    if [ ! -f "$ENV_FILE" ]; then
        print_warning ".env file missing. Creating from .env.example..."
        cp .env.example .env || touch .env
    fi

    docker compose -f "$COMPOSE_FILE" --profile "$profile" up -d
    
    print_info "Waiting for core dependencies to stabilize..."
    sleep 3
    
    print_success "LevPay is rising! System is coming online."
    show_status

    print_info "Attaching to global system pulse (Ctrl+C to switch to background mode)..."
    docker compose -f "$COMPOSE_FILE" --profile "$profile" logs -f
}

# Stop services
stop_services() {
    local mode="$1"
    if [ "$mode" == "clean" ]; then
        print_step "Performing Deep System Wipe (Volumes + Containers)..."
        docker compose -f "$COMPOSE_FILE" down -v --remove-orphans
        print_success "System cleaned. All data volumes purged."
    else
        print_step "Shutting down LevPay Ecosystem..."
        docker compose -f "$COMPOSE_FILE" down
        print_success "Services stopped successfully."
    fi
}

# Restart services
restart_services() {
    local service="$1"
    if [ -n "$service" ]; then
        print_step "Restarting Component: $service..."
        docker compose -f "$COMPOSE_FILE" restart "$service"
        print_success "Component '$service' restarted."
        
        print_info "Streaming pulse for $service (Ctrl+C to exit)..."
        docker compose -f "$COMPOSE_FILE" logs -f "$service"
    else
        print_step "Restarting Entire Ecosystem..."
        docker compose -f "$COMPOSE_FILE" restart
        print_success "System rebooted."
        
        print_info "Attaching to global pulse (Ctrl+C to exit)..."
        docker compose -f "$COMPOSE_FILE" logs -f --tail=100
    fi
}

# Rebuild services
rebuild_services() {
    local service="$1"
    if [ -n "$service" ]; then
        print_step "Rebuilding Component: $service..."
        docker compose -f "$COMPOSE_FILE" build "$service"
        docker compose -f "$COMPOSE_FILE" up -d --no-deps "$service"
        print_success "Component '$service' rebuilt and redeployed."
        
        print_info "Streaming pulse for $service (Ctrl+C to exit)..."
        docker compose -f "$COMPOSE_FILE" logs -f "$service"
    else
        print_step "Performing Full System Rebuild..."
        docker compose -f "$COMPOSE_FILE" build --parallel
        start_services
    fi
}

# Show status
show_status() {
    print_step "System Component Status:"
    echo ""
    docker compose -f "$COMPOSE_FILE" ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    echo ""
}

# Show logs
show_logs() {
    local service="$1"
    if [ -n "$service" ]; then
        print_step "Streaming Pulse: $service"
        docker compose -f "$COMPOSE_FILE" logs -f "$service"
    else
        print_step "Streaming Global System Pulse (Ctrl+C to exit)..."
        docker compose -f "$COMPOSE_FILE" logs -f --tail=100
    fi
}

# Health check
check_health() {
    print_step "Performing Deep Health Diagnostics..."
    echo ""
    
    local services=("app" "auth" "user" "wallet" "transaction" "kyc" "file" "notification" "billing" "admin" "cron")
    local ports=(5001 5012 5003 5004 5005 5006 5007 5008 5009 5010 5011) # Note: checking ports
    
    # Wait a bit if just started
    print_info "Polling service endpoints..."
    
    local healthy=0
    local total=${#services[@]}

    for i in "${!services[@]}"; do
        local svc=${services[$i]}
        
        # Determine port from docker-compose.dev.yaml mapping
        local port=0
        case $svc in
            app) port=5001 ;;
            auth) port=5002 ;;
            user) port=5003 ;;
            wallet) port=5004 ;;
            transaction) port=5005 ;;
            kyc) port=5006 ;;
            file) port=5007 ;;
            notification) port=5008 ;;
            billing) port=5009 ;;
            admin) port=5010 ;;
            cron) port=5011 ;;
        esac

        printf "  %-15s [%d] ... " "$svc" "$port"
        
        # Try up to 3 times for transient failures
        local status="OFFLINE"
        for try in {1..2}; do
            if curl -s -f --connect-timeout 2 "http://localhost:$port/health" > /dev/null; then
                status="ONLINE"
                break
            fi
            sleep 0.5
        done

        if [ "$status" == "ONLINE" ]; then
            echo -e "${GREEN}${BOLD}ONLINE${NC}"
            ((healthy++))
        else
            echo -e "${RED}${BOLD}OFFLINE${NC}"
        fi
    done
    
    printf "  %-15s [4003] ... " "gateway"
    if curl -s -f --connect-timeout 2 "http://localhost:4003/health" > /dev/null; then
        echo -e "${GREEN}${BOLD}ONLINE${NC}"
    else
        echo -e "${RED}${BOLD}OFFLINE${NC}"
    fi

    echo ""
    if [ $healthy -eq $total ]; then
        print_success "All systems are GO. LevPay is fully operational."
    else
        print_warning "System is partially operational ($healthy/$total services online)."
        echo "Check logs for offline services: ./manage.sh logs <service>"
    fi
    echo ""
}

# Registration Tool
register_user() {
    print_step "Interactive User Registration"
    
    # Get details
    read -p "Full Name: " name
    read -p "Email Address: " email
    read -s -p "Password: " password
    echo ""
    read -p "Role (user/merchant/admin) [default: user]: " role
    role=${role:-user}

    if [ -z "$email" ] || [ -z "$password" ]; then
        print_error "Email and Password are required."
        return 1
    fi

    print_info "Initiating registration for $email..."
    go run cmd/tools/register.go -name "$name" -email "$email" -password "$password" -role "$role"
}

# Main
main() {
    check_directory
    
    local cmd="${1:-help}"
    shift || true
    
    case "$cmd" in
        up)             print_header; start_services "full" ;;
        up-minimal)     print_header; start_services "minimal" ;;
        down)           print_header; stop_services ;;
        clean)          print_header; stop_services "clean" ;;
        restart)        print_header; restart_services "$1" ;;
        rebuild)        print_header; rebuild_services "$1" ;;
        status)         print_header; show_status ;;
        logs)           show_logs "$1" ;;
        health)         print_header; check_health ;;
        shell)          docker compose -f "$COMPOSE_FILE" exec "$1" sh ;;
        register)       print_header; register_user ;;
        help|--help|-h) show_usage ;;
        *)              print_error "Unknown directive: $cmd"; show_usage; exit 1 ;;
    esac
}

main "$@"
