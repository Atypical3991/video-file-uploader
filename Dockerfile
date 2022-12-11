#syntax=docker/dockerfile:1
#used Alpine flavour here for having lightweight container
FROM golang:1.19-alpine

ENV workdir=/app

#HOST should be the domain name of the web server
ENV HOST=localhost
#PORT on which the server is listening
ENV PORT=8080
#GOPATH for importing internal modules
ENV GOPATH=${workdir}/src
WORKDIR ${workdir}
#all the executables will go inside /bin
RUN mkdir -p /bin

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . ./

RUN go build -o /bin/city_os ./cmd/app/

EXPOSE ${PORT}

CMD ["/bin/city_os"]