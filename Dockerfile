FROM golang:1.21.0-alpine3.18 AS build

ENV CGO_ENABLED=0

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main

FROM gcr.io/distroless/static-debian11

COPY --from=build /app/main /

CMD ["/main"]