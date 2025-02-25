import React, { ReactElement } from 'react';

import { Avatar, Button, Flex } from '@patternfly/react-core';
import { useUserAuth } from 'containers/UserAuthProvider';
import { LogOut } from 'react-feather';

export default function UserInfo(): ReactElement {
  const { user, logout } = useUserAuth();
  return (
    <Flex alignItems={{ default: 'alignItemsCenter' }}>
      {user?.Picture ? (
        <Avatar alt={user?.Name || 'anonymous'} src={user.Picture} size="md" isBordered />
      ) : (
        <p>{user?.Name || 'anonymous'}</p>
      )}
      <Button onClick={logout} variant="control" icon={<LogOut size={16} />}>
        Logout
      </Button>
    </Flex>
  );
}
