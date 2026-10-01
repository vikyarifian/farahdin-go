# Build: docker build -t farahdin .
# Run:   docker run -p 8080:8080 --env-file .env farahdin   (.env must set POSTGRES_URL)

FROM node:22-alpine AS css
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY web ./web
RUN npm run css

FROM golang:1.25-alpine AS build
WORKDIR /src
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=css /src/web/static/css/app.css ./web/static/css/app.css
RUN templ generate && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/farahdin ./cmd/app \
    && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/import-convex ./cmd/import-convex

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 app
COPY --from=build /out/ /usr/local/bin/
USER app
ENV ADDR=:8080
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/usr/local/bin/farahdin"]
