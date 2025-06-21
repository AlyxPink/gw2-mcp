#!/bin/bash
# Post-installation script for gw2-mcp

# Create systemd service file if systemd is available
if command -v systemctl >/dev/null 2>&1; then
    cat > /etc/systemd/system/gw2-mcp.service << EOF
[Unit]
Description=Guild Wars 2 Model Context Provider Server
After=network.target

[Service]
Type=simple
User=nobody
Group=nobody
ExecStart=/usr/bin/gw2-mcp
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd and enable the service
    systemctl daemon-reload
    systemctl enable gw2-mcp.service
fi

echo "gw2-mcp installed successfully!"
echo "To start the service: sudo systemctl start gw2-mcp"
echo "To run manually: gw2-mcp"