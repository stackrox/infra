import React, { useState, ReactElement } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Divider, Flex, PageSection, Title } from '@patternfly/react-core';
import { DownloadIcon, TrashIcon } from '@patternfly/react-icons';
import { useQuery } from '@tanstack/react-query';

import { ClusterServiceApi, V1Status } from 'generated/client';
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

  const {
    isLoading: clusterInfoLoading,
    error: clusterInfoError,
    data: clusterInfoData,
  } = useQuery({
    queryKey: ['clusterInfo', clusterId],
    queryFn: () => clusterService.info(clusterId),
    refetchInterval: 10000,
  });

  const cluster = clusterInfoData?.data;
  const [deletionModalOpen, setDeletionModalOpen] = useState<boolean>(false);
  const [downloadArtifactsOpen, setDownloadArtifactsOpen] = useState<boolean>(false);

  const clusterIsReady = cluster?.Status === V1Status.Ready;

  const {
    isLoading: clusterLogsLoading,
    error: clusterLogsError,
    data: clusterLogsData,
  } = useQuery({
    queryKey: ['clusterLogs', clusterId],
    queryFn: () => clusterService.logs(clusterId),
    refetchInterval: 10000,
  });

  return (
    <>
      <div style={{ overflow: 'auto' }}>
        {clusterInfoLoading ? (
          <FullPageSpinner title="Loading cluster information" />
        ) : clusterInfoError || !cluster?.ID ? (
          <FullPageError message={clusterInfoError?.message || 'Unexpected server response'} />
        ) : (
          <PageSection style={{ overflow: 'auto' }}>
            <Flex direction={{ default: 'column' }}>
              <Flex justifyContent={{ default: 'justifyContentSpaceBetween' }}>
                <Title headingLevel="h1">
                  {cluster.ID}
                  {cluster.Description && ` (${cluster.Description})`} -{' '}
                  {cluster.Status || V1Status.Failed}
                </Title>
                {!!cluster && <MutableLifespan cluster={cluster} />}
              </Flex>
              {cluster.Connect && (
                <>
                  <Divider component="div" />
                  <ClusterConnect connect={cluster.Connect} />
                </>
              )}
              {cluster.URL && (
                <>
                  <Divider component="div" />
                  <span>
                    URL:{' '}
                    <a href={cluster.URL} target="_blank" rel="noreferrer">
                      {cluster.URL}
                    </a>
                  </span>
                </>
              )}
            </Flex>
          </PageSection>
        )}

        <PageSection>
          {clusterLogsLoading ? (
            <FullPageSpinner title="Loading cluster setup logs" />
          ) : clusterLogsError || !clusterLogsData?.data.Logs ? (
            <FullPageError message={clusterLogsError?.message || 'No logs found'} />
          ) : (
            <ClusterLogs logs={clusterLogsData.data.Logs} />
          )}
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
          clusterId={clusterId}
          onCancel={(): void => setDeletionModalOpen(false)}
          onDeleted={(): void => navigate('/')}
        />
      )}

      {downloadArtifactsOpen && (
        <DownloadArtifactsModal
          clusterId={clusterId}
          onClose={(): void => setDownloadArtifactsOpen(false)}
        />
      )}
    </>
  );
}
