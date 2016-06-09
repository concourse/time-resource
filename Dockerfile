FROM concourse/busyboxplus:base
ADD assets/ /opt/resource/

ENV ZONEINFO=/opt/resource/zoneinfo
