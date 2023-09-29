import React, { ReactElement, ReactNode } from 'react';
import { Modal, ModalVariant } from '@patternfly/react-core';

type Props = {
  isOpen: boolean;
  onRequestClose: () => void;
  header: string;
  children: ReactNode;
  buttons?: ReactNode;
};

export default function ({
  isOpen,
  onRequestClose,
  header,
  children,
  buttons,
}: Props): ReactElement {
  return (
    <Modal
      actions={buttons}
      isOpen={isOpen}
      onClose={onRequestClose}
      title={header}
      variant={ModalVariant.medium}
    >
      {children}
    </Modal>
  );
}
