# Use a versão oficial do Golang para compilar o código-fonte
FROM golang:1.22 as builder

WORKDIR /app

# Copie o código fonte
COPY . .

# Compile o aplicativo Go
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Use a imagem base leve para o estágio de execução
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copie o binário compilado do estágio de construção
COPY --from=builder /app/app .

# Exponha a porta que o serviço Go usa
EXPOSE 5002

# Execute o aplicativo Go
CMD ["./app"]
