SELECT
  environment,
  flavor,
  SUM(total_lifespan) / 60 / 60 / 24 AS total_days_consumed
FROM (
  # Finished clusters
  SELECT
    "finished_clusters" AS type,
    environment,
    flavor,
    SUM(lifespan) AS total_lifespan
  FROM (
    SELECT
      created.environment AS environment,
      created.workflowName AS workflowName,
      created.clusterID,
      created.flavor AS flavor,
      created.actor AS actor,
      TIMESTAMP_DIFF(deleted.deletionTimestamp, created.creationTimestamp, SECOND) AS lifespan
    FROM
      `infra_cluster_lifespan.infra_cluster_creation` AS created
    JOIN
      `infra_cluster_lifespan.infra_cluster_deletion` AS deleted
    ON
      created.workflowName = deleted.workflowName
    WHERE
      DATE_DIFF(CURRENT_DATE(), DATE(created.creationTimestamp), DAY) <= 30 )
  GROUP BY
    actor,
    flavor,
    environment
  UNION ALL (
    # Running clusters
    SELECT
      "running_clusters" AS type,
      environment,
      flavor,
      SUM(lifespan) AS total_lifespan
    FROM (
      SELECT
        created.environment AS environment,
        created.workflowName AS workflowName,
        created.clusterID,
        created.flavor AS flavor,
        created.actor AS actor,
        TIMESTAMP_DIFF(CURRENT_TIMESTAMP(), created.creationTimestamp, SECOND) AS lifespan
      FROM
        `infra_cluster_lifespan.infra_cluster_creation` AS created
      WHERE
        DATE_DIFF(CURRENT_DATE(), DATE(created.creationTimestamp), DAY) <= 30
        AND created.workflowName != ""
        AND created.workflowName NOT IN (
        SELECT
          workflowName
        FROM
          `infra_cluster_lifespan.infra_cluster_deletion`
        WHERE
          workflowName != "") )
    GROUP BY
      flavor,
      environment))
GROUP BY
  environment,
  flavor
