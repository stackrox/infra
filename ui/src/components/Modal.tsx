import React, { ReactElement, ReactNode } from 'react';
import { Modal as PatternFlyModal, ModalVariant } from '@patternfly/react-core';

type Props = {
  isOpen: boolean;
  onRequestClose: () => void;
  header: string;
  children: ReactNode;
  buttons?: ReactNode;
};

export default function Modal({
  isOpen,
  onRequestClose,
  header,
  children,
  buttons,
}: Props): ReactElement {
  return (
    <PatternFlyModal
      actions={buttons}
      isOpen={isOpen}
      onClose={onRequestClose}
      title={header}
      variant={ModalVariant.medium}
    >
      {children}
    </PatternFlyModal>
  );
}
