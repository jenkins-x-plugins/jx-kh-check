FROM scratch

ENTRYPOINT ["/jx-kh-check"]

COPY ./build/linux/jx-kh-check /jx-kh-check