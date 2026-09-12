import { useQuery } from '@tanstack/react-query';
import { useMemo } from 'react';

import { HttpUtil } from '@/utils';
import { parseMsg } from '@/utils/zodValidate';
import { Status } from '@/models/status';
import { StatusSchema } from '@/schemas/status';
import { keys } from '@/api/queryKeys';

const POLL_INTERVAL_MS = 2000;

async function fetchStatus(): Promise<Status> {
  const msg = await HttpUtil.get('/panel/api/server/status', undefined, { silent: true });
  if (!msg?.success) throw new Error(msg?.msg || 'Failed to fetch status');
  const validated = parseMsg(msg, StatusSchema, 'server/status');
  return new Status(validated.obj);
}

export function useStatusQuery() {
  const query = useQuery({
    queryKey: keys.server.status(),
    queryFn: fetchStatus,
    refetchInterval: POLL_INTERVAL_MS,
    refetchIntervalInBackground: false,
    staleTime: 0,
    retry: 3,
    retryDelay: (attempt) => Math.min(1000 * 2 ** attempt, 8000),
  });

  const status = useMemo(() => query.data ?? new Status(), [query.data]);
  const refresh = async () => {
    await query.refetch();
  };

  const hasData = query.data !== undefined;
  return {
    status,
    hasData,
    fetched: hasData || query.isError,
    // Only surface the error when we have NO data at all (first load).
    // On refetch failure (e.g. refresh click, xray restarting) keep showing
    // the last good dashboard instead of dropping to the error page.
    fetchError: !hasData && query.error ? (query.error as Error).message : '',
    refresh,
  };
}
