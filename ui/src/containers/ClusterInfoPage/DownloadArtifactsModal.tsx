import React, { ReactElement, useCallback } from 'react';

import { V1Cluster, ClusterServiceApi, V1Artifact } from 'generated/client';
import configuration from 'client/configuration';
import Modal from 'components/Modal';
import useApiQuery from 'client/useApiQuery';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  cluster: V1Cluster;
  onCancel: () => void;
};

export default function DownloadArtifactsModal({ cluster, onCancel }: Props): ReactElement {
  const fetchArtifacts = useCallback(() => clusterService.artifacts(cluster.ID || ''), [
    cluster.ID,
  ]);
  const { loading, error, data: artifacts } = useApiQuery(fetchArtifacts);

  return (
    <Modal isOpen onRequestClose={onCancel} header={`Artifacts for ${cluster.ID}`}>
      {loading && <p>Loading...</p>}
      {error && <p>Cannot load artifacts: `${error.message}`</p>}
      {artifacts?.Artifacts?.length === 0 && <p>This cluster has no artifacts</p>}
      {!!artifacts?.Artifacts?.length && <Artifacts artifacts={artifacts?.Artifacts || []} />}
    </Modal>
  );
}

type ArtifactsProps = {
  artifacts: V1Artifact[];
};

function Artifacts({ artifacts }: ArtifactsProps): ReactElement {
  return (
    <ul>
      {artifacts.map((artifact: V1Artifact) => (
        <li key={artifact.Name}>
          <a href={artifact.URL}>{artifact.Name}</a> - {artifact.Description}
        </li>
      ))}
    </ul>
  );
}
