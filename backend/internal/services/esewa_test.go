package services

import "testing"

// testEsewaSecret is eSewa's own published sandbox UAT secret -- the same
// value this app defaults ESEWA_SECRET_KEY to -- not a secret of this app's.
const testEsewaSecret = "8gBm/:&EnhH.1/q"

// TestEsewaSignatureMessage pins down the one detail eSewa's docs are easy
// to misread: the signed message is "name=value" pairs joined by commas,
// not the bare values. The exact string below is eSewa's own canonical
// illustration of the format on its ePay v2 docs page.
func TestEsewaSignatureMessage(t *testing.T) {
	message := esewaSignatureMessage(esewaRequestSignedFields, map[string]string{
		"total_amount":     "100",
		"transaction_uuid": "11-201-13",
		"product_code":     "EPAYTEST",
	})

	want := "total_amount=100,transaction_uuid=11-201-13,product_code=EPAYTEST"
	if message != want {
		t.Errorf("message = %q, want %q", message, want)
	}
}

// TestEsewaSign is the known-input/known-output case. The expected
// signature was not lifted from eSewa's own docs page -- its one fully
// worked example does not actually reproduce under the UAT secret it
// documents alongside it, which the next test's comment covers -- so this
// value was instead computed independently with two separate
// implementations (Python's hmac/hashlib and the openssl CLI, both
// agreeing) over eSewa's own canonical message-format example, and is
// pinned here as a regression guard.
func TestEsewaSign(t *testing.T) {
	message := "total_amount=100,transaction_uuid=11-201-13,product_code=EPAYTEST"
	want := "5DZywcrTKD0gia/rsSMcrRHmJl+4Tbol6S+lWgdJ94E="

	if got := esewaSign(testEsewaSecret, message); got != want {
		t.Errorf("esewaSign() = %q, want %q", got, want)
	}
}

// TestBuildEsewaFormFields checks the two things that would otherwise sink
// a real transaction silently: the amount string in the form is the exact
// same string the signature was computed over (not just numerically
// equal), and the signature itself matches what esewaSign produces for
// that message.
func TestBuildEsewaFormFields(t *testing.T) {
	form := buildEsewaFormFields(testEsewaSecret, "EPAYTEST", "11-201-13", 100, "https://example.com/success", "https://example.com/failure")

	if form.Amount != form.TotalAmount {
		t.Errorf("amount %q and total_amount %q must be the same string", form.Amount, form.TotalAmount)
	}
	if form.Amount != "100.00" {
		t.Errorf("amount = %q, want %q", form.Amount, "100.00")
	}

	wantMessage := "total_amount=100.00,transaction_uuid=11-201-13,product_code=EPAYTEST"
	wantSignature := esewaSign(testEsewaSecret, wantMessage)
	if form.Signature != wantSignature {
		t.Errorf("signature = %q, want %q (over %q)", form.Signature, wantSignature, wantMessage)
	}
	if form.SignedFieldNames != "total_amount,transaction_uuid,product_code" {
		t.Errorf("signed_field_names = %q", form.SignedFieldNames)
	}
}

// esewaResponseFields builds a fake success-callback payload, signed the
// same way eSewa signs its real ones, for tests to tamper with.
func esewaResponseFields(secret string) map[string]string {
	fields := map[string]string{
		"transaction_code": "0LD5CEH",
		"status":           "COMPLETE",
		"total_amount":     "1,000.0",
		"transaction_uuid": "240613-134231",
		"product_code":     "EPAYTEST",
	}
	signedFieldNames := []string{"transaction_code", "status", "total_amount", "transaction_uuid", "product_code", "signed_field_names"}
	fields["signed_field_names"] = "transaction_code,status,total_amount,transaction_uuid,product_code,signed_field_names"
	fields["signature"] = esewaSign(secret, esewaSignatureMessage(signedFieldNames, fields))
	return fields
}

// TestVerifyEsewaSignature covers both directions: a genuine response
// verifies, and a response with any single field changed after signing --
// whether that field is itself part of signed_field_names, like the status,
// or the signature field directly -- does not.
func TestVerifyEsewaSignature(t *testing.T) {
	t.Run("valid response verifies", func(t *testing.T) {
		fields := esewaResponseFields(testEsewaSecret)
		if !verifyEsewaSignature(testEsewaSecret, fields) {
			t.Error("a genuine response failed to verify")
		}
	})

	t.Run("tampered status is rejected", func(t *testing.T) {
		fields := esewaResponseFields(testEsewaSecret)
		fields["status"] = "PENDING"
		if verifyEsewaSignature(testEsewaSecret, fields) {
			t.Error("a tampered status field verified anyway")
		}
	})

	t.Run("tampered amount is rejected", func(t *testing.T) {
		fields := esewaResponseFields(testEsewaSecret)
		fields["total_amount"] = "1.0"
		if verifyEsewaSignature(testEsewaSecret, fields) {
			t.Error("a tampered total_amount field verified anyway")
		}
	})

	t.Run("forged signature is rejected", func(t *testing.T) {
		fields := esewaResponseFields(testEsewaSecret)
		fields["signature"] = "not-a-real-signature"
		if verifyEsewaSignature(testEsewaSecret, fields) {
			t.Error("a forged signature verified anyway")
		}
	})

	t.Run("wrong secret is rejected", func(t *testing.T) {
		fields := esewaResponseFields(testEsewaSecret)
		if verifyEsewaSignature("some-other-secret", fields) {
			t.Error("a response signed with a different secret verified anyway")
		}
	})
}

func TestGenerateTransactionUUID(t *testing.T) {
	a, err := generateTransactionUUID()
	if err != nil {
		t.Fatalf("generateTransactionUUID: %v", err)
	}
	b, err := generateTransactionUUID()
	if err != nil {
		t.Fatalf("generateTransactionUUID: %v", err)
	}

	if a == b {
		t.Error("two calls produced the same transaction_uuid")
	}

	for _, r := range a {
		isLetterOrDigit := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
		if !isLetterOrDigit && r != '-' {
			t.Fatalf("transaction_uuid %q contains %q, which is not a letter, digit, or hyphen", a, r)
		}
	}
}
