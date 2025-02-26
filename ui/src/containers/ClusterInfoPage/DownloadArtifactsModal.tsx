import React, { ReactElement } from 'react';
import { Button, ClipboardCopy, Flex, List, ListItem } from '@patternfly/react-core';

import { ClusterServiceApi, V1Artifact } from 'generated/client';
import configuration from 'client/configuration';
import Modal from 'components/Modal';
import assertDefined from 'utils/assertDefined';
import { useQuery } from '@tanstack/react-query';

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
  clusterId: string;
};

function Artifacts({ clusterId }: ArtifactsProps): ReactElement {
  const { isLoading: loading, error, data: rawData } = useQuery({
    queryKey: ['clusterArtifacts', clusterId],
    queryFn: () => clusterService.artifacts(clusterId || ''),
  });
  const artifacts = rawData?.data.Artifacts;

  if (loading) {
    return <p>Loading...</p>;
  }

  if (error) {
    return <p>Cannot load artifacts: {error.message}</p>;
  }

  if (artifacts?.length) {
    return (
      <div>
        <ArtifactsList artifacts={artifacts} />
        <Flex direction={{ default: 'column' }} spaceItems={{ default: 'spaceItemsSm' }}>
          <p>Note: You can download all artifacts at the command line with:</p>
          <ClipboardCopy isReadOnly hoverTip="Copy command" clickTip="Command copied!">
            {`infractl artifacts --download-dir=<some dir> ${clusterId ?? ''}`}
          </ClipboardCopy>
        </Flex>
      </div>
    );
  }

  return <p>There are no artifacts for this cluster.</p>;
}

type Props = {
  clusterId: string;
  onClose: () => void;
};

export default function DownloadArtifactsModal({ clusterId, onClose }: Props): ReactElement {
  assertDefined(clusterId);

  const closeButton = (
    <Button variant="primary" onClick={onClose}>
      Close
    </Button>
  );

  return (
    <Modal
      isOpen
      onRequestClose={onClose}
      header={`Artifacts for ${clusterId}`}
      buttons={[closeButton]}
    >
      <Artifacts clusterId={clusterId} />
    </Modal>
  );
}
