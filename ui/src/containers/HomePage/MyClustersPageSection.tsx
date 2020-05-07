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

function lifespanToDuration(lifespan: string): moment.Duration {
  // API returns lifespan in seconds, but it's being very explicit about it with `10800s` format...
  const matches = /\d+/.exec(lifespan);
  if (!matches || matches.length !== 1)
    throw new Error(`Unexpected server response for lifespan ${lifespan}`);

  const seconds = Number.parseInt(matches[0], 10);
  return moment.duration(seconds, 's');
}

function ClusterCards(): ReactElement {
  const { user } = useUserAuth();
  const { loading, error, data } = useApiQuery(fetchClusters);

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !data?.Clusters) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  // only this user clusters sorted in descending order by creation date
  const clusters = data.Clusters.filter((cluster) => cluster.Owner === user?.Email).sort((c1, c2) =>
    moment(c1.CreatedOn).isBefore(c2.CreatedOn) ? 1 : -1
  );

  if (clusters.length === 0) {
    return (
      <div className="m-6 flex w-full justify-center">
        <span className="text-4xl">You don&apos;t have any clusters created recently</span>
      </div>
    );
  }

  const cards = clusters.map((cluster) => {
    const expirationDate =
      cluster.Lifespan &&
      moment(cluster.CreatedOn).add(lifespanToDuration(cluster.Lifespan)).toDate();
    return (
      <LinkCard
        key={cluster.ID}
        to={`cluster/${cluster.ID}`}
        header={cluster.ID || 'No ID'}
        footer={cluster.Status && expirationDate && <Countdown targetDate={expirationDate} />}
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
