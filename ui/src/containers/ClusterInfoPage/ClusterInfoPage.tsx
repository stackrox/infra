import React, { useState, useEffect, useCallback, ReactElement } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Download, Trash2 } from 'react-feather';
<<<<<<< HEAD
import { Tooltip, TooltipOverlay } from '@stackrox/ui-components';
=======
import moment from 'moment';
>>>>>>> 1d73f24... optimistically update lifespan

import { ClusterServiceApi, V1Status } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import PageSection from 'components/PageSection';
import ClusterLifespanCountdown, { lifespanToDuration } from 'components/ClusterLifespanCountdown';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';
import ClusterLogs from './ClusterLogs';
import DeleteClusterModal from './DeleteClusterModal';

const clusterService = new ClusterServiceApi(configuration);

export default function ClusterInfoPage(): ReactElement {
  const navigate = useNavigate();
  const { clusterId } = useParams();
  const fetchClusterInfo = useCallback(() => clusterService.info(clusterId), [clusterId]);
  const { loading, error, data } = useApiQuery(fetchClusterInfo, { pollInterval: 10000 });
  const [deletionModalOpen, setDeletionModalOpen] = useState<boolean>(false);
  const [clientSideLifespan, setClientSideLifespan] = useState<string>('');

  useEffect(() => {
    if (data && data.Lifespan === clientSideLifespan) {
      // Clear the client side optimistic setting when the fetchClusterInfo poll catches up
      setClientSideLifespan('');
    }
  }, [data, clientSideLifespan]);

  const cluster = clientSideLifespan ? { ...data, Lifespan: clientSideLifespan } : data;

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error || !cluster?.ID) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  const modifyLifespan = async (notation: string, incOrDec: string): Promise<void> => {
    if (cluster?.Lifespan) {
      const current = lifespanToDuration(cluster.Lifespan);
      const delta = moment.duration(1, notation as moment.DurationInputArg2);
      const update = incOrDec === 'inc' ? current.add(delta) : current.subtract(delta);
      setClientSideLifespan(`${update.asSeconds()}s`);
      await clusterService.lifespan(clusterId, { Lifespan: `${update.asSeconds()}s` });
    }
  };

  const sectionHeader = (
    <div className="flex justify-between">
      <div>
        <span className="lowercase">{cluster.ID}</span>
        <span>
          {cluster.Description && ` (${cluster.Description})`} - {cluster.Status || 'FAILED'}
        </span>
      </div>
      <ClusterLifespanCountdown cluster={cluster} canModify onModify={modifyLifespan} />
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

<<<<<<< HEAD
        {data.Status &&
          data.Status === V1Status.READY && ( // show Delete only for running clusters
=======
        {cluster.Status &&
        cluster.Status === V1Status.READY && ( // show Delete only for running clusters
>>>>>>> 1d73f24... optimistically update lifespan
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
    </>
  );
}
