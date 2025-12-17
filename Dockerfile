FROM ubuntu:24.04
LABEL authors="M41den"

RUN mkdir /app
RUN apt update && apt install -y curl
COPY ./build/particle-linux-amd64 /app/particle
RUN chmod +x /app/particle


ENTRYPOINT ["/bin/bash", "-c", "/app/particle `echo $ARGS`"]