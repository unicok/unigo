FROM golang
ADD . /go/src/github.com/unicok/unigo/snowflake
RUN go install github.com/unicok/unigo/snowflake
RUN rm -rf pkg src
EXPOSE 50003
ENTRYPOINT ["snowflake"]