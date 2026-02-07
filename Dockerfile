FROM golang:1.25 AS backend

ARG VERSION=1.0.0
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN CGO_ENABLED=0 go build -ldflags \
  "-s -w -X github.com/akinimaginable/image-server/internal/constant.Version=${VERSION}" -o image-server

FROM gcr.io/distroless/static-debian13:nonroot

EXPOSE 8080/tcp
WORKDIR /app

COPY --from=backend /build/image-server /app/image-server

ENTRYPOINT ["/app/image-server"]
