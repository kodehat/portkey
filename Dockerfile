FROM node:20.11.1-alpine3.18 AS frontend

WORKDIR /usr/src/app

COPY package.json ./

RUN npm install --ignore-scripts --include=dev

COPY internal internal/
COPY assets assets/
COPY static static/
COPY tailwind.config.js ./

RUN npm run build

FROM golang:1.21.7-alpine3.19 AS backend

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --from=frontend /usr/src/app/static static/
COPY .git .git/
COPY *.go build.sh ./
COPY internal internal/

RUN apk add --no-cache git bash
RUN go install github.com/a-h/templ/cmd/templ@latest && templ generate
RUN bash build.sh

FROM alpine:3.19.1

LABEL org.opencontainers.image.authors='dev@codehat.de' \
      org.opencontainers.image.url='https://www.portkey.page' \
      org.opencontainers.image.documentation='https://github.com/kodehat/portkey' \
      org.opencontainers.image.source='https://github.com/kodehat/portkey' \
      org.opencontainers.image.vendor='kodehat' \
      org.opencontainers.image.licenses='AGPL-3.0'

WORKDIR /opt

COPY --from=backend /app/portkey /opt/app

RUN adduser -D -H nonroot && \
  chmod +x /opt/app

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT [ "./app" ]
