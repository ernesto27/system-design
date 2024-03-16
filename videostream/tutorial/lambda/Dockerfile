FROM golang:1.22 as build
WORKDIR /lambda
COPY go.mod go.sum ./
COPY main.go .
RUN go build -tags lambda.norpc -o main main.go

FROM public.ecr.aws/lambda/provided:al2023
RUN dnf -y install wget && dnf -y install tar && dnf -y install xz && \ 
    cd /usr/local/bin && \
    mkdir ffmpeg && \ 
    cd ffmpeg && \
    wget https://www.johnvansickle.com/ffmpeg/old-releases/ffmpeg-4.2.1-amd64-static.tar.xz && \
    tar xvf ffmpeg-4.2.1-amd64-static.tar.xz && \
    ln -s /usr/local/bin/ffmpeg/ffmpeg-4.2.1-amd64-static/ffmpeg /usr/bin/ffmpeg
COPY .env .env
COPY --from=build /lambda/main ./main
ENTRYPOINT [ "./main" ]