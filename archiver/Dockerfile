FROM golang
Add . /go/src/github.com/amorwilliams/bodoni/archiver
RUN go install github.com/amorwilliams/bodoni/archiver
RUN go install github.com/amorwilliams/bodoni/archiver/archiver
RUN go install github.com/amorwilliams/bodoni/archiver/replay
RUN rm -rf pkg src
RUN mkdir /data
VOLUME /data

