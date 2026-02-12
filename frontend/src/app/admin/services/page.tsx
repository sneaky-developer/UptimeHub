'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getAdminServices, createService, deleteService } from '@/lib/api';
import type { Service } from '@/types';
import { useState } from 'react';
import { Server, Plus, Trash2, ExternalLink, X } from 'lucide-react';

export default function ServicesPage() {
    const queryClient = useQueryClient();
    const [showForm, setShowForm] = useState(false);

    const { data: services, isLoading } = useQuery<Service[]>({
        queryKey: ['admin-services'],
        queryFn: () => getAdminServices().then((r) => r.data),
    });

    const createMutation = useMutation({
        mutationFn: (data: Record<string, unknown>) => createService(data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-services'] });
            setShowForm(false);
        },
    });

    const deleteMutation = useMutation({
        mutationFn: (id: string) => deleteService(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-services'] });
        },
    });

    const handleCreate = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        createMutation.mutate({
            name: formData.get('name'),
            url: formData.get('url'),
            group_name: formData.get('group_name'),
            check_interval: Number(formData.get('check_interval')) || 30,
            timeout: Number(formData.get('timeout')) || 10,
            retries: Number(formData.get('retries')) || 3,
            failure_threshold: Number(formData.get('failure_threshold')) || 3,
            is_public: formData.get('is_public') === 'on',
        });
    };

    return (
        <div className="space-y-6 animate-fade-in">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                        <Server className="w-7 h-7 text-brand-400" />
                        Services
                    </h1>
                    <p className="text-slate-400 mt-1">Manage monitored endpoints</p>
                </div>
                <button onClick={() => setShowForm(!showForm)} className="btn-primary flex items-center gap-2">
                    <Plus className="w-4 h-4" />
                    Add Service
                </button>
            </div>

            {/* Create Form */}
            {showForm && (
                <div className="glass-card p-6 animate-slide-up">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="text-lg font-semibold text-white">New Service</h3>
                        <button onClick={() => setShowForm(false)} className="text-slate-400 hover:text-white">
                            <X className="w-5 h-5" />
                        </button>
                    </div>
                    <form onSubmit={handleCreate} className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">Name</label>
                            <input name="name" className="input-field" placeholder="My API" required />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">URL</label>
                            <input name="url" className="input-field" placeholder="https://api.example.com/health" required />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">Group</label>
                            <input name="group_name" className="input-field" placeholder="Production" />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">Check Interval (s)</label>
                            <input name="check_interval" type="number" className="input-field" defaultValue={30} />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">Timeout (s)</label>
                            <input name="timeout" type="number" className="input-field" defaultValue={10} />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">Retries</label>
                            <input name="retries" type="number" className="input-field" defaultValue={3} />
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1">Failure Threshold</label>
                            <input name="failure_threshold" type="number" className="input-field" defaultValue={3} />
                        </div>
                        <div className="flex items-center gap-3 pt-6">
                            <input name="is_public" type="checkbox" defaultChecked className="w-4 h-4 rounded border-slate-600 text-brand-500 focus:ring-brand-500" />
                            <label className="text-sm text-slate-300">Show on public status page</label>
                        </div>
                        <div className="md:col-span-2">
                            <button type="submit" className="btn-primary" disabled={createMutation.isPending}>
                                {createMutation.isPending ? 'Creating...' : 'Create Service'}
                            </button>
                        </div>
                    </form>
                </div>
            )}

            {/* Services Table */}
            <div className="glass-card overflow-hidden">
                {isLoading ? (
                    <div className="p-8 text-center text-slate-400">Loading services...</div>
                ) : services && services.length > 0 ? (
                    <table className="w-full">
                        <thead>
                            <tr className="border-b border-slate-700/50">
                                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">Service</th>
                                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">URL</th>
                                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">Group</th>
                                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">Status</th>
                                <th className="text-left py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">Interval</th>
                                <th className="text-right py-4 px-6 text-xs font-semibold text-slate-400 uppercase tracking-wider">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-700/30">
                            {services.map((svc) => (
                                <tr key={svc.id} className="hover:bg-slate-800/30 transition-colors">
                                    <td className="py-4 px-6">
                                        <div className="flex items-center gap-3">
                                            <span className={`status-dot-${svc.status}`} />
                                            <span className="font-medium text-white">{svc.name}</span>
                                        </div>
                                    </td>
                                    <td className="py-4 px-6">
                                        <a href={svc.url} target="_blank" rel="noopener noreferrer" className="text-sm text-slate-400 hover:text-brand-400 flex items-center gap-1">
                                            {svc.url.length > 40 ? svc.url.slice(0, 40) + '...' : svc.url}
                                            <ExternalLink className="w-3 h-3" />
                                        </a>
                                    </td>
                                    <td className="py-4 px-6 text-sm text-slate-400">{svc.group_name || '—'}</td>
                                    <td className="py-4 px-6">
                                        <span className={`status-badge status-${svc.status}`}>{svc.status}</span>
                                    </td>
                                    <td className="py-4 px-6 text-sm text-slate-400 font-mono">{svc.check_interval}s</td>
                                    <td className="py-4 px-6 text-right">
                                        <button
                                            onClick={() => { if (confirm('Delete this service?')) deleteMutation.mutate(svc.id); }}
                                            className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
                                            title="Delete"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                ) : (
                    <div className="p-8 text-center text-slate-400">
                        No services configured. Click &ldquo;Add Service&rdquo; to get started.
                    </div>
                )}
            </div>
        </div>
    );
}
