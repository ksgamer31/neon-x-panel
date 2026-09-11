import { useState } from 'react';
import { Badge, Button, Card, ConfigProvider, Form, Input, Layout, Modal, Select, Space, Switch, Table, Tag, Typography, message, Popconfirm, Tooltip } from 'antd';
import { UserAddOutlined, SafetyCertificateOutlined, EyeOutlined, TeamOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { HttpUtil } from '@/utils';
import AppSidebar from '@/layouts/AppSidebar';
import { useTheme } from '@/hooks/useTheme';

type AdminUser = { id:number; username:string; role:string; enabled:boolean; displayName:string; inboundIds:string };

const ROLE_OPTS = [
  {value:'viewer',  label:'Viewer — read only'},
  {value:'creator', label:'Creator — create clients only'},
  {value:'editor',  label:'Editor — create + edit'},
  {value:'admin',   label:'Admin — full'},
  {value:'owner',   label:'Owner — full + manage owners'},
];
const ROLE_COLOR:Record<string,string>={owner:'#8b5cf6', admin:'#06ffa5', editor:'#22d3ee', creator:'#a78bfa', viewer:'#64748b'};
const ROLE_BG:Record<string,string>={owner:'rgba(139,92,246,0.15)', admin:'rgba(6,255,165,0.14)', editor:'rgba(34,211,238,0.14)', creator:'rgba(167,139,250,0.14)', viewer:'rgba(100,116,139,0.14)'};

const JSON_HEADERS = { headers: { 'Content-Type': 'application/json' } } as const;

async function fetchAdmins(): Promise<AdminUser[]>{
  const msg = await HttpUtil.post('/panel/api/users/list', {}, JSON_HEADERS) as any;
  if(!msg?.success) throw new Error(msg?.msg||'failed');
  return msg.obj as AdminUser[];
}

export default function AdminsPage(){
  const qc = useQueryClient();
  const { antdThemeConfig } = useTheme();
  const [msgApi, ctx]=message.useMessage();
  const {data, isLoading}=useQuery({queryKey:['admins'], queryFn: fetchAdmins});
  const [open, setOpen]=useState(false);
  const [form]=Form.useForm();

  const createMut = useMutation({
    mutationFn: async(v:any)=>{
      const r = await HttpUtil.post('/panel/api/users/create', v, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=>{ msgApi.success('Admin created'); qc.invalidateQueries({queryKey:['admins']}); setOpen(false); form.resetFields(); },
    onError:(e:any)=> msgApi.error(e.message||'error'),
  });
  const deleteMut = useMutation({
    mutationFn: async(id:number)=>{
      const r = await HttpUtil.post(`/panel/api/users/delete/${id}`, {}, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=>{ msgApi.success('Deleted'); qc.invalidateQueries({queryKey:['admins']}); },
    onError:(e:any)=> msgApi.error(e.message),
  });
  const toggleMut = useMutation({
    mutationFn: async(row:AdminUser)=>{
      const r = await HttpUtil.post(`/panel/api/users/update/${row.id}`, {enabled: !row.enabled}, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=> qc.invalidateQueries({queryKey:['admins']}),
    onError:(e:any)=> msgApi.error(e.message),
  });

  const total = data?.length ?? 0;
  const active = data?.filter(u=>u.enabled).length ?? 0;

  const cols:any[] = [
    {title:'#', dataIndex:'id', width:64, render:(v:number)=><span style={{color:'#94a3b8', fontWeight:600}}>#{v}</span>},
    {title:'User', dataIndex:'username', render:(v:string, row:AdminUser)=><Space><span style={{width:32,height:32,borderRadius:10,display:'grid',placeItems:'center',background:ROLE_BG[row.role]||'#1e293b',border:`1px solid ${ROLE_COLOR[row.role]||'#334155'}`,color:ROLE_COLOR[row.role]}}><TeamOutlined/></span><span><b style={{color:'#e2e8f0'}}>{v}</b><br/><span style={{fontSize:12,color:'#94a3b8'}}>{row.displayName||'—'}</span></span></Space>},
    {title:'Role', dataIndex:'role', width:140, render:(v:string)=><Tag style={{borderRadius:999, padding:'2px 10px', fontWeight:700, background:ROLE_BG[v]||'#1e293b', color:ROLE_COLOR[v]||'#94a3b8', border:`1px solid ${ROLE_COLOR[v]}40`}}>{v}</Tag>},
    {title:'Status', dataIndex:'enabled', width:110, render:(v:boolean,row:AdminUser)=><Switch checked={v} loading={toggleMut.isPending} onChange={()=>toggleMut.mutate(row)} checkedChildren="active" unCheckedChildren="disabled" style={v?{background:'#06ffa5'}:undefined} />},
    {title:'Inbound Access', dataIndex:'inboundIds', render:(v:string)=> v ? <Tooltip title={v}><Tag style={{maxWidth:160,overflow:'hidden',textOverflow:'ellipsis',background:'rgba(139,92,246,0.12)',border:'1px solid rgba(139,92,246,0.25)',color:'#c4b5fd'}}>{v}</Tag></Tooltip> : <Tag style={{borderRadius:999, background:'rgba(6,255,165,0.12)', color:'#6ee7b7', border:'1px solid rgba(6,255,165,0.22)'}}>all</Tag>},
    {title:'Actions', width:110, render:(_:any,row:AdminUser)=>(
      <Popconfirm title="Delete this admin?" description={row.username} okText="Delete" cancelText="Cancel" onConfirm={()=>deleteMut.mutate(row.id)}>
        <Button size="small" danger ghost style={{borderRadius:999}}>Delete</Button>
      </Popconfirm>
    )},
  ];

  return (
    <ConfigProvider theme={antdThemeConfig}>
      {ctx}
      <Layout style={{minHeight:'100vh'}}>
        <AppSidebar />
        <Layout className="content-shell">
          <Layout.Content id="content-layout" className="content-area" style={{padding:0, background:'transparent'}}>
            <div className="admins-page" style={{padding:24, maxWidth:1200, margin:'0 auto', width:'100%'}}>
              {/* Hero */}
              <div style={{borderRadius:20, padding:'20px 22px', marginBottom:16, background:'linear-gradient(135deg, #0a0e1a 0%, #1a1240 45%, #0f2a2a 100%)', border:'1px solid rgba(139,92,246,0.18)', boxShadow:'0 8px 32px rgba(0,0,0,0.35), 0 0 0 1px rgba(139,92,246,0.08), inset 0 1px 0 rgba(255,255,255,0.06)', position:'relative', overflow:'hidden'}}>
                <div style={{position:'absolute', inset:0, background:'radial-gradient(600px 220px at 20% 0%, rgba(139,92,246,0.18), transparent 60%), radial-gradient(500px 220px at 85% 100%, rgba(6,255,165,0.12), transparent 60%)', pointerEvents:'none'}}/>
                <div style={{position:'relative', display:'flex', alignItems:'center', justifyContent:'space-between', gap:16, flexWrap:'wrap'}}>
                  <div style={{display:'flex', gap:14, alignItems:'center'}}>
                    <div style={{width:46, height:46, borderRadius:14, display:'grid', placeItems:'center', background:'linear-gradient(135deg,#8b5cf6,#06ffa5)', boxShadow:'0 6px 20px rgba(139,92,246,0.35)', color:'#fff', fontSize:20}}><SafetyCertificateOutlined/></div>
                    <div>
                      <Typography.Title level={4} style={{margin:0, color:'#fff', lineHeight:1.1}}>
                        <span style={{background:'linear-gradient(135deg,#a78bfa 0%, #5eead4 100%)', WebkitBackgroundClip:'text', WebkitTextFillColor:'transparent'}}>Admin</span>
                        <span style={{color:'#64748b', fontWeight:500}}> — NEON X</span>
                      </Typography.Title>
                      <div style={{color:'#94a3b8', fontSize:13, marginTop:4}}>
                        <b style={{color:'#c4b5fd'}}>Viewer</b> read only · <b style={{color:'#a78bfa'}}>Creator</b> create only · <b style={{color:'#22d3ee'}}>Editor</b> create+edit · <b style={{color:'#6ee7b7'}}>Admin/Owner</b> full
                      </div>
                    </div>
                  </div>
                  <Space>
                    <span style={{display:'inline-flex', alignItems:'center', gap:8, padding:'8px 12px', borderRadius:999, background:'rgba(255,255,255,0.06)', border:'1px solid rgba(255,255,255,0.08)', color:'#e2e8f0'}}>
                      <Badge status="processing" color="#06ffa5" /> active: <b>{active}</b> / {total}
                    </span>
                    <Button type="primary" icon={<UserAddOutlined/>} onClick={()=>setOpen(true)} style={{height:40, borderRadius:999, padding:'0 18px', background:'linear-gradient(135deg,#7c3aed,#06ffa5)', border:'none', boxShadow:'0 6px 20px rgba(139,92,246,0.35)'}}>New Admin</Button>
                  </Space>
                </div>
                <div style={{position:'relative', display:'flex', gap:8, marginTop:14, flexWrap:'wrap'}}>
                  <span style={{fontSize:12, padding:'6px 10px', borderRadius:999, background:'rgba(139,92,246,0.12)', border:'1px solid rgba(139,92,246,0.22)', color:'#ddd6fe'}}><EyeOutlined/> limit by <code style={{background:'rgba(0,0,0,0.25)', padding:'1px 6px', borderRadius:6}}>inboundIds: [1,2]</code> — empty = all</span>
                  <span style={{fontSize:12, padding:'6px 10px', borderRadius:999, background:'rgba(6,255,165,0.10)', border:'1px solid rgba(6,255,165,0.20)', color:'#a7f3d0'}}>path: <code style={{background:'rgba(0,0,0,0.25)', padding:'1px 6px', borderRadius:6}}>/panel/admins</code></span>
                </div>
              </div>

              <Card style={{borderRadius:16, overflow:'hidden', border:'1px solid rgba(139,92,246,0.12)', boxShadow:'0 8px 32px rgba(0,0,0,0.25)'}} bodyStyle={{padding:0}}>
                <Table rowKey="id" loading={isLoading} dataSource={data||[]} columns={cols} pagination={false} size="middle" />
              </Card>
              <div style={{textAlign:'center', color:'#64748b', fontSize:12, marginTop:10}}>Only <b style={{color:'#c4b5fd'}}>owner</b> can create/delete owner/admin · change role from this table</div>
            </div>
          </Layout.Content>
        </Layout>
      </Layout>

      <Modal title={<span style={{display:'flex',gap:8,alignItems:'center'}}><span style={{width:28,height:28,borderRadius:8,display:'grid',placeItems:'center',background:'linear-gradient(135deg,#8b5cf6,#06ffa5)',color:'#fff'}}><UserAddOutlined/></span> New Admin</span>} open={open} onCancel={()=>setOpen(false)} footer={null} destroyOnClose width={520}>
        <Form form={form} layout="vertical" onFinish={(v)=>createMut.mutate(v)} style={{marginTop:12}}>
          <Form.Item name="username" label="Username" rules={[{required:true, message:'required'}]}><Input placeholder="e.g. sepehr" size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="password" label="Password" rules={[{required:true, message:'required'}]}><Input.Password placeholder="••••••••" size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="displayName" label="Display name"><Input placeholder="Support" size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="role" label="Role" initialValue="viewer"><Select options={ROLE_OPTS} size="large" /></Form.Item>
          <Form.Item name="inboundIds" label="Allowed inbounds" extra="empty = all · e.g. [1,3]"><Input placeholder='[] or [1,2]' size="large" style={{borderRadius:12}} /></Form.Item>
          <Button type="primary" htmlType="submit" loading={createMut.isPending} block size="large" style={{borderRadius:12, height:44, background:'linear-gradient(135deg,#7c3aed,#06ffa5)', border:'none'}}>Create Admin</Button>
        </Form>
      </Modal>
    </ConfigProvider>
  );
}
