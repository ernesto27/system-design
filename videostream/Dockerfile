# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.21 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./


RUN CGO_ENABLED=0 GOOS=linux go build -o /video-stream



FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /video-stream /video-stream
COPY .env ./

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/video-stream"]