/* eslint-disable react/jsx-props-no-spreading */
import React, { ReactElement } from 'react';
import { Link } from 'react-router-dom';
import { Button } from '@patternfly/react-core';
import { TerminalIcon } from '@patternfly/react-icons';

import AppHeaderLayout from 'components/AppHeaderLayout';
import ProductLogoTile from './ProductLogoTile';
import UserInfo from './UserInfo';

export default function AppHeader(): ReactElement {
  return (
    <AppHeaderLayout
      logo={<ProductLogoTile />}
      main={
        <Button
          component={(props) => <Link {...props} to="/downloads" />}
          variant="control"
          icon={<TerminalIcon />}
        >
          infractl
        </Button>
      }
      ending={<UserInfo />}
    />
  );
}

/* eslint-enable react/jsx-props-no-spreading */
