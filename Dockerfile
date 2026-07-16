FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go build -o /out/api ./cmd/api \
 && go build -o /out/payment-worker ./cmd/payment-worker \
 && go build -o /out/notification ./cmd/notification \
 && go build -o /out/cancellation-worker ./cmd/cancellation-worker

FROM alpine:3.21
RUN adduser -D app
WORKDIR /app
COPY --from=build /out/* /usr/local/bin/
COPY --from=build /src/public ./public
USER app
