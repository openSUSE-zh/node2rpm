[![Go Report Card](https://goreportcard.com/badge/github.com/openSUSE-zh/node2rpm)](https://goreportcard.com/report/github.com/openSUSE-zh/node2rpm)

# node2rpm

Next generation of node2rpm, a tool to package NodeJS modules for openSUSE.

It supports nodejs module bundle packaging and separate packaging

Package as bundle:

    node2rpm -pkg har-validator -ver 5.1.3 (without a version, latest will be used)

Package as bundle but split `punycode` to a separate package:

    node2rpm -pkg har-validator -ver 5.1.3 -exclude "punycode:^2.x"

Package as single package:

    node2rpm -pkg har-validator -bundle=false
