import React, { ReactElement, useState } from 'react';
import moment from 'moment';

import Lifespan, { lifespanToDuration } from 'components/Lifespan';
import { V1Cluster, ClusterServiceApi } from 'generated/client';
import configuration from 'client/configuration';
import InformationalModal from 'components/InformationalModal';

const clusterService = new ClusterServiceApi(configuration);

type Props = {
  cluster: V1Cluster;
};

const minimumClusterLifetime = moment.duration(15.5, 'm');

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
        <span className="pf-u-danger-color-100 pf-u-font-size-2xl">{message}</span>
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

    // lifespan is the number of seconds from the cluster's creation time until the time
    // it expires. Therefore, calculating the "now" moment expressed as a lifetime is a bit tricky.
    const now = moment.duration(
      Date.now() - moment(modifiedCluster.CreatedOn).toDate().getTime(),
      'ms',
    );
    const minimumLifespan = now.clone().add(minimumClusterLifetime);

    const delta = moment.duration(1, notation as moment.DurationInputArg2);
    let update = incOrDec === 'inc' ? current.clone().add(delta) : current.clone().subtract(delta);
    // Make sure a user cannot accidentally expire a cluster by reducing the duration to zero or below by clicking
    // "-" on the days or hours counter. However, allow reducing the time to zero or 1m via subsequent clicks on "-" for
    // the seconds counter, this should have a sufficiently low accident rate.
    if (incOrDec === 'dec' && delta > minimumClusterLifetime && update < minimumLifespan) {
      update = minimumLifespan;
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
          setTimeout(() => setClientSideLifespan(''), 20000) as unknown as number,
        );
      })
      .catch((err) => {
        setError(err);
        setClientSideLifespan('');
      });
  };

  return <Lifespan cluster={modifiedCluster} onModify={onModify} />;
}
