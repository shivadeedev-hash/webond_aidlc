// History page actions: Excel export of the filtered list and the monthly PDF
// statement (US-028). Both call the live API and hand the file to the browser's
// download flow. In mock mode the files do not exist, so the controls stay
// disabled with an explanatory tooltip.

import { useMemo, useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { historyService } from '../../services/history';
import { DATA_SOURCE } from '../../lib/constants';
import { ApiClientError } from '../../lib/errors';
import { saveFile } from '../../lib/saveFile';
import { useI18n } from '../../i18n/useI18n';
import type { HistoryQuery } from '../../types';
import Button from '../ui/Button';
import { recentMonths } from './months';

export default function HistoryDownloads({
  query,
  enabled = DATA_SOURCE === 'api',
}: {
  /** The active list filters; the export uses the same ones (without paging). */
  query: HistoryQuery;
  enabled?: boolean;
}): JSX.Element {
  const { t, locale } = useI18n();
  const months = useMemo(() => recentMonths(12, locale), [locale]);
  const [month, setMonth] = useState(months[0]?.value ?? '');

  const exporting = useMutation({
    mutationFn: () => historyService.exportTransactions(query),
    onSuccess: saveFile,
  });
  const statement = useMutation({
    mutationFn: () => historyService.downloadStatement(month),
    onSuccess: saveFile,
  });

  const error = exporting.error ?? statement.error;
  const errorText = error
    ? error instanceof ApiClientError
      ? error.message
      : t('page.history.downloadError')
    : null;
  const hint = enabled ? undefined : t('page.history.exportHint');

  return (
    <div className="flex flex-col items-start gap-2 sm:items-end">
      <div className="flex flex-wrap items-center gap-2">
        <span title={hint}>
          <Button
            variant="outline"
            disabled={!enabled || exporting.isPending}
            onClick={() => {
              statement.reset();
              exporting.mutate();
            }}
          >
            {exporting.isPending ? t('page.history.exporting') : t('page.history.export')}
          </Button>
        </span>

        <label className="flex items-center gap-2 text-sm">
          <span className="sr-only">{t('page.history.statementMonth')}</span>
          <select
            value={month}
            onChange={(e) => setMonth(e.target.value)}
            disabled={!enabled}
            className="min-h-[44px] rounded-md border border-border bg-surface px-2 disabled:opacity-60"
            aria-label={t('page.history.statementMonth')}
          >
            {months.map((m) => (
              <option key={m.value} value={m.value}>
                {m.label}
              </option>
            ))}
          </select>
        </label>
        <span title={hint}>
          <Button
            variant="outline"
            disabled={!enabled || statement.isPending || month === ''}
            onClick={() => {
              exporting.reset();
              statement.mutate();
            }}
          >
            {statement.isPending
              ? t('page.history.statementPreparing')
              : t('page.history.statement')}
          </Button>
        </span>
      </div>
      {errorText && (
        <p role="alert" className="text-sm text-danger">
          {errorText}
        </p>
      )}
    </div>
  );
}
