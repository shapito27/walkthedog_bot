FROM golang:1.18.3

WORKDIR /go/src/app
COPY . .

#RUN go mod init walkthedog
RUN go mod download

# Build
RUN go build -o /walkthedog-bot

CMD ["/walkthedog-bot"]