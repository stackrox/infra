import React, { ReactElement, useState } from 'react';
import { AlertCircle, CheckCircle } from 'react-feather';
import { ClipLoader } from 'react-spinners';

import { V1Cluster, ClusterServiceApi } from 'generated/client';
import configuration from 'client/configuration';
import Modal from 'components/Modal';

const clusterService = new ClusterServiceApi(configuration);

function ConfirmationModal(props: {
  header: string;
  message: string;
  success: boolean;
  onAcknowledged: () => void;
}): ReactElement {
  const { header, message, success, onAcknowledged } = props;

  const button = (
    <button type="button" className="btn btn-base" onClick={onAcknowledged}>
      OK
    </button>
  );

  return (
    <Modal isOpen onRequestClose={onAcknowledged} header={header} buttons={button}>
      <div className="flex items-center">
        {success ? (
          <CheckCircle size={16} className="mr-2 text-success-600" />
        ) : (
          <AlertCircle size={16} className="mr-2 text-alert-600" />
        )}
        <span className={`text-lg ${success ? 'text-success-600' : 'text-alert-600'}`}>
          {message}
        </span>
      </div>
    </Modal>
  );
}

type Props = {
  cluster: V1Cluster;
  isOpen: boolean;
  onCancel: () => void;
  onDeleted: () => void;
};

type RequestState = {
  processing: boolean;
  success: boolean;
  error?: any; // eslint-disable-line @typescript-eslint/no-explicit-any
};

export default function DeleteClusterModal({
  cluster,
  isOpen,
  onCancel,
  onDeleted,
}: Props): ReactElement | null {
  const [requestState, setRequestState] = useState<RequestState>({
    processing: false,
    success: false,
  });

  if (!cluster.ID) return null; // should never happen, better swagger definitions are needed

  if (requestState.processing) {
    return (
      <Modal isOpen onRequestClose={(): void => {}} header={`Deleting ${cluster.ID}...`}>
        <div className="flex mb-4 w-64 items-center justify-center">
          <ClipLoader size={32} color="currentColor" />
        </div>
      </Modal>
    );
  }

  if (requestState.error) {
    const message = `Cannot delete cluster. Server error occurred: "${requestState.error.message}".`;
    const onAcknowledged = (): void => {
      setRequestState({ processing: false, error: undefined, success: false });
      onCancel();
    };
    return (
      <ConfirmationModal
        header={`Failed to delete ${cluster.ID}!`}
        success={false}
        message={message}
        onAcknowledged={onAcknowledged}
      />
    );
  }

  if (requestState.success) {
    const message = `Cluster ${cluster.ID} was deleted.`;
    const onAcknowledged = (): void => {
      setRequestState({ processing: false, error: undefined, success: false });
      onDeleted();
    };
    return (
      <ConfirmationModal
        header={`Successfully deleted ${cluster.ID}!`}
        success
        message={message}
        onAcknowledged={onAcknowledged}
      />
    );
  }

  const onDelete = (): void => {
    if (!cluster.ID) return; // should never happen
    setRequestState({ processing: true, error: undefined, success: false });

    // eslint-disable-next-line no-underscore-dangle
    clusterService
      ._delete(`${cluster.ID}`)
      .then(() => {
        setRequestState({ processing: false, error: undefined, success: true });
      })
      .catch((error) => {
        setRequestState({ processing: false, error, success: false });
      });
  };

  const buttons = (
    <>
      <button type="button" className="btn btn-base mr-2" onClick={onDelete}>
        Yes
      </button>
      <button type="button" className="btn btn-base" onClick={onCancel}>
        Cancel
      </button>
    </>
  );

  return (
    <Modal
      isOpen={isOpen}
      onRequestClose={onCancel}
      header={`Are you sure you want delete ${cluster.ID}?`}
      buttons={buttons}
    >
      <span className="text-xl">This action cannot be undone.</span>
    </Modal>
  );
}
