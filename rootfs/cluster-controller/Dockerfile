FROM golang:1.8

WORKDIR /go/src/app
COPY . .

RUN curl https://glide.sh/get | sh

RUN ls -ll
RUN glide install

ENTRYPOINT go run ./*.go -kubeconf=""