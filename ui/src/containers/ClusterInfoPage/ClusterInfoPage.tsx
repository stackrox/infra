import React, { useState, useCallback, ReactElement } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Download, Trash2 } from 'react-feather';
import { Tooltip, TooltipOverlay } from '@stackrox/ui-components';

import { ClusterServiceApi, V1Status } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import PageSection from 'components/PageSection';
import ClusterLifespanCountdown from 'components/ClusterLifespanCountdown';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import ClusterLogs from './ClusterLogs';
import DeleteClusterModal from './DeleteClusterModal';

const clusterService = new ClusterServiceApi(configuration);

function modifyLifespan(notation: string, incOrDec: string): void {
  // eslint-disable-next-line no-console
  console.log(notation, incOrDec);
}

export default function ClusterInfoPage(): ReactElement {
  const navigate = useNavigate();
  const { clusterId } = useParams();
  const fetchClusterInfo = useCallback(() => clusterService.info(clusterId), [clusterId]);
  const { loading, error, data } = useApiQuery(fetchClusterInfo, { pollInterval: 10000 });
  const [deletionModalOpen, setDeletionModalOpen] = useState<boolean>(false);

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !data?.ID) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  const sectionHeader = (
    <div className="flex justify-between">
      <div>
        <span className="lowercase">{data.ID}</span>
        <span>
          {data.Description && ` (${data.Description})`} - {data.Status || 'FAILED'}
        </span>
      </div>
      <ClusterLifespanCountdown cluster={data} canModify onModify={modifyLifespan} />
    </div>
  );

  return (
    <>
      <PageSection header={sectionHeader}>
        <div className="flex flex-col">
          <ClusterLogs clusterId={clusterId} />
        </div>
      </PageSection>

      <div className=" flex border-base-400 border-t p-4">
        <Tooltip content={<TooltipOverlay>Not supported yet. Use infractl.</TooltipOverlay>}>
          <button className="btn btn-base" type="button">
            <Download size={16} className="mr-2" />
            Artifacts
          </button>
        </Tooltip>

        {data.Status &&
          data.Status === V1Status.READY && ( // show Delete only for running clusters
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
          cluster={data}
          onCancel={(): void => setDeletionModalOpen(false)}
          onDeleted={(): void => navigate('/')}
        />
      )}
    </>
  );
}
