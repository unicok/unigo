FROM golang
Add . /go/src/github.com/unicok/unigo/archiver
RUN go install github.com/unicok/unigo/archiver
RUN go install github.com/unicok/unigo/archiver/archiver
RUN go install github.com/unicok/unigo/archiver/replay
RUN rm -rf pkg src
RUN mkdir /data
VOLUME /data

