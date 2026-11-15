import { API_BASE_URL } from "@/lib/api";
import { authedFetch, readError } from "@/lib/auth";
import type {
  EsewaFormFields,
  EsewaInitiateResponse,
  PaymentConfigResponse,
  PaymentMode,
} from "@/types/esewa";
import type { Order, OrderResponse } from "@/types/order";

/**
 * getPaymentMode reads which checkout flow is actually active. It is public
 * -- no session needed -- since the cart page has to know this before a
 * visitor is necessarily signed in.
 */
export async function getPaymentMode(): Promise<PaymentMode> {
  const res = await fetch(`${API_BASE_URL}/config/payment`);
  if (!res.ok) {
    throw new Error(`Could not load payment config (${res.status})`);
  }
  const body: PaymentConfigResponse = await res.json();
  return body.data.mode;
}

export type EsewaInitiateResult = {
  orderId: number;
  form: EsewaFormFields;
  formUrl: string;
};

/**
 * initiateEsewaPayment prices the cart server-side and returns everything
 * needed to build and auto-submit the hidden form that actually sends the
 * user to eSewa. It does not touch the cart or the library -- that only
 * happens once the payment is verified.
 */
export async function initiateEsewaPayment(): Promise<EsewaInitiateResult> {
  const res = await authedFetch("/checkout/esewa/initiate", { method: "POST" });
  if (!res.ok) {
    throw await readError(res, "Could not start the eSewa payment");
  }
  const body: EsewaInitiateResponse = await res.json();
  return { orderId: body.data.order_id, form: body.data.form, formUrl: body.data.form_url };
}

/**
 * verifyEsewaPayment sends eSewa's base64 "data" query param (from
 * success_url) to the API, which checks its signature, independently
 * confirms it with eSewa, and completes the order exactly once. Safe to
 * call again for the same data -- the API's own idempotency is what a
 * refreshed success page relies on.
 */
export async function verifyEsewaPayment(data: string): Promise<Order> {
  const res = await authedFetch("/checkout/esewa/verify", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ data }),
  });
  if (!res.ok) {
    throw await readError(res, "Could not verify the eSewa payment");
  }
  const body: OrderResponse = await res.json();
  return body.data;
}

/**
 * failEsewaPayment marks a pending payment cancelled, leaving the cart
 * untouched so the user can retry from the cart page.
 */
export async function failEsewaPayment(transactionUuid: string): Promise<void> {
  const res = await authedFetch("/checkout/esewa/failure", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ transaction_uuid: transactionUuid }),
  });
  if (!res.ok) {
    throw await readError(res, "Could not record the cancelled payment");
  }
}

/**
 * submitEsewaForm builds a hidden HTML form for fields and immediately
 * submits it, navigating the browser to eSewa's own form URL -- a real
 * cross-origin POST, which is the one thing a client-side fetch cannot do
 * on the user's behalf for a redirect-based payment flow like this.
 */
export function submitEsewaForm(formUrl: string, fields: EsewaFormFields): void {
  const form = document.createElement("form");
  form.method = "POST";
  form.action = formUrl;
  form.style.display = "none";

  for (const [name, value] of Object.entries(fields)) {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = name;
    input.value = value;
    form.appendChild(input);
  }

  document.body.appendChild(form);
  form.submit();
}
