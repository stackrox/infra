import React, { ReactElement, ReactNode } from 'react';
import {
	Modal,
	ModalVariant
} from '@patternfly/react-core/deprecated';

type Props = {
  isOpen: boolean;
  onRequestClose: () => void;
  header: string;
  children: ReactNode;
  buttons?: ReactNode;
};

export default ({ isOpen, onRequestClose, header, children, buttons }: Props): ReactElement => (
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
