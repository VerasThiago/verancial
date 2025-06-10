import React, { useState, useEffect } from 'react';
import apiService, { UserDashboardStats, BankAccountStat } from '../services/api';
import '../styles/Dashboard.css';

const Dashboard: React.FC = () => {
  const [dashboardStats, setDashboardStats] = useState<UserDashboardStats | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      setIsLoading(true);
      const stats = await apiService.getDashboardStats();
      setDashboardStats(stats);
    } catch (err: any) {
      setError('Failed to load dashboard data. Please try again.');
      console.error('Dashboard error:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLogout = () => {
    apiService.logout();
  };

  const formatDaysOutdated = (days: number | null, transactionCount: number): string => {
    if (transactionCount === 0) return 'No data yet';
    if (days === null) return 'No recent data';
    if (days === 0) return 'Updated today';
    if (days === 1) return 'Updated yesterday';
    if (days <= 7) return `${days} days ago`;
    if (days <= 30) return `${days} days ago`;
    return `${days} days ago (stale)`;
  };

  const formatCurrency = (currency: string): string => {
    const currencyMap: { [key: string]: string } = {
      'USD': '$',
      'CAD': 'C$',
      'BRL': 'R$',
      'EUR': '€',
    };
    return currencyMap[currency] || currency;
  };

  if (isLoading) {
    return (
      <div className="dashboard-container">
        <div className="loading">Loading dashboard...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="dashboard-container">
        <div className="error-message">{error}</div>
        <button onClick={loadDashboardData} className="retry-button">
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <h1>Verancial Dashboard</h1>
        <button onClick={handleLogout} className="logout-button">
          Logout
        </button>
      </header>

      <main className="dashboard-content">
        <div className="stats-overview">
          <div className="stat-card">
            <h3>Bank Accounts</h3>
            <div className="stat-number">{dashboardStats?.total_bank_accounts || 0}</div>
          </div>
        </div>

        <section className="bank-accounts-section">
          <h2>Your Bank Accounts</h2>
          
          {dashboardStats?.bank_account_stats.length === 0 ? (
            <div className="empty-state">
              <p>No bank accounts found.</p>
              <p>Your connected bank accounts will appear here.</p>
            </div>
          ) : (
            <div className="bank-accounts-grid">
              {dashboardStats?.bank_account_stats.map((stat: BankAccountStat) => {
                const hasNoTransactions = stat.transaction_count === 0;
                const isFresh = stat.days_outdated !== null && stat.days_outdated <= 7;
                const isStale = stat.days_outdated !== null && stat.days_outdated > 30;
                
                return (
                  <div 
                    key={stat.bank_account.id} 
                    className="bank-account-card"
                    data-no-transactions={hasNoTransactions}
                    data-fresh={isFresh && !hasNoTransactions}
                    data-stale={isStale}
                  >
                    <div className="bank-info">
                      <h3>{stat.bank_account.display_name}</h3>
                      <div className="bank-details">
                        <span className="country-badge">
                          {stat.bank_account.country_code}
                        </span>
                        <span className="currency-badge">
                          {formatCurrency(stat.bank_account.currency)}
                        </span>
                      </div>
                    </div>
                    
                    <div className="transaction-stats">
                      <div className="stat-item">
                        <span className="stat-label">Transactions</span>
                        <span className={`stat-value ${hasNoTransactions ? 'transaction-count-zero' : ''}`}>
                          {stat.transaction_count}
                        </span>
                      </div>
                      
                      <div className="stat-item">
                        <span className="stat-label">Last Update</span>
                        <span className={`stat-value ${
                          hasNoTransactions ? '' : 
                          isFresh ? 'last-update-fresh' : 
                          isStale ? 'last-update-stale' : 
                          stat.days_outdated && stat.days_outdated > 7 ? 'outdated' : ''
                        }`}>
                          {formatDaysOutdated(stat.days_outdated, stat.transaction_count)}
                        </span>
                      </div>
                    </div>
                    
                    {isStale && !hasNoTransactions && (
                      <div className="warning-badge">
                        Data may be outdated
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </section>
      </main>
    </div>
  );
};

export default Dashboard; 