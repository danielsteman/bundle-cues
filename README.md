# bundles-cues

Cues for Databricks asset bundles. To idea is to parse the `databricks.yml` in the bundle root and union all the files that are referenced under the `include` key, and validate it against a CUE schema.

## Installation

### Brew

```bash
brew tap danielsteman/tap
brew install bundlecues
```

## Example usage

The following schema (`schema.cue`) checks if the the `id` of the `webhook_notifications` for `on_failure` is set to `"sup"`. This can be useful if you want to ensure that all jobs in a config are sending notifications to some channel whenever they fail.

```yaml
#Job: {
  // Other fields are allowed (open struct by default)
  ...
  webhook_notifications: {
    on_failure: [{
      id: "sup"
    }]
  }
}

targets: {
  prod: {
    resources: {
      jobs: [string]: #Job
    }
  }
}
```

The actual validation, according to the schema above, can be done like:

```bash
bundlecues validate schema.cue
```

Given that your current working directory is the root of the asset bundle (where `databricks.yml` is located).

## Build

```bash
GOOS=darwin GOARCH=arm64 go build -o bundlecues
tar -czvf bundlecues_darwin_arm64.tar.gz bundlecues

GOOS=darwin GOARCH=amd64 go build -o bundlecues
tar -czvf bundlecues_darwin_amd64.tar.gz bundlecues

GOOS=linux GOARCH=amd64 go build -o bundlecues
tar -czf bundlecues_linux_amd64.tar.gz bundlecues
```
