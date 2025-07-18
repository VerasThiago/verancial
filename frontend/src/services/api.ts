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

export interface UploadStatementRequest {
  bankId: string;
  fileName: string;
  fileData: string; // Base64 encoded CSV content
}

export interface Transaction {
  id: string;
  date: string;
  amount: number;
  payee: string;
  description: string;
  category: string;
  currency: string;
  bankid: string;
  metadata: Record<string, string>;
}

export interface ListTransactionsRequest {
  bankId: string;
  page?: number;
  pageSize?: number;
  uncategorized?: boolean;
}

export interface ListTransactionsResponse {
  status: string;
  data: Transaction[];
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
    const response: AxiosResponse<ApiResponse<UserDashboardStats>> = await this.api.get('/dashboard/user');
    return response.data.data!;
  }

  async getBankStats(bankId: string): Promise<BankAccountStat> {
    const response: AxiosResponse<ApiResponse<BankAccountStat>> = await this.api.get(`/bank/${bankId}`);
    return response.data.data!;
  }

  async uploadStatement(request: UploadStatementRequest): Promise<void> {
    // Get the current user token to extract user ID
    const token = this.getToken();
    if (!token) {
      throw new Error('No authentication token found');
    }

    // Decode JWT token to get user ID (simple base64 decode of payload)
    const payload = JSON.parse(atob(token.split('.')[1]));
    
    // Based on the Go JWT structure: JWTClaim has User field with ID
    const userId = payload.User?.id;

    if (!userId) {
      console.error('JWT payload structure:', payload);
      throw new Error('User ID not found in token. Expected User.id field.');
    }

    const reportRequest = {
      userid: userId,
      filedata: request.fileData,
      filename: request.fileName,
      bankid: request.bankId,
      asyncprocessing: false
    };

    await this.api.post('/report/process', reportRequest);
  }

  // Transaction methods
  async listTransactions(request: ListTransactionsRequest): Promise<Transaction[]> {
    // Ensure bankId is a clean UUID without any extra characters
    const cleanUUID = request.bankId.replace(/[^a-fA-F0-9-]/g, '');
    if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(cleanUUID)) {
      throw new Error('Invalid bank ID format. Must be a valid UUID.');
    }
    
    const params = new URLSearchParams({
      ...(request.page && { page: request.page.toString() }),
      ...(request.pageSize && { pageSize: request.pageSize.toString() }),
      ...(request.uncategorized && { uncategorized: request.uncategorized.toString() }),
    });

    const response: AxiosResponse<ListTransactionsResponse> = await this.api.get(
      `/transaction/list/${cleanUUID}?${params.toString()}`
    );
    return response.data.data;
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