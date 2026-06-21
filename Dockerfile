# --- build stage ---
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/dsa-sheet .

# --- run stage ---
FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/dsa-sheet ./dsa-sheet
ENV PORT=8080
ENV DATA_DIR=/app/data
RUN mkdir -p /app/data && chown -R app:app /app
USER app
EXPOSE 8080
ENTRYPOINT ["./dsa-sheet"]
