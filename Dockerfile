FROM golang:1 as build

WORKDIR /go/src/flow-aggregation
COPY . .
RUN go get ./...
RUN CGO_ENABLED=0 go build -o /server

# Final Stage
FROM gcr.io/distroless/static:nonroot
COPY --from=build --chown=nonroot:nonroot /server /server
# run as "non-root" user by id
USER 65532
ENTRYPOINT ["/server"]