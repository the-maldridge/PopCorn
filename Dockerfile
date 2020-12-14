FROM golang:1.15-alpine as build
WORKDIR /popcorn/
COPY . .
RUN go mod vendor && \
        CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /popcornd ./cmd/popcornd && \
        CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /pgrpcgw ./cmd/pgrpcgw && \
        apk add upx binutils && \
        strip /pgrpcgw && \
        strip /popcornd && \
        upx /pgrpcgw && \
        upx /popcornd && \
        ls -alh /pgrpcgw && \
        ls -alh /popcornd

FROM scratch
LABEL org.opencontainers.image.source https://github.com/the-maldridge/popcorn
ENTRYPOINT ["/popcornd"]
COPY --from=build /pgrpcgw /pgrpcgw
COPY --from=build /popcornd /popcornd
