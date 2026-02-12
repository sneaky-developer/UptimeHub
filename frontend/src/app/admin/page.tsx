'use client';

import { useQuery } from '@tanstack/react-query';
import { getAdminServices, getAdminAgents, getAdminIncidents } from '@/lib/api';
import type { Service, Agent, Incident } from '@/types';
import { Server, Cpu, AlertTriangle, CheckCircle2, XCircle, Activity } from 'lucide-react';

export default function AdminDashboard() {
    const { data: services } = useQuery<Service[]>({
        queryKey: ['admin-services'],
        queryFn: () => getAdminServices().then((r) => r.data),
    });

    const { data: agents } = useQuery<Agent[]>({
        queryKey: ['admin-agents'],
        queryFn: () => getAdminAgents().then((r) => r.data),
    });

    const { data: incidents } = useQuery<Incident[]>({
        queryKey: ['admin-incidents'],
        queryFn: () => getAdminIncidents().then((r) => r.data),
    });

    const servicesUp = services?.filter((s) => s.status === 'up').length || 0;
    const servicesDown = services?.filter((s) => s.status === 'down').length || 0;
    const activeAgents = agents?.filter((a) => a.status === 'active').length || 0;
    const activeIncidents = incidents?.filter((i) => i.status !== 'resolved').length || 0;

    const stats = [
        {
            label: 'Total Services',
            value: services?.length || 0,
            icon: Server,
            color: 'text-brand-400',
            bg: 'bg-brand-500/10',
        },
        {
            label: 'Services Up',
            value: servicesUp,
            icon: CheckCircle2,
            color: 'text-emerald-400',
            bg: 'bg-emerald-500/10',
        },
        {
            label: 'Services Down',
            value: servicesDown,
            icon: XCircle,
            color: 'text-red-400',
            bg: 'bg-red-500/10',
        },
        {
            label: 'Active Agents',
            value: activeAgents,
            icon: Cpu,
            color: 'text-sky-400',
            bg: 'bg-sky-500/10',
        },
        {
            label: 'Active Incidents',
            value: activeIncidents,
            icon: AlertTriangle,
            color: 'text-amber-400',
            bg: 'bg-amber-500/10',
        },
    ];

    return (
        <div className="space-y-8 animate-fade-in">
            <div>
                <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                    <Activity className="w-7 h-7 text-brand-400" />
                    Dashboard
                </h1>
                <p className="text-slate-400 mt-1">Overview of your monitoring system</p>
            </div>

            {/* Stats Grid */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
                {stats.map((stat) => (
                    <div key={stat.label} className="glass-card p-5">
                        <div className="flex items-center gap-3 mb-3">
                            <div className={`w-10 h-10 ${stat.bg} rounded-xl flex items-center justify-center`}>
                                <stat.icon className={`w-5 h-5 ${stat.color}`} />
                            </div>
                        </div>
                        <p className="text-3xl font-bold text-white">{stat.value}</p>
                        <p className="text-sm text-slate-400 mt-1">{stat.label}</p>
                    </div>
                ))}
            </div>

            {/* Recent Activity */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* Recent Incidents */}
                <div className="glass-card p-6">
                    <h3 className="text-lg font-semibold text-white mb-4">Recent Incidents</h3>
                    {incidents && incidents.length > 0 ? (
                        <div className="space-y-3">
                            {incidents.slice(0, 5).map((incident) => (
                                <div key={incident.id} className="flex items-center justify-between p-3 bg-slate-900/40 rounded-xl">
                                    <div>
                                        <p className="text-sm font-medium text-white">{incident.title}</p>
                                        <p className="text-xs text-slate-500">{new Date(incident.started_at).toLocaleDateString()}</p>
                                    </div>
                                    <span className={`status-badge ${incident.status === 'resolved' ? 'status-up' : 'status-degraded'}`}>
                                        {incident.status}
                                    </span>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <p className="text-sm text-slate-500">No incidents recorded</p>
                    )}
                </div>

                {/* Agents Status */}
                <div className="glass-card p-6">
                    <h3 className="text-lg font-semibold text-white mb-4">Agents</h3>
                    {agents && agents.length > 0 ? (
                        <div className="space-y-3">
                            {agents.map((agent) => (
                                <div key={agent.id} className="flex items-center justify-between p-3 bg-slate-900/40 rounded-xl">
                                    <div>
                                        <p className="text-sm font-medium text-white">{agent.name}</p>
                                        <p className="text-xs text-slate-500">{agent.cluster_name}</p>
                                    </div>
                                    <div className="text-right">
                                        <span className={`status-badge ${agent.status === 'active' ? 'status-up' : 'status-unknown'}`}>
                                            {agent.status}
                                        </span>
                                        {agent.last_heartbeat && (
                                            <p className="text-xs text-slate-500 mt-1">
                                                Last: {new Date(agent.last_heartbeat).toLocaleTimeString()}
                                            </p>
                                        )}
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <p className="text-sm text-slate-500">No agents registered</p>
                    )}
                </div>
            </div>
        </div>
    );
}
