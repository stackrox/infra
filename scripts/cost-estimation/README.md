# Cost Estimation

These scripts can be used to estimate the cost caused by infra clusters within the last 30 days.
They are available as a GitHub workflow with a manual dispatch.

If you have access privileges to BigQuery in the `stackrox-infra` project, you may also run the scripts directly like:

```bash
./scripts/cost-estimation/calculate-costs.sh
```

If you want to update the query (e.g. to exclude environments, change the observation window, ...), you may do this in the `total-time-consumed.sql` file.
