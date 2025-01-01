# temp container to build using gradle
FROM golang:1.20-bookworm
ENV APP_HOME=/usr/app/
WORKDIR $APP_HOME
COPY . .
RUN chmod +x build.sh && ./build.sh
RUN ./build.sh
# actual container
FROM alpine:3
ENV ARTIFACT_NAME=neo-plugin-master
ENV APP_HOME=/usr/app/
ARG USERNAME=neopluginmaster
WORKDIR $APP_HOME
COPY --from=0 $APP_HOME/$ARTIFACT_NAME .
RUN addgroup -S $USERNAME && \
    adduser -S $USERNAME -G $USERNAME && \
    chown $USERNAME:$USERNAME $ARTIFACT_NAME && \
    chmod +x $ARTIFACT_NAME
USER $USERNAME
EXPOSE 8069
EXPOSE 8080
CMD ["./neo-plugin-master"]