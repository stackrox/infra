import React, { useState, useCallback, ReactElement } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Flex, PageSection, Title } from '@patternfly/react-core';
import { DownloadIcon, TrashIcon } from '@patternfly/react-icons';

import { ClusterServiceApi, V1Status } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
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
  const { clusterId = '' } = useParams();
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

  const clusterIsReady = cluster.Status === V1Status.Ready;

  return (
    <>
      <div style={{ overflow: 'auto' }}>
        <PageSection className="pf-v6-u-h-100" style={{ overflow: 'auto' }}>
          <Flex direction={{ default: 'column' }}>
            <Flex justifyContent={{ default: 'justifyContentSpaceBetween' }}>
              <Title headingLevel="h1">
                {cluster.ID}
                {cluster.Description && ` (${cluster.Description})`} -{' '}
                {cluster.Status || V1Status.Failed}
              </Title>
              {!!cluster && <MutableLifespan cluster={cluster} />}
            </Flex>
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
          </Flex>
        </PageSection>

        <PageSection>
          <ClusterLogs clusterId={clusterId} />
        </PageSection>
      </div>

      <PageSection>
        <Flex justifyContent={{ default: 'justifyContentSpaceBetween' }}>
          <Button
            onClick={(): void => setDownloadArtifactsOpen(true)}
            isDisabled={!clusterIsReady}
            icon={<DownloadIcon />}
          >
            Artifacts
          </Button>

          <Button
            onClick={(): void => setDeletionModalOpen(true)}
            isDisabled={!clusterIsReady}
            icon={<TrashIcon />}
            variant="danger"
          >
            Delete
          </Button>
        </Flex>
      </PageSection>

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
