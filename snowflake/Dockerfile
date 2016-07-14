FROM golang
ADD . /go/src/github.com/amorwiliams/bodoni/snowflake
RUN go install github.com/amorwiliams/bodoni/snowflake
EXPOSE 50003
ENTRYPOINT ["snowflake"]