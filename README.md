# flow-aggregation

## Prequisites
These instructions assume you have already done the following:
- installed [Docker](https://docs.docker.com/get-docker/).
- cloned this repository

## Compilation

Use docker to build the container:
```sh
docker build -t kvg-server .
```

## Execution

Run the container with the following command:
```sh
docker run --net=host kvg-server
```

Note: the `--net=host` flag is intended to provide better performance than the
docker networking bridge. 

## Testing

You should be able to test the service by sending a request on localhost:8080:
```
curl -X POST "http://localhost:8080/flows" \
     -H 'Content-Type: application/json' \
     -d '[{"src_app":"foo","dest_app":"bar","vpc_id":"vpc-0","bytes_tx":100,"bytes_rx":300,"hour":1},{"src_app":"foo","dest_app":"bar","vpc_id":"vpc-0","bytes_tx":200,"bytes_rx":600,"hour":1},{"src_app":"baz","dest_app":"qux","vpc_id":"vpc-0","bytes_tx":100,"bytes_rx":500,"hour":1},{"src_app":"baz","dest_app":"qux","vpc_id":"vpc-0","bytes_tx":100,"bytes_rx":500,"hour":2},{"src_app":"baz","dest_app":"qux","vpc_id":"vpc-1","bytes_tx":100,"bytes_rx":500,"hour":2}]'

```