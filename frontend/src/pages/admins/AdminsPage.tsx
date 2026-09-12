import { useState } from 'react';
import { Badge, Button, Card, Col, ConfigProvider, Form, Input, InputNumber, Layout, Modal, Progress, Row, Select, Space, Statistic, Switch, Table, Tag, Typography, message, Popconfirm, Tooltip, Dropdown } from 'antd';
import { UserAddOutlined, SafetyCertificateOutlined, EyeOutlined, TeamOutlined, CloudOutlined, MoreOutlined, LinkOutlined, BellOutlined, FileTextOutlined } from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { HttpUtil } from '@/utils';
import AppSidebar from '@/layouts/AppSidebar';
import { useTheme } from '@/hooks/useTheme';
import { useNavigate } from 'react-router';

type AdminUser = {
  id:number; username:string; role:string; roleId:number; roleName:string; enabled:boolean;
  displayName:string; inboundIds:string; quotaGB:number; quotaUsed:number; quotaPct:number;
  totalClients:number; telegramId:string; supportUrl:string; profileTitle:string; subDomain:string; note:string;
  createdAt:number; updatedAt:number;
};
type Role = { id:number; name:string; slug:string; baseTier:string; builtIn:boolean; ownerRole:boolean; quotaGB:number; adminCount:number };

const TIER_COLOR:Record<string,string>={owner:'#8b5cf6', admin:'#06ffa5', editor:'#22d3ee', creator:'#a78bfa', viewer:'#64748b'};
const TIER_BG:Record<string,string>={owner:'rgba(139,92,246,0.15)', admin:'rgba(6,255,165,0.14)', editor:'rgba(34,211,238,0.14)', creator:'rgba(167,139,250,0.14)', viewer:'rgba(100,116,139,0.14)'};
const JSON_HEADERS = { headers: { 'Content-Type': 'application/json' } } as const;

function fmtGB(bytes:number){
  if(!bytes) return '0';
  return (bytes / (1024*1024*1024)).toFixed(2);
}

async function fetchAdmins(): Promise<AdminUser[]>{
  const msg = await HttpUtil.post('/panel/api/users/list', {}, JSON_HEADERS) as any;
  if(!msg?.success) throw new Error(msg?.msg||'failed');
  return (msg.obj as any[]).map((u:any)=> ({
    id: u.id, username: u.username, role: u.role, roleId: u.roleId||0, roleName: u.roleName||u.role,
    enabled: u.enabled, displayName: u.displayName||'', inboundIds: u.inboundIds||'',
    quotaGB: Number(u.quotaGB||0), quotaUsed: Number(u.quotaUsed||0), quotaPct: Number(u.quotaPct||0),
    totalClients: Number(u.totalClients||0), telegramId: u.telegramId||'', supportUrl: u.supportUrl||'',
    profileTitle: u.profileTitle||'', subDomain: u.subDomain||'', note: u.note||'',
    createdAt: u.createdAt||0, updatedAt: u.updatedAt||0,
  }));
}
async function fetchRoles(): Promise<Role[]>{
  try {
    const msg = await HttpUtil.get('/panel/api/roles/list', undefined, JSON_HEADERS) as any;
    if(msg?.success) return msg.obj as Role[];
  } catch { /* non-manager */ }
  return [];
}
async function fetchStats(): Promise<any>{
  try {
    const msg = await HttpUtil.get('/panel/api/users/stats', undefined, JSON_HEADERS) as any;
    if(msg?.success) return msg.obj;
  } catch { /* ignore */ }
  return null;
}

export default function AdminsPage(){
  const qc = useQueryClient();
  const navigate = useNavigate();
  const { antdThemeConfig } = useTheme();
  const [msgApi, ctx]=message.useMessage();
  const {data, isLoading}=useQuery({queryKey:['admins'], queryFn: fetchAdmins});
  const {data: roles}=useQuery({queryKey:['roles'], queryFn: fetchRoles});
  const {data: stats}=useQuery({queryKey:['admin-stats'], queryFn: fetchStats});
  const [open, setOpen]=useState(false);
  const [editing, setEditing]=useState<AdminUser|null>(null);
  const [form]=Form.useForm();

  const invalidate = ()=>{ qc.invalidateQueries({queryKey:['admins']}); qc.invalidateQueries({queryKey:['admin-stats']}); };

  const saveMut = useMutation({
    mutationFn: async(v:any)=>{
      const payload:any = { ...v };
      if(payload.quotaGB==null) payload.quotaGB=0;
      payload.quotaGB = Number(payload.quotaGB)||0;
      payload.roleId = Number(payload.roleId)||0;
      if(!payload.roleId) payload.role = v.tier;
      const url = editing ? `/panel/api/users/update/${editing.id}` : '/panel/api/users/create';
      const r = await HttpUtil.post(url, payload, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=>{ msgApi.success(editing?'Admin updated':'Admin created'); invalidate(); setOpen(false); setEditing(null); form.resetFields(); },
    onError:(e:any)=> msgApi.error(e.message||'error'),
  });
  const deleteMut = useMutation({
    mutationFn: async(id:number)=>{
      const r = await HttpUtil.post(`/panel/api/users/delete/${id}`, {}, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=>{ msgApi.success('Deleted'); invalidate(); },
    onError:(e:any)=> msgApi.error(e.message),
  });
  const toggleMut = useMutation({
    mutationFn: async(row:AdminUser)=>{
      const r = await HttpUtil.post(`/panel/api/users/update/${row.id}`, {enabled: !row.enabled}, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=> invalidate(),
    onError:(e:any)=> msgApi.error(e.message),
  });
  const resetUsageMut = useMutation({
    mutationFn: async(id:number)=>{
      const r = await HttpUtil.post(`/panel/api/users/resetUsage/${id}`, {}, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return r;
    },
    onSuccess:()=>{ msgApi.success('Usage reset'); invalidate(); },
    onError:(e:any)=> msgApi.error(e.message),
  });
  const clientsMut = useMutation({
    mutationFn: async({id, act}:{id:number; act:string})=>{
      const r = await HttpUtil.post(`/panel/api/users/clients/${act}/${id}`, {}, JSON_HEADERS) as any;
      if(!r?.success) throw new Error(r?.msg||'failed');
      return {act, count: (r.obj as any)?.count ?? 0};
    },
    onSuccess:({act, count})=>{ msgApi.success(`${act}: ${count} clients`); invalidate(); },
    onError:(e:any)=> msgApi.error(e.message),
  });

  const openCreate = ()=>{
    setEditing(null); form.resetFields();
    const firstRole = roles?.[0];
    form.setFieldsValue({tier:'viewer', roleId: firstRole?.id, quotaGB: firstRole?.quotaGB||0});
    setOpen(true);
  };
  const openEdit = (row:AdminUser)=>{
    setEditing(row);
    form.setFieldsValue({
      username: row.username, displayName: row.displayName, tier: row.role,
      roleId: row.roleId||undefined, inboundIds: row.inboundIds, quotaGB: row.quotaGB,
      telegramId: row.telegramId, supportUrl: row.supportUrl, profileTitle: row.profileTitle,
      subDomain: row.subDomain, note: row.note,
    });
    setOpen(true);
  };

  const total = stats?.totalAdmins ?? data?.length ?? 0;
  const active = stats?.activeAdmins ?? data?.filter(u=>u.enabled).length ?? 0;
  const disabled = stats?.disabledAdmins ?? (total - active);
  const limited = stats?.limitedAdmins ?? 0;

  const cols:any[] = [
    {title:'#', dataIndex:'id', width:52, render:(v:number)=><span style={{color:'#94a3b8', fontWeight:600}}>#{v}</span>},
    {title:'Admin', dataIndex:'username', width:200, render:(v:string, row:AdminUser)=>(
      <Space>
        <span style={{width:32,height:32,borderRadius:10,display:'grid',placeItems:'center',background:TIER_BG[row.role]||'#1e293b',border:`1px solid ${TIER_COLOR[row.role]||'#334155'}`,color:TIER_COLOR[row.role]}}><TeamOutlined/></span>
        <span><b style={{color: row.enabled?'#e2e8f0':'#64748b'}}>{v}</b><br/>
        <span style={{fontSize:12,color:'#94a3b8'}}>{row.profileTitle||row.displayName||'—'}{row.totalClients?` · ${row.totalClients} clients`:''}</span></span>
      </Space>
    )},
    {title:'Role', dataIndex:'roleId', width:170, render:(v:number, row:AdminUser)=>{
      const opts = (roles||[]).map(r=>({value:r.id, label:`${r.name} (${r.baseTier})`}));
      if(!opts.length) return <Tag style={{borderRadius:999}}>{row.roleName}</Tag>;
      return <Select value={v||undefined} placeholder={row.roleName} onChange={(roleId)=>saveMut.mutate({roleId} as any)} options={opts} size="small" style={{width:'100%'}} disabled={!!editing && false} />;
    }},
    {title:'Quota', width:190, render:(_:any,row:AdminUser)=>{
      const isUnlimited = !row.quotaGB || row.quotaGB===0;
      return (
        <div style={{width:170}}>
          <Tooltip title={isUnlimited ? 'نامحدود' : `${fmtGB(row.quotaUsed)} / ${row.quotaGB} GB`}>
            <Tag color={isUnlimited ? 'default' : row.quotaPct>=90 ? 'error' : row.quotaPct>=70 ? 'warning' : 'success'} style={{margin:0, borderRadius:999}} icon={<CloudOutlined/>}>
              {isUnlimited ? '∞ نامحدود' : `${row.quotaPct}%`}
            </Tag>
          </Tooltip>
          {!isUnlimited && (
            <>
              <Progress percent={row.quotaPct} size="small" showInfo={false} strokeColor={row.quotaPct>=90 ? '#ef4444' : row.quotaPct>=70 ? '#f59e0b' : '#06ffa5'} trailColor="rgba(255,255,255,0.08)" style={{margin:'4px 0 0'}} />
              <div style={{fontSize:11, color:'#94a3b8', marginTop:2}}>{fmtGB(row.quotaUsed)} / {row.quotaGB} GB</div>
            </>
          )}
        </div>
      );
    }},
    {title:'Contact', width:150, render:(_:any,row:AdminUser)=>(
      <Space direction="vertical" size={2}>
        {row.telegramId ? <span style={{fontSize:12,color:'#22d3ee'}}><BellOutlined/> {row.telegramId}</span> : <span style={{fontSize:12,color:'#475569'}}>—</span>}
        {row.supportUrl ? <Tooltip title={row.supportUrl}><a href={row.supportUrl} target="_blank" rel="noreferrer" style={{fontSize:12}}><LinkOutlined/> support</a></Tooltip> : null}
        {row.subDomain ? <span style={{fontSize:12,color:'#a78bfa'}}>{row.subDomain}</span> : null}
      </Space>
    )},
    {title:'Status', dataIndex:'enabled', width:96, render:(v:boolean,row:AdminUser)=><Switch checked={v} loading={toggleMut.isPending} onChange={()=>toggleMut.mutate(row)} checkedChildren="active" unCheckedChildren="disabled" style={v?{background:'#06ffa5'}:undefined} />},
    {title:'Actions', width:110, render:(_:any,row:AdminUser)=>(
      <Space>
        <Button size="small" ghost onClick={()=>openEdit(row)}>Edit</Button>
        <Dropdown menu={{items:[
          {key:'reset', label:'Reset usage', onClick:()=>resetUsageMut.mutate(row.id)},
          {key:'en', label:'Enable all clients', onClick:()=>clientsMut.mutate({id:row.id, act:'enable'})},
          {key:'dis', label:'Disable all clients', onClick:()=>clientsMut.mutate({id:row.id, act:'disable'})},
          {key:'rm', label:'Remove all clients', danger:true, onClick:()=>{
            Modal.confirm({title:`Remove ALL clients of ${row.username}?`, okText:'Remove', okType:'danger', cancelText:'Cancel',
              onOk:()=>clientsMut.mutate({id:row.id, act:'remove'})});
          }},
        ]}} trigger={['click']}>
          <Button size="small" ghost icon={<MoreOutlined/>} />
        </Dropdown>
        <Popconfirm title="Delete this admin?" description={row.username} okText="Delete" cancelText="Cancel" onConfirm={()=>deleteMut.mutate(row.id)}>
          <Button size="small" danger ghost>Del</Button>
        </Popconfirm>
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
            <div className="admins-page" style={{padding:24, maxWidth:1360, margin:'0 auto', width:'100%'}}>
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
                        <b style={{color:'#c4b5fd'}}>Viewer</b> read only · <b style={{color:'#a78bfa'}}>Creator</b> create only · <b style={{color:'#22d3ee'}}>Editor</b> create+edit · <b style={{color:'#6ee7b7'}}>Admin/Owner</b> full — سهمیه هر ادمین بر حسب GB
                      </div>
                    </div>
                  </div>
                  <Space>
                    <span style={{display:'inline-flex', alignItems:'center', gap:8, padding:'8px 12px', borderRadius:999, background:'rgba(255,255,255,0.06)', border:'1px solid rgba(255,255,255,0.08)', color:'#e2e8f0'}}>
                      <Badge status="processing" color="#06ffa5" /> active: <b>{active}</b> / {total}
                    </span>
                    <Button ghost onClick={()=>navigate('/panel/admin-roles')} icon={<SafetyCertificateOutlined/>} style={{borderRadius:999}}>Roles</Button>
                    <Button type="primary" icon={<UserAddOutlined/>} onClick={openCreate} style={{height:40, borderRadius:999, padding:'0 18px', background:'linear-gradient(135deg,#7c3aed,#06ffa5)', border:'none', boxShadow:'0 6px 20px rgba(139,92,246,0.35)'}}>New Admin</Button>
                  </Space>
                </div>
                <Row gutter={12} style={{position:'relative', marginTop:16}}>
                  <Col xs={12} sm={6}><Card size="small" style={{borderRadius:14, background:'rgba(255,255,255,0.04)', border:'1px solid rgba(255,255,255,0.08)'}}><Statistic title={<span style={{color:'#94a3b8'}}>Total</span>} value={total} valueStyle={{color:'#fff'}} prefix={<TeamOutlined/>} /></Card></Col>
                  <Col xs={12} sm={6}><Card size="small" style={{borderRadius:14, background:'rgba(6,255,165,0.06)', border:'1px solid rgba(6,255,165,0.2)'}}><Statistic title={<span style={{color:'#6ee7b7'}}>Active</span>} value={active} valueStyle={{color:'#6ee7b7'}} /></Card></Col>
                  <Col xs={12} sm={6}><Card size="small" style={{borderRadius:14, background:'rgba(100,116,139,0.08)', border:'1px solid rgba(100,116,139,0.25)'}}><Statistic title={<span style={{color:'#94a3b8'}}>Disabled</span>} value={disabled} valueStyle={{color:'#94a3b8'}} /></Card></Col>
                  <Col xs={12} sm={6}><Card size="small" style={{borderRadius:14, background:'rgba(239,68,68,0.06)', border:'1px solid rgba(239,68,68,0.25)'}}><Statistic title={<span style={{color:'#fca5a5'}}>Quota capped</span>} value={limited} valueStyle={{color:'#fca5a5'}} prefix={<CloudOutlined/>} /></Card></Col>
                </Row>
              </div>

              <Card style={{borderRadius:16, overflow:'hidden', border:'1px solid rgba(139,92,246,0.12)', boxShadow:'0 8px 32px rgba(0,0,0,0.25)'}} bodyStyle={{padding:0}}>
                <Table rowKey="id" loading={isLoading} dataSource={data||[]} columns={cols} pagination={false} size="middle" scroll={{x:1100}}
                  expandable={{expandedRowRender:(row:AdminUser)=>(
                    <div style={{display:'flex', gap:16, flexWrap:'wrap', fontSize:12, color:'#94a3b8'}}>
                      {row.note ? <span><FileTextOutlined/> {row.note}</span> : <span style={{color:'#475569'}}>no note</span>}
                      <span><EyeOutlined/> limit by <code style={{background:'rgba(0,0,0,0.25)', padding:'1px 6px', borderRadius:6}}>{row.inboundIds||'all'}</code></span>
                      {row.displayName ? <span>display: {row.displayName}</span> : null}
                    </div>
                  )}} />
              </Card>
              <div style={{textAlign:'center', color:'#64748b', fontSize:12, marginTop:10}}>Only <b style={{color:'#c4b5fd'}}>owner</b> can create/delete owner/admin · سهمیه 0 = نامحدود · با پر شدن سهمیه ادمین و همه کلاینت‌هایش خودکار disabled می‌شوند</div>
            </div>
          </Layout.Content>
        </Layout>
      </Layout>

      <Modal title={<span style={{display:'flex',gap:8,alignItems:'center'}}><span style={{width:28,height:28,borderRadius:8,display:'grid',placeItems:'center',background:'linear-gradient(135deg,#8b5cf6,#06ffa5)',color:'#fff'}}><UserAddOutlined/></span> {editing?`Edit — ${editing.username}`:'New Admin'}</span>}
        open={open} onCancel={()=>{setOpen(false); setEditing(null);}} footer={null} destroyOnClose width={620}>
        <Form form={form} layout="vertical" onFinish={(v)=>saveMut.mutate(v)} style={{marginTop:12}}>
          {!editing && <Form.Item name="username" label="Username" rules={[{required:true, message:'required'}]}><Input placeholder="e.g. sepehr" size="large" style={{borderRadius:12}} /></Form.Item>}
          {!editing && <Form.Item name="password" label="Password" rules={[{required:true, message:'required'}]}><Input.Password placeholder="••••••••" size="large" style={{borderRadius:12}} /></Form.Item>}
          <Form.Item name="displayName" label="Display name"><Input placeholder="Support" size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="roleId" label="Role" extra="Custom roles inherit their base tier access"><Select options={(roles||[]).map(r=>({value:r.id, label:`${r.name} (${r.baseTier})`}))} size="large" placeholder="select role" /></Form.Item>
          <Form.Item name="tier" label="Access tier (when no role selected)"><Select options={[
            {value:'viewer', label:'Viewer — read only'},
            {value:'creator', label:'Creator — create clients only'},
            {value:'editor', label:'Editor — create + edit'},
            {value:'admin', label:'Admin — full'},
            {value:'owner', label:'Owner — full + manage owners'},
          ]} size="large" /></Form.Item>
          <Form.Item name="inboundIds" label="Allowed inbounds" extra="empty = all · e.g. [1,3]"><Input placeholder="[] or [1,2]" size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="quotaGB" label="سهمیه (GB)" extra="0 = نامحدود — مجموع ترافیک کلاینت‌های این ادمین. پر شدن = خاموش خودکار">
            <InputNumber min={0} addonAfter="GB" placeholder="0 = ∞" size="large" style={{width:'100%', borderRadius:12}} />
          </Form.Item>
          <Typography.Title level={5} style={{color:'#c4b5fd', marginTop:8}}>Profile</Typography.Title>
          <Form.Item name="profileTitle" label="Profile title"><Input placeholder="e.g. Neon Support" size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="telegramId" label="Telegram ID"><Input placeholder="e.g. @support or numeric id" size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="supportUrl" label="Support URL"><Input placeholder="https://t.me/..." size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="subDomain" label="Subscription domain"><Input placeholder="sub.example.com" size="large" style={{borderRadius:12}} /></Form.Item>
          <Form.Item name="note" label="Note"><Input.TextArea rows={2} placeholder="internal note" style={{borderRadius:12}} /></Form.Item>
          <Button type="primary" htmlType="submit" loading={saveMut.isPending} block size="large" style={{borderRadius:12, height:44, background:'linear-gradient(135deg,#7c3aed,#06ffa5)', border:'none'}}>{editing?'Save':'Create Admin'}</Button>
        </Form>
      </Modal>
    </ConfigProvider>
  );
}
