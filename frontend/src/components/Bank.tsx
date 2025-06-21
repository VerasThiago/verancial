import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import apiService, { BankAccountStat, Transaction } from '../services/api';
import '../styles/Bank.css';

const Bank: React.FC = () => {
  const { bankId } = useParams<{ bankId: string }>();
  const navigate = useNavigate();
  const [bankData, setBankData] = useState<BankAccountStat | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingTransactions, setIsLoadingTransactions] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [showUploadModal, setShowUploadModal] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize] = useState(50);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    loadBankData();
  }, [bankId]);

  useEffect(() => {
    loadTransactions();
  }, [bankId, currentPage]);

  const loadBankData = async () => {
    try {
      setIsLoading(true);
      const bank = await apiService.getBankStats(bankId!);
      setBankData(bank);
    } catch (err: any) {
      setError('Failed to load bank data. Please try again.');
      console.error('Bank data error:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const loadTransactions = async () => {
    if (!bankId) return;
    
    try {
      setIsLoadingTransactions(true);
      console.log('Loading transactions for bankId:', bankId);
      const data = await apiService.listTransactions({
        bankId,
        page: currentPage,
        pageSize,
        uncategorized: true
      });
      setTransactions(data);
    } catch (err: any) {
      console.error('Failed to load transactions:', err);
    } finally {
      setIsLoadingTransactions(false);
    }
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) {
      setSelectedFile(null);
      return;
    }

    // Validate file type
    if (!file.name.toLowerCase().endsWith('.csv')) {
      alert('Please upload a CSV file');
      event.target.value = '';
      setSelectedFile(null);
      return;
    }

    setSelectedFile(file);
  };

  const showSuccessMessage = (message: string) => {
    setSuccessMessage(message);
    setTimeout(() => {
      setSuccessMessage(null);
    }, 3000); // Hide after 3 seconds
  };

  const handleUpload = async () => {
    if (!selectedFile) return;

    setIsUploading(true);
    try {
      // TODO: Change to presigned URL design
      const fileContent = await readFileAsText(selectedFile);
      const base64Content = btoa(fileContent); // Convert to base64
      
      await apiService.uploadStatement({
        bankId: bankId!,
        fileName: selectedFile.name,
        fileData: base64Content
      });
      
      // Reset file selection
      setSelectedFile(null);
      const fileInput = document.getElementById('statement-upload') as HTMLInputElement;
      if (fileInput) fileInput.value = '';
      
      // Show success message
      showSuccessMessage('Statement uploaded successfully! Refreshing data...');
      
      // Refresh both bank data and transactions
      await Promise.all([
        loadBankData(),
        loadTransactions()
      ]);
      
    } catch (err: any) {
      console.error('Upload error:', err);
      alert('Failed to upload statement. Please try again.');
    } finally {
      setIsUploading(false);
    }
  };

  const handleCancelUpload = () => {
    setSelectedFile(null);
    const fileInput = document.getElementById('statement-upload') as HTMLInputElement;
    if (fileInput) fileInput.value = '';
  };

  const readFileAsText = (file: File): Promise<string> => {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result as string);
      reader.onerror = reject;
      reader.readAsText(file);
    });
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

  const formatDaysOutdated = (days: number | null, transactionCount: number): string => {
    if (transactionCount === 0) return 'No data yet';
    if (days === null) return 'No recent data';
    if (days === 0) return 'Updated today';
    if (days === 1) return 'Updated yesterday';
    if (days <= 7) return `${days} days ago`;
    if (days <= 30) return `${days} days ago`;
    return `${days} days ago (stale)`;
  };

  const handleCategoryChange = (transactionId: string, category: string) => {
    // Implement category change logic
  };

  if (isLoading) {
    return (
      <div className="bank-container">
        <div className="loading">Loading bank details...</div>
      </div>
    );
  }

  if (error || !bankData) {
    return (
      <div className="bank-container">
        <div className="error-message">{error || 'Bank not found'}</div>
        <button onClick={() => navigate('/dashboard')} className="back-button">
          Back to Dashboard
        </button>
      </div>
    );
  }

  return (
    <div className="bank-container">
      {successMessage && (
        <div className="success-message">
          {successMessage}
        </div>
      )}
      <header className="bank-header">
        <button onClick={() => navigate('/dashboard')} className="back-button">
          ← Back to Dashboard
        </button>
        <h1>{bankData.bank_account.display_name}</h1>
      </header>

      <main className="bank-content">
        <div className="bank-info-card">
          <div className="bank-details">
            <div className="detail-row">
              <span className="detail-label">Country:</span>
              <span className="country-badge">{bankData.bank_account.country_code}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Currency:</span>
              <span className="currency-badge">{formatCurrency(bankData.bank_account.currency)}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Transactions:</span>
              <span className={`detail-value ${bankData.transaction_count === 0 ? 'no-data' : ''}`}>
                {bankData.transaction_count}
              </span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Last Update:</span>
              <span className={`detail-value ${
                bankData.transaction_count === 0 ? '' : 
                bankData.days_outdated && bankData.days_outdated <= 7 ? 'fresh' : 
                bankData.days_outdated && bankData.days_outdated > 30 ? 'stale' : ''
              }`}>
                {formatDaysOutdated(bankData.days_outdated, bankData.transaction_count)}
              </span>
            </div>
          </div>
          <button className="upload-modal-trigger" onClick={() => setShowUploadModal(true)}>
            Upload Statement
          </button>
        </div>

        <div className="transactions-section">
          <h2>Uncategorized Transactions</h2>
          {isLoadingTransactions ? (
            <div className="loading">Loading transactions...</div>
          ) : transactions.length === 0 ? (
            <div className="no-transactions">No uncategorized transactions found</div>
          ) : (
            <>
              <div className="transactions-table">
                <table>
                  <thead>
                    <tr>
                      <th>Date</th>
                      <th>Description</th>
                      <th>Amount</th>
                      <th>Category</th>
                    </tr>
                  </thead>
                  <tbody>
                    {transactions.map((transaction) => (
                      <tr key={transaction.id}>
                        <td>{new Date(transaction.date).toLocaleDateString()}</td>
                        <td>{transaction.description}</td>
                        <td className={transaction.amount < 0 ? 'negative' : 'positive'}>
                          {formatCurrency(transaction.currency)}{Math.abs(transaction.amount).toFixed(2)}
                        </td>
                        <td>
                          <select
                            value={transaction.category || ''}
                            onChange={(e) => handleCategoryChange(transaction.id, e.target.value)}
                          >
                            <option value="">Select category</option>
                            {/* TODO: Add category options */}
                          </select>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <div className="pagination">
                <button
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                >
                  Previous
                </button>
                <span>Page {currentPage}</span>
                <button
                  onClick={() => setCurrentPage(p => p + 1)}
                  disabled={transactions.length < pageSize}
                >
                  Next
                </button>
              </div>
            </>
          )}
        </div>

        {/* Modal Dialog for Upload */}
        {showUploadModal && (
          <div className="modal-overlay">
            <div className="modal-content">
              <button className="modal-close" onClick={() => { setShowUploadModal(false); handleCancelUpload(); }}>&times;</button>
              <div className="upload-section modal-upload-section">
                <h2>Upload Statement</h2>
                <p>Upload your CSV bank statement to update transaction data</p>
                <div className="upload-area">
                  <input
                    type="file"
                    id="statement-upload"
                    accept=".csv"
                    onChange={handleFileSelect}
                    disabled={isUploading}
                    className="file-input"
                  />
                  <label htmlFor="statement-upload" className={`upload-button ${isUploading || selectedFile ? 'has-file' : ''}`}>
                    {selectedFile ? selectedFile.name : 'Choose CSV File'}
                  </label>
                  {selectedFile && (
                    <div className="file-actions">
                      <button 
                        onClick={async () => { await handleUpload(); setShowUploadModal(false); }} 
                        disabled={isUploading}
                        className={`upload-submit-button ${isUploading ? 'uploading' : ''}`}
                      >
                        {isUploading ? 'Uploading...' : 'Upload Statement'}
                      </button>
                      <button 
                        onClick={handleCancelUpload} 
                        disabled={isUploading}
                        className="cancel-button"
                      >
                        Cancel
                      </button>
                    </div>
                  )}
                </div>
                <div className="upload-note">
                  <p><strong>Note:</strong> Please upload CSV files only. The file should contain your bank statement data.</p>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
};

export default Bank; 