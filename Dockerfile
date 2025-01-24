FROM node:23.6.1-alpine3.21 AS frontend

WORKDIR /usr/src/app

COPY package.json ./

RUN npm install --ignore-scripts --include=dev

COPY internal internal/
COPY assets assets/
COPY static static/
COPY tailwind.config.js ./

RUN npm run build

FROM golang:1.23.5-alpine3.21 AS backend

ARG VERSION=dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --from=frontend /usr/src/app/static static/
COPY .git .git/
COPY *.go build.sh ./
COPY internal internal/

# "curl" is added only for Docker healthchecks!
RUN apk add --no-cache bash curl git && \
  go install github.com/a-h/templ/cmd/templ@v0.2.793 && templ generate && \
  ./build.sh -v "$VERSION"

FROM alpine:3.21.2

LABEL org.opencontainers.image.authors='dev@codehat.de' \
      org.opencontainers.image.url='https://www.portkey.page' \
      org.opencontainers.image.documentation='https://github.com/kodehat/portkey' \
      org.opencontainers.image.source='https://github.com/kodehat/portkey' \
      org.opencontainers.image.vendor='kodehat' \
      org.opencontainers.image.licenses='AGPL-3.0'

WORKDIR /opt

COPY --from=backend /app/portkey ./app
# Provide a default config. Can be overwritten by mounting as volume by user.
COPY config.yml .

RUN apk add --no-cache tzdata && \
  adduser -D -H nonroot && \
  chmod +x ./app

EXPOSE 3000/tcp

USER nonroot:nonroot

ENTRYPOINT [ "./app" ]
