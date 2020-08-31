import React, { ReactElement, useState, useEffect } from 'react';
import moment from 'moment';

import Lifespan, { lifespanToDuration } from 'components/Lifespan';
import { V1Cluster, ClusterServiceApi } from 'generated/client';
import useApiOperation from 'client/useApiOperation';
import configuration from 'client/configuration';
import FullPageError from 'components/FullPageError';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  cluster: V1Cluster;
};

export default function MutableLifespan({ cluster }: Props): ReactElement {
  const [clientSideLifespan, setClientSideLifespan] = useState<string>('');
  const [updateLifespan, { error }] = useApiOperation((notation: string, incOrDec: string) => {
    if (cluster.ID && cluster.Lifespan) {
      const current = lifespanToDuration(cluster.Lifespan);
      const delta = moment.duration(1, notation as moment.DurationInputArg2);
      const update = incOrDec === 'inc' ? current.add(delta) : current.subtract(delta);
      setClientSideLifespan(`${update.asSeconds()}s`);
      return clusterService.lifespan(cluster.ID, { Lifespan: `${update.asSeconds()}s` });
    }
  });

  useEffect(() => {
    if (cluster && cluster.Lifespan === clientSideLifespan) {
      // Clear the client side optimistic setting when the fetchClusterInfo poll catches up
      setClientSideLifespan('');
    }
  }, [cluster, clientSideLifespan]);

  if (error) {
    return <FullPageError message={error?.message || 'Unexpected server response'} />;
  }

  const modifiedCluster = clientSideLifespan
    ? { ...cluster, Lifespan: clientSideLifespan }
    : cluster;

  return <Lifespan cluster={modifiedCluster} canModify onModify={updateLifespan} />;
}
