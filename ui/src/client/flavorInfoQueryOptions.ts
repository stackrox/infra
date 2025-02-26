import { FlavorServiceApi, V1Flavor } from 'generated/client';
import configuration from './configuration';
import { QueryClient } from '@tanstack/react-query';

const flavorService = new FlavorServiceApi(configuration);

export function flavorInfoQueryOptions(flavorId: string) {
  return {
    queryKey: ['flavorInfo', flavorId],
    queryFn: () => flavorService.info(flavorId),
    staleTime: 10 * 60 * 1000,
  };
}

export async function prefetchFlavors(queryClient: QueryClient, flavors: V1Flavor[]) {
  for (const { ID } of flavors) {
    await queryClient.prefetchQuery(flavorInfoQueryOptions(ID ?? ''));
  }
}
