/* eslint-disable react/jsx-props-no-spreading */
import React, { ReactElement } from 'react';
import { Link } from 'react-router-dom';
import { Avatar, Button, Flex } from '@patternfly/react-core';
import { OutlinedHandPointRightIcon, TerminalIcon } from '@patternfly/react-icons';

import AppHeaderLayout from 'components/AppHeaderLayout';
import RHACSLogo from 'components/RHACSLogo';
import { useUserAuth } from 'containers/UserAuthProvider';

export default function AppHeader(): ReactElement {
  const { user, logout } = useUserAuth();
  return (
    <AppHeaderLayout
      logo={
        <Link to="/">
          <RHACSLogo />
        </Link>
      }
      main={
        <Button
          component={(props) => <Link {...props} to="/downloads" />}
          variant="control"
          icon={<TerminalIcon />}
        >
          infractl
        </Button>
      }
      ending={
        <Flex alignItems={{ default: 'alignItemsCenter' }}>
          {user?.Picture ? (
            <Avatar alt={user?.Name || 'anonymous'} src={user.Picture} size="md" isBordered />
          ) : (
            <p>{user?.Name || 'anonymous'}</p>
          )}
          <Button onClick={logout} variant="control" icon={<OutlinedHandPointRightIcon />}>
            Logout
          </Button>
        </Flex>
      }
    />
  );
}

/* eslint-enable react/jsx-props-no-spreading */
