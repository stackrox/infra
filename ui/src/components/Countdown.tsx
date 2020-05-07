import React, { useState, useEffect, ReactElement } from 'react';
import moment from 'moment';

import Tooltip from 'components/Tooltip';
import TooltipOverlay from 'components/TooltipOverlay';

function calcDuration(targetDate: Date): moment.Duration {
  return moment.duration(moment(targetDate).diff(moment()));
}

function formatDuration(duration: moment.Duration): string {
  // everything will be negative if it's a negative duration (i.e. expired)
  const expiredMultiplier = duration.asMilliseconds() <= 0 ? -1 : 1;
  if (expiredMultiplier * duration.asDays() > 30) {
    return expiredMultiplier > 0 ? 'More than 30 days remain' : 'Expired more than 30 days ago';
  }

  const days = expiredMultiplier * duration.days();
  const hours = expiredMultiplier * duration.hours();
  const minutes = expiredMultiplier * duration.minutes();

  const timeStr = `${days > 0 ? `${days}d ` : ''}${hours}h ${minutes}m`;
  return expiredMultiplier > 0 ? `${timeStr} remain` : `Expired ${timeStr} ago`;
}

type Props = {
  targetDate: Date;
  className?: string;
};

export default function Countdown({ targetDate, className = '' }: Props): ReactElement {
  const [duration, setDuration] = useState<moment.Duration>(calcDuration(targetDate));

  useEffect(() => {
    const timer = setInterval(() => {
      setDuration(calcDuration(targetDate));
    }, 10000);
    return (): void => clearInterval(timer);
  }, [targetDate]);

  return (
    <Tooltip
      content={<TooltipOverlay>{`Expiration: ${moment(targetDate).format('LLL')}`}</TooltipOverlay>}
    >
      <div className={className}>{formatDuration(duration)}</div>
    </Tooltip>
  );
}
