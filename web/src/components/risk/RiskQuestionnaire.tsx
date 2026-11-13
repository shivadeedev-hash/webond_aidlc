// Risk-suitability questionnaire (US-007). Five questions scored 1–5; the mean,
// rounded, yields a Risk Level 1–5. Phase-1 mock has no profile service, so the
// computed result is surfaced via onComplete for the page to display.

import { useState } from 'react';
import type { RiskLevel } from '../../types';
import { useI18n } from '../../i18n/useI18n';
import Button from '../ui/Button';

interface Question {
  id: string;
  text: string;
  options: { label: string; score: number }[];
}

const QUESTIONS: Question[] = [
  {
    id: 'q1',
    text: 'How long do you plan to keep your money invested?',
    options: [
      { label: 'Less than 1 year', score: 1 },
      { label: '1–3 years', score: 2 },
      { label: '3–5 years', score: 3 },
      { label: '5–10 years', score: 4 },
      { label: 'More than 10 years', score: 5 },
    ],
  },
  {
    id: 'q2',
    text: 'If your portfolio dropped 20% in a month, you would…',
    options: [
      { label: 'Sell everything', score: 1 },
      { label: 'Sell some', score: 2 },
      { label: 'Do nothing', score: 3 },
      { label: 'Buy a little more', score: 4 },
      { label: 'Buy a lot more', score: 5 },
    ],
  },
  {
    id: 'q3',
    text: 'Which best describes your investment goal?',
    options: [
      { label: 'Preserve capital', score: 1 },
      { label: 'Steady income', score: 2 },
      { label: 'Balanced growth', score: 3 },
      { label: 'Growth', score: 4 },
      { label: 'Aggressive growth', score: 5 },
    ],
  },
  {
    id: 'q4',
    text: 'How much investment experience do you have?',
    options: [
      { label: 'None', score: 1 },
      { label: 'A little', score: 2 },
      { label: 'Moderate', score: 3 },
      { label: 'Experienced', score: 4 },
      { label: 'Very experienced', score: 5 },
    ],
  },
  {
    id: 'q5',
    text: 'What share of your savings is this investment?',
    options: [
      { label: 'Almost all', score: 1 },
      { label: 'More than half', score: 2 },
      { label: 'About half', score: 3 },
      { label: 'Less than half', score: 4 },
      { label: 'A small portion', score: 5 },
    ],
  },
];

export default function RiskQuestionnaire({
  onComplete,
  onCancel,
}: {
  onComplete: (level: RiskLevel, answers: Record<string, number>) => void;
  onCancel: () => void;
}): JSX.Element {
  const { t } = useI18n();
  const [answers, setAnswers] = useState<Record<string, number>>({});

  const allAnswered = QUESTIONS.every((q) => answers[q.id] !== undefined);

  function submit(): void {
    const scores = QUESTIONS.map((q) => answers[q.id]);
    const mean = scores.reduce((s, n) => s + n, 0) / scores.length;
    const level = Math.min(5, Math.max(1, Math.round(mean))) as RiskLevel;
    onComplete(level, answers);
  }

  return (
    <div className="flex flex-col gap-5">
      <h3 className="text-base font-semibold">{t('page.profile.riskQuestionnaire')}</h3>
      {QUESTIONS.map((q, i) => (
        <fieldset key={q.id} className="flex flex-col gap-2">
          <legend className="text-sm font-medium">
            {i + 1}. {q.text}
          </legend>
          <div className="flex flex-col gap-1.5">
            {q.options.map((opt) => (
              <label key={opt.score} className="flex items-center gap-2 text-sm">
                <input
                  type="radio"
                  name={q.id}
                  value={opt.score}
                  checked={answers[q.id] === opt.score}
                  onChange={() => setAnswers((a) => ({ ...a, [q.id]: opt.score }))}
                  className="h-4 w-4 accent-[color:var(--color-primary)]"
                />
                <span>{opt.label}</span>
              </label>
            ))}
          </div>
        </fieldset>
      ))}

      <div className="flex gap-3">
        <Button variant="outline" className="flex-1" onClick={onCancel}>
          {t('common.cancel')}
        </Button>
        <Button variant="primary" className="flex-1" onClick={submit} disabled={!allAnswered}>
          {t('common.save')}
        </Button>
      </div>
    </div>
  );
}
