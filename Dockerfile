FROM golang:latest as build
WORKDIR /app
COPY . .
RUN go build main.go

FROM ubuntu:latest
RUN apt update
RUN apt install ffmpeg -y
ENV telegram_token your_token
COPY test.mp4 .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/main .
RUN apt-get -y install curl
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
RUN chmod a+rx /usr/local/bin/yt-dlp
RUN apt install python-is-python3 -y
CMD ["./main"]