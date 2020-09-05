FROM golang:latest as build

ENV GO111MODULE=on

WORKDIR /app/server

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build

FROM alpine:latest as server

WORKDIR /app/server

COPY --from=build /app/server/api-gateway ./

RUN chmod +x ./api-gateway

EXPOSE 10000

CMD [ "./api-gateway" ]