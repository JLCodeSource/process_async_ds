FROM golang:alpine AS build

WORKDIR /usr/src/app

copy . .

RUN go mod download

RUN go build -o /usr/src/app/process_processed


FROM gcr.io/distroless/base-debian10

WORKDIR /usr/src/app

COPY --from=build /usr/src/app/process_processed /usr/src/app/process_processed

USER nonroot:nonroot

ENTRYPOINT ["/usr/src/app/process_processed"]