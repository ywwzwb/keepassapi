FROM golang:1.9-alpine
RUN apk --no-cache add git
WORKDIR /go/src/keepassapi/
RUN go get github.com/gorilla/mux && go get github.com/tobischo/gokeepasslib
COPY ./* ./
RUN go build

FROM alpine:latest
COPY --from=0 /go/src/keepassapi/keepassapi /bin/keepassapi
ENV   PORT=8000 PATH="/keepass/keepass.kdbx"
EXPOSE $PORT
CMD ["/bin/keepassapi"]