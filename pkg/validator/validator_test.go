package validator_test

import (
	"testing"

	pkgvalidator "github.com/faqihyugos/coffee-pos/pkg/validator"
)

type TestInput struct {
	Name  string `json:"name"  validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
	Price int64  `json:"price" validate:"required,min=1"`
}

func TestValidate_ValidInput(t *testing.T) {
	v := pkgvalidator.New()

	errs := v.Validate(TestInput{
		Name:  "Kopi Susu",
		Email: "kopi@coffeeshop.id",
		Price: 25000,
	})

	if errs != nil {
		t.Errorf("expected nil, got errors: %v", errs)
	}
}

func TestValidate_AllFieldsEmpty(t *testing.T) {
	v := pkgvalidator.New()

	errs := v.Validate(TestInput{})

	if errs == nil {
		t.Fatal("expected errors, got nil")
	}

	expectedKeys := []string{"name", "email", "price"}
	for _, key := range expectedKeys {
		if _, ok := errs[key]; !ok {
			t.Errorf("expected error for field %q, but not found in map", key)
		}
	}

	if len(errs) != len(expectedKeys) {
		t.Errorf("expected %d errors, got %d: %v", len(expectedKeys), len(errs), errs)
	}
}

func TestValidate_InvalidEmail(t *testing.T) {
	v := pkgvalidator.New()

	errs := v.Validate(TestInput{
		Name:  "Kopi Susu",
		Email: "bukan-email",
		Price: 25000,
	})

	if errs == nil {
		t.Fatal("expected errors, got nil")
	}

	msg, ok := errs["email"]
	if !ok {
		t.Fatal("expected error for field \"email\", but not found")
	}

	const want = "format email tidak valid"
	if msg != want {
		t.Errorf("email error: want %q, got %q", want, msg)
	}
}

func TestValidate_NameTooShort(t *testing.T) {
	v := pkgvalidator.New()

	errs := v.Validate(TestInput{
		Name:  "A",
		Email: "kopi@coffeeshop.id",
		Price: 25000,
	})

	if errs == nil {
		t.Fatal("expected errors, got nil")
	}

	msg, ok := errs["name"]
	if !ok {
		t.Fatal("expected error for field \"name\", but not found")
	}

	const want = "minimal 2 karakter"
	if msg != want {
		t.Errorf("name error: want %q, got %q", want, msg)
	}
}
