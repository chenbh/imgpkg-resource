FROM golang:1.17 as builder
WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
ENV CGO_ENABLED 0
RUN go build -o /assets/in ./cmd/in
RUN go build -o /assets/out ./cmd/out
RUN go build -o /assets/check ./cmd/check
RUN chmod +x /assets/*

FROM paketobuildpacks/run:tiny-cnb AS resource
COPY --from=builder assets/ /opt/resource/
USER root
