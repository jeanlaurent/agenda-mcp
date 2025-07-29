FROM golang:1.24-alpine AS gobuilder
RUN  --mount=type=cache,target=/var/cache/apk apk add --no-cache git ca-certificates tzdata
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o agenda-mcp main.go

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/agenda-mcp /agenda-mcp
EXPOSE 8080
VOLUME ["/data"]
WORKDIR /data
CMD ["/agenda-mcp", "mcp"] 