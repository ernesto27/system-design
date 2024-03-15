FROM golang:1.22 as build
WORKDIR /lambda
COPY go.mod go.sum ./
COPY main.go .
RUN go build -tags lambda.norpc -o main main.go

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=build /lambda/main ./main
ENTRYPOINT [ "./main" ]