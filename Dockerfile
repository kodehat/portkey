FROM node:23.8.0-alpine3.21 AS frontend

WORKDIR /usr/src/app

COPY package.json ./

RUN npm install --ignore-scripts --include=dev

COPY internal internal/
COPY assets assets/
COPY static static/

RUN npm run build

FROM golang:1.24.0-alpine3.21 AS backend

ARG VERSION=dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --from=frontend /usr/src/app/static static/
COPY .git .git/
COPY *.go build.sh ./
COPY internal internal/

RUN apk add --no-cache bash curl git && \
  go install github.com/a-h/templ/cmd/templ@v0.3.833 && templ generate && \
  ./build.sh -v "$VERSION"

FROM alpine:3.21.3

LABEL org.opencontainers.image.authors='dev@codehat.de' \
      org.opencontainers.image.url='https://www.portkey.page' \
      org.opencontainers.image.documentation='https://github.com/kodehat/portkey' \
      org.opencontainers.image.source='https://github.com/kodehat/portkey' \
      org.opencontainers.image.vendor='kodehat' \
      org.opencontainers.image.licenses='AGPL-3.0'

WORKDIR /opt

COPY --from=backend /app/portkey ./portkey
# Provide a default config. Can be overwritten by mounting as volume by user.
COPY config.yml .

# "curl" is added only for Docker healthchecks!
RUN apk add --no-cache tzdata curl && \
  adduser -D -H nonroot && \
  chmod +x ./portkey

EXPOSE 3000/tcp

USER nonroot:nonroot

ENTRYPOINT [ "/opt/portkey" ]
CMD [ "--config-path=/opt/" ]
