'use client';

import { useQuery } from '@tanstack/react-query';
import { getAdminAgents } from '@/lib/api';
import type { Agent } from '@/types';
import { Cpu, Clock, Globe } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';

export default function AgentsPage() {
    const { data: agents, isLoading } = useQuery<Agent[]>({
        queryKey: ['admin-agents'],
        queryFn: () => getAdminAgents().then((r) => r.data),
    });

    return (
        <div className="space-y-6 animate-fade-in">
            <div>
                <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                    <Cpu className="w-7 h-7 text-sky-400" />
                    Agents
                </h1>
                <p className="text-slate-400 mt-1">Registered worker agents across your clusters</p>
            </div>

            {isLoading ? (
                <div className="glass-card p-8 text-center text-slate-400">Loading agents...</div>
            ) : agents && agents.length > 0 ? (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {agents.map((agent) => (
                        <div key={agent.id} className="glass-card-hover p-6">
                            <div className="flex items-start justify-between mb-4">
                                <div className="flex items-center gap-3">
                                    <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${agent.status === 'active' ? 'bg-emerald-500/15' : 'bg-slate-700/50'
                                        }`}>
                                        <Cpu className={`w-5 h-5 ${agent.status === 'active' ? 'text-emerald-400' : 'text-slate-500'
                                            }`} />
                                    </div>
                                    <div>
                                        <h3 className="font-semibold text-white">{agent.name}</h3>
                                        <div className="flex items-center gap-1.5 mt-0.5">
                                            <Globe className="w-3 h-3 text-slate-500" />
                                            <span className="text-sm text-slate-400">{agent.cluster_name}</span>
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
                                <div className="text-xs text-slate-500 font-mono">
                                    ID: {agent.id}
                                </div>
                                <div className="text-xs text-slate-500 font-mono">
                                    Registered: {new Date(agent.created_at).toLocaleDateString()}
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            ) : (
                <div className="glass-card p-8 text-center text-slate-400">
                    No agents registered yet. Deploy an agent to your Kubernetes cluster to get started.
                </div>
            )}
        </div>
    );
}
