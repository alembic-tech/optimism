FROM --platform=$BUILDPLATFORM golang:1.19.9-alpine3.16 as builder

ARG VERSION=v0.0.0

RUN apk add --no-cache make gcc musl-dev linux-headers git jq bash

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app/

RUN go mod download

WORKDIR /app/da/dac
# build op-node with the shared go.mod & go.sum files
COPY ./op-service /app/op-service
COPY ./da /app/da
COPY ./.git /app/.git

ARG TARGETOS TARGETARCH

RUN make dac-member VERSION="$VERSION" GOOS=$TARGETOS GOARCH=$TARGETARCH

FROM alpine:3.16

COPY --from=builder /app/da/dac/bin/dac-member /usr/local/bin

CMD ["dac-member"]
