import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter } from 'react-router-dom';
import Dashboard from './Dashboard';
import apiService, { UserDashboardStats } from '../services/api';

jest.mock('../services/api');
const mockedApiService = apiService as jest.Mocked<typeof apiService>;

const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}));

const renderDashboard = () =>
  render(
    <MemoryRouter>
      <Dashboard />
    </MemoryRouter>
  );

const baseStats: UserDashboardStats = {
  total_bank_accounts: 1,
  bank_account_stats: [
    {
      bank_account: {
        id: 'bank-1',
        name: 'scotiabank',
        display_name: 'Scotiabank Chequing',
        country_code: 'CA',
        currency: 'CAD',
        is_active: true,
      },
      transaction_count: 5,
      last_transaction: '2024-01-01',
      days_outdated: 2,
    },
  ],
};

describe('Dashboard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('shows a loading state while stats are being fetched', async () => {
    let resolveStats: (v: UserDashboardStats) => void = () => {};
    mockedApiService.getDashboardStats.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveStats = resolve;
      })
    );

    renderDashboard();

    expect(screen.getByText(/loading dashboard/i)).toBeInTheDocument();
    resolveStats(baseStats);
    await waitFor(() => expect(screen.queryByText(/loading dashboard/i)).not.toBeInTheDocument());
  });

  it('renders bank account stats once loaded', async () => {
    mockedApiService.getDashboardStats.mockResolvedValueOnce(baseStats);

    renderDashboard();

    expect(await screen.findByText('Scotiabank Chequing')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
    expect(screen.getByText('CA')).toBeInTheDocument();
  });

  it('renders an empty state when there are no bank accounts', async () => {
    mockedApiService.getDashboardStats.mockResolvedValueOnce({
      total_bank_accounts: 0,
      bank_account_stats: [],
    });

    renderDashboard();

    expect(await screen.findByText(/no bank accounts found/i)).toBeInTheDocument();
  });

  it('shows an error state with a retry button when the fetch fails, and retries on click', async () => {
    mockedApiService.getDashboardStats
      .mockRejectedValueOnce(new Error('boom'))
      .mockResolvedValueOnce(baseStats);

    renderDashboard();

    expect(await screen.findByText(/failed to load dashboard data/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /retry/i }));

    expect(await screen.findByText('Scotiabank Chequing')).toBeInTheDocument();
    expect(mockedApiService.getDashboardStats).toHaveBeenCalledTimes(2);
  });

  it('navigates to /bank/:id when a bank account card is clicked', async () => {
    mockedApiService.getDashboardStats.mockResolvedValueOnce(baseStats);

    renderDashboard();

    fireEvent.click(await screen.findByText('Scotiabank Chequing'));

    expect(mockNavigate).toHaveBeenCalledWith('/bank/bank-1');
  });

  it('calls apiService.logout when the logout button is clicked', async () => {
    mockedApiService.getDashboardStats.mockResolvedValueOnce(baseStats);

    renderDashboard();

    fireEvent.click(await screen.findByRole('button', { name: /logout/i }));

    expect(mockedApiService.logout).toHaveBeenCalledTimes(1);
  });

  it('flags a stale bank account (days_outdated > 30) with a warning badge', async () => {
    mockedApiService.getDashboardStats.mockResolvedValueOnce({
      total_bank_accounts: 1,
      bank_account_stats: [
        {
          ...baseStats.bank_account_stats[0],
          days_outdated: 45,
        },
      ],
    });

    renderDashboard();

    expect(await screen.findByText(/data may be outdated/i)).toBeInTheDocument();
    expect(screen.getByText('45 days ago (stale)')).toBeInTheDocument();
  });

  it('shows "No data yet" for a bank account with zero transactions', async () => {
    mockedApiService.getDashboardStats.mockResolvedValueOnce({
      total_bank_accounts: 1,
      bank_account_stats: [
        {
          ...baseStats.bank_account_stats[0],
          transaction_count: 0,
          days_outdated: null,
        },
      ],
    });

    renderDashboard();

    expect(await screen.findByText('No data yet')).toBeInTheDocument();
  });

  it.each([
    [0, 'Updated today'],
    [1, 'Updated yesterday'],
    [null, 'No recent data'],
    [15, '15 days ago'],
  ])('formats last-update text for days_outdated=%p as "%s"', async (daysOutdated, expected) => {
    mockedApiService.getDashboardStats.mockResolvedValueOnce({
      total_bank_accounts: 1,
      bank_account_stats: [
        { ...baseStats.bank_account_stats[0], days_outdated: daysOutdated },
      ],
    });

    renderDashboard();

    expect(await screen.findByText(expected)).toBeInTheDocument();
  });

  it('falls back to the raw currency code for an unmapped currency', async () => {
    mockedApiService.getDashboardStats.mockResolvedValueOnce({
      total_bank_accounts: 1,
      bank_account_stats: [
        {
          ...baseStats.bank_account_stats[0],
          bank_account: { ...baseStats.bank_account_stats[0].bank_account, currency: 'JPY' },
        },
      ],
    });

    renderDashboard();

    expect(await screen.findByText('JPY')).toBeInTheDocument();
  });
});
