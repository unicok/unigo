FROM golang
ADD . /go/src/auth
RUN go install auth
EXPOSE 50006
ENTRYPOINT ["bash"]