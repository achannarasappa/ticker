FROM golang:1.20-alpine

# Set environment variable for terminal colors
ENV TERM=xterm-256color

# Set the working directory
WORKDIR /ticker

# Download dependencies (this happens once, not every build)
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Expose the volume for configuration file
VOLUME ["/ticker/.ticker.yaml"]

# Default entrypoint
CMD ["sh"]

