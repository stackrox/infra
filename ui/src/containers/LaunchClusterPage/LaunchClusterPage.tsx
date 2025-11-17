import React, { ReactElement } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { PageSection, Title } from '@patternfly/react-core';
import { useQuery, useQueryClient } from '@tanstack/react-query';

import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import { flavorInfoQueryOptions } from 'client/flavorInfoQueryOptions';
import ClusterForm from './ClusterForm';

export default function LaunchClusterPage(): ReactElement {
  const { flavorId = '' } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { isLoading: loading, error, data: rawData } = useQuery(flavorInfoQueryOptions(flavorId));
  const data = rawData?.data;
  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !data?.Name || !data?.Parameters) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  return (
    <PageSection>
      <Title headingLevel="h1" className="pf-v6-u-mb-xl">
        {`Launch "${data.Name}" Cluster (${data?.Availability || 'Alpha'})`}
      </Title>
      <ClusterForm
        flavorId={flavorId}
        flavorParameters={data.Parameters}
        onClusterCreated={(clusterId) => {
          queryClient.invalidateQueries({ queryKey: ['clusters'] });
          navigate(`/cluster/${clusterId}`);
        }}
      />
    </PageSection>
  );
}
