FROM busybox:glibc

COPY server /bin/api

ENTRYPOINT ["/bin/api"]

EXPOSE 8080