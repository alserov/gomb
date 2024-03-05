FROM golang:alpine as builder

WORKDIR /build

COPY go.mod go.sum ./
COPY . ./
COPY cmd/main.go ./

RUN go mod download
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o ./bin

FROM scratch

COPY --from=builder /build/bin /bin
COPY --from=builder /build/config /config

CMD ["/bin"]
