FROM golang:1.24-alpine

# Instalar ffmpeg
RUN apk add --no-cache ffmpeg

# Criar diretório de trabalho
WORKDIR /app

# Copiar arquivos
COPY . .

# Instalar dependências
RUN go mod tidy

# Criar diretórios necessários
RUN mkdir -p uploads outputs temp

# Executar aplicação
CMD ["go", "run", "main.go"]