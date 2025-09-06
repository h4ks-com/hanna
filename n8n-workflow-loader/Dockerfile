FROM alpine:latest

# Install curl, jq, bash, and docker for CLI operations
RUN apk --no-cache add curl jq bash docker-cli

# Copy the workflow loader script
COPY load-workflows.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/load-workflows.sh

# Copy the workflows directory
COPY workflows/ /workflows/

CMD ["/usr/local/bin/load-workflows.sh"]