# README

## Monitoring and Managing UPS Power with Docker and Automation

After implementing the initial setup, our goal is to monitor the UPS and automatically turn off devices powered by it when necessary.

With this we have several options, however my approach is to create another docker container that has a webhook and waits for a request from Prometheus Alert Manager. Alert manager will check the nut_status value and if that is equal 2 it will send a curl request to a docker container running on Pinute. After that the docker container will ssh into every target and shut down them

To ssh we create ssh keys inside Pinute, go to the targets and add `no-pty,no-X11-forwarding,command="sudo /root/off" ssh-rsa ...` in target's `.ssh/authorized_keys`.
Then we create a script in `/root/` called `off` with `halt -f -f -p` to turn off the target

This information is forwarded to the telegram bot by making a request to a special endpoint, in order to telegram bot forward that message to me. In the end an email is sent with all the logs.

## Docker

- Example of Dockerfile

``` yml
FROM alpine:latest

RUN apk update && apk add --no-cache \
    curl \
    openssh-client \
    bash


ENV SENDEREMAIL=EMAIL \
    SENDERPASS=PASS \
    GKTOKEN=GKTOKEN

RUN echo "StrictHostKeyChecking no" >> /etc/ssh/ssh_config

WORKDIR /app

COPY . /app

CMD ["./nutgmonitor"]
```

- Example of docker-compose.yml

``` yml
services:
  nutgmonitor:
    build:
      context: .  # Path to the Dockerfile
    container_name: nutgmonitor-container
    privileged: true  # Run the container in privileged mode
    restart: always
    environment:
      SENDEREMAIL: SENDEREMAIL
      SENDERPASS: SENDERPASS
      GKTOKEN: GKTOKEN
    volumes:
      - ./logs:/app/logs
      - /home/brun0/nut/nut/upslog.txt:/app/logs/upslog/upslog.txt
    ports:
      - "9999:9999"
```

## Build

We can just run `make build` (adjust the TARGET_ARCH inside `Makefile`) and then we can run as a common binary with `./nutgmonitor`

However I am using Docker. 

``` console
$ make run
$ docker compose up -d --build
```
