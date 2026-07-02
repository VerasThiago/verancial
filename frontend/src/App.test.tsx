import React from 'react';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import App from './App';
import apiService from './services/api';

jest.mock('./services/api');
jest.mock('./components/Login', () => () => <div>Login Page</div>);
jest.mock('./components/Dashboard', () => () => <div>Dashboard Page</div>);
jest.mock('./components/Bank', () => () => <div>Bank Page</div>);

const mockedApiService = apiService as jest.Mocked<typeof apiService>;

const setPath = (path: string) => {
  window.history.pushState({}, '', path);
};

describe('App', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('renders the login page at /login regardless of auth state', () => {
    mockedApiService.isAuthenticated.mockReturnValue(false);
    setPath('/login');

    render(<App />);

    expect(screen.getByText('Login Page')).toBeInTheDocument();
  });

  it('redirects unauthenticated users away from /dashboard to /login', () => {
    mockedApiService.isAuthenticated.mockReturnValue(false);
    setPath('/dashboard');

    render(<App />);

    expect(screen.getByText('Login Page')).toBeInTheDocument();
    expect(screen.queryByText('Dashboard Page')).not.toBeInTheDocument();
  });

  it('renders the dashboard for authenticated users at /dashboard', () => {
    mockedApiService.isAuthenticated.mockReturnValue(true);
    setPath('/dashboard');

    render(<App />);

    expect(screen.getByText('Dashboard Page')).toBeInTheDocument();
  });

  it('redirects unauthenticated users away from /bank/:bankId to /login', () => {
    mockedApiService.isAuthenticated.mockReturnValue(false);
    setPath('/bank/11111111-1111-1111-1111-111111111111');

    render(<App />);

    expect(screen.getByText('Login Page')).toBeInTheDocument();
  });

  it('renders the bank page for authenticated users at /bank/:bankId', () => {
    mockedApiService.isAuthenticated.mockReturnValue(true);
    setPath('/bank/11111111-1111-1111-1111-111111111111');

    render(<App />);

    expect(screen.getByText('Bank Page')).toBeInTheDocument();
  });

  it('redirects the root path to /dashboard', () => {
    mockedApiService.isAuthenticated.mockReturnValue(true);
    setPath('/');

    render(<App />);

    expect(screen.getByText('Dashboard Page')).toBeInTheDocument();
  });
});
