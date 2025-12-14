import axios from 'axios';
import type { AxiosInstance } from 'axios';

export interface DriftEvent {
  id: string;
  resource_id: string;
  resource_type: string;
  type: 'created' | 'modified' | 'deleted';
  severity: 'low' | 'medium' | 'high' | 'critical';
  timestamp: string;
  root_cause?: {
    user_identity: string;
    event_name: string;
  };
}

export interface ImpactAnalysisResult {
  drift_event_id: string;
  affected_resource_count: number;
  blast_radius: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  recommendations: string[];
}

export interface DriftStats {
  total_count: number;
  by_severity: Record<string, number>;
  by_type: Record<string, number>;
  by_resource_type: Record<string, number>;
}

export interface ImpactStats {
  total_analyzed: number;
  avg_blast_radius: number;
  avg_affected_resources: number;
  top_affected_resource_types: Record<string, number>;
}

export interface HighImpactDrift {
  id: string;
  resource_id: string;
  resource_type: string;
  drift_type: string;
  timestamp: string;
  affected_resource_count: number;
  blast_radius: number;
  severity: string;
}

class APIClient {
  private client: AxiosInstance;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.client = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  }

  // Health check
  async health(): Promise<any> {
    const response = await this.client.get('/health');
    return response.data;
  }

  // Drift events
  async listDrifts(params?: {
    limit?: number;
    resource_type?: string;
    severity?: string;
    drift_type?: string;
    user_identity?: string;
    start_time?: string;
    end_time?: string;
  }): Promise<{ drifts: DriftEvent[]; count: number }> {
    const response = await this.client.get('/api/v1/drifts', { params });
    return response.data;
  }

  async getDrift(id: string): Promise<DriftEvent> {
    const response = await this.client.get(`/api/v1/drifts/${id}`);
    return response.data;
  }

  async getDriftStats(days: number = 7): Promise<{ stats: DriftStats; days: number }> {
    const response = await this.client.get('/api/v1/drifts/stats', {
      params: { days },
    });
    return response.data;
  }

  // Impact analysis
  async listImpactAnalysis(params?: {
    limit?: number;
    severity?: string;
    min_blast_radius?: number;
    min_affected_resources?: number;
    start_time?: string;
    end_time?: string;
  }): Promise<{ results: ImpactAnalysisResult[]; count: number }> {
    const response = await this.client.get('/api/v1/impact', { params });
    return response.data;
  }

  async getImpactAnalysis(driftEventId: string): Promise<ImpactAnalysisResult> {
    const response = await this.client.get(`/api/v1/impact/${driftEventId}`);
    return response.data;
  }

  async getImpactStats(days: number = 7): Promise<{ stats: ImpactStats; days: number }> {
    const response = await this.client.get('/api/v1/impact/stats', {
      params: { days },
    });
    return response.data;
  }

  async getHighImpactDrifts(days: number = 7, limit: number = 50): Promise<{
    drifts: HighImpactDrift[];
    count: number;
    days: number;
  }> {
    const response = await this.client.get('/api/v1/impact/high', {
      params: { days, limit },
    });
    return response.data;
  }
}

// Export singleton instance
export const apiClient = new APIClient();
export default apiClient;
