FROM nginx:alpine

# Install certbot for Let's Encrypt
RUN apk add --no-cache certbot openssl bash

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config and init script
COPY default.conf /etc/nginx/conf.d/
COPY init-letsencrypt.sh /

# Make init script executable
RUN chmod +x /init-letsencrypt.sh

# Create directories for Let's Encrypt
RUN mkdir -p /etc/letsencrypt
RUN mkdir -p /var/www/certbot

# Expose ports
EXPOSE 80 443

# Start nginx with Let's Encrypt initialization
CMD ["/init-letsencrypt.sh"]
