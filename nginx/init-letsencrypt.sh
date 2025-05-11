#!/bin/bash

# This script will set up Let's Encrypt certificates for the domain

# Exit on error
set -e

# Domain and email for Let's Encrypt
domains=(${DOMAIN:-localhost})
email="${EMAIL:-admin@example.com}"
staging=0 # Set to 1 if you're testing your setup to avoid hitting request limits

# Get the server IP from environment or use localhost
server_ip=${SERVER_IP:-localhost}

# Replace domain.com with the actual domain in nginx config
sed -i "s/domain.com/$server_ip/g" /etc/nginx/conf.d/default.conf

# Create required directories
mkdir -p /etc/letsencrypt/live/$server_ip
mkdir -p /var/www/certbot

# Check if certificates already exist
if [ -d "/etc/letsencrypt/live/$server_ip" ]; then
  echo "Certificates already exist, skipping certificate generation"
  exit 0
fi

# Create dummy certificates
echo "Creating dummy certificates for $server_ip..."
openssl req -x509 -nodes -newkey rsa:2048 -days 1 \
  -keyout "/etc/letsencrypt/live/$server_ip/privkey.pem" \
  -out "/etc/letsencrypt/live/$server_ip/fullchain.pem" \
  -subj "/CN=$server_ip"

echo "Starting nginx..."
nginx -g "daemon off;" &
nginx_pid=$!

# Wait for nginx to start
sleep 5

# Request Let's Encrypt certificate
echo "Requesting Let's Encrypt certificate for $server_ip..."
if [ $staging -eq 1 ]; then
  staging_arg="--staging"
else
  staging_arg=""
fi

certbot certonly --webroot -w /var/www/certbot \
  $staging_arg \
  --email $email \
  -d $server_ip \
  --agree-tos \
  --force-renewal \
  --non-interactive

# Reload nginx
echo "Reloading nginx..."
kill -HUP $nginx_pid

# Wait for nginx to reload
sleep 5

echo "Let's Encrypt certificates have been successfully set up!"
