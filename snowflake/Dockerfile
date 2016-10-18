FROM golang
ADD . /go/src/snowflake
RUN go install snowflake
# RUN rm -rf pkg src
EXPOSE 50003
ENTRYPOINT ["snowflake"]