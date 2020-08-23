import React, { ReactElement } from 'react';

import { Avatar } from '@stackrox/ui-components';
import { useUserAuth } from 'containers/UserAuthProvider';
import { LogOut } from 'react-feather';

export default function UserInfo(): ReactElement {
  const { user, logout } = useUserAuth();
  return (
    <div className="flex flex-row h-full w-full items-center mr-2">
      <Avatar
        name={user?.Name}
        imageSrc={user?.Picture}
        className="flex w-12 h-12 justify-center items-center mr-2"
      />
      <button onClick={logout} type="button" className="btn btn-base">
        <LogOut size={16} className="mr-2" />
        Logout
      </button>
    </div>
  );
}
