FROM alpine:latest

RUN apk update && apk add --no-cache \
    curl \
    openssh-client \
    bash


ENV SENDEREMAIL=YOUREMAIL \
    SENDERPASS=YOURPASS \
    GKTOKEN=GKTOKEN

RUN echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config

RUN apk add --no-cache tzdata
ENV TZ=Europe/Lisbon

WORKDIR /app

COPY . /app

CMD ["./nutgmonitor"]