# octodns-verifier

 * `config/expected.yaml` contains a set of entries that we expect to match.
 * `config/branch.yaml` contains a proposed set of changes that we expect not to match.
 * `verify` is a placeholder script for your solution.
 
 ## Instructions
 
Welcome!

1.  Please check the open issues in this repository.

    You can navigate there using the `Issues` tab-bar above.  It should be directly beneath the repository name.
2.  There should be just a single open issue.  Assign it to yourself.

    The issue has instructions on what we would like to see.
3.  Hack away, using this repository to store your work.

    Committing once with the solution is *ok*, but we'd love to see your incremental process.
    
    We suggest taking a look at [GitHub flow](https://guides.github.com/introduction/flow/) to structure your commits.
4.  [Submit a pull request](https://help.github.com/articles/creating-a-pull-request/) once you are happy with your work. **Please leave your pull request open: we'll discuss it as a part of the interview process.**

## How to run

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