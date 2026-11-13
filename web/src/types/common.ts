// Shared primitives and the API envelope used across all domain DTOs.
// Money/unit values are transported as DECIMAL strings (api-spec.md) to avoid
// float precision loss; the UI parses them via lib/format when displaying.

/** A DECIMAL(18,4) value serialized as a string, e.g. "1250400.0000". */
export type Decimal = string;

/** ISO-8601 instant, e.g. "2026-10-28T03:00:00Z". */
export type ISODateTime = string;

/** ISO-8601 calendar date, e.g. "2026-10-28". */
export type ISODate = string;

export type AssetType = 'fund' | 'bond' | 'gold';

export type OrderSide = 'buy' | 'sell';

/** Success envelope payload metadata (pagination etc.). */
export interface Meta {
  requestId?: string;
  page?: number;
  pageSize?: number;
  total?: number;
  hasMore?: boolean;
}

/** Typed error shape returned inside the API error envelope. */
export interface ApiErrorBody {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

/** Success response envelope: { data, meta? }. */
export interface ApiSuccess<T> {
  data: T;
  meta?: Meta;
}

/** Error response envelope: { error, meta? }. */
export interface ApiError {
  error: ApiErrorBody;
  meta?: Meta;
}

export type ApiEnvelope<T> = ApiSuccess<T> | ApiError;

/** A page of rows plus pagination metadata (used by list services). */
export interface Paginated<T> {
  items: T[];
  meta: Required<Pick<Meta, 'page' | 'pageSize' | 'total' | 'hasMore'>>;
}

/** A file received from the API (xlsx export, pdf statement). */
export interface DownloadedFile {
  blob: Blob;
  filename: string;
}
