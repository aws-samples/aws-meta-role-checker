FROM golang
LABEL maintainer="evisb@amazon.com"
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build
CMD ["./meta-role-checker"]