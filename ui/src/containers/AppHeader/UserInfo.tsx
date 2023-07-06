import React, { ReactElement } from 'react';

import { Avatar } from '@patternfly/react-core';
import { useUserAuth } from 'containers/UserAuthProvider';
import { LogOut } from 'react-feather';

export default function UserInfo(): ReactElement {
  const { user, logout } = useUserAuth();
  return (
    <div className="flex flex-row h-full w-full items-center mr-2">
      {user?.Picture ? (
        <Avatar
          alt={user?.Name || 'anonymous'}
          src={user.Picture}
          size="md"
          border="dark"
          className="flex justify-center items-center mr-2"
        />
      ) : (
        <p className="flex justify-center items-center mr-2">{user?.Name || 'anonymous'}</p>
      )}
      <button onClick={logout} type="button" className="btn btn-base">
        <LogOut size={16} className="mr-2" />
        Logout
      </button>
    </div>
  );
}
