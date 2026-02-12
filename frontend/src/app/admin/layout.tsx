'use client';

import { useEffect, useState, type ReactNode } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import {
    Activity,
    LayoutDashboard,
    Server,
    AlertTriangle,
    Cpu,
    LogOut,
} from 'lucide-react';

const navItems = [
    { href: '/admin', icon: LayoutDashboard, label: 'Dashboard' },
    { href: '/admin/services', icon: Server, label: 'Services' },
    { href: '/admin/incidents', icon: AlertTriangle, label: 'Incidents' },
    { href: '/admin/agents', icon: Cpu, label: 'Agents' },
];

export default function AdminLayout({ children }: { children: ReactNode }) {
    const router = useRouter();
    const pathname = usePathname();
    const [user, setUser] = useState<{ name: string; email: string } | null>(null);

    useEffect(() => {
        // Skip auth check on login page
        if (pathname === '/admin/login') return;

        const token = localStorage.getItem('uptimehub_token');
        const userData = localStorage.getItem('uptimehub_user');

        if (!token) {
            router.push('/admin/login');
            return;
        }

        if (userData) {
            setUser(JSON.parse(userData));
        }
    }, [pathname, router]);

    // Don't wrap login page in admin layout
    if (pathname === '/admin/login') {
        return <>{children}</>;
    }

    const handleLogout = () => {
        localStorage.removeItem('uptimehub_token');
        localStorage.removeItem('uptimehub_user');
        router.push('/admin/login');
    };

    return (
        <div className="min-h-screen flex">
            {/* Sidebar */}
            <aside className="w-64 bg-slate-900/80 border-r border-slate-800 flex flex-col">
                {/* Logo */}
                <div className="p-6 border-b border-slate-800">
                    <Link href="/admin" className="flex items-center gap-3">
                        <div className="w-9 h-9 bg-gradient-to-br from-brand-500 to-brand-700 rounded-xl flex items-center justify-center">
                            <Activity className="w-4.5 h-4.5 text-white" />
                        </div>
                        <div>
                            <h1 className="font-bold text-white text-sm">UptimeHub</h1>
                            <p className="text-xs text-slate-500">Admin Panel</p>
                        </div>
                    </Link>
                </div>

                {/* Navigation */}
                <nav className="flex-1 p-4 space-y-1">
                    {navItems.map((item) => {
                        const isActive = pathname === item.href;
                        return (
                            <Link
                                key={item.href}
                                href={item.href}
                                className={`flex items-center gap-3 px-4 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${isActive
                                        ? 'bg-brand-600/20 text-brand-400 border border-brand-500/20'
                                        : 'text-slate-400 hover:text-white hover:bg-slate-800/50'
                                    }`}
                            >
                                <item.icon className="w-4.5 h-4.5" />
                                {item.label}
                            </Link>
                        );
                    })}
                </nav>

                {/* User info + Logout */}
                <div className="p-4 border-t border-slate-800">
                    <div className="flex items-center justify-between px-4 py-2">
                        <div>
                            <p className="text-sm font-medium text-white">{user?.name || 'Admin'}</p>
                            <p className="text-xs text-slate-500">{user?.email}</p>
                        </div>
                        <button
                            onClick={handleLogout}
                            className="p-2 text-slate-400 hover:text-red-400 transition-colors rounded-lg hover:bg-red-500/10"
                            title="Logout"
                        >
                            <LogOut className="w-4 h-4" />
                        </button>
                    </div>
                </div>
            </aside>

            {/* Main Content */}
            <main className="flex-1 overflow-y-auto">
                <div className="p-8">{children}</div>
            </main>
        </div>
    );
}
