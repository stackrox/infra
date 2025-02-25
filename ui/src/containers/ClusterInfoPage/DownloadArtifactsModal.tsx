import React, { ReactElement, useCallback } from 'react';
import { Button, ClipboardCopy, Flex, List, ListItem } from '@patternfly/react-core';

import { V1Cluster, ClusterServiceApi, V1Artifact } from 'generated/client';
import configuration from 'client/configuration';
import Modal from 'components/Modal';
import useApiQuery from 'client/useApiQuery';
import assertDefined from 'utils/assertDefined';

const clusterService = new ClusterServiceApi(configuration);

type ArtifactsListProps = {
  artifacts: V1Artifact[];
};

function ArtifactsList({ artifacts }: ArtifactsListProps): ReactElement {
  return (
    <List className="pf-v6-u-mb-md">
      {artifacts
        .sort((a, b) => {
          if (a.Description && !b.Description) {
            return -1;
          }
          if (!a.Description && b.Description) {
            return 1;
          }
          return 0;
        })
        .map((artifact) => (
          <ListItem key={artifact.URL}>
            <a href={artifact.URL}>{artifact.Name}</a> - {artifact.Description}
          </ListItem>
        ))}
    </List>
  );
}

type ArtifactsProps = {
  cluster: V1Cluster;
};

function Artifacts({ cluster }: ArtifactsProps): ReactElement {
  const fetchArtifacts = useCallback(() => clusterService.artifacts(cluster.ID || ''), [
    cluster.ID,
  ]);
  const { loading, error, data: artifacts } = useApiQuery(fetchArtifacts);

  if (loading) {
    return <p>Loading...</p>;
  }

  if (error) {
    return <p>Cannot load artifacts: {error.message}</p>;
  }

  if (artifacts?.Artifacts?.length) {
    return (
      <div>
        <ArtifactsList artifacts={artifacts.Artifacts} />
        <Flex direction={{ default: 'column' }} spaceItems={{ default: 'spaceItemsSm' }}>
          <p>Note: You can download all artifacts at the command line with:</p>
          <ClipboardCopy isReadOnly hoverTip="Copy command" clickTip="Command copied!">
            {`infractl artifacts --download-dir=<some dir> ${cluster.ID ?? ''}`}
          </ClipboardCopy>
        </Flex>
      </div>
    );
  }

  return <p>There are no artifacts for this cluster.</p>;
}

type Props = {
  cluster: V1Cluster;
  onClose: () => void;
};

export default function DownloadArtifactsModal({ cluster, onClose }: Props): ReactElement {
  assertDefined(cluster.ID);

  const closeButton = (
    <Button variant="primary" onClick={onClose}>
      Close
    </Button>
  );

  return (
    <Modal
      isOpen
      onRequestClose={onClose}
      header={`Artifacts for ${cluster.ID}`}
      buttons={[closeButton]}
    >
      <Artifacts cluster={cluster} />
    </Modal>
  );
}
