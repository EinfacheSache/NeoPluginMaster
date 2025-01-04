# Temp container to build using Go for ARM/ARM64
FROM --platform=$BUILDPLATFORM golang:1.23-bookworm AS builder

# Set environment variables for ARM architecture
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ENV APP_HOME=/usr/app/
WORKDIR $APP_HOME

# Copy application source code
COPY . .

# Ensure the build script is executable
RUN chmod +x build.sh

# Build the application for the target architecture
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} ./build.sh

# Actual runtime container
FROM --platform=$TARGETPLATFORM alpine:3
ENV ARTIFACT_NAME=neo-plugin-master
ENV APP_HOME=/usr/app/
ARG USERNAME=neopluginmaster
WORKDIR $APP_HOME

# Copy the built application from the builder container
COPY --from=builder $APP_HOME/$ARTIFACT_NAME .

# Create a user and set permissions
RUN addgroup -S $USERNAME && \
    adduser -S $USERNAME -G $USERNAME && \
    chown $USERNAME:$USERNAME $ARTIFACT_NAME && \
    chmod +x $ARTIFACT_NAME

# Switch to the non-root user
USER $USERNAME

# Expose application ports
EXPOSE 8069
EXPOSE 8080

# Run the application
CMD ["./neo-plugin-master"]
