FROM node:20.11.1-alpine3.18 as frontend

WORKDIR /usr/src/app

COPY package.json ./

RUN npm install

COPY pkg/components pkg/components/
COPY assets assets/
COPY static static/

RUN npm run build

FROM golang:1.21.7-alpine3.19 as backend

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --from=frontend /usr/src/app/static static/
COPY .git .git/
COPY *.go build.sh ./
COPY pkg/components pkg/components/

RUN apk add --no-cache git bash
RUN go install github.com/a-h/templ/cmd/templ@latest && templ generate
RUN bash build.sh

FROM alpine:3.19.1

WORKDIR /opt

RUN adduser -D -H nonroot

COPY --chown=nonroot:nonroot --from=backend /app/thisismy.cloud /opt/app

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT [ "./app" ]
