FROM golang:latest as builder
WORKDIR /app
COPY . .
RUN make init && make build

FROM scratch
COPY --from=builder /app/server .
COPY --from=builder /app/env.json .
CMD ["./server"]