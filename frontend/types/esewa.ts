/** EsewaFormFields is every field that must be POSTed to eSewa's form URL. */
export type EsewaFormFields = {
  amount: string;
  tax_amount: string;
  total_amount: string;
  transaction_uuid: string;
  product_code: string;
  product_service_charge: string;
  product_delivery_charge: string;
  success_url: string;
  failure_url: string;
  signed_field_names: string;
  signature: string;
};

export type EsewaInitiateResponse = {
  data: {
    order_id: number;
    form: EsewaFormFields;
    form_url: string;
  };
};

export type PaymentMode = "mock" | "esewa";

export type PaymentConfigResponse = {
  data: {
    mode: PaymentMode;
  };
};
