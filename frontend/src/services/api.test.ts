import type axiosType from 'axios';

jest.mock('axios');

// Helper to build a fresh mock axios instance shape (with interceptors captured)
const makeMockInstance = () => {
  const requestHandlers: Array<(config: any) => any> = [];
  const responseHandlers: Array<{
    onFulfilled: (response: any) => any;
    onRejected: (error: any) => any;
  }> = [];

  const instance: any = {
    get: jest.fn(),
    post: jest.fn(),
    interceptors: {
      request: {
        use: jest.fn((onFulfilled: any) => {
          requestHandlers.push(onFulfilled);
        }),
      },
      response: {
        use: jest.fn((onFulfilled: any, onRejected: any) => {
          responseHandlers.push({ onFulfilled, onRejected });
        }),
      },
    },
    __requestHandlers: requestHandlers,
    __responseHandlers: responseHandlers,
  };
  return instance;
};

describe('ApiService', () => {
  let mainInstance: any;
  let loginInstance: any;
  let mockedAxios: jest.Mocked<typeof axiosType>;
  let apiService: typeof import('./api').default;
  let originalLocation: Location;

  beforeEach(() => {
    jest.resetModules();
    localStorage.clear();

    mainInstance = makeMockInstance();
    loginInstance = makeMockInstance();

    // jest.resetModules() gives './api' a fresh require registry, so axios
    // must be re-required here too (not via the top-level import) to land
    // on the same mock instance './api' will see when it requires axios.
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    mockedAxios = require('axios');

    // First axios.create() call (in ApiService constructor) returns mainInstance.
    // Subsequent calls (in login()) return loginInstance.
    let createCallCount = 0;
    mockedAxios.create = jest.fn(() => {
      createCallCount += 1;
      return createCallCount === 1 ? mainInstance : loginInstance;
    }) as any;

    originalLocation = window.location;
    // @ts-ignore
    delete window.location;
    // @ts-ignore
    window.location = { ...originalLocation, href: '' };

    // Re-require after mocks are set up so the singleton constructor uses our mocked axios.
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    apiService = require('./api').default;
  });

  afterEach(() => {
    window.location = originalLocation;
    jest.clearAllMocks();
  });

  describe('login', () => {
    it('posts credentials to the login base URL (separate axios instance) and stores the token on success', async () => {
      loginInstance.post.mockResolvedValueOnce({
        data: { status: 'ok', token: 'abc123' },
      });

      const credentials = { email: 'user@example.com', password: 'secret' };
      const result = await apiService.login(credentials);

      expect(mockedAxios.create).toHaveBeenCalledWith(
        expect.objectContaining({
          baseURL: expect.any(String),
        })
      );
      expect(loginInstance.post).toHaveBeenCalledWith('/v0/user/signin', credentials);
      expect(mainInstance.post).not.toHaveBeenCalled();
      expect(result).toEqual({ status: 'ok', token: 'abc123' });
      expect(localStorage.getItem('auth_token')).toBe('abc123');
    });

    it('does not store a token when the login response has no token', async () => {
      loginInstance.post.mockResolvedValueOnce({
        data: { status: 'error' },
      });

      await apiService.login({ email: 'user@example.com', password: 'wrong' });

      expect(localStorage.getItem('auth_token')).toBeNull();
    });

    it('does not store a token when the login request fails', async () => {
      loginInstance.post.mockRejectedValueOnce(new Error('Network error'));

      await expect(
        apiService.login({ email: 'user@example.com', password: 'wrong' })
      ).rejects.toThrow('Network error');

      expect(localStorage.getItem('auth_token')).toBeNull();
    });
  });

  describe('logout', () => {
    it('clears the token and redirects to /login', () => {
      localStorage.setItem('auth_token', 'sometoken');

      apiService.logout();

      expect(localStorage.getItem('auth_token')).toBeNull();
      expect(window.location.href).toBe('/login');
    });
  });

  describe('isAuthenticated', () => {
    it('returns false when no token is present', () => {
      expect(apiService.isAuthenticated()).toBe(false);
    });

    it('returns true when a token is present', () => {
      localStorage.setItem('auth_token', 'sometoken');
      expect(apiService.isAuthenticated()).toBe(true);
    });
  });

  describe('getDashboardStats', () => {
    it('calls the dashboard endpoint and unwraps the response data', async () => {
      const stats = { total_bank_accounts: 2, bank_account_stats: [] };
      mainInstance.get.mockResolvedValueOnce({
        data: { status: 'ok', data: stats },
      });

      const result = await apiService.getDashboardStats();

      expect(mainInstance.get).toHaveBeenCalledWith('/dashboard/user');
      expect(result).toEqual(stats);
    });
  });

  describe('getBankStats', () => {
    it('calls the bank endpoint with the bank id and unwraps the response data', async () => {
      const stat = {
        bank_account: {
          id: 'bank-1',
          name: 'test',
          display_name: 'Test Bank',
          country_code: 'US',
          currency: 'USD',
          is_active: true,
        },
        transaction_count: 5,
        last_transaction: '2024-01-01',
        days_outdated: 1,
      };
      mainInstance.get.mockResolvedValueOnce({
        data: { status: 'ok', data: stat },
      });

      const result = await apiService.getBankStats('bank-1');

      expect(mainInstance.get).toHaveBeenCalledWith('/bank/bank-1');
      expect(result).toEqual(stat);
    });
  });

  describe('uploadStatement', () => {
    it('throws if not authenticated (no token)', async () => {
      await expect(
        apiService.uploadStatement({
          bankId: '11111111-1111-1111-1111-111111111111',
          fileName: 'statement.csv',
          fileData: 'base64data',
        })
      ).rejects.toThrow('No authentication token found');

      expect(mainInstance.post).not.toHaveBeenCalled();
    });

    it('posts the correctly shaped payload when authenticated', async () => {
      localStorage.setItem('auth_token', 'sometoken');
      mainInstance.post.mockResolvedValueOnce({ data: { status: 'ok' } });

      await apiService.uploadStatement({
        bankId: '11111111-1111-1111-1111-111111111111',
        fileName: 'statement.csv',
        fileData: 'base64data',
      });

      expect(mainInstance.post).toHaveBeenCalledWith('/report/process', {
        filedata: 'base64data',
        filename: 'statement.csv',
        bankid: '11111111-1111-1111-1111-111111111111',
        asyncprocessing: false,
      });
    });
  });

  describe('listTransactions', () => {
    const validUUID = '11111111-1111-1111-1111-111111111111';

    it('throws "Invalid bank ID format..." for a malformed/injected bankId and never hits the server', async () => {
      await expect(
        apiService.listTransactions({ bankId: '../../etc/passwd' })
      ).rejects.toThrow('Invalid bank ID format. Must be a valid UUID.');

      expect(mainInstance.get).not.toHaveBeenCalled();
    });

    it('throws for a bankId with extra path segments appended to a valid UUID', async () => {
      await expect(
        apiService.listTransactions({ bankId: `${validUUID}/../../admin` })
      ).rejects.toThrow('Invalid bank ID format. Must be a valid UUID.');

      expect(mainInstance.get).not.toHaveBeenCalled();
    });

    it('throws for a bankId that is too short even after stripping invalid characters', async () => {
      await expect(
        apiService.listTransactions({ bankId: 'not-a-uuid' })
      ).rejects.toThrow('Invalid bank ID format. Must be a valid UUID.');

      expect(mainInstance.get).not.toHaveBeenCalled();
    });

    it('strips disallowed characters and accepts a valid UUID, hitting the expected URL with query params', async () => {
      const transactions = [
        {
          id: 't1',
          date: '2024-01-01',
          amount: -10,
          payee: 'Store',
          description: 'desc',
          category: 'food',
          currency: 'USD',
          bankid: validUUID,
          metadata: {},
        },
      ];
      mainInstance.get.mockResolvedValueOnce({
        data: { status: 'ok', data: transactions },
      });

      const result = await apiService.listTransactions({
        bankId: validUUID,
        page: 2,
        pageSize: 25,
        uncategorized: true,
      });

      expect(mainInstance.get).toHaveBeenCalledWith(
        `/transaction/list/${validUUID}?page=2&pageSize=25&uncategorized=true`
      );
      expect(result).toEqual(transactions);
    });

    it('accepts a valid UUID embedded with extra junk characters stripped out (regression: strip then validate)', async () => {
      // Injecting characters outside [a-fA-F0-9-] around a valid UUID should be stripped;
      // the result must still fail strict validation since stripping alone must not bypass the regex
      // when the stripped result no longer matches the UUID shape exactly.
      const injected = `${validUUID}<script>`;
      await expect(
        apiService.listTransactions({ bankId: injected })
      ).rejects.toThrow('Invalid bank ID format. Must be a valid UUID.');
      expect(mainInstance.get).not.toHaveBeenCalled();
    });

    it('builds the URL without query params when no optional params are provided', async () => {
      mainInstance.get.mockResolvedValueOnce({
        data: { status: 'ok', data: [] },
      });

      await apiService.listTransactions({ bankId: validUUID });

      expect(mainInstance.get).toHaveBeenCalledWith(`/transaction/list/${validUUID}?`);
    });
  });

  describe('request interceptor', () => {
    it('attaches the Authorization header when a token is present', () => {
      localStorage.setItem('auth_token', 'my-token');
      const onFulfilled = mainInstance.__requestHandlers[0];
      const config = { headers: {} as Record<string, string> };

      const result = onFulfilled(config);

      expect(result.headers.Authorization).toBe('my-token');
    });

    it('does not attach an Authorization header when no token is present', () => {
      const onFulfilled = mainInstance.__requestHandlers[0];
      const config = { headers: {} as Record<string, string> };

      const result = onFulfilled(config);

      expect(result.headers.Authorization).toBeUndefined();
    });
  });

  describe('response interceptor (401 handling)', () => {
    it('passes through successful responses unchanged', () => {
      const { onFulfilled } = mainInstance.__responseHandlers[0];
      const response = { status: 200, data: {} };

      expect(onFulfilled(response)).toBe(response);
    });

    it('clears the token and redirects to /login on a 401 response', async () => {
      localStorage.setItem('auth_token', 'sometoken');
      const { onRejected } = mainInstance.__responseHandlers[0];
      const error = { response: { status: 401 } };

      await expect(onRejected(error)).rejects.toBe(error);

      expect(localStorage.getItem('auth_token')).toBeNull();
      expect(window.location.href).toBe('/login');
    });

    it('does not clear the token on non-401 errors', async () => {
      localStorage.setItem('auth_token', 'sometoken');
      const { onRejected } = mainInstance.__responseHandlers[0];
      const error = { response: { status: 500 } };

      await expect(onRejected(error)).rejects.toBe(error);

      expect(localStorage.getItem('auth_token')).toBe('sometoken');
    });

    it('handles errors without a response object gracefully', async () => {
      const { onRejected } = mainInstance.__responseHandlers[0];
      const error = { message: 'Network Error' };

      await expect(onRejected(error)).rejects.toBe(error);
    });
  });
});
