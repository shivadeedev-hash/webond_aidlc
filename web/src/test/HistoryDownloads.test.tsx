import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import HistoryDownloads from '../components/history/HistoryDownloads';
import { ApiClientError } from '../lib/errors';
import { renderWithProviders } from './renderApp';

const { exportTransactions, downloadStatement, saveFile } = vi.hoisted(() => ({
  exportTransactions: vi.fn(),
  downloadStatement: vi.fn(),
  saveFile: vi.fn(),
}));

vi.mock('../services/history', () => ({
  historyService: { exportTransactions, downloadStatement },
}));
vi.mock('../lib/saveFile', () => ({ saveFile }));

const FILE = { blob: new Blob(['x']), filename: 'webond-transactions-20261028.xlsx' };

describe('HistoryDownloads (US-028)', () => {
  beforeEach(() => {
    exportTransactions.mockReset();
    downloadStatement.mockReset();
    saveFile.mockReset();
  });

  it('exports with the active filters and hands the file to the browser', async () => {
    exportTransactions.mockResolvedValueOnce(FILE);
    renderWithProviders(
      <HistoryDownloads query={{ tab: 'orders', q: 'abc', sort: 'date' }} enabled />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Export Excel' }));

    await waitFor(() => expect(saveFile).toHaveBeenCalledTimes(1));
    expect(exportTransactions.mock.calls[0][0]).toEqual({ tab: 'orders', q: 'abc', sort: 'date' });
    expect(saveFile.mock.calls[0][0]).toEqual(FILE);
  });

  it('downloads the statement for the month picked in the list', async () => {
    downloadStatement.mockResolvedValueOnce({ blob: new Blob(['%PDF']), filename: 's.pdf' });
    renderWithProviders(<HistoryDownloads query={{}} enabled />);

    const select = screen.getByRole('combobox', { name: 'Statement month' });
    const options = screen.getAllByRole('option') as HTMLOptionElement[];
    expect(options).toHaveLength(12);
    const older = options[2].value;
    fireEvent.change(select, { target: { value: older } });
    fireEvent.click(screen.getByRole('button', { name: 'Statement' }));

    await waitFor(() => expect(downloadStatement).toHaveBeenCalledWith(older));
    await waitFor(() => expect(saveFile).toHaveBeenCalledTimes(1));
    expect(older).toMatch(/^\d{6}$/);
  });

  it('shows the API error instead of failing silently', async () => {
    exportTransactions.mockRejectedValueOnce(new ApiClientError('EXPORT_TOO_LARGE'));
    renderWithProviders(<HistoryDownloads query={{}} enabled />);

    fireEvent.click(screen.getByRole('button', { name: 'Export Excel' }));

    const alert = await screen.findByRole('alert');
    expect(alert.textContent?.length ?? 0).toBeGreaterThan(0);
    expect(saveFile).not.toHaveBeenCalled();
  });

  it('falls back to a generic message for unexpected errors', async () => {
    downloadStatement.mockRejectedValueOnce(new Error('boom'));
    renderWithProviders(<HistoryDownloads query={{}} enabled />);

    fireEvent.click(screen.getByRole('button', { name: 'Statement' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Download failed');
  });

  it('is disabled with an explanation when there is no live API (mock mode)', () => {
    renderWithProviders(<HistoryDownloads query={{}} enabled={false} />);

    expect(screen.getByRole('button', { name: 'Export Excel' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Statement' })).toBeDisabled();
    expect(screen.getByRole('combobox', { name: 'Statement month' })).toBeDisabled();
    expect(screen.getAllByTitle('Enabled once the live API is connected.').length).toBeGreaterThan(
      0,
    );
  });
});
