FROM golang:1.22 as build
WORKDIR /helloworld
# Copy dependencies list
COPY go.mod go.sum ./
# Build with optional lambda.norpc tag
COPY main.go .
RUN go build -tags lambda.norpc -o main main.go
# Copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2023
RUN ls /etc/os-release
RUN dnf -y install wget && dnf -y install tar && dnf -y install xz && \ 
    cd /usr/local/bin && \
    mkdir ffmpeg && \ 
    cd ffmpeg && \
    wget https://www.johnvansickle.com/ffmpeg/old-releases/ffmpeg-4.2.1-amd64-static.tar.xz && \
    tar xvf ffmpeg-4.2.1-amd64-static.tar.xz && \
    ln -s /usr/local/bin/ffmpeg/ffmpeg-4.2.1-amd64-static/ffmpeg /usr/bin/ffmpeg
RUN ls
COPY .env .env
COPY --from=build /helloworld/main ./main
ENTRYPOINT [ "./main" ]