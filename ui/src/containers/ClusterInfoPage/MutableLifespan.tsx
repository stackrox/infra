import React, { ReactElement, useState } from 'react';
import moment from 'moment';

import Lifespan, { lifespanToDuration } from 'components/Lifespan';
import { V1Cluster, ClusterServiceApi } from 'generated/client';
import configuration from 'client/configuration';
import InformationalModal from 'components/InformationalModal';
import { AlertCircle } from 'react-feather';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  cluster: V1Cluster;
};

const minimumClusterLifetime = moment.duration(15, 'm');

export default function MutableLifespan({ cluster }: Props): ReactElement {
  const [clientSideLifespan, setClientSideLifespan] = useState<string>();
  const [clearClientSideUpdate, setClearClientSideUpdate] = useState<number>();
  const [error, setError] = useState<Error | null>(null);

  if (error) {
    const message = `Cannot change the cluster lifespan. A server error occurred: "${error.message}".`;
    return (
      <InformationalModal
        header="Cannot change the cluster lifespan"
        onAcknowledged={(): void => setError(null)}
      >
        <div className="flex items-center">
          <AlertCircle size={16} className="mr-2 text-alert-600" />
          <span className="text-lg text-alert-600">{message}</span>
        </div>
      </InformationalModal>
    );
  }

  const modifiedCluster = clientSideLifespan
    ? { ...cluster, Lifespan: clientSideLifespan }
    : cluster;

  const onModify = (notation: string, incOrDec: string): void => {
    const lifespan = modifiedCluster.Lifespan;
    if (!lifespan || !modifiedCluster.ID) return;
    const current = lifespanToDuration(lifespan);
    const delta = moment.duration(1, notation as moment.DurationInputArg2);
    let update = incOrDec === 'inc' ? current.add(delta) : current.subtract(delta);
    // Make sure a user cannot accidentally expire a cluster by reducing the duration to zero or below by clicking
    // "-" on the days or hours counter. However, allow reducing the time to zero or 1m via subsequent clicks on "-" for
    // the seconds counter, this should have a sufficiently low accident rate.
    if (incOrDec === 'dec' && delta > minimumClusterLifetime && update < minimumClusterLifetime) {
      update = minimumClusterLifetime;
      if (update > current) {
        return; // no change, the deletion protection should never increase the cluster's lifetime.
      }
    }
    setClientSideLifespan(`${update.asSeconds()}s`);
    clusterService
      .lifespan(modifiedCluster.ID, { Lifespan: `${update.asSeconds()}s` })
      .then(() => {
        if (clearClientSideUpdate) clearTimeout(clearClientSideUpdate);
        setClearClientSideUpdate(
          (setTimeout(() => setClientSideLifespan(''), 20000) as unknown) as number
        );
      })
      .catch((err) => {
        setError(err);
        setClientSideLifespan('');
      });
  };

  return <Lifespan cluster={modifiedCluster} onModify={onModify} />;
}
