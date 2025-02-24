/* eslint-disable no-nested-ternary */
/* eslint-disable jsx-a11y/label-has-associated-control */
import React, { ReactElement } from 'react';
import { useSearchParams } from 'react-router-dom';
import { AxiosPromise } from 'axios';
import { Flex, Gallery, GalleryItem, Label, PageSection, Title } from '@patternfly/react-core';

import { V1FlavorListResponse, FlavorServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import LinkCard from 'components/LinkCard';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import assertDefined from 'utils/assertDefined';

const flavorService = new FlavorServiceApi(configuration);

const fetchFlavors = (): AxiosPromise<V1FlavorListResponse> => flavorService.list(false);
const fetchAllFlavors = (): AxiosPromise<V1FlavorListResponse> => flavorService.list(true);

type FlavorCardsProps = {
  showAllFlavors: boolean;
};

function FlavorCards({ showAllFlavors = false }: FlavorCardsProps): ReactElement {
  const { loading, error, data } = useApiQuery(showAllFlavors ? fetchAllFlavors : fetchFlavors);

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !data?.Flavors) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  const cards = data.Flavors.map((flavor) => {
    assertDefined(flavor.ID); // swagger definitions are too permitting
    return (
      <GalleryItem>
        <LinkCard
          key={flavor.ID}
          to={`launch/${flavor.ID}`}
          header={
            <Flex
              alignItems={{ default: 'alignItemsCenter' }}
              justifyContent={{ default: 'justifyContentSpaceBetween' }}
            >
              <span>{flavor.Name || 'Unnamed'}</span>
              <Label
                color={
                  flavor.Availability === 'default'
                    ? 'blue'
                    : flavor.Availability === 'stable'
                    ? 'green'
                    : flavor.Availability === 'beta'
                    ? 'orange'
                    : 'orangered'
                }
              >
                {flavor.Availability || 'Alpha'}
              </Label>
            </Flex>
          }
        >
          <p>{flavor.Description}</p>
        </LinkCard>
      </GalleryItem>
    );
  });
  return <>{cards}</>;
}

export default function LaunchPageSection(): ReactElement {
  const [searchParams, setSearchParams] = useSearchParams();
  const showAllFlavors = searchParams.get('showAllFlavors') === 'true';

  function toggleFlavorFilter() {
    const newSearchParams = new URLSearchParams(searchParams);
    newSearchParams.set('showAllFlavors', (!showAllFlavors).toString());
    setSearchParams(newSearchParams);
  }

  const headerText = showAllFlavors ? 'All Flavors' : 'My Flavors';
  const flavorFilterToggle = (
    <span className="flex items-center">
      <label htmlFor="flavor-filter-toggle" className="mr-2 text-lg">
        Show All Flavors
      </label>
      <input
        type="checkbox"
        id="flavor-filter-toggle"
        checked={showAllFlavors}
        onChange={toggleFlavorFilter}
        className="w-4 h-4 rounded-sm"
      />
    </span>
  );

  const header = (
    <div className="flex justify-between items-center ">
      <Title headingLevel="h2">{headerText}</Title>
      {flavorFilterToggle}
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
          <FlavorCards showAllFlavors={showAllFlavors} />
        </Gallery>
      </PageSection>
    </>
  );
}

/* eslint-enable jsx-a11y/label-has-associated-control */
/* eslint-enable no-nested-ternary */
