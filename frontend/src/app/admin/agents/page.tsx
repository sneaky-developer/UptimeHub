'use client';

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAdminAgents, getAgentGroups, createAgentGroup, deleteAgentGroup } from '@/lib/api';
import type { Agent, AgentGroup } from '@/types';
import { Cpu, Clock, Globe, Plus, Trash2, Key, Copy, Check, Server } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

export default function AgentsPage() {
    const [activeTab, setActiveTab] = useState<'workers' | 'groups'>('workers');
    const [isCreatingGroup, setIsCreatingGroup] = useState(false);
    const [newGroupName, setNewGroupName] = useState('');
    const [createdToken, setCreatedToken] = useState<string | null>(null);
    const [copiedToken, setCopiedToken] = useState(false);
    const queryClient = useQueryClient();

    const { data: agents, isLoading: isLoadingAgents } = useQuery<Agent[]>({
        queryKey: ['admin-agents'],
        queryFn: () => getAdminAgents().then((r) => r.data),
    });

    const { data: groups, isLoading: isLoadingGroups } = useQuery<AgentGroup[]>({
        queryKey: ['admin-agent-groups'],
        queryFn: () => getAgentGroups().then((r) => r.data),
    });

    const createGroupMutation = useMutation({
        mutationFn: (name: string) => createAgentGroup(name).then((r) => r.data as AgentGroup),
        onSuccess: (data) => {
            setCreatedToken(data.token!);
            setNewGroupName('');
            queryClient.invalidateQueries({ queryKey: ['admin-agent-groups'] });
        },
    });

    const deleteGroupMutation = useMutation({
        mutationFn: (id: string) => deleteAgentGroup(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-agent-groups'] });
            queryClient.invalidateQueries({ queryKey: ['admin-agents'] });
        },
    });

    const handleCopyToken = () => {
        if (createdToken) {
            navigator.clipboard.writeText(createdToken);
            setCopiedToken(true);
            setTimeout(() => setCopiedToken(false), 2000);
        }
    };

    return (
        <div className="space-y-6 animate-fade-in">
            <div className="flex flex-col md:flex-row md:items-end justify-between gap-4">
                <div>
                    <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                        <Cpu className="w-7 h-7 text-sky-400" />
                        Infrastructure Limits
                    </h1>
                    <p className="text-slate-400 mt-1">Manage agent groups and active worker nodes</p>
                </div>
                
                {/* Tabs */}
                <div className="flex bg-slate-900/50 p-1 rounded-lg border border-white/5">
                    <button
                        onClick={() => setActiveTab('workers')}
                        className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${
                            activeTab === 'workers' ? 'bg-sky-500/20 text-sky-400' : 'text-slate-400 hover:text-white'
                        }`}
                    >
                        Active Workers
                    </button>
                    <button
                        onClick={() => setActiveTab('groups')}
                        className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${
                            activeTab === 'groups' ? 'bg-sky-500/20 text-sky-400' : 'text-slate-400 hover:text-white'
                        }`}
                    >
                        Agent Groups
                    </button>
                </div>
            </div>

            {/* TAB: WORKERS */}
            {activeTab === 'workers' && (
                <div className="space-y-4">
                    {isLoadingAgents ? (
                        <div className="glass-card p-8 text-center text-slate-400">Loading agents...</div>
                    ) : agents && agents.length > 0 ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {agents.map((agent) => (
                                <div key={agent.id} className="glass-card-hover p-6">
                                    <div className="flex items-start justify-between mb-4">
                                        <div className="flex items-center gap-3">
                                            <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${agent.status === 'active' ? 'bg-emerald-500/15' : 'bg-slate-700/50'
                                                }`}>
                                                <Server className={`w-5 h-5 ${agent.status === 'active' ? 'text-emerald-400' : 'text-slate-500'
                                                    }`} />
                                            </div>
                                            <div>
                                                <h3 className="font-semibold text-white">{agent.name}</h3>
                                                <div className="flex items-center gap-1.5 mt-0.5">
                                                    <Globe className="w-3 h-3 text-slate-500" />
                                                    <span className="text-sm text-slate-400">{agent.group?.name || 'Unknown Group'}</span>
                                                </div>
                                            </div>
                                        </div>
                                        <span className={`status-badge ${agent.status === 'active' ? 'status-up' : agent.status === 'pending' ? 'status-degraded' : 'status-unknown'}`}>
                                            {agent.status}
                                        </span>
                                    </div>

                                    <div className="space-y-2 text-sm">
                                        <div className="flex items-center gap-2 text-slate-400">
                                            <Clock className="w-3.5 h-3.5" />
                                            <span>
                                                Last heartbeat:{' '}
                                                {agent.last_heartbeat
                                                    ? formatDistanceToNow(new Date(agent.last_heartbeat), { addSuffix: true })
                                                    : 'Never'}
                                            </span>
                                        </div>
                                        <div className="text-xs text-slate-500 font-mono text-ellipsis overflow-hidden">
                                            ID: {agent.id}
                                        </div>
                                        <div className="text-xs text-slate-500">
                                            Registered: {new Date(agent.created_at).toLocaleDateString()}
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="glass-card p-12 flex flex-col items-center justify-center text-center">
                            <Server className="w-12 h-12 text-slate-500 mb-4" />
                            <h3 className="text-lg font-medium text-white mb-2">No Worker Nodes</h3>
                            <p className="text-slate-400 max-w-sm">
                                You don't have any active agents reporting to the master. Go to Agent Groups to create an enrollment token.
                            </p>
                        </div>
                    )}
                </div>
            )}

            {/* TAB: GROUPS */}
            {activeTab === 'groups' && (
                <div className="space-y-6">
                    {/* Header Action */}
                    <div className="flex justify-end">
                        <button
                            onClick={() => setIsCreatingGroup(!isCreatingGroup)}
                            className="btn-primary"
                        >
                            <Plus className="w-4 h-4 mr-2" />
                            Create Group
                        </button>
                    </div>

                    {/* Create Group Form/Result */}
                    {isCreatingGroup && (
                        <div className="glass-card p-6 border-sky-500/30">
                            <h3 className="text-lg font-medium text-white mb-4">Create New Agent Group</h3>
                            
                            {!createdToken ? (
                                <div className="flex gap-3">
                                    <input
                                        type="text"
                                        placeholder="Group Name (e.g. us-east-production)"
                                        value={newGroupName}
                                        onChange={(e) => setNewGroupName(e.target.value)}
                                        className="input-field flex-1"
                                    />
                                    <button
                                        onClick={() => createGroupMutation.mutate(newGroupName)}
                                        disabled={!newGroupName || createGroupMutation.isPending}
                                        className="btn-primary"
                                    >
                                        Generate Token
                                    </button>
                                </div>
                            ) : (
                                <div className="space-y-4">
                                    <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-lg">
                                        <p className="text-emerald-400 text-sm mb-2 font-medium">Group created successfully! Copy this Enrollment Token now.</p>
                                        <p className="text-slate-400 text-xs mb-3">You will need to pass this as the <code className="text-sky-400">ENROLLMENT_TOKEN</code> environment variable when starting Agent worker nodes in this group. You will not be able to see this token again.</p>
                                        
                                        <div className="flex items-center gap-2">
                                            <code className="flex-1 p-3 bg-slate-900 rounded border border-white/10 text-white font-mono text-sm break-all">
                                                {createdToken}
                                            </code>
                                            <button
                                                onClick={handleCopyToken}
                                                className="p-3 bg-slate-800 hover:bg-slate-700 rounded transition-colors text-white"
                                                title="Copy to clipboard"
                                            >
                                                {copiedToken ? <Check className="w-5 h-5 text-emerald-400" /> : <Copy className="w-5 h-5" />}
                                            </button>
                                        </div>
                                    </div>
                                    <button onClick={() => { setCreatedToken(null); setIsCreatingGroup(false); }} className="btn-secondary">
                                        Done
                                    </button>
                                </div>
                            )}
                        </div>
                    )}

                    {/* Groups List */}
                    <div className="glass-card overflow-hidden">
                        <div className="overflow-x-auto">
                            <table className="w-full text-left">
                                <thead>
                                    <tr className="border-b border-slate-700/50 bg-slate-800/30">
                                        <th className="px-6 py-4 text-sm font-semibold text-slate-300">Group Name</th>
                                        <th className="px-6 py-4 text-sm font-semibold text-slate-300">Created Date</th>
                                        <th className="px-6 py-4 text-right text-sm font-semibold text-slate-300">Actions</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-700/50">
                                    {isLoadingGroups ? (
                                        <tr>
                                            <td colSpan={3} className="px-6 py-8 text-center text-slate-400">Loading groups...</td>
                                        </tr>
                                    ) : groups && groups.length > 0 ? (
                                        groups.map((group) => (
                                            <tr key={group.id} className="hover:bg-slate-800/30 transition-colors">
                                                <td className="px-6 py-4">
                                                    <div className="flex items-center gap-3">
                                                        <Key className="w-4 h-4 text-sky-400" />
                                                        <span className="font-medium text-white">{group.name}</span>
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 text-sm text-slate-400">
                                                    {new Date(group.created_at).toLocaleDateString()}
                                                </td>
                                                <td className="px-6 py-4 text-right">
                                                    <button
                                                        onClick={() => {
                                                            if (confirm(`Are you sure you want to delete ${group.name}? This will instantly kick offline all agents in this group.`)) {
                                                                deleteGroupMutation.mutate(group.id);
                                                            }
                                                        }}
                                                        className="p-2 text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 rounded-lg transition-colors"
                                                        title="Delete Group & Revoke Access"
                                                    >
                                                        <Trash2 className="w-4 h-4" />
                                                    </button>
                                                </td>
                                            </tr>
                                        ))
                                    ) : (
                                        <tr>
                                            <td colSpan={3} className="px-6 py-12 text-center text-slate-400">
                                                No agent groups configured. Create one to get your enrollment token.
                                            </td>
                                        </tr>
                                    )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
