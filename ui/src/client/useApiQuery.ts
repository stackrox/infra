import { useEffect, useState } from 'react';
import { AxiosError, AxiosPromise } from 'axios';

export interface DataFetcher<T> {
  (): AxiosPromise<T>;
}

export interface RequestState<T> {
  /** whether the fetching is in progress */
  loading: boolean;
  /** occurred error (if request failed) */
  error?: AxiosError<T>;
  /** returned data (if request succeeded) */
  data?: T;
}

/**
 * Takes out the boilerplate of handling loading and error handling, as well as extracting
 * data from Axios response.
 *
 * @template T
 * @param {DataFetcher<T>} fetcher callback that makes an API call.
 *   **Important**: fetcher instance should NOT be recreated on every component render
 * @returns {RequestState<T>} the state of the request
 */
export default function useApiQuery<T>(fetcher: DataFetcher<T>): RequestState<T> {
  // setting `loading: true` from the beginning as that the intention of the hook
  // to start making the request right away on component mounting through `useEffect`,
  // yet React hook execution model doesn't guarantee synchronous execution of `useEffect`.
  const [requestState, setRequestState] = useState<RequestState<T>>({ loading: true });
  useEffect(() => {
    setRequestState({ loading: true });

    let isCancelled = false;

    fetcher()
      .then((response) => {
        if (!isCancelled) {
          setRequestState({ loading: false, error: undefined, data: response.data });
        }
      })
      .catch((error) => {
        if (!isCancelled) {
          setRequestState({ loading: false, error, data: undefined });
        }
      });

    return (): void => {
      isCancelled = true;
    };
  }, [fetcher]);

  return requestState;
}
