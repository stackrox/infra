import React, { ReactElement } from 'react';
import { Button } from '@patternfly/react-core';
import { MinusCircleIcon, PlusCircleIcon } from '@patternfly/react-icons';
import moment from 'moment';

function calcDuration(targetDate: Date): moment.Duration {
  return moment.duration(moment(targetDate).diff(moment()));
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
    <span className="pf-v6-u-display-inline-flex pf-v6-u-flex-direction-column pf-v6-u-align-items-center">
      <span className="pf-v6-u-font-weight-bold">
        {`${value}`.padStart(2, '0')}
        {notation}
      </span>
      <span className="pf-v6-u-display-inline-flex">
        <Button
          icon={<PlusCircleIcon />}
          size="sm"
          variant="control"
          className="pf-v6-u-p-0 pf-v6-u-mr-xs"
          aria-label="Increment"
          onClick={(): void => {
            onChange(notation, 'inc');
          }}
        />
        <Button
          icon={<MinusCircleIcon />}
          size="sm"
          variant="control"
          className="pf-v6-u-p-0"
          aria-label="Decrement"
          onClick={(): void => {
            onChange(notation, 'dec');
          }}
        />
      </span>
    </span>
  );
}

type FormatDurationProps = {
  duration: moment.Duration;
  onModify?: (notation: string, incOrDec: string) => void;
};

function FormatDuration({ duration, onModify }: FormatDurationProps): ReactElement {
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
  if (!onModify) {
    return <span>{timeStr} remains</span>;
  }

  return (
    <span>
      <ModifiableTimeUnit value={days} notation="d" onChange={onModify} />{' '}
      <ModifiableTimeUnit value={hours} notation="h" onChange={onModify} />{' '}
      <ModifiableTimeUnit value={minutes} notation="m" onChange={onModify} /> remains
    </span>
  );
}

type CountdownProps = {
  targetDate: Date;
  onModify?: (notation: string, incOrDec: string) => void;
};

export default function Countdown({ targetDate, onModify }: CountdownProps): ReactElement {
  const duration = calcDuration(targetDate);

  return (
    <div>
      {`Expiration: ${moment(targetDate).format('LLL')}`}
      <span>&nbsp;-&nbsp;</span>
      <FormatDuration duration={duration} onModify={onModify} />
    </div>
  );
}
