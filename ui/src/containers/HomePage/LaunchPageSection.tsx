/* eslint-disable no-nested-ternary */
/* eslint-disable jsx-a11y/label-has-associated-control */
import React, { ReactElement } from 'react';
import { useSearchParams } from 'react-router-dom';
import { AxiosPromise } from 'axios';
import {
  Checkbox,
  Flex,
  Gallery,
  GalleryItem,
  Label,
  PageSection,
  Title,
} from '@patternfly/react-core';

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
    return <FullPageSpinner title="Loading available cluster flavors" />;
  }

  if (error || !data?.Flavors) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  const cards = data.Flavors.map((flavor) => {
    assertDefined(flavor.ID); // swagger definitions are too permitting
    return (
      <GalleryItem key={flavor.ID}>
        <LinkCard
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

export default function LaunchPageSection(): ReactElement {
  const [searchParams, setSearchParams] = useSearchParams();
  const showAllFlavors = searchParams.get('showAllFlavors') === 'true';

  function toggleFlavorFilter() {
    const newSearchParams = new URLSearchParams(searchParams);
    newSearchParams.set('showAllFlavors', (!showAllFlavors).toString());
    setSearchParams(newSearchParams);
  }

  return (
    <>
      <PageSection>
        <Flex
          justifyContent={{ default: 'justifyContentSpaceBetween' }}
          alignItems={{ default: 'alignItemsCenter' }}
        >
          <Title headingLevel="h2">{showAllFlavors ? 'All Flavors' : 'My Flavors'}</Title>
          <Checkbox
            labelPosition="start"
            label="Show All Flavors"
            id="flavor-filter-toggle"
            name="flavor-filter-toggle"
            isChecked={showAllFlavors}
            onChange={toggleFlavorFilter}
          />
        </Flex>
      </PageSection>
      <PageSection>
        <FlavorCards showAllFlavors={showAllFlavors} />
      </PageSection>
    </>
  );
}

/* eslint-enable jsx-a11y/label-has-associated-control */
/* eslint-enable no-nested-ternary */
