'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getNotificationChannels, createNotificationChannel, updateNotificationChannel, deleteNotificationChannel, testNotificationChannel } from '@/lib/api';
import type { NotificationChannel } from '@/types';
import { useState } from 'react';
import { Bell, Plus, Trash2, CheckCircle, AlertTriangle, Send, Pencil, X } from 'lucide-react';

export default function AlertsPage() {
    const queryClient = useQueryClient();
    const [showForm, setShowForm] = useState(false);
    const [channelType, setChannelType] = useState<'email' | 'slack' | 'webhook'>('email');
    const [editingChannel, setEditingChannel] = useState<NotificationChannel | null>(null);

    const { data: channels, isLoading } = useQuery<NotificationChannel[]>({
        queryKey: ['admin-channels'],
        queryFn: () => getNotificationChannels().then((r) => r.data),
    });

    const createMutation = useMutation({
        mutationFn: createNotificationChannel,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-channels'] });
            setShowForm(false);
            setEditingChannel(null);
        },
    });

    const updateMutation = useMutation({
        mutationFn: ({ id, data }: { id: string; data: Record<string, unknown> }) => updateNotificationChannel(id, data),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-channels'] });
            setShowForm(false);
            setEditingChannel(null);
        },
    });

    const deleteMutation = useMutation({
        mutationFn: deleteNotificationChannel,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['admin-channels'] });
        },
    });

    const testMutation = useMutation({
        mutationFn: testNotificationChannel,
        onSuccess: () => alert('Test alert sent!'),
        onError: (err) => alert('Test failed: ' + err),
    });

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        const formData = new FormData(e.currentTarget);
        const name = formData.get('name') as string;

        let config = {};
        if (channelType === 'email') {
            config = {
                host: formData.get('host'),
                port: Number(formData.get('port')),
                user: formData.get('user'),
                password: formData.get('password'),
                from: formData.get('from'),
                to: (formData.get('to') as string).split(',').map(s => s.trim()),
            };
        } else if (channelType === 'slack') {
            config = {
                webhook_url: formData.get('webhook_url'),
                channel: formData.get('channel'),
            };
        } else if (channelType === 'webhook') {
            config = {
                url: formData.get('url'),
            };
        }

        const payload = {
            name,
            type: channelType,
            config,
            enabled: formData.get('enabled') === 'on',
        };

        if (editingChannel) {
            updateMutation.mutate({ id: editingChannel.id, data: payload });
        } else {
            createMutation.mutate(payload);
        }
    };

    const handleEdit = (channel: NotificationChannel) => {
        setEditingChannel(channel);
        setChannelType(channel.type as any);
        setShowForm(true);
        window.scrollTo({ top: 0, behavior: 'smooth' });
    };

    const handleNew = () => {
        setEditingChannel(null);
        setChannelType('email');
        setShowForm(true);
    };

    return (
        <div className="space-y-6 animate-fade-in">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white flex items-center gap-3">
                        <Bell className="w-7 h-7 text-brand-400" />
                        Alerting
                    </h1>
                    <p className="text-slate-400 mt-1">Configure notification channels for incidents</p>
                </div>
                <button onClick={handleNew} className="btn-primary flex items-center gap-2">
                    <Plus className="w-4 h-4" />
                    New Channel
                </button>
            </div>

            {/* Create Form */}
            {showForm && (
                <div className="glass-card p-6 animate-slide-up">
                    <div className="flex items-center justify-between mb-4">
                        <h3 className="text-lg font-semibold text-white">
                            {editingChannel ? 'Edit Channel' : 'New Channel'}
                        </h3>
                        <button 
                            onClick={() => { setShowForm(false); setEditingChannel(null); }} 
                            className="text-slate-400 hover:text-white transition-colors"
                        >
                            <X className="w-5 h-5" />
                        </button>
                    </div>

                    <div className="flex gap-4 mb-6">
                        {['email', 'slack', 'webhook'].map((t) => (
                            <button
                                key={t}
                                type="button"
                                onClick={() => setChannelType(t as any)}
                                disabled={editingChannel !== null}
                                className={`px-4 py-2 rounded-lg text-sm font-medium capitalize transition-colors ${channelType === t ? 'bg-brand-500 text-white' : 'bg-slate-800 text-slate-400 hover:text-white'
                                    } ${editingChannel !== null && channelType !== t ? 'opacity-50 cursor-not-allowed' : ''}`}
                            >
                                {t}
                            </button>
                        ))}
                    </div>

                    <form key={editingChannel ? editingChannel.id : 'new-form'} onSubmit={handleSubmit} className="space-y-4">
                        <div className="flex flex-col md:flex-row md:items-start gap-4">
                            <div className="flex-1">
                                <label className="block text-sm font-medium text-slate-300 mb-1">Channel Name</label>
                                <input name="name" className="input-field" placeholder="My Team Alerts" defaultValue={editingChannel?.name} required />
                            </div>
                            <div className="flex items-center gap-3 md:mt-8">
                                <input name="enabled" type="checkbox" defaultChecked={editingChannel ? editingChannel.enabled : true} className="w-4 h-4 rounded border-slate-600 text-brand-500 focus:ring-brand-500" />
                                <label className="text-sm font-medium text-slate-300">Enabled for alerts</label>
                            </div>
                        </div>

                        {channelType === 'email' && (
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">SMTP Host</label>
                                    <input name="host" className="input-field" placeholder="smtp.gmail.com" defaultValue={(editingChannel?.config as any)?.host} required />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">SMTP Port</label>
                                    <input name="port" type="number" className="input-field" placeholder="587" defaultValue={(editingChannel?.config as any)?.port} required />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">Username</label>
                                    <input name="user" className="input-field" placeholder="user@example.com" defaultValue={(editingChannel?.config as any)?.user} />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">Password</label>
                                    <input name="password" type="password" className="input-field" placeholder="password" defaultValue={(editingChannel?.config as any)?.password} />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">From Email</label>
                                    <input name="from" className="input-field" placeholder="alerts@uptimehub.com" defaultValue={(editingChannel?.config as any)?.from} required />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">To Email(s)</label>
                                    <input name="to" className="input-field" placeholder="team@example.com, oncall@example.com" defaultValue={(editingChannel?.config as any)?.to?.join(', ')} required />
                                    <p className="text-xs text-slate-500 mt-1">Comma separated</p>
                                </div>
                            </div>
                        )}

                        {channelType === 'slack' && (
                            <div className="grid grid-cols-1 gap-4">
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">Webhook URL</label>
                                    <input name="webhook_url" className="input-field" placeholder="https://hooks.slack.com/services/..." defaultValue={(editingChannel?.config as any)?.webhook_url} required />
                                </div>
                                <div>
                                    <label className="block text-sm font-medium text-slate-300 mb-1">Channel (Optional)</label>
                                    <input name="channel" className="input-field" placeholder="#alerts" defaultValue={(editingChannel?.config as any)?.channel} />
                                </div>
                            </div>
                        )}

                        {channelType === 'webhook' && (
                            <div>
                                <label className="block text-sm font-medium text-slate-300 mb-1">Target URL</label>
                                <input name="url" className="input-field" placeholder="https://api.pagerduty.com/..." defaultValue={(editingChannel?.config as any)?.url} required />
                            </div>
                        )}

                        <div className="flex justify-end pt-4">
                            <button type="submit" className="btn-primary" disabled={createMutation.isPending || updateMutation.isPending}>
                                {createMutation.isPending || updateMutation.isPending ? 'Saving...' : 'Save Channel'}
                            </button>
                        </div>
                    </form>
                </div>
            )}

            {/* Channels List */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {isLoading ? (
                    <div className="col-span-full h-32 flex items-center justify-center text-slate-400">Loading...</div>
                ) : channels?.map((channel) => (
                    <div key={channel.id} className="glass-card p-6 flex flex-col justify-between group">
                        <div>
                            <div className="flex items-center justify-between mb-2">
                                <span className={`text-xs font-mono px-2 py-1 rounded bg-slate-800 text-slate-300 uppercase`}>
                                    {channel.type}
                                </span>
                                <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                    <button
                                        onClick={() => handleEdit(channel)}
                                        className="text-slate-400 hover:text-blue-400 transition-colors"
                                        title="Edit"
                                      >
                                          <Pencil className="w-4 h-4" />
                                      </button>
                                    <button
                                        onClick={() => testMutation.mutate(channel.id)}
                                        className="text-slate-400 hover:text-brand-400 transition-colors"
                                        title="Test Alert"
                                    >
                                        <Send className="w-4 h-4" />
                                    </button>
                                    <button
                                        onClick={() => { if (confirm('Delete?')) deleteMutation.mutate(channel.id); }}
                                        className="text-slate-400 hover:text-red-400 transition-colors"
                                        title="Delete"
                                    >
                                        <Trash2 className="w-4 h-4" />
                                    </button>
                                </div>
                            </div>
                            <h3 className="text-lg font-semibold text-white mb-1">{channel.name}</h3>
                            <div className="text-sm text-slate-400 truncate">
                                {channel.type === 'email' && (channel.config.to as string[]).join(', ')}
                                {channel.type === 'slack' && (channel.config.channel || 'Webhook')}
                                {channel.type === 'webhook' && channel.config.url}
                            </div>
                        </div>
                        <div className="mt-4 pt-4 border-t border-slate-700/50 flex items-center gap-2">
                            {channel.enabled ? (
                                <span className="flex items-center gap-2 text-xs text-brand-400">
                                    <CheckCircle className="w-3 h-3" /> Enabled
                                </span>
                            ) : (
                                <span className="flex items-center gap-2 text-xs text-slate-400">
                                    <AlertTriangle className="w-3 h-3" /> Disabled
                                </span>
                            )}
                        </div>
                    </div>
                ))}

                {!isLoading && channels?.length === 0 && (
                    <div className="col-span-full py-12 text-center text-slate-400 border border-dashed border-slate-800 rounded-xl">
                        No notification channels configured.
                    </div>
                )}
            </div>
        </div>
    );
}
