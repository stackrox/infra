import React, { useContext, createContext, ReactElement, ReactNode } from 'react';
import { AxiosPromise } from 'axios';

import { V1User, V1WhoamiResponse, UserServiceApi } from 'generated/client';
import useApiQuery from 'client/useApiQuery';
import configuration from 'client/configuration';
import FullPageSpinner from 'components/FullPageSpinner';
import FullPageError from 'components/FullPageError';

const userService = new UserServiceApi(configuration);

const fetchWhoami = (): AxiosPromise<V1WhoamiResponse> => userService.whoami();

const logout = (): void => {
  window.location.href = '/logout';
};

export interface UserAuthContextData {
  user?: V1User;
  logout: () => void;
}

const UserAuthContext = createContext({ logout });

export const useUserAuth = (): UserAuthContextData => useContext(UserAuthContext);

type Props = {
  children: ReactNode;
};

export default function UserAuthProvider({ children }: Props): ReactElement | null {
  const { loading, error, data } = useApiQuery(fetchWhoami);

  if (loading) {
    return <FullPageSpinner />;
  }

  if (error) {
    return (
      <FullPageError message="Unexpected error while authenticating. Please reach out to the service support team." />
    );
  }

  if (!data?.User) {
    // assuming we're not authenticated
    if (process.env.NODE_ENV === 'development') {
      // TODO-ivan: temporal "solution", until backend supports dynamically defined redirect URI
      return (
        <FullPageError message="For now, please add token cookie to the app through browser dev tools. Then refresh the page." />
      );
    }
    window.location.href = '/login';
    return null;
  }

  const contextValue: UserAuthContextData = {
    user: data.User,
    logout,
  };

  return <UserAuthContext.Provider value={contextValue}>{children}</UserAuthContext.Provider>;
}
