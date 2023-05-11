FROM golang:1.20 AS build
EXPOSE 8080
EXPOSE 8084
EXPOSE 8089
WORKDIR /go/src/orchy
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/orchy .
FROM scratch
COPY --from=build /go/bin/orchy /bin/orchy
ENTRYPOINT ["/bin/orchy"]