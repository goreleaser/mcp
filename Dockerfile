FROM scratch
ARG TARGETPLATFORM
ENTRYPOINT [ "/usr/bin/example" ]
COPY $TARGETPLATFORM/example /usr/bin/example
