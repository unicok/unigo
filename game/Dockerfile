FROM golang
ADD . /go/src/game
RUN go install game
EXPOSE 51000
ENTRYPOINT ["game"]