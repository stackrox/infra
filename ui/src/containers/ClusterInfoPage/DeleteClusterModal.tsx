import React, { ReactElement } from 'react';
import { AlertCircle, CheckCircle } from 'react-feather';
import { ClipLoader } from 'react-spinners';

import { V1Cluster, ClusterServiceApi } from 'generated/client';
import configuration from 'client/configuration';
import useApiOperation from 'client/useApiOperation';
import Modal from 'components/Modal';
import InformationalModal from 'components/InformationalModal';
import assertDefined from 'utils/assertDefined';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  cluster: V1Cluster;
  onCancel: () => void;
  onDeleted: () => void;
};

export default function DeleteClusterModal({ cluster, onCancel, onDeleted }: Props): ReactElement {
  const [deleteCluster, { called, loading, error }] = useApiOperation(() => {
    assertDefined(cluster.ID); // swagger definitions are too permitting
    return clusterService.clusterServiceDelete(cluster.ID); // eslint-disable-line no-underscore-dangle
  });

  assertDefined(cluster.ID); // swagger definitions are too permitting

  if (!called) {
    // waiting for user confirmation
    const buttons = (
      <>
        <button type="button" className="btn btn-base mr-2" onClick={deleteCluster}>
          Yes
        </button>
        <button type="button" className="btn btn-base" onClick={onCancel}>
          Cancel
        </button>
      </>
    );

    return (
      <Modal
        isOpen
        onRequestClose={onCancel}
        header={`Are you sure you want to delete ${cluster.ID}?`}
        buttons={buttons}
      >
        <span className="text-xl">This action cannot be undone.</span>
      </Modal>
    );
  }

  if (loading) {
    // waiting for server response
    return (
      <Modal isOpen onRequestClose={(): void => {}} header={`Deleting ${cluster.ID}...`}>
        <div className="flex mb-4 w-64 items-center justify-center">
          <ClipLoader size={32} color="currentColor" />
        </div>
      </Modal>
    );
  }

  if (error) {
    // operation failed
    const message = `Cannot delete cluster. Server error occurred: "${error.message}".`;
    return (
      <InformationalModal header={`Failed to delete ${cluster.ID}!`} onAcknowledged={onCancel}>
        <div className="flex items-center">
          <AlertCircle size={16} className="mr-2 text-alert-600" />
          <span className="text-lg text-alert-600">{message}</span>
        </div>
      </InformationalModal>
    );
  }

  // no need to check for data response from the server, "no error happened" means operation was successful
  const message = `Cluster ${cluster.ID} is being destroyed now.`;
  return (
    <InformationalModal header={`Successfully deleted ${cluster.ID}!`} onAcknowledged={onDeleted}>
      <div className="flex items-center">
        <CheckCircle size={16} className="mr-2 text-success-600" />
        <span className="text-lg text-success-600">{message}</span>
      </div>
    </InformationalModal>
  );
}
