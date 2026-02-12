'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAdminIncidents, createIncident, updateIncident } from '@/lib/api';
import type { Incident } from '@/types';
import { useState } from 'react';
import { AlertTriangle, Plus, X } from 'lucide-react';

export default function IncidentsPage() {
    const queryClient = useQueryClient();
    const [showForm, setShowForm] = useState(false);
    const [updatingId, setUpdatingId] = useState<string | null>(null);
    const [updateStatus, setUpdateStatus] = useState('');
    const [updateMessage, setUpdateMessage] = useState('');

    const { data: incidents, isLoading } = useQuery<Incident[]>({
        queryKey: ['admin-incidents'],
        queryFn: () => getAdminIncidents().then((r) => r.data),
    });

    const createMutation = useMutation({
        mutationFn: (data: Record<string, unknown>) => createIncident(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-incidents'] });
            setShowForm(false);
        },
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string; data: Record<string, unknown> }) => updateIncident(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-incidents'] });
            setUpdatingId(null);
            setUpdateStatus('');
            setUpdateMessage('');
        },
    });

    const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        createMutation.mutate({
            title: formData.get('title'),
            description: formData.get('description'),
            severity: formData.get('severity'),
            status: formData.get('status'),
        });
    };

    const handleUpdate = (id: string) => {
        updateMutation.mutate({
            id,
            data: { status: updateStatus, message: updateMessage },
        });
    };

    const statusColors: Record<string, string> = {
        investigating: 'status-down',
        identified: 'status-degraded',
        monitoring: 'status-partial',
        resolved: 'status-up',
    };

    return (
        <div className="space-y-6 animate-fade-in">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                        <AlertTriangle className="w-7 h-7 text-amber-400" />
                        Incidents
                    </h1>
                    <p className="text-slate-400 mt-1">Manage service incidents</p>
                </div>
                <button onClick={() => setShowForm(!showForm)} className="btn-primary flex items-center gap-2">
                    <Plus className="w-4 h-4" />
                    Create Incident
                </button>
            </div>

            {/* Create Form */}
            {showForm && (
                <div className="glass-card p-6 animate-slide-up">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="text-lg font-semibold text-white">New Incident</h3>
                        <button onClick={() => setShowForm(false)} className="text-slate-400 hover:text-white">
                            <X className="w-5 h-5" />
                        </button>
                    </div>
                    <form onSubmit={handleCreate} className="space-y-4">
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">Title</label>
                            <input name="title" className="input-field" placeholder="API experiencing elevated latency" required />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">Description</label>
                            <textarea name="description" className="input-field min-h-[80px]" placeholder="We are investigating elevated response times..." />
                        </div>
                        <div className="grid grid-cols-2 gap-4">
                            <div>
                                <label className="block text-sm font-medium text-slate-300 mb-1">Severity</label>
                                <select name="severity" className="input-field" defaultValue="minor">
                                    <option value="minor">Minor</option>
                                    <option value="major">Major</option>
                                    <option value="critical">Critical</option>
                                </select>
                            </div>
                            <div>
                                <label className="block text-sm font-medium text-slate-300 mb-1">Status</label>
                                <select name="status" className="input-field" defaultValue="investigating">
                                    <option value="investigating">Investigating</option>
                                    <option value="identified">Identified</option>
                                    <option value="monitoring">Monitoring</option>
                                    <option value="resolved">Resolved</option>
                                </select>
                            </div>
                        </div>
                        <button type="submit" className="btn-primary" disabled={createMutation.isPending}>
                            {createMutation.isPending ? 'Creating...' : 'Create Incident'}
                        </button>
                    </form>
                </div>
            )}

            {/* Incidents List */}
            {isLoading ? (
                <div className="glass-card p-8 text-center text-slate-400">Loading incidents...</div>
            ) : incidents && incidents.length > 0 ? (
                <div className="space-y-4">
                    {incidents.map((incident) => (
                        <div key={incident.id} className="glass-card p-6">
                            <div className="flex items-start justify-between mb-3">
                                <div>
                                    <h3 className="text-lg font-semibold text-white">{incident.title}</h3>
                                    <p className="text-sm text-slate-400 mt-1">{incident.description}</p>
                                </div>
                                <div className="flex items-center gap-2">
                                    <span className={`status-badge ${statusColors[incident.status] || 'status-unknown'}`}>
                                        {incident.status}
                                    </span>
                                    <span className={`status-badge ${incident.severity === 'critical' ? 'status-down' : incident.severity === 'major' ? 'status-degraded' : 'status-partial'}`}>
                                        {incident.severity}
                                    </span>
                                </div>
                            </div>

                            <div className="text-xs text-slate-500 font-mono mb-3">
                                Started: {new Date(incident.started_at).toLocaleString()}
                                {incident.resolved_at && ` — Resolved: ${new Date(incident.resolved_at).toLocaleString()}`}
                            </div>

                            {/* Timeline */}
                            {incident.updates && incident.updates.length > 0 && (
                                <div className="border-t border-slate-700/50 pt-3 mt-3 space-y-2">
                                    {incident.updates.map((update) => (
                                        <div key={update.id} className="flex gap-3 text-sm">
                                            <span className={`status-badge text-xs ${statusColors[update.status] || 'status-unknown'}`}>
                                                {update.status}
                                            </span>
                                            <span className="text-slate-300 flex-1">{update.message}</span>
                                            <span className="text-slate-500 text-xs font-mono whitespace-nowrap">
                                                {new Date(update.created_at).toLocaleTimeString()}
                                            </span>
                                        </div>
                                    ))}
                                </div>
                            )}

                            {/* Update form */}
                            {incident.status !== 'resolved' && (
                                <div className="border-t border-slate-700/50 pt-3 mt-3">
                                    {updatingId === incident.id ? (
                                        <div className="space-y-3">
                                            <select
                                                value={updateStatus}
                                                onChange={(e) => setUpdateStatus(e.target.value)}
                                                className="input-field"
                                            >
                                                <option value="">Select status...</option>
                                                <option value="investigating">Investigating</option>
                                                <option value="identified">Identified</option>
                                                <option value="monitoring">Monitoring</option>
                                                <option value="resolved">Resolved</option>
                                            </select>
                                            <textarea
                                                value={updateMessage}
                                                onChange={(e) => setUpdateMessage(e.target.value)}
                                                className="input-field min-h-[60px]"
                                                placeholder="Update message..."
                                            />
                                            <div className="flex gap-2">
                                                <button
                                                    onClick={() => handleUpdate(incident.id)}
                                                    className="btn-primary text-sm"
                                                    disabled={!updateStatus || !updateMessage || updateMutation.isPending}
                                                >
                                                    Post Update
                                                </button>
                                                <button onClick={() => setUpdatingId(null)} className="btn-secondary text-sm">
                                                    Cancel
                                                </button>
                                            </div>
                                        </div>
                                    ) : (
                                        <button
                                            onClick={() => setUpdatingId(incident.id)}
                                            className="btn-secondary text-sm"
                                        >
                                            Post Update
                                        </button>
                                    )}
                                </div>
                            )}
                        </div>
                    ))}
                </div>
            ) : (
                <div className="glass-card p-8 text-center text-slate-400">
                    No incidents recorded. All systems operational!
                </div>
            )}
        </div>
    );
}
