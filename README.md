# bundles-cues

Cues for Databricks asset bundles. To idea is to parse the `databricks.yml` in the bundle root and union all the files that are referenced under the `include` key, and validate it against a CUE schema.

## Example usage

```bash
cue vet merged.yaml validate.cue
```
