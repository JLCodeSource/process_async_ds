FROM golang:alpine AS build

WORKDIR /usr/src/app

RUN apk update && apk upgrade && apk add bash

COPY gbr /usr/bin/gbr

RUN chmod +x /usr/bin/gbr

RUN go install github.com/hhatto/gocloc/cmd/gocloc@latest

COPY . .

RUN go mod download

RUN go test -v ./...

RUN go test -race ./...

RUN go test -cover ./...

RUN gocloc .

RUN go build -o /usr/src/app/process_processed

FROM gcr.io/distroless/base-debian10

WORKDIR /usr/src/app

COPY --from=build /usr/src/app/process_processed /usr/src/app/process_processed

USER nonroot:nonroot

ENTRYPOINT ["/usr/src/app/process_processed"]
