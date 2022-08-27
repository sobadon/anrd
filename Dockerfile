FROM golang:1.19 as builder

WORKDIR /workspace

COPY go.mod go.sum ./

COPY . .

RUN make build


FROM alpine as downloader

RUN cd /tmp && \
    wget https://www.johnvansickle.com/ffmpeg/old-releases/ffmpeg-4.4.1-amd64-static.tar.xz && \
    tar Jxvf ffmpeg-4.4.1-amd64-static.tar.xz && \
    cp ffmpeg-4.4.1-amd64-static/ffmpeg /usr/local/bin/ && \
    rm ffmpeg-4.4.1-amd64-static.tar.xz


FROM gcr.io/distroless/base-debian11:debug-nonroot as runner

WORKDIR /app

COPY --from=builder /workspace/anrd /app/
COPY --from=downloader /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg

ENTRYPOINT [ "/app/anrd run" ]
