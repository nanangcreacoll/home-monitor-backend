FROM golang:1.24-alpine AS build

WORKDIR /build

COPY . .
RUN go build -o app .

FROM alpine:3.21

WORKDIR /app

COPY --from=build /build/app /main
COPY --from=build /build/database/migrations /app/database/migrations

EXPOSE 8080

CMD ["/main"]