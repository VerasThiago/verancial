import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import '@testing-library/jest-dom';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import Bank from './Bank';
import apiService, { BankAccountStat, Transaction } from '../services/api';

jest.mock('../services/api');
const mockedApiService = apiService as jest.Mocked<typeof apiService>;

const bankId = '11111111-1111-1111-1111-111111111111';

const baseBankStat: BankAccountStat = {
  bank_account: {
    id: bankId,
    name: 'scotiabank',
    display_name: 'Scotiabank Chequing',
    country_code: 'CA',
    currency: 'CAD',
    is_active: true,
  },
  transaction_count: 3,
  last_transaction: '2024-01-01',
  days_outdated: 1,
};

const baseTransaction: Transaction = {
  id: 'tx-1',
  date: '2024-01-01T00:00:00Z',
  amount: -12.5,
  payee: 'Coffee Shop',
  description: 'Coffee Shop purchase',
  category: '',
  currency: 'CAD',
  bankid: bankId,
  metadata: {},
};

const renderBank = () =>
  render(
    <MemoryRouter initialEntries={[`/bank/${bankId}`]}>
      <Routes>
        <Route path="/bank/:bankId" element={<Bank />} />
      </Routes>
    </MemoryRouter>
  );

describe('Bank', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockedApiService.listTransactions.mockResolvedValue([]);
    window.alert = jest.fn();
  });

  it('shows a loading state while bank data is being fetched', async () => {
    let resolveBank: (v: BankAccountStat) => void = () => {};
    mockedApiService.getBankStats.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveBank = resolve;
      })
    );

    renderBank();

    expect(screen.getByText(/loading bank details/i)).toBeInTheDocument();
    resolveBank(baseBankStat);
    await waitFor(() => expect(screen.queryByText(/loading bank details/i)).not.toBeInTheDocument());
  });

  it('renders bank details once loaded', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);

    renderBank();

    expect(await screen.findByText('Scotiabank Chequing')).toBeInTheDocument();
    expect(screen.getByText('CA')).toBeInTheDocument();
    expect(mockedApiService.getBankStats).toHaveBeenCalledWith(bankId);
  });

  it('shows an error state with a back button when bank data fails to load, and clears on retry via navigation back', async () => {
    mockedApiService.getBankStats.mockRejectedValueOnce(new Error('boom'));

    renderBank();

    expect(await screen.findByText(/failed to load bank data/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /back to dashboard/i })).toBeInTheDocument();
  });

  it('loads uncategorized transactions for the bank and renders them in a table', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);
    mockedApiService.listTransactions.mockResolvedValueOnce([baseTransaction]);

    renderBank();

    await screen.findByText('Scotiabank Chequing');
    expect(await screen.findByText('Coffee Shop purchase')).toBeInTheDocument();
    expect(mockedApiService.listTransactions).toHaveBeenCalledWith({
      bankId,
      page: 1,
      pageSize: 50,
      uncategorized: true,
    });
  });

  it('shows an empty state when there are no uncategorized transactions', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);
    mockedApiService.listTransactions.mockResolvedValueOnce([]);

    renderBank();

    expect(await screen.findByText(/no uncategorized transactions found/i)).toBeInTheDocument();
  });

  it('advances to the next page and refetches with the new page number', async () => {
    mockedApiService.getBankStats.mockResolvedValue(baseBankStat);
    mockedApiService.listTransactions.mockResolvedValue(
      Array.from({ length: 50 }, (_, i) => ({ ...baseTransaction, id: `tx-${i}` }))
    );

    renderBank();

    await screen.findByText('Scotiabank Chequing');
    await waitFor(() => expect(mockedApiService.listTransactions).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1 })
    ));

    fireEvent.click(screen.getByRole('button', { name: /next/i }));

    await waitFor(() =>
      expect(mockedApiService.listTransactions).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 })
      )
    );
  });

  it('rejects a non-CSV file selection with an alert and does not set a selected file', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(screen.getByRole('button', { name: /upload statement/i }));

    const file = new File(['not a csv'], 'statement.txt', { type: 'text/plain' });
    const input = document.getElementById('statement-upload') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [file] } });

    expect(window.alert).toHaveBeenCalledWith('Please upload a CSV file');
    expect(screen.queryByText('statement.txt')).not.toBeInTheDocument();
  });

  it('accepts a CSV file, uploads it, and shows a success message', async () => {
    mockedApiService.getBankStats.mockResolvedValue(baseBankStat);
    mockedApiService.uploadStatement.mockResolvedValueOnce(undefined);

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(screen.getByRole('button', { name: /upload statement/i }));

    const file = new File(['date,amount\n2024-01-01,10'], 'statement.csv', { type: 'text/csv' });
    const input = document.getElementById('statement-upload') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [file] } });

    expect(await screen.findByText('statement.csv')).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole('button', { name: /^upload statement$/i }).slice(-1)[0]);

    await waitFor(() => expect(mockedApiService.uploadStatement).toHaveBeenCalledWith(
      expect.objectContaining({ bankId, fileName: 'statement.csv' })
    ));
    expect(await screen.findByText(/statement uploaded successfully/i)).toBeInTheDocument();
  });

  it('alerts and does not crash when the upload request fails', async () => {
    mockedApiService.getBankStats.mockResolvedValue(baseBankStat);
    mockedApiService.uploadStatement.mockRejectedValueOnce(new Error('upload failed'));

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(screen.getByRole('button', { name: /upload statement/i }));

    const file = new File(['date,amount\n2024-01-01,10'], 'statement.csv', { type: 'text/csv' });
    const input = document.getElementById('statement-upload') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [file] } });

    await screen.findByText('statement.csv');
    fireEvent.click(screen.getAllByRole('button', { name: /^upload statement$/i }).slice(-1)[0]);

    await waitFor(() =>
      expect(window.alert).toHaveBeenCalledWith('Failed to upload statement. Please try again.')
    );
  });

  it('cancel button clears the selected file without uploading', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(screen.getByRole('button', { name: /upload statement/i }));

    const file = new File(['x'], 'statement.csv', { type: 'text/csv' });
    const input = document.getElementById('statement-upload') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [file] } });

    expect(await screen.findByText('statement.csv')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));

    expect(mockedApiService.uploadStatement).not.toHaveBeenCalled();
  });

  it('the back-to-dashboard button on the error screen navigates away from the bank route', async () => {
    mockedApiService.getBankStats.mockRejectedValueOnce(new Error('boom'));

    renderBank();
    await screen.findByText(/failed to load bank data/i);

    fireEvent.click(screen.getByRole('button', { name: /back to dashboard/i }));

    expect(screen.queryByText(/failed to load bank data/i)).not.toBeInTheDocument();
  });

  it('the back-to-dashboard button in the header navigates to /dashboard', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(screen.getByRole('button', { name: /back to dashboard/i }));

    // Bank unmounts once the route changes away from /bank/:bankId.
    expect(screen.queryByText('Scotiabank Chequing')).not.toBeInTheDocument();
  });

  it('logs and does not crash when loading transactions fails', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);
    mockedApiService.listTransactions.mockReset();
    mockedApiService.listTransactions.mockRejectedValueOnce(new Error('boom'));

    renderBank();

    expect(await screen.findByText('Scotiabank Chequing')).toBeInTheDocument();
    expect(await screen.findByText(/no uncategorized transactions found/i)).toBeInTheDocument();
  });

  it('clears the selection when the file input is cleared without choosing a file', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(screen.getByRole('button', { name: /upload statement/i }));

    const input = document.getElementById('statement-upload') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [] } });

    expect(screen.getByText('Choose CSV File')).toBeInTheDocument();
  });

  it('closing the modal via the × button resets the selected file', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(screen.getByRole('button', { name: /upload statement/i }));

    const file = new File(['x'], 'statement.csv', { type: 'text/csv' });
    const input = document.getElementById('statement-upload') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [file] } });
    expect(await screen.findByText('statement.csv')).toBeInTheDocument();

    fireEvent.click(screen.getByText('×'));

    fireEvent.click(screen.getByRole('button', { name: /upload statement/i }));
    expect(screen.getByText('Choose CSV File')).toBeInTheDocument();
  });

  it('changing a transaction category select does not throw (handler is a documented no-op)', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);
    mockedApiService.listTransactions.mockReset();
    mockedApiService.listTransactions.mockResolvedValueOnce([baseTransaction]);

    renderBank();
    await screen.findByText('Coffee Shop purchase');

    fireEvent.change(screen.getByDisplayValue('Select category'), { target: { value: 'food' } });
  });

  it('the Previous button steps back a page, floored at page 1', async () => {
    mockedApiService.getBankStats.mockResolvedValue(baseBankStat);
    mockedApiService.listTransactions.mockReset();
    mockedApiService.listTransactions.mockResolvedValue(
      Array.from({ length: 50 }, (_, i) => ({ ...baseTransaction, id: `tx-${i}` }))
    );

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(await screen.findByRole('button', { name: /next/i }));
    await waitFor(() =>
      expect(mockedApiService.listTransactions).toHaveBeenCalledWith(expect.objectContaining({ page: 2 }))
    );

    fireEvent.click(await screen.findByRole('button', { name: /previous/i }));
    await waitFor(() =>
      expect(mockedApiService.listTransactions).toHaveBeenCalledWith(expect.objectContaining({ page: 1 }))
    );
  });

  it('the success message auto-hides after 3 seconds', async () => {
    jest.useFakeTimers();
    mockedApiService.getBankStats.mockResolvedValue(baseBankStat);
    mockedApiService.uploadStatement.mockResolvedValueOnce(undefined);

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    fireEvent.click(screen.getByRole('button', { name: /upload statement/i }));
    const file = new File(['x'], 'statement.csv', { type: 'text/csv' });
    const input = document.getElementById('statement-upload') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [file] } });
    fireEvent.click(screen.getAllByRole('button', { name: /^upload statement$/i }).slice(-1)[0]);

    expect(await screen.findByText(/statement uploaded successfully/i)).toBeInTheDocument();

    act(() => {
      jest.advanceTimersByTime(3000);
    });

    expect(screen.queryByText(/statement uploaded successfully/i)).not.toBeInTheDocument();
    jest.useRealTimers();
  });

  it.each([
    [{ transaction_count: 0, days_outdated: null }, 'No data yet'],
    [{ transaction_count: 2, days_outdated: 0 }, 'Updated today'],
    [{ transaction_count: 2, days_outdated: 7 }, '7 days ago'],
    [{ transaction_count: 2, days_outdated: 30 }, '30 days ago'],
    [{ transaction_count: 2, days_outdated: 31 }, '31 days ago (stale)'],
  ])('formats last-update text for %j as %s', async (overrides, expected) => {
    mockedApiService.getBankStats.mockResolvedValueOnce({ ...baseBankStat, ...overrides });

    renderBank();

    expect(await screen.findByText(expected)).toBeInTheDocument();
  });

  it.each([
    [2, 'fresh'],
    [15, ''],
    [45, 'stale'],
  ])('applies the last-update css class for days_outdated=%i (expect "%s")', async (daysOutdated, expectedClass) => {
    mockedApiService.getBankStats.mockResolvedValueOnce({
      ...baseBankStat,
      days_outdated: daysOutdated,
    });

    renderBank();
    await screen.findByText('Scotiabank Chequing');

    const detailValue = screen.getByText(/days ago/).closest('span');
    expect(detailValue).toHaveClass(`detail-value ${expectedClass}`.trim());
  });

  it('renders a positive transaction amount without the negative css class', async () => {
    mockedApiService.getBankStats.mockResolvedValueOnce(baseBankStat);
    mockedApiService.listTransactions.mockReset();
    mockedApiService.listTransactions.mockResolvedValueOnce([{ ...baseTransaction, amount: 25 }]);

    renderBank();
    await screen.findByText('Coffee Shop purchase');

    const amountCell = screen.getByText('C$25.00');
    expect(amountCell).toHaveClass('positive');
  });
});
