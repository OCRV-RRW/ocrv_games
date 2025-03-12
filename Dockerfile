#FROM golang:alpine AS builder
#
#WORKDIR /build
#
#ADD go.mod .
#
#COPY . .
#
#RUN go build -o app main.
#
#FROM alpine
#
#CMD ["/build/app"]

# STEP-1
# build app from source

FROM golang:1.24.1-alpine3.21 AS builder

WORKDIR /mysource

COPY ./vendor ./vendor
COPY ./go.mod ./go.sum ./main.go ./.env ./
COPY ./internal ./internal
COPY ./docs ./docs

RUN go build -o app ./main.go

# STEP-2
# make container

FROM alpine:3.21

WORKDIR /myapp

COPY --from=builder /mysource ./

CMD [ "/myapp/app" ]