import axios, { AxiosInstance, AxiosResponse } from 'axios';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  status: string;
  token: string;
}

export interface BankAccount {
  id: string;
  name: string;
  display_name: string;
  country_code: string;
  currency: string;
  is_active: boolean;
}

export interface BankAccountStat {
  bank_account: BankAccount;
  transaction_count: number;
  last_transaction: string | null;
  days_outdated: number | null;
}

export interface UserDashboardStats {
  total_bank_accounts: number;
  bank_account_stats: BankAccountStat[];
}

export interface ApiResponse<T> {
  status: string;
  data?: T;
  message?: string;
}

class ApiService {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: process.env.REACT_APP_API_URL || '/api/v0',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include JWT token
    this.api.interceptors.request.use((config) => {
      const token = this.getToken();
      if (token) {
        config.headers.Authorization = token;
      }
      return config;
    });

    // Add response interceptor to handle authentication errors
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          this.removeToken();
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Authentication methods
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    // Use a separate axios instance for login to avoid auth interceptors
    const loginApi = axios.create({
      baseURL: process.env.REACT_APP_LOGIN_URL || '/login',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const response: AxiosResponse<LoginResponse> = await loginApi.post(
      '/v0/user/signin',
      credentials
    );
    
    if (response.data.token) {
      this.setToken(response.data.token);
    }
    
    return response.data;
  }

  logout(): void {
    this.removeToken();
    window.location.href = '/login';
  }

  // Dashboard methods
  async getDashboardStats(): Promise<UserDashboardStats> {
    const response: AxiosResponse<ApiResponse<UserDashboardStats>> = await this.api.get('/dashboard');
    return response.data.data!;
  }





  // Token management
  private setToken(token: string): void {
    localStorage.setItem('auth_token', token);
  }

  private getToken(): string | null {
    return localStorage.getItem('auth_token');
  }

  private removeToken(): void {
    localStorage.removeItem('auth_token');
  }

  isAuthenticated(): boolean {
    return this.getToken() !== null;
  }
}

export default new ApiService(); 