FROM golang
ADD . /go/src/agent
RUN go install agent
# RUN rm -rf pkg src
EXPOSE 8888 8888/udp
ENTRYPOINT ["agent"]
