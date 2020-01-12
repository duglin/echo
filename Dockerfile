FROM golang:alpine
COPY echo.go /
RUN go build -o /echo /echo.go
FROM alpine
RUN apk add curl
COPY --from=0 /echo /echo
CMD /echo
