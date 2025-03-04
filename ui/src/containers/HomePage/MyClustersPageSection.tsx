/* eslint-disable jsx-a11y/label-has-associated-control */
import React, { ReactElement } from 'react';
import { useSearchParams } from 'react-router-dom';
import moment from 'moment';
import {
  Bullseye,
  Flex,
  Gallery,
  GalleryItem,
  Icon,
  PageSection,
  Switch,
  Title,
} from '@patternfly/react-core';
import { StarIcon } from '@patternfly/react-icons';

import { ClusterServiceApi } from 'generated/client';
import configuration from 'client/configuration';
import { useUserAuth } from 'containers/UserAuthProvider';
import LinkCard from 'components/LinkCard';
import Lifespan from 'components/Lifespan';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import assertDefined from 'utils/assertDefined';
import { useQuery } from '@tanstack/react-query';

const clusterService = new ClusterServiceApi(configuration);

function NoClustersMessage(): ReactElement {
  return (
    <Bullseye className="pf-v6-u-w-100 pf-v6-u-h-100">
      <Title headingLevel="h3">You haven‘t created any clusters recently</Title>
    </Bullseye>
  );
}

type ClusterCardsProps = {
  showAllClusters: boolean;
};

function ClusterCards({ showAllClusters = false }: ClusterCardsProps): ReactElement {
  const { user } = useUserAuth();

  const myClustersRequest = useQuery({
    queryKey: ['clusters', false],
    queryFn: () => clusterService.list(false),
  });
  const allClustersRequest = useQuery({
    queryKey: ['clusters', true],
    queryFn: () => clusterService.list(true),
  });
  const requestToUse = showAllClusters ? allClustersRequest : myClustersRequest;
  const { isLoading: loading, error, data: rawData } = requestToUse;
  const data = rawData?.data;

  if (loading) {
    return <FullPageSpinner title="Loading infra clusters" />;
  }

  if (error || !data) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  if (!data.Clusters) {
    return <NoClustersMessage />;
  }

  // sorted in descending order by creation date
  const sortedClusters = data.Clusters.sort((c1, c2) =>
    moment(c1.CreatedOn).isBefore(c2.CreatedOn) ? 1 : -1
  );

  if (sortedClusters.length === 0) {
    return <NoClustersMessage />;
  }

  const cards = sortedClusters.map((cluster) => {
    assertDefined(cluster.ID);

    const isMyCluster = cluster.Owner === user?.Email;
    return (
      <GalleryItem key={cluster.ID}>
        <LinkCard
          to={`cluster/${cluster.ID}`}
          header={
            <Flex justifyContent={{ default: 'justifyContentSpaceBetween' }}>
              <span>{cluster.ID || 'No ID'}</span>
              {isMyCluster && (
                <Icon>
                  <StarIcon className="pf-v6-u-icon-color-favorite" />
                </Icon>
              )}
            </Flex>
          }
          footer={cluster.Status && <Lifespan cluster={cluster} />}
        >
          {cluster.Description && <p>Description: {cluster.Description}</p>}
          <p>Status: {cluster.Status || 'FAILED'}</p>
          <p>Flavor: {cluster.Flavor || 'Unknown'}</p>
        </LinkCard>
      </GalleryItem>
    );
  });

  return (
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
      {cards}
    </Gallery>
  );
}

export default function MyClustersPageSection(): ReactElement {
  const [searchParams, setSearchParams] = useSearchParams();
  const showAllClusters = searchParams.get('showAllClusters') === 'true';

  function toggleClusterFilter() {
    const newSearchParams = new URLSearchParams(searchParams);
    newSearchParams.set('showAllClusters', (!showAllClusters).toString());
    setSearchParams(newSearchParams);
  }

  return (
    <>
      <PageSection>
        <Flex
          justifyContent={{ default: 'justifyContentSpaceBetween' }}
          alignItems={{ default: 'alignItemsCenter' }}
        >
          <Title headingLevel="h2">{showAllClusters ? 'All Clusters' : 'My Clusters'}</Title>
          <Switch
            label="Show All Clusters"
            id="cluster-filter-toggle"
            name="cluster-filter-toggle"
            isChecked={showAllClusters}
            onChange={toggleClusterFilter}
          />
        </Flex>
      </PageSection>
      <PageSection>
        <ClusterCards showAllClusters={showAllClusters} />
      </PageSection>
    </>
  );
}

/* eslint-enable jsx-a11y/label-has-associated-control */
