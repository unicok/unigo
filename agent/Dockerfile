FROM golang
ADD . /go/src/github.com/unicok/unigo/agent
RUN go install github.com/unicok/unigo/agent
RUN rm -rf pkg src
EXPOSE 8888 8888/udp
ENTRYPOINT ["agent"]