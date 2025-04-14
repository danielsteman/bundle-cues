# bundles-cues

Cues for Databricks asset bundles. To idea is to parse the `databricks.yml` in the bundle root and union all the files that are referenced under the `include` key, and validate it against a CUE schema.

## Installation

### Brew

```bash
brew tap danielsteman/tap
brew install bundlecues
```

## Example usage

```bash
cue vet merged.yaml validate.cue
```

## Build

```bash
GOOS=darwin GOARCH=arm64 go build -o bundlecues
tar -czvf bundlecues_darwin_arm64.tar.gz bundlecues

GOOS=darwin GOARCH=amd64 go build -o bundlecues
tar -czvf bundlecues_darwin_amd64.tar.gz bundlecues
```
