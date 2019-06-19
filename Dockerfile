FROM golang:1.7.5
RUN go get gopkg.in/telegram-bot-api.v4
WORKDIR /server/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM scratch  
WORKDIR /root/
COPY --from=0 /server/app .
CMD ["./app"]  