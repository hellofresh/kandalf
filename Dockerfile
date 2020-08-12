FROM ubuntu:20.04

ADD kandalf /bin/kandalf
RUN chmod a+x /bin/kandalf

RUN mkdir -p /etc/kandalf/conf
ADD assets/pipes.yml /etc/kandalf/conf/

# Use nobody user + group
USER 65534:65534

ENTRYPOINT ["/bin/kandalf"]

# just to have it
RUN ["/bin/kandalf", "--version"]
