FROM golang:1.17-buster AS builders

ENV FFMPEG_VERSION=4.4 BUILD_PREFIX=/opt/ffmpeg
ENV FFMPEG_ROOT=/opt/ffmpeg
ENV LD_LIBRARY_PATH=/opt/ffmpeg/lib:/usr/local/lib64:/usr/local/lib
ENV PKG_CONFIG_PATH=/opt/ffmpeg/lib/pkgconfig:/usr/local/lib64/pkgconfig:/usr/local/lib/pkgconfig
ENV CGO_CPPFLAGS -I/opt/ffmpeg/include
#ENV CGO_CXXFLAGS "--std=c++1z"
ENV CGO_LDFLAGS="-L/opt/ffmpeg/lib -lavcodec -lavformat -lavutil -lswscale -lswresample -lavdevice -lavfilter"

RUN apk add --update build-base curl nasm tar bzip2 pkgconfig gcc libc-dev musl-dev \
  zlib-dev openssl-dev yasm-dev lame-dev libogg-dev \
  x264-dev libvpx-dev libvorbis-dev x265-dev  \
  freetype-dev libass-dev libwebp-dev libtheora-dev \
  opus-dev && \
  DIR=$(mktemp -d) && cd ${DIR} && \
  curl -s http://ffmpeg.org/releases/ffmpeg-${FFMPEG_VERSION}.tar.gz | tar zxvf - -C . && \
  cd ffmpeg-${FFMPEG_VERSION} && \
  ./configure \
  --enable-shared --enable-version3 --prefix=${BUILD_PREFIX} && \
  make && \
  make install && \
  rm -rf ${DIR} && \
  apk del build-base curl tar bzip2 x264 openssl nasm && rm -rf /var/cache/apk/* && \
  ln -s /opt/ffmpeg/bin/ffmpeg /usr/bin/ffmpeg

ADD . /tmp/app
WORKDIR /tmp/app

RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o minion cmd/main.go

FROM alpine

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /tmp/app/minion .
COPY --from=builder /tmp/app/config.yml .
COPY --from=builder /tmp/app/cert.pem .
COPY --from=builder /tmp/app/key.pem .
COPY --from=builder /opt/ffmpeg/bin /usr/local/bin
COPY --from=builder /opt/ffmpeg/share /usr/local/share
COPY --from=builder /opt/ffmpeg/include /usr/local/include
COPY --from=builder /opt/ffmpeg/lib /usr/local/lib
      
EXPOSE 8080

RUN chmod +x ./minion
ENTRYPOINT ["./minion"]
#CMD ["./minion"]
