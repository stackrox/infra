import { FlavorServiceApi, V1Flavor } from 'generated/client';
import { QueryClient } from '@tanstack/react-query';
import configuration from './configuration';

const flavorService = new FlavorServiceApi(configuration);

export function flavorInfoQueryOptions(flavorId: string) {
  return {
    queryKey: ['flavorInfo', flavorId],
    queryFn: () => flavorService.info(flavorId),
    staleTime: 60 * 60 * 1000, // One hour - this info almost never changes
  };
}

export function prefetchFlavors(queryClient: QueryClient, flavors: V1Flavor[]) {
  flavors.forEach(async ({ ID }) => {
    await queryClient.prefetchQuery(flavorInfoQueryOptions(ID ?? ''));
  });
}
