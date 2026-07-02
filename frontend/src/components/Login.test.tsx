import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import '@testing-library/jest-dom';
import Login from './Login';
import apiService from '../services/api';

jest.mock('../services/api');
const mockedApiService = apiService as jest.Mocked<typeof apiService>;

describe('Login', () => {
  let originalLocation: Location;

  beforeEach(() => {
    jest.clearAllMocks();
    originalLocation = window.location;
    // @ts-ignore
    delete window.location;
    // @ts-ignore
    window.location = { ...originalLocation, href: '' };
  });

  afterEach(() => {
    window.location = originalLocation;
  });

  it('renders email and password fields and a submit button', () => {
    render(<Login />);

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('submits the entered credentials and redirects to /dashboard on success', async () => {
    mockedApiService.login.mockResolvedValueOnce({ status: 'success', token: 'tok' });

    render(<Login />);

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'user@example.com' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'hunter2' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(mockedApiService.login).toHaveBeenCalledWith({
        email: 'user@example.com',
        password: 'hunter2',
      });
    });

    await waitFor(() => expect(window.location.href).toBe('/dashboard'));
  });

  it('shows a loading state while the login request is in flight', async () => {
    let resolveLogin: (value: any) => void = () => {};
    mockedApiService.login.mockReturnValueOnce(
      new Promise((resolve) => {
        resolveLogin = resolve;
      })
    );

    render(<Login />);

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'user@example.com' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'hunter2' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    expect(await screen.findByRole('button', { name: /signing in/i })).toBeDisabled();

    resolveLogin({ status: 'success', token: 'tok' });
    await waitFor(() => expect(window.location.href).toBe('/dashboard'));
  });

  it('shows the server-provided error message when login fails', async () => {
    mockedApiService.login.mockRejectedValueOnce({
      response: { data: { message: 'Invalid credentials' } },
    });

    render(<Login />);

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'user@example.com' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'wrong' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    expect(await screen.findByText('Invalid credentials')).toBeInTheDocument();
    expect(window.location.href).toBe('');
  });

  it('shows a generic error message when the failure has no server message', async () => {
    mockedApiService.login.mockRejectedValueOnce(new Error('network down'));

    render(<Login />);

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: 'user@example.com' } });
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: 'wrong' } });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    expect(await screen.findByText('Login failed. Please try again.')).toBeInTheDocument();
  });
});
