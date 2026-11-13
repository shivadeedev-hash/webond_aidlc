// Typed API error codes and an error-code → user message map (api-spec.md).
// The UI renders ApiClientError.message; services throw it on a non-OK envelope.

export type ApiErrorCode =
  | 'INSUFFICIENT_BALANCE'
  | 'INSUFFICIENT_HOLDING'
  | 'BOND_OFFER_INVALID'
  | 'QUOTE_EXPIRED'
  | 'CANCEL_NOT_ALLOWED'
  | 'RANGE_TOO_LARGE'
  | 'EXPORT_TOO_LARGE'
  | 'TXN_NOT_FOUND'
  | 'DUPLICATE_EMAIL'
  | 'RISK_ACK_REQUIRED'
  | 'SUITABILITY_REQUIRED'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'VALIDATION_ERROR'
  | 'PRICE_UNAVAILABLE'
  | 'NETWORK_ERROR'
  | 'UNKNOWN';

/** HTTP status associated with each error code (api-spec.md). */
export const ERROR_STATUS: Record<ApiErrorCode, number> = {
  INSUFFICIENT_BALANCE: 422,
  INSUFFICIENT_HOLDING: 422,
  BOND_OFFER_INVALID: 422,
  QUOTE_EXPIRED: 409,
  CANCEL_NOT_ALLOWED: 409,
  RANGE_TOO_LARGE: 400,
  EXPORT_TOO_LARGE: 400,
  TXN_NOT_FOUND: 404,
  DUPLICATE_EMAIL: 409,
  RISK_ACK_REQUIRED: 422,
  SUITABILITY_REQUIRED: 403,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  VALIDATION_ERROR: 400,
  PRICE_UNAVAILABLE: 409,
  NETWORK_ERROR: 0,
  UNKNOWN: 500,
};

/**
 * Default, translation-ready English messages per code. i18n keys mirror the
 * code as `errors.<CODE>`; see lib/i18n.ts. These defaults are the fallback.
 */
export const ERROR_MESSAGES: Record<ApiErrorCode, string> = {
  INSUFFICIENT_BALANCE: 'Insufficient wallet balance for this order.',
  INSUFFICIENT_HOLDING: 'You do not hold enough of this asset to sell.',
  BOND_OFFER_INVALID: 'This bond is not currently available for the requested amount.',
  QUOTE_EXPIRED: 'Price quote expired, please re-quote.',
  CANCEL_NOT_ALLOWED: 'This order can no longer be cancelled.',
  RANGE_TOO_LARGE: 'The selected date range is too large (max 12 months).',
  EXPORT_TOO_LARGE: 'Too many rows to export (max 10,000). Narrow your filters.',
  TXN_NOT_FOUND: 'Transaction not found.',
  DUPLICATE_EMAIL: 'An account with this email already exists.',
  RISK_ACK_REQUIRED: 'Please acknowledge the risk notice to continue.',
  SUITABILITY_REQUIRED: 'Please complete your risk assessment before trading.',
  UNAUTHORIZED: 'Please sign in to continue.',
  FORBIDDEN: 'You do not have permission to perform this action.',
  VALIDATION_ERROR: 'Please check the highlighted fields and try again.',
  PRICE_UNAVAILABLE: 'Price is temporarily unavailable, please try again shortly.',
  NETWORK_ERROR: 'Network error. Please check your connection and try again.',
  UNKNOWN: 'Something went wrong. Please try again.',
};

function isApiErrorCode(code: string): code is ApiErrorCode {
  return Object.prototype.hasOwnProperty.call(ERROR_STATUS, code);
}

/** Normalize any string code to a known ApiErrorCode (fallback: UNKNOWN). */
export function toApiErrorCode(code: string | undefined | null): ApiErrorCode {
  if (code && isApiErrorCode(code)) return code;
  return 'UNKNOWN';
}

/** Resolve a user-facing message for an error code. */
export function messageForCode(code: string | undefined | null): string {
  return ERROR_MESSAGES[toApiErrorCode(code)];
}

/** Error thrown by the API client / mock services on a non-OK result. */
export class ApiClientError extends Error {
  readonly code: ApiErrorCode;
  readonly status: number;
  readonly details?: Record<string, unknown>;
  readonly requestId?: string;

  constructor(
    code: string | undefined | null,
    message?: string,
    options?: { status?: number; details?: Record<string, unknown>; requestId?: string },
  ) {
    const normalized = toApiErrorCode(code);
    super(message || ERROR_MESSAGES[normalized]);
    this.name = 'ApiClientError';
    this.code = normalized;
    this.status = options?.status ?? ERROR_STATUS[normalized];
    this.details = options?.details;
    this.requestId = options?.requestId;
  }
}
