import { useState } from 'react';
import { Button, Card, ConfigProvider, Form, Input, InputNumber, Layout, Modal, Select, Space, Table, Tag, Typography, message, Popconfirm, Switch } from 'antd';
import { PlusOutlined, CopyOutlined, DeleteOutlined, SafetyCertificateOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { HttpUtil } from '@/utils';
import AppSidebar from '@/layouts/AppSidebar';
import { useTheme } from '@/hooks/useTheme';

type Role = { id:number; name:string; slug:string; baseTier:string; builtIn:boolean; ownerRole:boolean; quotaGB:number; permissions:any; limits:any; features:any; access:any; adminCount:number; createdAt:number; updatedAt:number };

const TIER_OPTS = [
  {value:'viewer', label:'Viewer — read only'},
  {value:'creator', label:'Creator — create clients only'},
  {value:'editor', label:'Editor — create + edit'},
  {value:'admin', label:'Admin — full clients+inbounds'},
];
const TIER_COLOR:Record<string,string>={owner:'#8b5cf6', admin:'#06ffa5', editor:'#22d3ee', creator:'#a78bfa', viewer:'#64748b'};
const JSON_HEADERS = { headers: { 'Content-Type': 'application/json' } } as const;

async function fetchRoles(): Promise<Role[]>{
  const msg = await HttpUtil.get('/panel/api/roles/list', undefined, JSON_HEADERS) as any;
  if(!msg?.success) throw new Error(msg?.msg||'failed');
  return msg.obj as Role[];
}

export default function RolesPage(){
  const qc = useQueryClient();
  const { antdThemeConfig } = useTheme();
  const [msgApi, ctx]=message.useMessage();
  const {data, isLoading}=useQuery({queryKey:['roles'], queryFn: fetchRoles});
  const [open, setOpen]=useState(false);
  const [editing, setEditing]=useState<Role|null>(null);
  const [form]=Form.useForm();

  const saveMut = useMutation({
    mutationFn: async(v:any)=>{
      const payload = {
        name: v.name, baseTier: v.baseTier, quotaGB: Number(v.quotaGB)||0,
        permissions: v.permissions ? JSON.parse(v.permissions) : {},
        limits: v.limits ? JSON.parse(v.limits) : {},
        features: { blockLimitedAdmins: v.blockLimited!==false, disconnectUsersWhenLimited: v.disconnect!==false },
        access: v.access ? JSON.parse(v.access) : {},
      };
      const url = editing ? `/panel/api/roles/update/${editing.id}` : '/panel/api/roles/create';
      const r = await HttpUtil.post(url, payload, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=>{ msgApi.success(editing?'Role updated':'Role created'); qc.invalidateQueries({queryKey:['roles']}); setOpen(false); setEditing(null); form.resetFields(); },
    onError:(e:any)=> msgApi.error(e.message||'error'),
  });
  const dupMut = useMutation({
    mutationFn: async(id:number)=>{
      const r = await HttpUtil.post(`/panel/api/roles/duplicate/${id}`, {}, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=>{ msgApi.success('Duplicated'); qc.invalidateQueries({queryKey:['roles']}); },
    onError:(e:any)=> msgApi.error(e.message),
  });
  const delMut = useMutation({
    mutationFn: async(id:number)=>{
      const r = await HttpUtil.post(`/panel/api/roles/remove/${id}`, {}, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=>{ msgApi.success('Deleted'); qc.invalidateQueries({queryKey:['roles']}); },
    onError:(e:any)=> msgApi.error(e.message),
  });

  const openCreate = ()=>{ setEditing(null); form.resetFields(); form.setFieldsValue({baseTier:'viewer', quotaGB:0, blockLimited:true, disconnect:true}); setOpen(true); };
  const openEdit = (row:Role)=>{
    setEditing(row);
    form.setFieldsValue({
      name: row.builtIn && !row.ownerRole ? undefined : row.name,
      baseTier: row.baseTier, quotaGB: row.quotaGB||0,
      blockLimited: (row.features as any)?.blockLimitedAdmins!==false,
      disconnect: (row.features as any)?.disconnectUsersWhenLimited!==false,
      permissions: JSON.stringify(row.permissions||{}),
      limits: JSON.stringify(row.limits||{}),
      access: JSON.stringify(row.access||{}),
    });
    setOpen(true);
  };

  const cols:any[] = [
    {title:'Role', dataIndex:'name', render:(v:string,row:Role)=><Space><span style={{width:32,height:32,borderRadius:10,display:'grid',placeItems:'center',background:'rgba(139,92,246,0.12)',border:'1px solid rgba(139,92,246,0.3)',color:'#c4b5fd'}}><SafetyCertificateOutlined/></span><span><b style={{color:'#e2e8f0'}}>{v}</b><br/><span style={{fontSize:12,color:'#64748b'}}>{row.slug}{row.builtIn?' · built-in':''}{row.ownerRole?' · owner':''}</span></span></Space>},
    {title:'Base access', dataIndex:'baseTier', width:170, render:(v:string)=><Tag style={{borderRadius:999, background:'rgba(255,255,255,0.05)', border:`1px solid ${TIER_COLOR[v]||'#334155'}`, color:TIER_COLOR[v]||'#94a3b8'}}>{v}</Tag>},
    {title:'Default quota', dataIndex:'quotaGB', width:120, render:(v:number)=> v? `${v} GB` : <span style={{color:'#64748b'}}>∞</span>},
    {title:'Admins', dataIndex:'adminCount', width:80, render:(v:number)=><b style={{color:'#e2e8f0'}}>{v}</b>},
    {title:'Actions', width:220, render:(_:any,row:Role)=>(
      <Space>
        <Button size="small" ghost onClick={()=>openEdit(row)}>Edit</Button>
        {!row.ownerRole && <Button size="small" ghost icon={<CopyOutlined/>} onClick={()=>dupMut.mutate(row.id)}>Copy</Button>}
        {!row.builtIn && !row.ownerRole && (
          <Popconfirm title="Delete role?" description={row.name} okText="Delete" cancelText="Cancel" onConfirm={()=>delMut.mutate(row.id)}>
            <Button size="small" danger ghost icon={<DeleteOutlined/>} />
          </Popconfirm>
        )}
      </Space>
    )},
  ];

  return (
    <ConfigProvider theme={antdThemeConfig}>
      {ctx}
      <Layout style={{minHeight:'100vh'}}>
        <AppSidebar />
        <Layout className="content-shell">
          <Layout.Content id="content-layout" className="content-area" style={{padding:0, background:'transparent'}}>
            <div style={{padding:24, maxWidth:1100, margin:'0 auto', width:'100%'}}>
              <div style={{borderRadius:20, padding:'20px 22px', marginBottom:16, background:'linear-gradient(135deg, #0a0e1a 0%, #1a1240 45%, #0f2a2a 100%)', border:'1px solid rgba(139,92,246,0.18)', boxShadow:'0 8px 32px rgba(0,0,0,0.35)', position:'relative', overflow:'hidden'}}>
                <div style={{position:'absolute', inset:0, background:'radial-gradient(600px 220px at 20% 0%, rgba(139,92,246,0.18), transparent 60%), radial-gradient(500px 220px at 85% 100%, rgba(6,255,165,0.12), transparent 60%)', pointerEvents:'none'}}/>
                <div style={{position:'relative', display:'flex', alignItems:'center', justifyContent:'space-between', gap:16, flexWrap:'wrap'}}>
                  <div style={{display:'flex', gap:14, alignItems:'center'}}>
                    <div style={{width:46, height:46, borderRadius:14, display:'grid', placeItems:'center', background:'linear-gradient(135deg,#8b5cf6,#06ffa5)', color:'#fff', fontSize:20}}><SafetyCertificateOutlined/></div>
                    <div>
                      <Typography.Title level={4} style={{margin:0, color:'#fff'}}>Roles <span style={{color:'#64748b', fontWeight:500}}>— NEON X</span></Typography.Title>
                      <div style={{color:'#94a3b8', fontSize:13, marginTop:4}}>Custom roles inherit a base tier · quota default applies to new admins</div>
                    </div>
                  </div>
                  <Button type="primary" icon={<PlusOutlined/>} onClick={openCreate} style={{height:40, borderRadius:999, background:'linear-gradient(135deg,#7c3aed,#06ffa5)', border:'none'}}>New Role</Button>
                </div>
              </div>
              <Card style={{borderRadius:16, overflow:'hidden', border:'1px solid rgba(139,92,246,0.12)'}} bodyStyle={{padding:0}}>
                <Table rowKey="id" loading={isLoading} dataSource={data||[]} columns={cols} pagination={false} size="middle" scroll={{x:800}} />
              </Card>
            </div>
          </Layout.Content>
        </Layout>
      </Layout>
      <Modal title={editing?`Edit Role — ${editing.name}`:'New Role'} open={open} onCancel={()=>{setOpen(false); setEditing(null);}} footer={null} destroyOnClose width={560}>
        <Form form={form} layout="vertical" onFinish={(v)=>saveMut.mutate(v)} style={{marginTop:12}}>
          {(!editing || (!editing.builtIn && !editing.ownerRole)) && (
            <Form.Item name="name" label="Role name" rules={[{required:true, message:'required'}]}><Input placeholder="e.g. Support L1" /></Form.Item>
          )}
          <Form.Item name="baseTier" label="Base access tier" rules={[{required:true}]}><Select options={TIER_OPTS} /></Form.Item>
          <Form.Item name="quotaGB" label="Default quota (GB)" extra="0 = unlimited — applied to new admins of this role" initialValue={0}><InputNumber min={0} addonAfter="GB" style={{width:'100%'}} /></Form.Item>
          <Space style={{display:'flex', marginBottom:16}}>
            <Form.Item name="blockLimited" label="Disable admin at quota" valuePropName="checked" style={{margin:0}}><Switch/></Form.Item>
            <Form.Item name="disconnect" label="Disable clients at quota" valuePropName="checked" style={{margin:0}}><Switch/></Form.Item>
          </Space>
          <Button type="primary" htmlType="submit" loading={saveMut.isPending} block size="large" style={{borderRadius:12, background:'linear-gradient(135deg,#7c3aed,#06ffa5)', border:'none'}}>{editing?'Save':'Create Role'}</Button>
        </Form>
      </Modal>
    </ConfigProvider>
  );
}

