FROM alpine:3.23.4
RUN adduser -D sample
COPY ./sampleapp /sampleapp
RUN chmod +x /sampleapp && mkdir -p /data && chown sample:sample /sampleapp /data
USER sample
WORKDIR /
ENTRYPOINT ["/sampleapp"]