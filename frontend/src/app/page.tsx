'use client';

import { useQuery } from '@tanstack/react-query';
import { getStatus, getPublicIncidents, getMaintenance } from '@/lib/api';
import type { ServiceStatus, Incident, MaintenanceWindow } from '@/types';
import { Activity, AlertTriangle, CheckCircle2, XCircle, Clock, Wrench } from 'lucide-react';

export default function StatusPage() {
    const { data: services, isLoading: servicesLoading } = useQuery<ServiceStatus[]>({
        queryKey: ['status'],
        queryFn: () => getStatus().then((r) => r.data),
    });

    const { data: incidents } = useQuery<Incident[]>({
        queryKey: ['incidents'],
        queryFn: () => getPublicIncidents().then((r) => r.data),
    });

    const { data: maintenance } = useQuery<MaintenanceWindow[]>({
        queryKey: ['maintenance'],
        queryFn: () => getMaintenance().then((r) => r.data),
    });

    const activeIncidents = incidents?.filter((i) => i.status !== 'resolved') || [];
    const recentIncidents = incidents?.filter((i) => i.status === 'resolved').slice(0, 10) || [];

    // Group services by group_name
    const groupedServices = services?.reduce(
        (acc, svc) => {
            const group = svc.group_name || 'Services';
            if (!acc[group]) acc[group] = [];
            acc[group].push(svc);
            return acc;
        },
        {} as Record<string, ServiceStatus[]>
    );

    // Overall status
    const overallStatus = services?.every((s) => s.status === 'up')
        ? 'operational'
        : services?.some((s) => s.status === 'down')
            ? 'outage'
            : services?.some((s) => s.status === 'degraded')
                ? 'degraded'
                : 'unknown';

    return (
        <div className="min-h-screen">
            {/* Header */}
            <header className="border-b border-slate-800">
                <div className="max-w-4xl mx-auto px-6 py-6 flex items-center justify-between">
                    <div className="flex items-center gap-3">
                        <div className="w-10 h-10 bg-gradient-to-br from-brand-500 to-brand-700 rounded-xl flex items-center justify-center">
                            <Activity className="w-5 h-5 text-white" />
                        </div>
                        <h1 className="text-xl font-bold text-white">UptimeHub</h1>
                    </div>
                    <a
                        href="/admin/login"
                        className="text-sm text-slate-500 hover:text-slate-300 transition-colors"
                    >
                        Admin
                    </a>
                </div>
            </header>

            <main className="max-w-4xl mx-auto px-6 py-10 space-y-10 animate-fade-in">
                {/* Overall Status Banner */}
                <div className="glass-card p-8 text-center">
                    {servicesLoading ? (
                        <div className="flex items-center justify-center gap-3">
                            <div className="w-5 h-5 border-2 border-brand-500 border-t-transparent rounded-full animate-spin" />
                            <span className="text-slate-400">Loading status...</span>
                        </div>
                    ) : (
                        <>
                            <div className="flex items-center justify-center gap-3 mb-2">
                                {overallStatus === 'operational' && (
                                    <CheckCircle2 className="w-8 h-8 text-emerald-400" />
                                )}
                                {overallStatus === 'outage' && (
                                    <XCircle className="w-8 h-8 text-red-400" />
                                )}
                                {overallStatus === 'degraded' && (
                                    <AlertTriangle className="w-8 h-8 text-amber-400" />
                                )}
                            </div>
                            <h2 className="text-2xl font-bold">
                                {overallStatus === 'operational' && 'All Systems Operational'}
                                {overallStatus === 'outage' && 'Service Outage Detected'}
                                {overallStatus === 'degraded' && 'Partial Service Degradation'}
                                {overallStatus === 'unknown' && 'Status Unknown'}
                            </h2>
                            <p className="text-slate-400 mt-1 text-sm">
                                Last updated: {new Date().toLocaleTimeString()}
                            </p>
                        </>
                    )}
                </div>

                {/* Active Incidents */}
                {activeIncidents.length > 0 && (
                    <section>
                        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                            <AlertTriangle className="w-5 h-5 text-amber-400" />
                            Active Incidents
                        </h3>
                        <div className="space-y-3">
                            {activeIncidents.map((incident) => (
                                <div key={incident.id} className="glass-card p-5 border-l-4 border-l-amber-500">
                                    <div className="flex items-start justify-between">
                                        <div>
                                            <h4 className="font-semibold text-white">{incident.title}</h4>
                                            <p className="text-sm text-slate-400 mt-1">{incident.description}</p>
                                        </div>
                                        <span className={`status-badge status-${incident.severity === 'critical' ? 'down' : incident.severity === 'major' ? 'degraded' : 'partial'}`}>
                                            {incident.severity}
                                        </span>
                                    </div>
                                    {incident.updates && incident.updates.length > 0 && (
                                        <div className="mt-4 space-y-2 border-t border-slate-700/50 pt-3">
                                            {incident.updates.map((update) => (
                                                <div key={update.id} className="flex gap-3 text-sm">
                                                    <span className="text-slate-500 whitespace-nowrap font-mono text-xs mt-0.5">
                                                        {new Date(update.created_at).toLocaleString()}
                                                    </span>
                                                    <span className="text-slate-300">{update.message}</span>
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>
                    </section>
                )}

                {/* Maintenance */}
                {maintenance && maintenance.length > 0 && (
                    <section>
                        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                            <Wrench className="w-5 h-5 text-blue-400" />
                            Scheduled Maintenance
                        </h3>
                        <div className="space-y-3">
                            {maintenance.map((m) => (
                                <div key={m.id} className="glass-card p-5 border-l-4 border-l-blue-500">
                                    <h4 className="font-semibold text-white">{m.title}</h4>
                                    <p className="text-sm text-slate-400 mt-1">{m.description}</p>
                                    <p className="text-xs text-slate-500 mt-2 font-mono">
                                        {new Date(m.scheduled_start).toLocaleString()} — {new Date(m.scheduled_end).toLocaleString()}
                                    </p>
                                </div>
                            ))}
                        </div>
                    </section>
                )}

                {/* Service Groups */}
                {groupedServices &&
                    Object.entries(groupedServices).map(([group, svcs]) => (
                        <section key={group} className="animate-slide-up">
                            <h3 className="text-lg font-semibold text-white mb-4">{group}</h3>
                            <div className="glass-card divide-y divide-slate-700/50">
                                {svcs.map((svc) => (
                                    <div key={svc.id} className="p-5 flex items-center justify-between">
                                        <div className="flex items-center gap-3">
                                            <span className={`status-dot-${svc.status}`} />
                                            <span className="font-medium text-white">{svc.name}</span>
                                        </div>
                                        <div className="flex items-center gap-4">
                                            {/* Uptime bars */}
                                            {svc.history && svc.history.length > 0 && (
                                                <div className="hidden sm:flex items-end gap-0.5" title="90-day uptime history">
                                                    {svc.history.slice(-90).map((day, i) => (
                                                        <div
                                                            key={i}
                                                            className={`uptime-bar ${day.uptime_pct >= 99
                                                                    ? 'bg-emerald-500'
                                                                    : day.uptime_pct >= 95
                                                                        ? 'bg-amber-500'
                                                                        : 'bg-red-500'
                                                                }`}
                                                            title={`${day.date}: ${day.uptime_pct}%`}
                                                        />
                                                    ))}
                                                </div>
                                            )}
                                            <div className="text-right">
                                                <span className={`status-badge status-${svc.status}`}>
                                                    {svc.status}
                                                </span>
                                                <p className="text-xs text-slate-500 mt-1 font-mono">
                                                    {svc.uptime_pct.toFixed(2)}% uptime
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </section>
                    ))}

                {/* Recent Incidents */}
                {recentIncidents.length > 0 && (
                    <section>
                        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                            <Clock className="w-5 h-5 text-slate-400" />
                            Past Incidents
                        </h3>
                        <div className="space-y-3">
                            {recentIncidents.map((incident) => (
                                <div key={incident.id} className="glass-card p-5">
                                    <div className="flex items-start justify-between">
                                        <div>
                                            <h4 className="font-medium text-white">{incident.title}</h4>
                                            <p className="text-xs text-slate-500 mt-1 font-mono">
                                                {new Date(incident.started_at).toLocaleDateString()} —{' '}
                                                {incident.resolved_at
                                                    ? new Date(incident.resolved_at).toLocaleDateString()
                                                    : 'Ongoing'}
                                            </p>
                                        </div>
                                        <span className="status-badge status-up">Resolved</span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </section>
                )}

                {/* Footer */}
                <footer className="text-center py-8 border-t border-slate-800">
                    <p className="text-sm text-slate-500">
                        Powered by <span className="text-brand-400 font-semibold">UptimeHub</span>
                    </p>
                </footer>
            </main>
        </div>
    );
}
