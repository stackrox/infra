import React, { ReactElement } from 'react';
import { AxiosPromise } from 'axios';
import moment from 'moment';

import { V1ClusterListResponse, ClusterServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import { useUserAuth } from 'containers/UserAuthProvider';
import PageSection from 'components/PageSection';
import LinkCard from 'components/LinkCard';
import ClusterLifespanCountdown from 'components/ClusterLifespanCountdown';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';

const clusterService = new ClusterServiceApi(configuration);

const fetchClusters = (): AxiosPromise<V1ClusterListResponse> => clusterService.list();

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
    return (
      <LinkCard
        key={cluster.ID}
        to={`cluster/${cluster.ID}`}
        header={cluster.ID || 'No ID'}
        footer={cluster.Status && <ClusterLifespanCountdown cluster={cluster} canModify={false} />}
        className="m-2"
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
      <div className="flex flex-wrap -m-2">
        <ClusterCards />
      </div>
    </PageSection>
  );
}
