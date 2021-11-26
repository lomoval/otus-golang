package hw09structvalidator

import (
	"encoding/json"
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
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Regexp struct {
		Str1 string `validate:"regexp:\\d+|len:5"`
	}

	Float struct {
		F1 float32 `validate:"min:5.5"`
		F2 float64 `validate:"min:5.5|max:6.6"`
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

	Nested struct {
		Response `validate:"nested"`
		User     `validate:"nested"`
		Num      int8 `validate:"min:5|max:6"`
	}
)

func TestValidateCorrectValues(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in:          App{Version: "1.0.0"},
			expectedErr: nil,
		},
		{
			in: Nested{
				Response: Response{Code: 500, Body: ""},
				User: User{
					ID:     "e6bb694e-0669-4ad6-9920-41162213e92d",
					Name:   "",
					Age:    36,
					Email:  "test1@test.ru",
					Role:   "stuff",
					Phones: []string{"12345678901", "09876543210"},
				},
				Num: 5,
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "e6bb694e-0669-4ad6-9920-41162213e92d",
				Name:   "",
				Age:    18,
				Email:  "test@test.ru",
				Role:   "stuff",
				Phones: []string{"12345678901", "09876543210"},
				meta:   []byte{1},
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "e6bb694e-0669-4ad6-9920-41162213e92d",
				Name:   "",
				Age:    50,
				Email:  "test@test.ru",
				Role:   "admin",
				Phones: []string{"12345678901", "09876543210"},
				meta:   []byte{1},
			},
			expectedErr: nil,
		},
		{
			in:          Response{Code: 404, Body: ""},
			expectedErr: nil,
		},
		{
			in: Token{
				Header:    []byte{1},
				Payload:   []byte{1, 2},
				Signature: []byte{1, 2, 3},
			},
			expectedErr: nil,
		},
		{
			in:          Float{F1: 5.5, F2: 6.1},
			expectedErr: nil,
		},
		{
			in:          Float{F1: 15.5, F2: 6.6},
			expectedErr: nil,
		},
		{
			in: Regexp{Str1: "12345"},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			if tt.expectedErr == nil {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestValidateIncorrectValues(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: App{Version: "1.0"},
			expectedErr: ValidationErrors{
				ValidationError{Field: "Version", Err: ErrValidateIncorrectLen},
			},
		},
		{
			in: User{
				ID:     "1",
				Name:   "testName",
				Age:    5,
				Email:  "incorrect-email",
				Role:   "user",
				Phones: []string{"1234567890"},
				meta:   []byte{1},
			},
			expectedErr: ValidationErrors{
				ValidationError{Field: "ID", Err: ErrValidateIncorrectLen},
				ValidationError{Field: "Age", Err: ErrValidateIncorrectNumeric},
				ValidationError{Field: "Email", Err: ErrValidateNotMatchRegexp},
				ValidationError{Field: "Role", Err: ErrValidateNotFoundInList},
				ValidationError{Field: "Phones", Err: ErrValidateIncorrectLen},
			},
		},
		{
			in: Response{Code: 0, Body: ""},
			expectedErr: ValidationErrors{
				ValidationError{Field: "Code", Err: ErrValidateNotFoundInList},
			},
		},
		{
			in: Float{F1: 5.4, F2: 6.1},
			expectedErr: ValidationErrors{
				ValidationError{Field: "F1", Err: ErrValidateIncorrectNumeric},
			},
		},
		{
			in: Float{F1: 15.5, F2: 6.7},
			expectedErr: ValidationErrors{
				ValidationError{Field: "F2", Err: ErrValidateIncorrectNumeric},
			},
		},
		{
			in: Regexp{Str1: "123456"},
			expectedErr: ValidationErrors{
				ValidationError{Field: "Str1", Err: ErrValidateIncorrectLen},
			},
		},
		{
			in: Regexp{Str1: "12a45"},
			expectedErr: ValidationErrors{
				ValidationError{Field: "Str1", Err: ErrValidateNotMatchRegexp},
			},
		},
		{
			in: Regexp{Str1: "123a456"},
			expectedErr: ValidationErrors{
				ValidationError{Field: "Str1", Err: ErrValidateIncorrectLen},
				ValidationError{Field: "Str1", Err: ErrValidateNotMatchRegexp},
			},
		},
		{
			in: Nested{
				Response: Response{Code: 600, Body: ""},
				User: User{
					ID:     "6bb694e-0669-4ad6-9920-41162213e92d",
					Name:   "",
					Age:    51,
					Email:  "test1test.ru",
					Role:   "no",
					Phones: []string{"1234568901", "098765432"},
				},
				Num: 10,
			},
			expectedErr: ValidationErrors{
				ValidationError{Field: "ID", Err: ErrValidateIncorrectLen},
				ValidationError{Field: "Age", Err: ErrValidateIncorrectNumeric},
				ValidationError{Field: "Email", Err: ErrValidateNotMatchRegexp},
				ValidationError{Field: "Role", Err: ErrValidateNotFoundInList},
				ValidationError{Field: "Phones", Err: ErrValidateIncorrectLen},
				ValidationError{Field: "Phones", Err: ErrValidateIncorrectLen},
				ValidationError{Field: "Code", Err: ErrValidateNotFoundInList},
				ValidationError{Field: "Num", Err: ErrValidateIncorrectNumeric},
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			if tt.expectedErr == nil {
				require.NoError(t, err)
				return
			}

			require.EqualError(t, tt.expectedErr, err.Error())
		})
	}
}

func TestValidateIncorrectInput(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in:          nil,
			expectedErr: ErrIncorrectStruct,
		},
		{
			in:          1,
			expectedErr: ErrIncorrectStruct,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)

			if tt.expectedErr == nil {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, tt.expectedErr, err)
		})
	}
}
