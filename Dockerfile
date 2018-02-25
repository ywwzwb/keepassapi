FROM golang:1.9-alpine
RUN apk --no-cache add git
WORKDIR /go/src/keepassapi/
RUN go get github.com/gorilla/mux && go get github.com/tobischo/gokeepasslib
COPY ./ ./
RUN  ls -al && go build

FROM alpine:latest
COPY --from=0 /go/src/keepassapi/keepassapi /bin/keepassapi
ENV   KEEPASS_PORT=8000 KEEPASS_DBPATH="/keepass/keepass.kdbx"
EXPOSE $KEEPASS_PORT
CMD ["/bin/keepassapi"]