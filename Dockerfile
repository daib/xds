FROM golang AS builder

RUN mkdir -p /usr/src/app
WORKDIR /usr/src/app

#RUN apk add --no-cache git
RUN apt update

RUN apt install git

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 go build -v -o /bin/xds

FROM envoy:latest

COPY --from=builder /bin/xds /bin/xds

ENTRYPOINT ["/bin/xds"]
