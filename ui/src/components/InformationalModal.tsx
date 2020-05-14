import React, { ReactElement, ReactNode } from 'react';

import Modal from 'components/Modal';

type Props = {
  header: string;
  /** the body of the modal */
  children: ReactNode;
  onAcknowledged: () => void;
};

/**
 * The component to show an informational message in a modal dialog. It's preferred to use it over `Modal`
 * for cases when user is only expected to acknowledge the informative message (not make a choice etc.)
 *
 * @see {@link components/Modal}
 * @param {Props} props
 */
export default function InformationalModal({
  header,
  children,
  onAcknowledged,
}: Props): ReactElement {
  const button = (
    <button type="button" className="btn btn-base" onClick={onAcknowledged}>
      OK
    </button>
  );

  return (
    <Modal isOpen onRequestClose={onAcknowledged} header={header} buttons={button}>
      {children}
    </Modal>
  );
}
