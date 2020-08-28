import React, { ReactElement } from 'react';
import moment from 'moment';

import { V1Cluster } from 'generated/client';
import Countdown from 'components/Countdown';

/**
 * Converts backend returned lifespan to the moment duration
 * @param lifespan lifespan string the way it comes from the backend
 * @returns duration object
 * @throws error if it cannot parse lifespan string
 */
function lifespanToDuration(lifespan: string): moment.Duration {
  // API returns lifespan in seconds, but it's being very explicit about it with `10800s` format...
  const matches = /\d+/.exec(lifespan);
  if (!matches || matches.length !== 1)
    throw new Error(`Unexpected server response for lifespan ${lifespan}`);

  const seconds = Number.parseInt(matches[0], 10);
  return moment.duration(seconds, 's');
}

type Props = {
  cluster: V1Cluster;
  canModify: boolean;
};

export default function ClusterLifespanCountdown({ cluster, canModify }: Props): ReactElement {
  let expirationDate: Date | null = null;
  try {
    const duration = lifespanToDuration(cluster.Lifespan || '0s');
    expirationDate =
      duration.asMilliseconds() === 0 ? null : moment(cluster.CreatedOn).add(duration).toDate();
  } catch (e) {
    // should never happen, ignore, we'll just show N/A for expiration
    // TODO: eventually log the error to the backend
  }

  return expirationDate ? (
    <Countdown targetDate={expirationDate} canModify={canModify} />
  ) : (
    <>Expiration: N/A</>
  );
}
