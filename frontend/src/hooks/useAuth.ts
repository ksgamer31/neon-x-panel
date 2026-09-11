import { useQuery } from '@tanstack/react-query';
import { HttpUtil } from '@/utils';

export type PanelRole = 'owner' | 'admin' | 'editor' | 'creator' | 'viewer' | '';

export interface Me {
  id: number;
  username: string;
  role: PanelRole;
  enabled: boolean;
  displayName: string;
  inboundIds: string;
}

export function useAuth() {
  const q = useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      const msg = await HttpUtil.get<{ id:number; username:string; role:string; enabled:boolean; displayName:string; inboundIds:string }>('/panel/api/users/me', undefined, { silent:true });
      if (msg?.success && msg.obj) {
        const o = msg.obj as any;
        return { id: o.id, username: o.username, role: (o.role || 'owner') as PanelRole, enabled: !!o.enabled, displayName: o.displayName||'', inboundIds: o.inboundIds||'' } as Me;
      }
      return null as Me | null;
    },
    staleTime: 30000,
    retry: false,
  });
  const role: PanelRole = (q.data?.role as PanelRole) || 'owner';
  const isOwner = role === 'owner' || role === '';
  const canCreateClient = role === 'owner' || role === 'admin' || role === 'editor' || role === 'creator';
  const canEditClient = role === 'owner' || role === 'admin' || role === 'editor';
  const canDeleteClient = role === 'owner' || role === 'admin';
  const canCreateInbound = role === 'owner' || role === 'admin' || role === 'editor';
  const canEditInbound = role === 'owner' || role === 'admin' || role === 'editor';
  const canDeleteInbound = role === 'owner' || role === 'admin';
  const canAccessSettings = isOwner;
  const canAccessNodes = isOwner;
  const canAccessHosts = isOwner;
  const canAccessXray = isOwner;
  const canManageAdmins = isOwner || role === 'admin';
  const showInbounds = role !== 'creator';
  const readOnly = role === 'viewer';
  return { me: q.data, role, isOwner, loading: q.isLoading, canCreateClient, canEditClient, canDeleteClient, canCreateInbound, canEditInbound, canDeleteInbound, canAccessSettings, canAccessNodes, canAccessHosts, canAccessXray, canManageAdmins, showInbounds, readOnly };
}
