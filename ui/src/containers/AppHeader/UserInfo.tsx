import React, { ReactElement } from 'react';

import Avatar from 'components/Avatar';
import { useUserAuth } from 'containers/UserAuthProvider';

export default function UserInfo(): ReactElement {
  const { user, logout } = useUserAuth();
  return (
    <div className="flex flex-row h-full w-full items-center mr-2">
      <Avatar name={user?.Name} imageSrc={user?.Picture} className="mr-2" />
      <button onClick={logout} type="button" className="btn btn-base">
        Logout
      </button>
    </div>
  );
}
