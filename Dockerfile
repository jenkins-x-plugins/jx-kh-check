FROM scratch

ENTRYPOINT ["/jx-kcheck"]

COPY ./build/linux/jx-kcheck /jx-kcheck