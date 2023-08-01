# Release process

1. Create a new tag that follows semantic versioning:

```bash
$ tag=v0.3.0
$ git tag -s "${tag}" -m "${tag}"
$ git push origin "${tag}"
```

2. Make sure you are using goreleaser >= v1.19.2
3. Run `$ goreleaser release --rm-dist`

