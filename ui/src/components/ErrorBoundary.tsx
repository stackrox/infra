import React, { ReactNode, Component, ReactElement } from 'react';
import { useLocation } from 'react-router-dom';
import { Location } from 'history';
import { XSquare } from 'react-feather';

type Props = {
  message: string;
  children: ReactNode;
};

type PropsWithLocation = Props & {
  location: Location;
};

type State = {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  error: any;
  errorLocation: Location | null;
};

class ErrorBoundary extends Component<PropsWithLocation, State> {
  constructor(props: PropsWithLocation) {
    super(props);
    this.state = {
      error: null,
      errorLocation: null,
    };
  }

  static getDerivedStateFromProps(nextProps: PropsWithLocation, state: State): State | null {
    if (state.error && nextProps.location !== state.errorLocation) {
      // stop showing error on location change to allow user to navigate after error happens
      return { error: null, errorLocation: null };
    }
    return null;
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  componentDidCatch(error: any): void {
    const { location } = this.props;
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    this.setState({ error, errorLocation: location });
  }

  render(): ReactElement {
    const { message, children } = this.props;
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    const { error } = this.state;

    if (error) {
      return (
        <div className="flex h-full items-center justify-center text-base-600">
          <XSquare size="48" />
          {/* eslint-disable-next-line @typescript-eslint/no-unsafe-member-access */}
          <p className="ml-2 text-lg">{message || error.message || 'Unexpected error occurred'}</p>
        </div>
      );
    }

    return <>{children}</>;
  }
}

export default function WithLocation({ children, message }: Props): ReactElement {
  const location = useLocation();
  return (
    <ErrorBoundary message={message} location={location}>
      {children}
    </ErrorBoundary>
  );
}
