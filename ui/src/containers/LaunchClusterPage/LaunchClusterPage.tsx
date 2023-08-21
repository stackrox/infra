import React, { useCallback, ReactElement } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

import { FlavorServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import PageSection from 'components/PageSection';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import ClusterForm from './ClusterForm';

const flavorService = new FlavorServiceApi(configuration);

export default function LaunchClusterPage(): ReactElement {
  const { flavorId = '' } = useParams();
  const navigate = useNavigate();
  const fetchFlavorInfo = useCallback(() => flavorService.info(flavorId), [flavorId]);
  const { loading, error, data } = useApiQuery(fetchFlavorInfo);

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !data?.Name || !data?.Parameters) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  return (
    <PageSection header={`Launch "${data.Name}" Cluster (${data?.Availability || 'Alpha'})`}>
      <ClusterForm
        flavorId={flavorId}
        flavorParameters={data.Parameters}
        onClusterCreated={(clusterId): void => navigate(`/cluster/${clusterId}`)}
      />
    </PageSection>
  );
}
