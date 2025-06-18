import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import apiService, { BankAccountStat } from '../services/api';
import '../styles/Bank.css';

const Bank: React.FC = () => {
  const { bankId } = useParams<{ bankId: string }>();
  const navigate = useNavigate();
  const [bankData, setBankData] = useState<BankAccountStat | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [showUploadModal, setShowUploadModal] = useState(false);

  useEffect(() => {
    loadBankData();
  }, [bankId]);

  const loadBankData = async () => {
    try {
      setIsLoading(true);
      const stats = await apiService.getDashboardStats();
      const bank = stats.bank_account_stats.find(stat => stat.bank_account.id === bankId);
      
      if (!bank) {
        setError('Bank account not found');
        return;
      }
      
      setBankData(bank);
    } catch (err: any) {
      setError('Failed to load bank data. Please try again.');
      console.error('Bank data error:', err);
    } finally {
      setIsLoading(false);
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

      alert('Statement uploaded successfully!');
      
      // Reset file selection
      setSelectedFile(null);
      const fileInput = document.getElementById('statement-upload') as HTMLInputElement;
      if (fileInput) fileInput.value = '';
      
      // Refresh bank data
      await loadBankData();
      
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