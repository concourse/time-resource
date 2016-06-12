FROM concourse/buildroot:base
ADD assets/ /opt/resource/

ENV ZONEINFO=/opt/resource/zoneinfo
