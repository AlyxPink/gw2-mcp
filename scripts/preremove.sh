#!/bin/bash
# Pre-removal script for gw2-mcp

# Stop and disable the service if systemd is available
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet gw2-mcp.service; then
        echo "Stopping gw2-mcp service..."
        systemctl stop gw2-mcp.service
    fi
    
    if systemctl is-enabled --quiet gw2-mcp.service; then
        echo "Disabling gw2-mcp service..."
        systemctl disable gw2-mcp.service
    fi
    
    # Remove the service file
    if [ -f /etc/systemd/system/gw2-mcp.service ]; then
        rm -f /etc/systemd/system/gw2-mcp.service
        systemctl daemon-reload
    fi
fi

echo "gw2-mcp service stopped and disabled."