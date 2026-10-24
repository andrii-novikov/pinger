# Container image for the Pi stack (homelab repo, service "pinger"). Runs in interval mode.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -o /out/pinger .

FROM alpine:3.21
# ca-certificates: HTTPS to api.telegram.org
# iputils-ping:    pinger shells out to `ping -c 1`
# tzdata:          local times in logs when TZ is set
RUN apk add --no-cache ca-certificates tzdata iputils-ping
COPY --from=build /out/pinger /usr/local/bin/pinger
USER nobody
# State (/tmp/.pinger) lives in the container's writable layer: it survives restarts and reboots and is only
# reset when the container is recreated, which just rebaselines "last seen up".
ENV PINGER_INTERVAL=60s
ENTRYPOINT ["/usr/local/bin/pinger"]
