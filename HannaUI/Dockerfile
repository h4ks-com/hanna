# HannaUI Docker Container
FROM nginx:alpine

# Set working directory
WORKDIR /usr/share/nginx/html

# Copy all web files
COPY *.html ./
COPY *.css ./
COPY *.js ./
COPY *.md ./

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Create environment injection script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Install envsubst for environment variable substitution
RUN apk add --no-cache gettext

# Expose port 80
EXPOSE 80

# Use custom entrypoint to inject environment variables
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
