# Cost Estimation

These scripts can be used to estimate the cost caused by infra clusters within the last 30 days.
They are available as a GitHub workflow with a manual dispatch.

If you have access privileges to BigQuery in the `acs-team-automation` project, you may also run the scripts directly like:

```bash
./scripts/cost-estimation/calculate-costs.sh
```

If you want to update the query (e.g. to exclude environments, change the observation window, ...), you may do this in the `total-time-consumed.sql` file.

The `DAILY_COST_MAP` are estimations based on the flavor default values, for more details see [this spreadsheet](https://docs.google.com/spreadsheets/d/1NsaEOOfJ2pMqgR-1-j-as1lb6qJqYOhCD4PW569AWkY/edit#gid=0).
