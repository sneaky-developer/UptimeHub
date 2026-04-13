// ─── API Response Types ─────────────────────────────────────────────

export interface ServiceStatus {
    id: string;
    name: string;
    status: 'up' | 'down' | 'degraded' | 'unknown';
    group_name: string;
    uptime_pct: number;
    history?: UptimeDayStatus[];
}

export interface UptimeDayStatus {
    date: string;
    uptime_pct: number;
    status: 'up' | 'down' | 'degraded' | 'partial';
}

export interface Incident {
    id: string;
    service_id: string | null;
    title: string;
    description: string;
    status: 'investigating' | 'identified' | 'monitoring' | 'resolved';
    severity: 'minor' | 'major' | 'critical';
    started_at: string;
    resolved_at: string | null;
    is_manual: boolean;
    created_at: string;
    updated_at: string;
    updates?: IncidentUpdate[];
    service?: ServiceBasic;
}

export interface IncidentUpdate {
    id: string;
    incident_id: string;
    status: string;
    message: string;
    created_at: string;
}

export interface ServiceBasic {
    id: string;
    name: string;
    status: string;
}

export interface MaintenanceWindow {
    id: string;
    title: string;
    description: string;
    service_ids: string[];
    scheduled_start: string;
    scheduled_end: string;
    created_at: string;
}

export interface AgentGroup {
    id: string;
    name: string;
    token?: string; // Only returned on creation
    created_at: string;
}

export interface Agent {
    id: string;
    name: string;
    group?: AgentGroup;
    status: 'pending' | 'active' | 'inactive';
    last_heartbeat: string | null;
    metadata: Record<string, unknown>;
    created_at: string;
}

export interface Service {
    id: string;
    agent_id: string | null;
    name: string;
    url: string;
    check_interval: number;
    timeout: number;
    retries: number;
    failure_threshold: number;
    status: string;
    is_public: boolean;
    group_name: string;
    created_at: string;
    updated_at: string;
    agent?: Agent;
}

export interface LoginResponse {
    token: string;
    name: string;
    email: string;
}

export interface NotificationChannel {
    id: string;
    name: string;
    type: 'email' | 'slack' | 'webhook';
    config: Record<string, any>;
    enabled: boolean;
    created_at: string;
    updated_at: string;
}
