FROM --platform=$BUILDPLATFORM golang:1.20-bookworm AS builder

# Build setup for ARM
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV APP_HOME=/usr/app/
WORKDIR $APP_HOME
COPY . .
RUN chmod +x build.sh && ./build.sh

FROM --platform=$TARGETPLATFORM arm64v8/alpine:3
ENV ARTIFACT_NAME=neo-plugin-master
ENV APP_HOME=/usr/app/
ARG USERNAME=neopluginmaster
WORKDIR $APP_HOME
COPY --from=builder $APP_HOME/$ARTIFACT_NAME .
RUN addgroup -S $USERNAME && \
    adduser -S $USERNAME -G $USERNAME && \
    chown $USERNAME:$USERNAME $ARTIFACT_NAME && \
    chmod +x $ARTIFACT_NAME
USER $USERNAME
EXPOSE 8069
EXPOSE 8080
CMD ["./neo-plugin-master"]