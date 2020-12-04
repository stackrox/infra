import { useCallback, useState, useEffect, useRef } from 'react';
import { AxiosError } from 'axios';

import { ApiCaller, RequestState } from './useApiQuery';

/**
 * React hook that helps to deal with API requests that are typically invoked on a user action
 * (like updating an API resource on user saving a state). Opposite to `useApiQuery`, which sends
 * the fetching query right away on component being mounted, this hook returns a callback function
 * that can be passed as a handler to `onClick` or be called from inside the corresponding user
 * action handler.
 *
 * @template T
 * @param {ApiCaller<T>} requester callback that makes an API request to perform an operation
 * @returns {[operation, RequestState<T>]} callback and the state of the request
 */
export default function useApiOperation<T>(requester: ApiCaller<T>): [() => void, RequestState<T>] {
  const isMounted = useRef(true);
  const [requestState, setRequestState] = useState<RequestState<T>>({
    called: false,
    loading: false,
  });

  useEffect(() => {
    return (): void => {
      isMounted.current = false;
    };
  }, [isMounted]);

  const operation = useCallback(() => {
    setRequestState({ called: true, loading: true });

    requester()
      .then((response) => {
        if (isMounted.current) {
          setRequestState({ called: true, loading: false, error: undefined, data: response.data });
        }
      })
      .catch((error: AxiosError<T>) => {
        if (isMounted.current) {
          setRequestState({ called: true, loading: false, error, data: undefined });
        }
      });
  }, [requester]);

  return [operation, requestState];
}
