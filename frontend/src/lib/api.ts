import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// ─── Public API (no auth) ───────────────────────────────────────────

export const publicApi = axios.create({
    baseURL: API_URL,
    headers: { 'Content-Type': 'application/json' },
});

// ─── Admin API (JWT auth) ───────────────────────────────────────────

export const adminApi = axios.create({
    baseURL: API_URL,
    headers: { 'Content-Type': 'application/json' },
});

// Add JWT token to admin requests
adminApi.interceptors.request.use((config) => {
    if (typeof window !== 'undefined') {
        const token = localStorage.getItem('uptimehub_token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
    }
    return config;
});

// Handle 401 responses
adminApi.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401 && typeof window !== 'undefined') {
            localStorage.removeItem('uptimehub_token');
            window.location.href = '/admin/login';
        }
        return Promise.reject(error);
    }
);

// ─── API Functions ──────────────────────────────────────────────────

// Public
export const getStatus = () => publicApi.get('/api/status');
export const getServiceHistory = (id: string) => publicApi.get(`/api/status/${id}/history`);
export const getPublicIncidents = () => publicApi.get('/api/incidents');
export const getMaintenance = () => publicApi.get('/api/maintenance');

// Auth
export const login = (email: string, password: string) =>
    publicApi.post('/api/admin/login', { email, password });

// Admin — Services
export const getAdminServices = () => adminApi.get('/api/admin/services');
export const createService = (data: Record<string, unknown>) =>
    adminApi.post('/api/admin/services', data);
export const updateService = (id: string, data: Record<string, unknown>) =>
    adminApi.put(`/api/admin/services/${id}`, data);
export const deleteService = (id: string) =>
    adminApi.delete(`/api/admin/services/${id}`);

// Admin — Agents
export const getAdminAgents = () => adminApi.get('/api/admin/agents');

// Admin — Incidents
export const getAdminIncidents = () => adminApi.get('/api/admin/incidents');
export const createIncident = (data: Record<string, unknown>) =>
    adminApi.post('/api/admin/incidents', data);
export const updateIncident = (id: string, data: Record<string, unknown>) =>
    adminApi.put(`/api/admin/incidents/${id}`, data);

// Admin — Maintenance
export const getAdminMaintenance = () => adminApi.get('/api/admin/maintenance');
export const createMaintenance = (data: Record<string, unknown>) =>
    adminApi.post('/api/admin/maintenance', data);

// Admin — Alerts
export const getNotificationChannels = () => adminApi.get('/api/admin/alerts/channels');
export const createNotificationChannel = (data: Record<string, unknown>) =>
    adminApi.post('/api/admin/alerts/channels', data);
export const updateNotificationChannel = (id: string, data: Record<string, unknown>) =>
    adminApi.put(`/api/admin/alerts/channels/${id}`, data);
export const deleteNotificationChannel = (id: string) =>
    adminApi.delete(`/api/admin/alerts/channels/${id}`);
export const testNotificationChannel = (id: string) =>
    adminApi.post(`/api/admin/alerts/channels/${id}/test`);
