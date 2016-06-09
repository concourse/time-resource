FROM concourse/busyboxplus:base
ADD assets/ /opt/resource/

ADD zoneinfo/ /opt/resource/zoneinfo
ENV ZONEINFO=/opt/resource
