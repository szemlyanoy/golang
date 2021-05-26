FROM golang:1.15-alpine AS base
WORKDIR /src/
COPY ./src/* /src/
ENV CGO_ENABLED=0 
RUN go build
RUN ls -latr /src/

FROM scratch
#FROM golang:1.15-alpine
WORKDIR /opt/goweb/
COPY --from=base /src/goweb ./goweb
CMD ["./goweb"]