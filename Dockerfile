FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/sky-crawler .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/sky-crawler /sky-crawler
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/sky-crawler"]
