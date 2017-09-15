FROM alpine
RUN apk add --update bash tzdata
ADD assets/ /opt/resource/
