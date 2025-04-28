FROM golang:1.24

WORKDIR /home/isucon/webapp/go

COPY . .
RUN go build -o webapp .

CMD ["/home/isucon/webapp/go/webapp"]

