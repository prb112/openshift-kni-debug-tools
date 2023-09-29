FROM registry.access.redhat.com/ubi9/openssl:latest
RUN microdnf install -y hwdata && \
    microdnf clean -y all
COPY _output /usr/local/bin
COPY run.sh /run.sh
COPY help.sh /help.sh
# no tool is more important than others
ENTRYPOINT ["/run.sh"]
CMD ["/help.sh"]
