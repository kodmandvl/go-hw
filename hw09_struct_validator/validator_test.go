package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	BadStruct1 struct {
		Name  string `validate:"len:qwerty"`
		Email string
	}

	BadStruct2 struct {
		Name  string
		Email string `validate:"regexp:^[a-z)$"`
	}

	SomeInt int

	SomeString string
)

// Успешные кейсы.
func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "dfa60991-f3c5-457d-b059-9efa977ab716", // UUID, 36 characters
				Name:   "Ivan Ivanov",
				Age:    33,
				Email:  "ivanivanov@google.com",
				Role:   "stuff",
				Phones: []string{"79001234567", "79007654321"},
			},
			expectedErr: nil,
		},
		{
			in: App{
				Version: "1.3.0",
			},
			expectedErr: nil,
		},
		{
			in: Token{
				Header:    []byte("some header"),
				Payload:   []byte("some payload"),
				Signature: []byte("some signature"),
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 200,
				Body: "Success message",
			},
			expectedErr: nil,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()
			err := Validate(tt.in)
			require.Equal(t, err, tt.expectedErr)
		})
	}
}

// Ошибочные кейсы (ошибки валидации, ошибка компиляции regexp, ошибка в теге).
func TestValidate2(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "dfa60991-f3c5-457d-b059-9efa977ab716",
				Name:   "Ivan Ivanov",
				Age:    33,
				Email:  "ivan!ivanov@google.com",
				Role:   "admin",
				Phones: []string{"79001234567"},
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "Email",
					Err:   fmt.Errorf("%w: ^\\w+@\\w+\\.\\w+$", ErrRegexp),
				},
			},
		},
		{
			in: User{
				ID:     "dfa60991-f3c5-457d-b059-9efa977ab716",
				Name:   "Ivan Ivanov",
				Age:    99,
				Email:  "ivanivanov@google.com",
				Role:   "hunter",
				Phones: []string{"7900"},
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "Age",
					Err:   fmt.Errorf("%w: must be less than or equal to 50", ErrMax),
				},
				ValidationError{
					Field: "Role",
					Err:   fmt.Errorf("%w: must be one of [admin stuff]", ErrIn),
				},
				ValidationError{
					Field: "Phones[0]",
					Err:   fmt.Errorf("%w: expected 11, actual 4", ErrLen),
				},
			},
		},
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz7890",
				Name:   "Ivan Ivanov",
				Age:    14,
				Email:  "ivan,ivanov@google.com",
				Role:   "admin",
				Phones: []string{"7900", "79001234567", "78005678"},
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "ID",
					Err:   fmt.Errorf("%w: expected 36, actual 30", ErrLen),
				},
				ValidationError{
					Field: "Age",
					Err:   fmt.Errorf("%w: must be greater than or equal to 18", ErrMin),
				},
				ValidationError{
					Field: "Email",
					Err:   fmt.Errorf("%w: ^\\w+@\\w+\\.\\w+$", ErrRegexp),
				},
				ValidationError{
					Field: "Phones[0]",
					Err:   fmt.Errorf("%w: expected 11, actual 4", ErrLen),
				},
				ValidationError{
					Field: "Phones[2]",
					Err:   fmt.Errorf("%w: expected 11, actual 8", ErrLen),
				},
			},
		},
		{
			in: App{
				Version: "1.3.0.4",
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "Version",
					Err:   fmt.Errorf("%w: expected 5, actual 7", ErrLen),
				},
			},
		},
		{
			in: Response{
				Code: 777,
				Body: "Wonderful",
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "Code",
					Err:   fmt.Errorf("%w: must be one of [200 404 500]", ErrIn),
				},
			},
		},
		{
			in: BadStruct1{
				Name:  "Ivan Ivanov",
				Email: "ivanivanov@google.com",
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "Name",
					Err:   fmt.Errorf("%w: qwerty", ErrInvalidTag),
				},
			},
		},
		{
			in: BadStruct2{
				Name:  "Ivan Ivanov",
				Email: "ivanivanov@google.com",
			},
			expectedErr: ValidationErrors{
				ValidationError{
					Field: "Email",
					Err:   fmt.Errorf("%w: ^[a-z)$", ErrInvalidRegexp),
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()
			err := Validate(tt.in)
			if !errors.As(err, &ValidationErrors{}) {
				t.Errorf("unexpected error: actual %v, desired %v", err, tt.expectedErr)
			} else {
				require.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

// Ошибочные кейсы (ошибка при проверке, что интерфейс - не структура).
func TestValidate3(t *testing.T) {
	SomeMap := map[string]string{
		"A": "Atomicity",
		"C": "Consistency",
		"I": "Isolation",
		"D": "Durability",
	}
	SomeSlice := []string{"Atomicity", "Consistency", "Isolation", "Durability"}
	var sv SomeString = "Some String"
	var iv SomeInt = 50
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in:          SomeMap,
			expectedErr: ErrNotStruct,
		},
		{
			in:          SomeSlice,
			expectedErr: ErrNotStruct,
		},
		{
			in:          sv,
			expectedErr: ErrNotStruct,
		},
		{
			in:          iv,
			expectedErr: ErrNotStruct,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()
			err := Validate(tt.in)
			require.ErrorIs(t, err, ErrNotStruct)
		})
	}
}
