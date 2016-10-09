FROM golang
ADD . /go/src/chat
RUN go install chat
RUN mkdir /data
VOLUME /data
EXPOSE 50008
ENTRYPOINT ["chat"]