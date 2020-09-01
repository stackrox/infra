import React, { ReactElement, useCallback } from 'react';

import { V1Cluster, ClusterServiceApi, V1Artifact } from 'generated/client';
import configuration from 'client/configuration';
import Modal from 'components/Modal';
import useApiQuery from 'client/useApiQuery';
import { X } from 'react-feather';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  cluster: V1Cluster;
  onClose: () => void;
};

export default function DownloadArtifactsModal({ cluster, onClose }: Props): ReactElement {
  const closeButton = (
    <button type="button" className="btn btn-base" onClick={onClose}>
      <X size={16} className="mr-2" /> Close
    </button>
  );

  return (
    <Modal
      isOpen
      onRequestClose={onClose}
      header={`Artifacts for ${cluster.ID}`}
      buttons={closeButton}
    >
      <Artifacts cluster={cluster} />
    </Modal>
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
      <>
        <ArtifactsList artifacts={artifacts.Artifacts} />
        <p>Note: You can download all artifacts at the command line with:</p>
        <code>infractl artifacts --download-dir=&lt;some dir&gt; {cluster.ID}</code>
      </>
    );
  }

  return <p>There are no artifacts for this cluster.</p>;
}

type ArtifactsListProps = {
  artifacts: V1Artifact[];
};

function ArtifactsList({ artifacts }: ArtifactsListProps): ReactElement {
  return (
    <ul className="list-disc ml-5">
      {artifacts.map((artifact: V1Artifact) => (
        <li key={artifact.Name}>
          <a href={artifact.URL} className="underline text-blue-500">
            {artifact.Name}
          </a>{' '}
          - {artifact.Description}
        </li>
      ))}
    </ul>
  );
}
