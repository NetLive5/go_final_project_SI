FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p /out/data \
	&& CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/scheduler .

FROM scratch

WORKDIR /app
COPY --from=build /out/scheduler /app/scheduler
COPY --from=build /src/web /app/web
COPY --from=build --chown=1000:1000 /out/data /data

ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db

EXPOSE 7540
USER 1000:1000
ENTRYPOINT ["/app/scheduler"]
