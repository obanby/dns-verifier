# dns-verifier

 * `config/expected.yaml` contains a set of entries that we expect to match.
 * `config/branch.yaml` contains a proposed set of changes that we expect not to match.
 * `verify` is the entry point command
 
# How to run

If you are running locally please follow the following steps:

```shell
  go mod download
  cd ./cmd/verify
  go build .
  ./verify <config-file> <domain> <nameserver>
```

If you are using docker

```shell
    docker build -t verifyier .
    docker run -it  verifyier <config-file> <domain> <nameserver>
```

Please note that in docker the local config directory will be located in `/app/config`

`docker run -it verifyier /app/config/branch.yaml dns-exercise.dev ns-1775.awsdns-29.co.uk`
