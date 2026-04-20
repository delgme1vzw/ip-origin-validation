# The build stage
FROM golang:1.20.14 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/*.go

# The run stage
FROM scratch
WORKDIR /app
# Copy CA certificates - don't need it for this project since no outbound http requests
# but good practice for future projects
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/api .
EXPOSE 8080
CMD ["./api"]
