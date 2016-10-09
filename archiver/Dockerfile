FROM golang
Add . /go/src/archiver
RUN archiver
RUN archiver/archiver
RUN archiver/replay
RUN rm -rf pkg src
RUN mkdir /data
VOLUME /data

