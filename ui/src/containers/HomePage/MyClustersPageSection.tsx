/* eslint-disable jsx-a11y/label-has-associated-control */
import React, { useState, ReactElement } from 'react';
import { AxiosPromise } from 'axios';
import moment from 'moment';

import { V1ClusterListResponse, ClusterServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import { useUserAuth } from 'containers/UserAuthProvider';
import PageSection from 'components/PageSection';
import LinkCard from 'components/LinkCard';
import Lifespan from 'components/Lifespan';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import assertDefined from 'utils/assertDefined';

const clusterService = new ClusterServiceApi(configuration);

const FETCH_ALL_CLUSTERS = true;
const fetchClusters = (): AxiosPromise<V1ClusterListResponse> =>
  clusterService.list(FETCH_ALL_CLUSTERS);

function NoClustersMessage(): ReactElement {
  return (
    <div className="m-6 flex w-full justify-center">
      <span className="text-4xl">You haven‘t created any clusters recently.</span>
    </div>
  );
}

type ClusterCardsProps = {
  showAllClusters: boolean;
};

function ClusterCards({ showAllClusters = false }: ClusterCardsProps): ReactElement {
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

  // choose whether to show all or just this user's clusters
  const clustersToShow = showAllClusters
    ? data.Clusters
    : data.Clusters.filter((cluster) => cluster.Owner === user?.Email);
  // sorted in descending order by creation date
  const sortedClusters = clustersToShow.sort((c1, c2) =>
    moment(c1.CreatedOn).isBefore(c2.CreatedOn) ? 1 : -1
  );

  if (sortedClusters.length === 0) {
    return <NoClustersMessage />;
  }

  const cards = sortedClusters.map((cluster) => {
    assertDefined(cluster.ID);

    const extraCardClass = showAllClusters && cluster.Owner === user?.Email ? 'bg-base-200' : '';
    return (
      <LinkCard
        key={cluster.ID}
        to={`cluster/${cluster.ID}`}
        header={cluster.ID || 'No ID'}
        footer={cluster.Status && <Lifespan cluster={cluster} />}
        className={`m-2 ${extraCardClass}`}
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
  const [showAllClusters, setShowAllClusters] = useState(false);

  function toggleClusterFilter() {
    setShowAllClusters(!showAllClusters);
  }

  const headerText = showAllClusters ? 'All Clusters' : 'My Clusters';
  const clusterFilterToggle = (
    <span className="flex items-center">
      <label htmlFor="cluster-filter-toggle" className="mr-2 text-lg">
        Show All Clusters
      </label>
      <input
        type="checkbox"
        id="cluster-filter-toggle"
        checked={showAllClusters}
        onChange={toggleClusterFilter}
        className="w-4 h-4 rounded-sm"
      />
    </span>
  );

  const header = (
    <div className="flex justify-between items-center ">
      <span>{headerText}</span>
      {clusterFilterToggle}
    </div>
  );
  return (
    <PageSection header={header}>
      <div className="flex flex-wrap -m-2">
        <ClusterCards showAllClusters={showAllClusters} />
      </div>
    </PageSection>
  );
}

/* eslint-enable jsx-a11y/label-has-associated-control */
