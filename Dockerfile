FROM scratch

ENTRYPOINT ["/jx-health"]

COPY ./build/linux/jx-health /jx-health