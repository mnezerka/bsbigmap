FROM golang:alpine AS builder

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"
CMD ["./gobigmap"]

FROM alpine:latest AS alpine
COPY --from=builder /app/gobigmap /app/gobigmap
COPY --from=builder /app/providers.csv /app/providers.csv
COPY --from=builder /app/html /app/html
COPY --from=builder /app/js /app/js
WORKDIR /app/
EXPOSE 9090
CMD ["./gobigmap"]
