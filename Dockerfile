# Builder
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalacja certyfikatów SSL, które są niezbędne by aplikacja mogła wykonywać zapytania HTTPS do API Open-Meteo.
RUN apk --no-cache add ca-certificates

# Kopiowanie kodu źródłowego
COPY main.go .

# Optymalizacja cache: Inicjalizacja modułu i kompilacja
# CGO_ENABLED=0 wymusza kompilację statyczną, dzięki czemu plik binarny zadziała na pustym obrazie.
RUN go mod init weatherapp && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o weatherapp main.go

# Docelowy obraz
FROM scratch

# Etykieta - autor obrazu
LABEL org.opencontainers.image.authors="Jakub Gwozdowski"

# Kopiowanie certyfikatów SSL z etapu buildera
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Kopiowanie skompilowanej aplikacji z etapu buildera
COPY --from=builder /app/weatherapp /weatherapp

# Deklaracja portu
EXPOSE 8080

# HEALTHCHECK dla warstwy scratch
# Zamiast curl, mechanizm z użyciem flagi '-health'.
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/weatherapp", "-health"]

# Uruchomienie głównego procesu 
CMD ["/weatherapp"]