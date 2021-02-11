import React, { useCallback, ReactElement } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

import { FlavorServiceApi, V1Flavor } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import PageSection from 'components/PageSection';
import ErrorBoundary from 'components/ErrorBoundary';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import { getOneClickFlavor } from 'utils/poc.utils';
import ClusterForm from './ClusterForm';

const flavorService = new FlavorServiceApi(configuration);

export default function LaunchClusterPage(): ReactElement {
  const { flavorId } = useParams();
  const navigate = useNavigate();
  const fetchFlavorInfo = useCallback(() => flavorService.info(flavorId), [flavorId]);
  const { loading, error, data } = useApiQuery(fetchFlavorInfo);

  if (loading) {
    return <FullPageSpinner />;
  }

  const pocData =
    flavorId === 'one-click-release-demo' ? (getOneClickFlavor() as V1Flavor) : (data as V1Flavor);

  if (flavorId !== 'one-click-release-demo' && (error || !pocData?.Name || !pocData?.Parameters)) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  if (!pocData.Parameters || !pocData.Name) {
    return <FullPageError message="Missing cluster flavor params" />;
  }

  return (
    <PageSection header={`Launch "${pocData.Name}" Cluster (${pocData?.Availability || 'Alpha'})`}>
      <ErrorBoundary message="UI doesn't support this flavor yet. Use infractl instead.">
        <ClusterForm
          flavorId={flavorId}
          flavorParameters={pocData.Parameters}
          onClusterCreated={(clusterId): void => navigate(`/cluster/${clusterId}`)}
        />
      </ErrorBoundary>
    </PageSection>
  );
}
