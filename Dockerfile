FROM ubuntu:latest

RUN apt-get update
RUN apt-get install -y net-tools curl iputils-ping
