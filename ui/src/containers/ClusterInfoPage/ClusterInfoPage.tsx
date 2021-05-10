import React, { useState, useCallback, ReactElement } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Download, Trash2 } from 'react-feather';

import { ClusterServiceApi, V1Status } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import PageSection from 'components/PageSection';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import ClusterLogs from './ClusterLogs';
import ClusterConnect from './ClusterConnect';
import DeleteClusterModal from './DeleteClusterModal';
import DownloadArtifactsModal from './DownloadArtifactsModal';
import MutableLifespan from './MutableLifespan';

const clusterService = new ClusterServiceApi(configuration);

export default function ClusterInfoPage(): ReactElement {
  const navigate = useNavigate();
  const { clusterId } = useParams();
  const fetchClusterInfo = useCallback(() => clusterService.info(clusterId), [clusterId]);
  const { loading, error, data: cluster } = useApiQuery(fetchClusterInfo, { pollInterval: 10000 });
  const [deletionModalOpen, setDeletionModalOpen] = useState<boolean>(false);
  const [downloadArtifactsOpen, setDownloadArtifactsOpen] = useState<boolean>(false);

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !cluster?.ID) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  const sectionHeader = (
    <div className="flex flex-col space-y-2">
      <div className="flex justify-between">
        <div>
          <span className="lowercase">{cluster.ID}</span>
          <span>
            {cluster.Description && ` (${cluster.Description})`} -{' '}
            {cluster.Status || V1Status.Failed}
          </span>
        </div>
        {!!cluster && <MutableLifespan cluster={cluster} />}
      </div>
      {cluster.Connect && <ClusterConnect connect={cluster.Connect} />}
      {cluster.URL && (
        <span className="text-base normal-case">
          URL:{' '}
          <a
            href={cluster.URL}
            className="underline text-blue-500"
            target="_blank"
            rel="noreferrer"
          >
            {cluster.URL}
          </a>
        </span>
      )}
    </div>
  );

  const clusterIsReady = cluster.Status === V1Status.Ready;

  return (
    <>
      <PageSection header={sectionHeader}>
        <div className="flex flex-col">
          <ClusterLogs clusterId={clusterId} />
        </div>
      </PageSection>

      <div className=" flex border-base-400 border-t p-4">
        <button
          className="btn btn-base"
          type="button"
          onClick={(): void => setDownloadArtifactsOpen(true)}
          disabled={!clusterIsReady}
        >
          <Download size={16} className="mr-2" />
          Artifacts
        </button>

        <button
          className="btn btn-base ml-auto"
          type="button"
          onClick={(): void => setDeletionModalOpen(true)}
          disabled={!clusterIsReady}
        >
          <Trash2 size={16} className="mr-2" />
          Delete
        </button>
      </div>

      {deletionModalOpen && (
        <DeleteClusterModal
          cluster={cluster}
          onCancel={(): void => setDeletionModalOpen(false)}
          onDeleted={(): void => navigate('/')}
        />
      )}

      {downloadArtifactsOpen && (
        <DownloadArtifactsModal
          cluster={cluster}
          onClose={(): void => setDownloadArtifactsOpen(false)}
        />
      )}
    </>
  );
}
