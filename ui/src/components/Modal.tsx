import React, { ReactElement, ReactNode } from 'react';
import ReactModal from 'react-modal';

type Props = {
  isOpen: boolean;
  onRequestClose: () => void;
  header: string;
  children: ReactNode;
  buttons?: ReactNode;
  className?: string;
};

export default function Modal({
  isOpen,
  onRequestClose,
  header,
  children,
  buttons,
  className = '',
}: Props): ReactElement {
  return (
    <ReactModal
      isOpen={isOpen}
      onRequestClose={onRequestClose}
      contentLabel="Modal"
      ariaHideApp={false}
      overlayClassName="ReactModal__Overlay react-modal-overlay p-4 flex shadow-lg rounded-sm"
      className={`ReactModal__Content mx-auto my-0 flex flex-col self-center bg-base-100 max-h-full transition ${className}`}
    >
      <h3 className="py-2 px-1 mb-4 text-2xl border-b border-base-400">{header}</h3>
      <div className="flex flex-col items-center px-2">
        {children}
        {buttons && <div className="py-4">{buttons}</div>}
      </div>
    </ReactModal>
  );
}
