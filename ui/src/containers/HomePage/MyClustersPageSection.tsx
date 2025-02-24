/* eslint-disable jsx-a11y/label-has-associated-control */
import React, { ReactElement } from 'react';
import { useSearchParams } from 'react-router-dom';
import { AxiosPromise } from 'axios';
import moment from 'moment';
import { Gallery, GalleryItem, PageSection, Title } from '@patternfly/react-core';

import { V1ClusterListResponse, ClusterServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import { useUserAuth } from 'containers/UserAuthProvider';
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
      <GalleryItem>
        <LinkCard
          key={cluster.ID}
          to={`cluster/${cluster.ID}`}
          header={cluster.ID || 'No ID'}
          footer={cluster.Status && <Lifespan cluster={cluster} />}
          className={extraCardClass}
        >
          {cluster.Description && <p>Description: {cluster.Description}</p>}
          <p>Status: {cluster.Status || 'FAILED'}</p>
          <p>Flavor: {cluster.Flavor || 'Unknown'}</p>
        </LinkCard>
      </GalleryItem>
    );
  });
  return <>{cards}</>;
}

export default function MyClustersPageSection(): ReactElement {
  const [searchParams, setSearchParams] = useSearchParams();
  const showAllClusters = searchParams.get('showAllClusters') === 'true';
  function toggleClusterFilter() {
    const newSearchParams = new URLSearchParams(searchParams);
    newSearchParams.set('showAllClusters', (!showAllClusters).toString());
    setSearchParams(newSearchParams);
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
      <Title headingLevel="h2">{headerText}</Title>
      {clusterFilterToggle}
    </div>
  );
  return (
    <>
      <PageSection>{header}</PageSection>
      <PageSection>
        <Gallery
          hasGutter
          minWidths={{
            default: '100%',
            md: '100px',
            xl: '300px',
          }}
          maxWidths={{
            md: '200px',
            xl: '1fr',
          }}
        >
          <ClusterCards showAllClusters={showAllClusters} />
        </Gallery>
      </PageSection>
    </>
  );
}

/* eslint-enable jsx-a11y/label-has-associated-control */
