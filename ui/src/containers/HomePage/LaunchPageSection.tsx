import React, { ReactElement } from 'react';
import { AxiosPromise } from 'axios';

import { V1FlavorListResponse, FlavorServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import PageSection from 'components/PageSection';
import LinkCard from 'components/LinkCard';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import assertDefined from 'utils/assertDefined';

const flavorService = new FlavorServiceApi(configuration);

const fetchFlavors = (): AxiosPromise<V1FlavorListResponse> => flavorService.list();

function FlavorCards(): ReactElement {
  const { loading, error, data } = useApiQuery(fetchFlavors);

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !data?.Flavors) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  const cards = data.Flavors.map((flavor) => {
    assertDefined(flavor.ID); // swagger definitions are too permitting
    return (
      <LinkCard
        key={flavor.ID}
        to={`launch/${flavor.ID}`}
        header={flavor.Name || 'Unnamed'}
        footer={<span className="capitalize">{flavor.Availability || 'Alpha'}</span>}
        className="m-2"
      >
        <p className="text-lg">{flavor.Description}</p>
      </LinkCard>
    );
  });
  return <>{cards}</>;
}

export default function LaunchPageSection(): ReactElement {
  return (
    <PageSection header="Launch Cluster">
      <div className="flex flex-wrap -m-2">
        <FlavorCards />
      </div>
    </PageSection>
  );
}
