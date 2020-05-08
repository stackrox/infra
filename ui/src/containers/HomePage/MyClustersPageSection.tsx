import React, { ReactElement } from 'react';
import { AxiosPromise } from 'axios';
import moment from 'moment';

import { V1ClusterListResponse, ClusterServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import { useUserAuth } from 'containers/UserAuthProvider';
import PageSection from 'components/PageSection';
import LinkCard from 'components/LinkCard';
import Countdown from 'components/Countdown';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';

const clusterService = new ClusterServiceApi(configuration);

const fetchClusters = (): AxiosPromise<V1ClusterListResponse> => clusterService.list();

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

function NoClustersMessage(): ReactElement {
  return (
    <div className="m-6 flex w-full justify-center">
      <span className="text-4xl">You haven‘t created any clusters recently.</span>
    </div>
  );
}

function ClusterCards(): ReactElement {
  const { user } = useUserAuth();
  const { loading, error, data } = useApiQuery(fetchClusters);

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !data) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  if (!data.Clusters) {
    return <NoClustersMessage />;
  }

  // only this user clusters sorted in descending order by creation date
  const clusters = data.Clusters.filter((cluster) => cluster.Owner === user?.Email).sort((c1, c2) =>
    moment(c1.CreatedOn).isBefore(c2.CreatedOn) ? 1 : -1
  );

  if (clusters.length === 0) {
    return <NoClustersMessage />;
  }

  const cards = clusters.map((cluster) => {
    let expirationDate: Date | null = null;
    try {
      expirationDate = !cluster.Lifespan
        ? null
        : moment(cluster.CreatedOn).add(lifespanToDuration(cluster.Lifespan)).toDate();
    } catch (e) {
      // should never happen, ignore, we'll just show N/A for expiration
      // TODO: eventually log the error to the backend
    }

    return (
      <LinkCard
        key={cluster.ID}
        to={`cluster/${cluster.ID}`}
        header={cluster.ID || 'No ID'}
        footer={
          (cluster.Status && expirationDate && <Countdown targetDate={expirationDate} />) ||
          'Expiration: N/A'
        }
      >
        {cluster.Description && (
          <span className="mb-2 text-lg">Description: {cluster.Description}</span>
        )}
        <span className="text-lg">Status: {cluster.Status || 'FAILED'}</span>
      </LinkCard>
    );
  });
  return <>{cards}</>;
}

export default function LaunchPageSection(): ReactElement {
  return (
    <PageSection header="My Clusters">
      <div className="flex flex-wrap">
        <ClusterCards />
      </div>
    </PageSection>
  );
}
