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
import DeleteClusterModal from './DeleteClusterModal';
import DownloadArtifactsModal from './DownloadArtifactsModal';
// eslint-disable import/no-named-as-default
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
    <div className="flex justify-between">
      <div>
        <span className="lowercase">{cluster.ID}</span>
        <span>
          {cluster.Description && ` (${cluster.Description})`} - {cluster.Status || 'FAILED'}
        </span>
      </div>
      {!!cluster && <MutableLifespan cluster={cluster} />}
    </div>
  );

  const clusterIsReady = cluster.Status && cluster.Status === V1Status.READY;

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
        >
          <Download size={16} className="mr-2" />
          Artifacts
        </button>

        {clusterIsReady && (
          <button
            className="btn btn-base ml-auto"
            type="button"
            onClick={(): void => setDeletionModalOpen(true)}
          >
            <Trash2 size={16} className="mr-2" />
            Delete
          </button>
        )}
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
          onCancel={(): void => setDownloadArtifactsOpen(false)}
        />
      )}
    </>
  );
}
