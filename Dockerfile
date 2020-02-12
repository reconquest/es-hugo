FROM golang:1.13

ARG version

WORKDIR /build
COPY / .

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build \
    -ldflags="-X=main.version=${version:-unknown}" \
    -o /app \
    ./

FROM node:13.8-alpine

WORKDIR /
COPY --from=0 /app .

RUN apk add --update --no-cache npm git openssh ca-certificates
RUN npm install -g kovetskiy/hugo-elasticsearch

CMD ["/app"]
