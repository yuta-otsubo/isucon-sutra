ARG SV_IMAGE="692859926955.dkr.ecr.ap-northeast-1.amazonaws.com/stg/benchmarker:supervisor"

FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN go build -o /app/benchmarker -ldflags "-s -w"

FROM ${SV_IMAGE} AS supervisor

FROM debian:bookworm-slim
RUN apt-get update && apt install -y ca-certificates openssl curl unzip jq
RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
  unzip awscliv2.zip && \
  ./aws/install && \
  rm -rf awscliv2.zip ./aws/
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
WORKDIR /app
COPY --from=builder /app/benchmarker /app/benchmarker
COPY --from=supervisor /usr/local/bin/isuxportal-supervisor /app/supervisor
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["/app/supervisor", "/app/benchmarker"]
