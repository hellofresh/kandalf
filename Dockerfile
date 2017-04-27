FROM alpine
ADD dist/kandalf_linux-amd64 /
RUN mkdir -p /etc/kandalf/conf
ADD assets/pipes.yml /etc/kandalf/conf/
ENTRYPOINT ["/kandalf_linux-amd64"]
