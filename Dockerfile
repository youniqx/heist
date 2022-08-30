FROM gcr.io/distroless/static:nonroot
WORKDIR /

ARG BINARY_PATH=bin/heist
COPY --chown=65532:65532 "${BINARY_PATH}" /bin/heist

USER 65532:65532
ENTRYPOINT ["/bin/heist"]