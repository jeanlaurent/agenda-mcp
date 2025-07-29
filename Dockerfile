FROM golang:1.24-alpine AS gobuilder
WORKDIR /app
RUN --mount=type=bind,target=. \
    --mount=type=cache,target=/root/.cache \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o /agenda-mcp .

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
COPY --from=gobuilder /agenda-mcp /
EXPOSE 8080
VOLUME ["/data"]
WORKDIR /data
ENTRYPOINT ["/agenda-mcp"] 
CMD ["mcp"] 