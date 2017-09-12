FROM alpine
RUN apk add --update bash
ADD assets/ /opt/resource/
