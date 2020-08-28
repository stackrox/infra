import React, { useState, useEffect, ReactElement } from 'react';
import moment from 'moment';
import { PlusCircle, MinusCircle } from 'react-feather';
import { Tooltip, TooltipOverlay } from '@stackrox/ui-components';

function calcDuration(targetDate: Date): moment.Duration {
  return moment.duration(moment(targetDate).diff(moment()));
}

type Props = {
  targetDate: Date;
  className?: string;
  canModify: boolean;
  onModify?: (notation: string, incOrDec: string) => void;
};

export default function Countdown({
  targetDate,
  className = '',
  canModify,
  onModify,
}: Props): ReactElement {
  const [duration, setDuration] = useState<moment.Duration>(calcDuration(targetDate));

  useEffect(() => {
    const timer = setInterval(() => {
      setDuration(calcDuration(targetDate));
    }, 10000);
    return (): void => clearInterval(timer);
  }, [targetDate]);

  return (
    <Tooltip
      placement="left"
      content={<TooltipOverlay>{`Expiration: ${moment(targetDate).format('LLL')}`}</TooltipOverlay>}
    >
      <div className={className}>
        <FormatDuration duration={duration} canModify={canModify} onModify={onModify} />
      </div>
    </Tooltip>
  );
}

type FormatDurationProps = {
  duration: moment.Duration;
  canModify: boolean;
  onModify?: (notation: string, incOrDec: string) => void;
};

function FormatDuration({
  duration,
  canModify = true,
  onModify,
}: FormatDurationProps): ReactElement {
  // everything will be negative if it's a negative duration (i.e. expired)
  const expiredMultiplier = duration.asMilliseconds() <= 0 ? -1 : 1;
  if (expiredMultiplier * duration.asDays() > 30) {
    return expiredMultiplier > 0 ? (
      <span>More than 30 days remain</span>
    ) : (
      <span>Expired more than 30 days ago</span>
    );
  }

  const days = expiredMultiplier * duration.days();
  const hours = expiredMultiplier * duration.hours();
  const minutes = expiredMultiplier * duration.minutes();

  const timeStr = `${days > 0 ? `${days}d ` : ''}${hours}h ${minutes}m`;
  if (expiredMultiplier <= 0) {
    return <span>Expired {timeStr} ago</span>;
  }
  if (!canModify) {
    return <span>{timeStr} remains</span>;
  }

  return (
    <span>
      {days > 0 && <ModifiableTimeUnit value={days} notation="d" onChange={onModify} />}{' '}
      <ModifiableTimeUnit value={hours} notation="h" onChange={onModify} />{' '}
      <ModifiableTimeUnit value={minutes} notation="m" onChange={onModify} /> remains
    </span>
  );
}

type ModifiableTimeUnitProps = {
  notation: string;
  value: number;
  onChange?: (notation: string, incOrDec: string) => void;
};

function ModifiableTimeUnit({
  notation,
  value,
  onChange = (): void => {},
}: ModifiableTimeUnitProps): ReactElement {
  return (
    <span className="inline-flex flex-col items-center">
      <span>
        {value}
        {notation}
      </span>
      <span className="inline-flex text-sm normal-case">
        <PlusCircle className="mr-2" size={12} onClick={(): void => onChange(notation, 'inc')} />
        <MinusCircle size={12} onClick={(): void => onChange(notation, 'dec')} />
      </span>
    </span>
  );
}
