package providercommand

import (
	"errors"
	"reflect"
	"testing"
)

func TestEveryCommandFieldParticipatesInReplayIdentity(t *testing.T) {
	original := Command{OperationID: "op_1", AttemptID: "attempt_1", ProviderConnectionID: "conn_eu", Provider: "simulator_a", AmountMinor: 12900, Currency: "USD", PaymentTokenRef: "opaque_token_reference", CaptureMethod: "automatic"}
	if err := VerifyReplay(original, original); err != nil {
		t.Fatal(err)
	}
	// Reflection makes a future newly added field fail this gate if it is not
	// serialized into the digest; the implementation remains the book's code.
	for i := 0; i < reflect.TypeOf(original).NumField(); i++ {
		incoming := original
		field := reflect.ValueOf(&incoming).Elem().Field(i)
		switch field.Kind() {
		case reflect.String:
			field.SetString(field.String() + "_changed")
		case reflect.Int64:
			field.SetInt(field.Int() + 1)
		default:
			t.Fatalf("add a mutation for field %s", reflect.TypeOf(original).Field(i).Name)
		}
		if err := VerifyReplay(original, incoming); !errors.Is(err, ErrOperationConflict) {
			t.Fatalf("field %s drift not rejected: %v", reflect.TypeOf(original).Field(i).Name, err)
		}
	}
}
