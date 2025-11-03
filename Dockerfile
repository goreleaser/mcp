FROM scratch
ARG TARGETPLATFORM
ENTRYPOINT [ "/usr/bin/goreleaser-mcp" ]
COPY $TARGETPLATFORM/goreleaser-mcp /usr/bin/goreleaser-mcp
