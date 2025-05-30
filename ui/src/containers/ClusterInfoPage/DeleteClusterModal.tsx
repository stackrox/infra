import React, { ReactElement } from 'react';
import { Alert, Button } from '@patternfly/react-core';

import { ClusterServiceApi } from 'generated/client';
import configuration from 'client/configuration';
import useApiOperation from 'client/useApiOperation';
import Modal from 'components/Modal';
import InformationalModal from 'components/InformationalModal';
import assertDefined from 'utils/assertDefined';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  clusterId: string;
  onCancel: () => void;
  onDeleted: () => void;
};

export default function DeleteClusterModal({
  clusterId,
  onCancel,
  onDeleted,
}: Props): ReactElement {
  const [deleteCluster, { called, loading, error }] = useApiOperation(() => {
    assertDefined(clusterId); // swagger definitions are too permitting
    return clusterService._delete(clusterId); // eslint-disable-line no-underscore-dangle
  });

  assertDefined(clusterId); // swagger definitions are too permitting

  if (!called) {
    // waiting for user confirmation
    const buttons = [
      <Button variant="danger" onClick={deleteCluster}>
        Yes
      </Button>,
      <Button variant="link" onClick={onCancel}>
        Cancel
      </Button>,
    ];

    return (
      <Modal
        isOpen
        onRequestClose={onCancel}
        header={`Are you sure you want to delete ${clusterId}?`}
        buttons={buttons}
      >
        <Alert isInline variant="danger" title="This will permanently delete the cluster" />
      </Modal>
    );
  }

  if (loading) {
    const message = `Cluster ${clusterId} is being destroyed now.`;
    // waiting for server response
    return (
      <Modal isOpen onRequestClose={(): void => {}} header={`Deleting ${clusterId}...`}>
        <Alert isInline variant="info" title={message} />
      </Modal>
    );
  }

  if (error) {
    // operation failed
    const message = `Could not delete cluster. Server error occurred: "${error.message}".`;
    return (
      <InformationalModal header={`Failed to delete ${clusterId}!`} onAcknowledged={onCancel}>
        <Alert isInline variant="warning" title={message} />
      </InformationalModal>
    );
  }

  // no need to check for data response from the server, "no error happened" means operation was successful
  return (
    <InformationalModal header={`Successfully deleted ${clusterId}!`} onAcknowledged={onDeleted}>
      <Alert isInline variant="success" title="The cluster was deleted" />
    </InformationalModal>
  );
}
