export interface Agent {
  id: string;
  name: string;
  description: string;
  version: string;
  hostname: string;
  status: 'online' | 'offline';
  last_seen_at?: string;
  created_at: string;
  cpu_usage_percent?: number;
  memory_used_bytes?: number;
  memory_total_bytes?: number;
  disk_used_bytes?: number;
  disk_total_bytes?: number;
  running_containers?: number;
}
