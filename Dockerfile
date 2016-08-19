FROM alpine:3.4

RUN mkdir /app
COPY app files/ static/ /app/

EXPOSE 8080
CMD ["/app/app"]